package main

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http/httputil"
	"strings"
)

func padZip(lo *[]byte) {
	for !bytes.Equal((*lo)[len(*lo)-3:], []byte{0, 0, 0}) {
		*lo = append(*lo, 0)
	}
}
func defalteData(data []byte) (resData []byte, err error) {
	// padZip(&data)
	b := bytes.NewBuffer(data)

	r := flate.NewReader(b)

	defer r.Close()
	resData, err = ioutil.ReadAll(r)

	if err != nil {
		if len(resData) >= 0 {
			fmt.Printf("\nEncountered an error (%v) in defalteData the data. Continuing anyways since the uncompressed data was non-empty... \n", err)
			err = nil
		} else {
			fmt.Printf("\nEncountered an error (%v) in defalteData the data. Ignoring payload... \n", err)
			return
		}
	}
	return
}

func reflateData(data []byte) (compressedData []byte, err error) {

	var b bytes.Buffer
	flateWrite, err := flate.NewWriter(&b, flate.BestCompression)
	if err != nil {
		fmt.Printf("\nerr in initializing Flate writer : %v \n", err)
		return
	}
	defer flateWrite.Close()

	_, err = flateWrite.Write(data)
	if err != nil {
		fmt.Printf("\nFail to re-zip the data. gz.Write: %v \n", err)
		return
	}

	if err = flateWrite.Flush(); err != nil {
		fmt.Printf("\nFail to re-zip the data. gz.Flush: %v \n", err)
		return
	}

	compressedData = b.Bytes()

	return
}

func unchunkData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)
	r := ioutil.NopCloser(httputil.NewChunkedReader(b))
	defer r.Close()
	resData, err = ioutil.ReadAll(r)

	if err != nil {
		if len(resData) >= 0 {
			fmt.Printf("\nEncountered an error (%v) in unchunkData the data. Continuing anyways since the uncompressed data was non-empty... \n", err)
			err = nil
		} else {
			fmt.Printf("\nEncountered an error (%v) in unchunkData the data. Ignoring payload... \n", err)
			return
		}
	}
	return
}

func gUnzipData(data []byte) (resData []byte, err error) {
	// padZip(&data)
	b := bytes.NewBuffer(data)

	// var r io.Reader
	r, err := gzip.NewReader(b)
	if err != nil {
		fmt.Printf("\nFail to init the unzipper for the the data: %v \n", err)
		return
	}
	defer r.Close()
	resData, err = ioutil.ReadAll(r)

	if err != nil {
		if len(resData) >= 0 {
			fmt.Printf("\nEncountered an error (%v) in gunzipping the data. Continuing anyways since the uncompressed data was non-empty... \n", err)
			err = nil
		} else {
			fmt.Printf("\nEncountered an error (%v) in gunzipping the data. Ignoring payload... \n", err)
			return
		}
	}
	return
}

func gZipData(data []byte) (compressedData []byte, err error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	defer gz.Close()

	_, err = gz.Write(data)
	if err != nil {
		fmt.Printf("\nFail to re-zip the data. gz.Write: %v \n", err)
		return
	}

	if err = gz.Flush(); err != nil {
		fmt.Printf("\nFail to re-zip the data. gz.Flush: %v \n", err)
		return
	}

	if err = gz.Close(); err != nil {
		fmt.Printf("\nFail to re-zip the data. gz.Close: %v \n", err)
		return
	}

	compressedData = b.Bytes()
	return
}

func getHeaderValue(headerKey string, headerStr *string) (headerValue string, err error) {
	cLen := headerKey + ": "
	splHEADER := strings.Split(*headerStr, "\r\n") //
	for _, hdr := range splHEADER {
		idx := strings.Index(hdr, cLen)
		if idx != -1 {
			return strings.TrimSpace(hdr[len(cLen):]), nil
		}
	}
	return headerValue, errors.New("Key not found")
}

func addAccessControl(headerStr *string, accessControl *string) {
	idx := strings.Index(*headerStr, *accessControl)
	if idx != -1 {
		return
	}
	*headerStr = strings.TrimSpace(*headerStr) + "\r\n" + *accessControl + "\r\n\r\n"
}

func httpDataHandler(_data *[]byte, encoding *string, javaScript *string, chunked *bool, verbose bool) (markAsModified bool) {
	defer func() {
		if r := recover(); r != nil {
			if verbose {
				fmt.Println("Recovered from panic in modifyHTTPResonse", r)
			}
			markAsModified = false
		}
	}()
	modifyHTTPResonse(_data, encoding, javaScript, chunked)
	return true
}

func modifyHTTPResonse(_data *[]byte, encoding *string, javaScript *string, chunked *bool) {
	sep := "\r\n\r\n"
	content := *_data
	strData := string(content)
	residx := strings.Index(strData, sep)
	var HEADER string
	if residx != -1 {
		HEADER = string(content[:residx+len(sep)])
		content = content[residx+len(sep):]
	} else {
		panic("error parsing HTTP header")
	}

	if len(content) == 0 {
		panic("Unable to handle empty HTTP body")
	}

	var uncompressedData []byte
	var uncompressedDataErr error
	doGzip := *encoding == "gzip"
	doFlate := *encoding == "deflate"

	if *chunked {
		content, uncompressedDataErr = unchunkData(content)
		if uncompressedDataErr != nil {
			fmt.Printf("\nError unchunking contents: %v \n", uncompressedDataErr)
			panic(uncompressedDataErr)
		}
	}

	if doGzip {
		uncompressedData, uncompressedDataErr = gUnzipData(content)
		if uncompressedDataErr != nil {
			fmt.Printf("\nError unzipping contents: %v \n", uncompressedDataErr)
			panic(uncompressedDataErr)
		}
	} else if doFlate {
		uncompressedData, uncompressedDataErr = defalteData(content)
		if uncompressedDataErr != nil {
			fmt.Printf("\nError deflating contents: %v \n", uncompressedDataErr)
			panic(uncompressedDataErr)
		}
	} else {
		uncompressedData = content
	}

	repl := strings.Replace(string(uncompressedData), "<head>", "<head><script>"+*javaScript+"</script>", 1)
	repl = strings.Replace(repl, "<HEAD>", "<HEAD><script>"+*javaScript+"</script>", 1)

	var recomp []byte
	var recompErr error

	if doGzip {
		recomp, recompErr = gZipData([]byte(repl))
		if recompErr != nil {
			fmt.Printf("\nError recompressing contents: %v \n", recompErr)
			panic(recompErr)
		}
	} else if doFlate {
		recomp, recompErr = reflateData([]byte(repl))
		if recompErr != nil {
			fmt.Printf("\nError reflating contents: %v \n", recompErr)
			panic(recompErr)
		}
	} else {
		recomp = []byte(repl)
	}

	*_data = append([]byte(HEADER), recomp...)
}
