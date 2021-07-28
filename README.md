[![PkgGoDev](https://pkg.go.dev/badge/github.com/D-Thatcher/modify-tcp)](https://pkg.go.dev/github.com/D-Thatcher/modify-tcp)
[![License MIT](https://img.shields.io/badge/license-MIT-lightgrey.svg?style=flat)](https://github.com/D-Thatcher/modify-tcp)
[![Go Report Card](https://goreportcard.com/badge/github.com/d-thatcher/modify-tcp)](https://goreportcard.com/report/github.com/d-thatcher/modify-tcp)

# modify-tcp

A lightweight CLI tool for modifying TCP data on the fly, without needing any proxies or virtual machines. The program is built upon [libnetfilter_queue](https://www.netfilter.org/projects/libnetfilter_queue) and supports intercepting requests and responses for arbitrary connections on the specified network interface. The CLI tool focuses on enabling the insertion of Javascript into inbound HTTP and WebSocket responses, with support for various encodings.

## Overview

Assuming your device is [set up](#getting-started), run
```
$ sudo ./modify-tcp -handle-iptables -iface <interface>
```

Open up another terminal (or browser) and query an HTTP site
```
$ curl http://www.climvis.org/content/global.htm

<html>
<head><script>console.log('Hello from modify-tcp!')</script>
<meta http-equiv="Content-Language" content="en-us">
<meta http-equiv="Content-Type" content="text/html; charset=windows-1252">
<title>Global Climate Animations</title>
</head>
<body>
<div align="center">
  <center>
  <table border="0" width="595" id="table1">
  ...
```
If you queried from a browser, you'll notice a message in the console `Hello from modify-tcp!`. To see the WebSocket in action, open up the [testSocket.html](doc/assets/testSocket.html) file in a browser. It will ping ws://echo.websocket.org/ and you'll also notice the same message in the console `Hello from modify-tcp!`


## Supports:
* HTTP response intercepts and script insertions
* WebSocket response intercepts and script insertions
* Additional modules can be added to support other protocols as all network traffic is routed through the NFQUEUE. If the queue overflows, traffic will not be dropped and it will bypass the queue
* IPv4 & IPv6 over wireless or ethernet
* Multiple content encodings, such as gzip, deflate and chunked - Note this feature is still in testing
 

## Getting started:
* Install libnetfilter `$ apt-get install libnetfilter-queue-dev`
* Find a network interface you want to reroute. Try running `$ ip addr` 
* Ensure your firewall will allow an iptable modification. If you pass the -handle-iptables flag, it will add the iptable route for you (and remove it when you stop modify-tcp)
* If you prefer to manually change it, here's the default rule that's used in modify-tcp: `$ iptables -A INPUT -i <IFACE> -j NFQUEUE --queue-num 1`
* Start the program `$ sudo ./modify-tcp -handle-iptables -iface <IFACE>`


## Using the CLI tool

    Usage of ./modify-tcp:
        -handle-iptables
              If true, the device's iptables will be updated to redirect traffic on the <iface> to the NFQUEUE
        -iface string
              The network interface that will have its traffic redirected to the NFQUEUE (E.g. wlp0s20f3)
        -javascript string
              The Javascript that will be inserted into HTTP responses (default "console.log('Hello from modify-tcp!')")
        -override-ufw
              If true, the device's UFW firewall will disabled
        -queue-buffer-size int
              The socket buffer size for receiving packets from nfnetlink_queue (default 16777216)
        -queue-len int
              The max number of packets the NFQUEUE will hold (default 1000)
        -queue-num int
              The NFQUEUE number (default 1)
        -verbose
              If true, updates will be logged to STDOUT (default true)
      

## Acknowledgements
* [netfilter](https://www.netfilter.org/projects/libnetfilter_queue/)
* [Telefonica's go binding to libnetfilter_queue](https://github.com/Telefonica/nfqueue)




