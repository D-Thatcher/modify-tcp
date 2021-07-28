package main

import (
	"flag"
	"fmt"
	"github.com/Telefonica/nfqueue"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

// var script string = `fetch('https://api.weather.gov/alerts/active')`
var scriptDefault string = `console.log('Hello from modify-tcp!')`

// Keep the websocket script well below the chunk size threshold of 127 (websocket chunking is not currently handled in the project)
var wsScript string = `<img id="photo" onerror="` + scriptDefault + `" src="img.jpg" width="1" height="1"/>`

var iface *string = flag.String("iface", "", "The network interface that will have its traffic redirected to the NFQUEUE (E.g. wlp0s20f3)")
var queueNumber *int = flag.Int("queue-num", 1, "The NFQUEUE number")
var disableUFW *bool = flag.Bool("override-ufw", false, "If true, the device's UFW firewall will disabled")
var handleIpTable *bool = flag.Bool("handle-iptables", false, "If true, the device's iptables will be updated to redirect traffic on the <iface> to the NFQUEUE")
var verbose *bool = flag.Bool("verbose", true, "If true, updates will be logged to STDOUT")
var script *string = flag.String("javascript", scriptDefault, "The Javascript that will be inserted into HTTP responses")
var maxPacketsInQueue *int = flag.Int("queue-len", 1000, "The max number of packets the NFQUEUE will hold")
var queueBufferSize *int = flag.Int("queue-buffer-size", 16*1024*1024, "The socket buffer size for receiving packets from nfnetlink_queue")

// webSocketMap persists a state for expected inbound websocket data frames
var webSocketMap map[string]struct{} = make(map[string]struct{})

func printLn(s ...string) {
	if *verbose {
		fmt.Println(s)
	}
}

type Queue struct {
	id    uint16
	queue *nfqueue.Queue
}

// FailOpen flag will make it non-blocking if the queue overflows
func NewQueue(id uint16) *Queue {
	q := &Queue{
		id: id,
	}
	queueCfg := &nfqueue.QueueConfig{
		MaxPackets: uint32(*maxPacketsInQueue),
		BufferSize: uint32(*queueBufferSize),
		QueueFlags: []nfqueue.QueueFlag{nfqueue.FailOpen},
	}
	q.queue = nfqueue.NewQueue(q.id, q, queueCfg)
	return q
}

// Start the queue.
func (q *Queue) Start() error {
	return q.queue.Start()
}

// Stop the queue.
func (q *Queue) Stop() error {
	return q.queue.Stop()
}

// Keep it stateless; another method could have already assigned and modified the data
func acceptableWebSocketUpgrade(tcp *layers.TCP, data *[]byte, strData *string) (doInsert bool) {
	if tcp == nil || tcp.SrcPort != 80 {
		return false
	}
	if len(*data) == 0 {
		*data = tcp.LayerPayload()
		*strData = string(*data)
	}
	if len(*strData) == 0 {
		return false
	}
	return strings.Contains(*strData, "websocket") && strings.HasPrefix(*strData, "HTTP/1.1 101")
}

func getWebSocketId(packet *gopacket.Packet, tcp *layers.TCP) (wsId string) {
	ipv4Layer := (*packet).Layer(layers.LayerTypeIPv4)
	ipv6Layer := (*packet).Layer(layers.LayerTypeIPv6)
	idWebSocket := ""
	sp := strconv.Itoa(int(tcp.SrcPort))
	dp := strconv.Itoa(int(tcp.DstPort))

	if ipv4Layer != nil {
		nl4 := ipv4Layer.(*layers.IPv4)
		idWebSocket = idWebSocket + string(nl4.SrcIP.String()) + ":" + sp + "-" + string(nl4.DstIP.String()) + ":" + dp
	} else if ipv6Layer != nil {
		nl6 := ipv6Layer.(*layers.IPv6)
		idWebSocket = idWebSocket + string(nl6.SrcIP.String()) + ":" + sp + "-" + string(nl6.DstIP.String()) + ":" + dp
	}
	return idWebSocket
}

// Keep it stateless; another method could have already assigned and modified the data
func acceptableWebSocketFrame(packet *gopacket.Packet, tcp *layers.TCP, data *[]byte, strData *string) (doInsert bool) {
	if tcp == nil || tcp.SrcPort != 80 {
		return false
	}
	if len(*data) == 0 {
		*data = tcp.LayerPayload()
		*strData = string(*data)
	}
	if len(*strData) == 0 {
		return false
	}
	idWebSocket := getWebSocketId(packet, tcp)
	_, present := webSocketMap[idWebSocket]
	return present
}

// Keep it stateless; another method could have already assigned and modified the data
func acceptableHTTP(tcp *layers.TCP, data *[]byte, strData *string) (doInsert bool) {
	if tcp == nil || tcp.SrcPort != 80 {
		return false
	}
	if len(*data) == 0 {
		*data = tcp.LayerPayload()
		*strData = string(*data)
	}
	if len(*strData) == 0 {
		return false
	}
	return strings.HasPrefix(*strData, "HTTP/1.1") && (strings.Contains(*strData, "html") || strings.Contains(*strData, "HTML"))
}

func calcOpCodeWS(payload *[]byte) (res byte) {
	if len(*payload) < 2 {
		return 8
	}
	return (*payload)[0] & 0b00001111
}

func calcContentLengthWS(payload *[]byte) (res byte) {
	if len(*payload) < 2 {
		return 0
	}
	return (*payload)[1] & 0b01111111
}

func finishWS(opcode byte) (res bool) {
	return opcode == 0x08
}
func textOrBinaryWS(opcode byte) (res bool) {
	return (opcode == 0x01) || (opcode == 0x02)
}

// Handle a nfqueue packet. It implements nfqueue.PacketHandler interface.
func (q *Queue) Handle(p *nfqueue.Packet) {
	var packet gopacket.Packet
	// @TODO does it matter what layer type we specify here? IPv6 layer is still visible in the decode anyways
	packet = gopacket.NewPacket(p.Buffer, layers.LayerTypeIPv4, gopacket.DecodeOptions{Lazy: true, NoCopy: true})

	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp := tcpLayer.(*layers.TCP)
		var data []byte
		var strData string
		var serialzeToModify bool = false

		if acceptableWebSocketUpgrade(tcp, &data, &strData) {
			var strct struct{}
			idWebSocket := getWebSocketId(&packet, tcp)
			webSocketMap[idWebSocket] = strct

		} else if acceptableHTTP(tcp, &data, &strData) {
			printLn("Inserting Javascript into HTTP response...")
			// Should isolate just the content-encoding value from the header, but this'll do for now
			encoding := ""
			chunked := strings.Contains(strData, "chunked")

			if strings.Contains(strData, "gzip") {
				encoding = "gzip"
			} else if strings.Contains(strData, "deflate") {
				encoding = "deflate"
			}
			insertedCorrectly := httpDataHandler(&data, &encoding, &scriptDefault, &chunked, *verbose)
			if !insertedCorrectly {
				p.Accept()
				return
			}
			serialzeToModify = true
		} else if acceptableWebSocketFrame(&packet, tcp, &data, &strData) {
			opCode := calcOpCodeWS(&data)

			if finishWS(opCode) {
				idWebSocket := getWebSocketId(&packet, tcp)
				delete(webSocketMap, idWebSocket)

			} else if textOrBinaryWS(opCode) {
				var modData []byte
				data[1] = byte(len(wsScript) + len(data) - 2)
				modData = append(data, []byte(wsScript)...)
				data = modData
				serialzeToModify = true
			} else {
				fmt.Printf("Unhandled state of websocket! OpCode: %v \n", opCode)
			}
		}
		if serialzeToModify {
			var eth *layers.Ethernet
			var nl4 *layers.IPv4
			var nl6 *layers.IPv6

			ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
			ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
			ipv6Layer := packet.Layer(layers.LayerTypeIPv6)

			var packetTopLvl []gopacket.SerializableLayer

			if ethernetLayer != nil {
				eth = ethernetLayer.(*layers.Ethernet)
				packetTopLvl = append(packetTopLvl, eth)
			}
			if ipv4Layer != nil {
				nl4 = ipv4Layer.(*layers.IPv4)
				tcp.SetNetworkLayerForChecksum(nl4)
				packetTopLvl = append(packetTopLvl, nl4)
			} else if ipv6Layer != nil {
				nl6 = ipv6Layer.(*layers.IPv6)
				tcp.SetNetworkLayerForChecksum(nl6)
				packetTopLvl = append(packetTopLvl, nl6)
			}

			packetTopLvl = append(packetTopLvl, tcp, gopacket.Payload([]byte(data)))
			buf := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{
				FixLengths:       true,
				ComputeChecksums: true,
			}
			if err := gopacket.SerializeLayers(buf, opts, packetTopLvl...); err != nil {
				fmt.Println(err)
			} else {
				p.Modify(buf.Bytes())
				return
			}
		}
	}
	p.Accept()
}

func main() {
	flag.Parse()
	if len(*iface) == 0 {
		fmt.Println("A network interface must be flagged. To show available interfaces run `ip addr`")
		os.Exit(1)
	}
	if strings.Contains(*script, "<script>") {
		fmt.Println("Script elements will be automatically added where applicable. Do not include them in the --javascript parameter")
		os.Exit(1)
	}
	queueNumberInt := uint16(*queueNumber)
	queueNumberStr := strconv.Itoa(*queueNumber)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	if *disableUFW {
		printLn("Disabling UFW...")
		changeUFW(false)
	}

	if *handleIpTable {
		printLn("Redirecting traffic to NFQUEUE...")
		changeIpTableRule(true, *iface, queueNumberStr)
	}
	printLn("Actively monitoring traffic...")
	q := NewQueue(queueNumberInt)
	go q.Start()
	<-c
	printLn("Exiting cleanly...")
	if *handleIpTable {
		printLn("Restoring network traffic to initial state...")
		changeIpTableRule(false, *iface, queueNumberStr)
	}
	if *disableUFW {
		printLn("Enabling UFW...")
		changeUFW(true)
	}
	printLn("Done.")
	os.Exit(1)
}
