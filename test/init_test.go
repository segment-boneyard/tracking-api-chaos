package test

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gohttp/response"
)

func init() {
	response.Pretty = false
}

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// Helper methods for constructing http.Requests
func post(path, body string) *http.Request {
	r := bytes.NewReader([]byte(body))
	req, err := http.NewRequest("POST", "http://api.test"+path, r)
	check(err)
	return req
}

func postGzip(path, body string) *http.Request {
	b := new(bytes.Buffer)
	z := gzip.NewWriter(b)
	z.Write([]byte(body))
	z.Close()

	req, err := http.NewRequest("POST", "http://api.test"+path, b)
	check(err)
	req.Header.Set("Content-Encoding", "gzip")
	return req
}

func options(path string) *http.Request {
	req, err := http.NewRequest("OPTIONS", "http://api.test"+path, nil)
	check(err)
	return req
}

func put(path, body string) *http.Request {
	r := bytes.NewReader([]byte(body))
	req, err := http.NewRequest("PUT", "http://api.test"+path, r)
	check(err)
	return req
}

func get(path, body string) *http.Request {
	data := base64.URLEncoding.EncodeToString([]byte(body))

	if strings.Contains(path, "?") {
		path += "&data=" + data
	} else {
		path += "?data=" + data
	}

	req, err := http.NewRequest("GET", "http://api.test"+path, nil)
	check(err)
	return req
}

func query(path, query string) *http.Request {
	req, err := http.NewRequest("GET", "http://api.test"+path+"?"+query, nil)
	check(err)
	return req
}

func fixture(name string) []byte {
	b, err := ioutil.ReadFile("fixtures/" + name)
	check(err)
	return b
}
