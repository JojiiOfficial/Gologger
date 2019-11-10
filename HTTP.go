package main

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strings"
)

func request(url, file string, data []byte, ignoreCert bool) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCert},
	}
	client := &http.Client{Transport: tr}
	var addFile string
	if strings.HasSuffix(url, "/") {
		if strings.HasPrefix(file, "/") {
			file = file[1:]
		}
		addFile = url + file
	} else {
		if strings.HasPrefix(file, "/") {
			file = file[1:]
		}
		addFile = url + "/" + file
	}
	resp, err := client.Post(addFile, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(d), nil
}
