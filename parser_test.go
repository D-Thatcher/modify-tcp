package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestInsertScript(t *testing.T) {
	content, _ := ioutil.ReadFile("doc/assets/test.cap")
	script := `console.log('Hello from modify-tcp!')`
	s := "gzip"
	c := false
	modifyHTTPResonse(&content, &s, &script, &c)
	sep := "\r\n\r\n"
	strData := string(content)
	residx := strings.Index(strData, sep)
	if residx != -1 {
		content = content[residx+len(sep):]
	} else {
		panic("error parsing HTTP header in test.cap")
	}
	unzipped, err := gUnzipData(content)
	if err != nil {
		fmt.Printf("Error evaluating TestInsertScript %v", err)
	}

	if !strings.Contains(string(unzipped), script) {
		t.Errorf("httpDataHandler failed to insert javascript")
	}
}
