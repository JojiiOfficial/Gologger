package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func request(url, file string, data []byte, ignoreCert bool, timeout time.Duration) (string, error) {
	if timeout == 0*time.Second {
		timeout = http.DefaultClient.Timeout
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCert},
		IdleConnTimeout: timeout,
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
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	resp.Body.Close()
	return string(d), nil
}
