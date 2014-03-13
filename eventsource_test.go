// Copyright 2014 Matthias Kalb, Railsmechanic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eventsource

import (
	"bytes"
	"github.com/gorilla/mux"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	ModeAll     = "all"
	ModeNoid    = "noid"
	ModeNoevent = "noevent"
	ModeNodata  = "nodata"
)

type testEventSource struct {
	eventSource EventSource
	testServer  *httptest.Server
}

// Helper for building the EventSource environment
func setupEventSource(t *testing.T, settings *Settings) *testEventSource {
	es := New(settings)
	if es == nil {
		t.Error("Unable to setup EventSource")
	}

	return &testEventSource{
		eventSource: es,
		testServer:  httptest.NewServer(es.Router()),
	}
}

// Helper to properly shutdown the EventSource environment
func (es *testEventSource) closeEventSource() {
	es.eventSource.Stop()
	es.testServer.Close()
}

// Helper for reading EventSource responses
func readResponse(t *testing.T, conn net.Conn) []byte {
	resp := make([]byte, 1024)
	if _, err := conn.Read(resp); err != nil && err != io.EOF {
		t.Error(err)
	}
	return resp
}

// Helper for joining an EventSource channel
func (es *testEventSource) joinChannel(t *testing.T, channel string) (net.Conn, []byte) {
	conn, err := net.Dial("tcp", strings.Replace(es.testServer.URL, "http://", "", 1))
	if err != nil {
		t.Error(err)
	}

	if _, err := conn.Write([]byte("GET /" + channel + " HTTP/1.1\n\n")); err != nil {
		t.Error(err)
	}

	return conn, readResponse(t, conn)
}

// Helper to compare EventSource responses
func expectResponse(t *testing.T, conn net.Conn, expectedResponse string) {
	time.Sleep(100 * time.Millisecond)
	if resp := readResponse(t, conn); !strings.Contains(string(resp), expectedResponse) {
		t.Errorf("Expected response:\n%s\n and got:\n%s\n", expectedResponse, resp)
	}
}

// Helper function to build EventMessages
func buildMessageData(messageType string) io.Reader {
	var messageStream io.Reader
	switch messageType {
	case ModeAll:
		messageStream = strings.NewReader("{\"id\":1,\"event\":\"foo\",\"data\":\"bar\"}")
	case ModeNoid:
		messageStream = strings.NewReader("{\"event\":\"foo\",\"data\":\"bar\"}")
	case ModeNoevent:
		messageStream = strings.NewReader("{\"id\":1,\"data\":\"bar\"}")
	case ModeNodata:
		messageStream = strings.NewReader("{\"id\":1,\"event\":\"foo\"}")
	}
	return messageStream
}

func TestRouter(t *testing.T) {
	es := New(nil)
	router := es.Router()
	var match mux.RouteMatch

	// Testing Router with a GET Request and a proper formated channel name
	req, err := http.NewRequest("GET", "http://127.0.0.1/default", nil)
	if err != nil {
		t.Error(err)
	}

	if !router.Match(req, &match) {
		t.Error("Method 'GET' on is not allowed for channel 'default'")
	}

	// Testing Router with a POST Request and a proper formated channel name
	req, err = http.NewRequest("POST", "http://127.0.0.1/default", nil)
	if err != nil {
		t.Error(err)
	}

	if !router.Match(req, &match) {
		t.Error("Method 'POST' is not allowed for channel name 'default'")
	}

	// Testing Router with a DELETE Request and a proper formated channel name
	req, err = http.NewRequest("DELETE", "http://127.0.0.1/default", nil)
	if err != nil {
		t.Error(err)
	}

	if !router.Match(req, &match) {
		t.Error("Method 'DELETE' is not allowed for channel name 'default'")
	}

	// Testing Router with a PUT Request and a proper formated channel name
	req, err = http.NewRequest("PUT", "http://127.0.0.1/default", nil)
	if err != nil {
		t.Error(err)
	}

	if router.Match(req, &match) {
		t.Error("Method 'PUT' is not allowed for channel name 'default'")
	}

	// Testing Router with a GET Request and a wrong formated channel name
	req, err = http.NewRequest("GET", "http://127.0.0.1/DEFAULT", nil)
	if err != nil {
		t.Error(err)
	}

	if router.Match(req, &match) {
		t.Error("Method 'GET' on is not allowed wrong formated for channel name 'DEFAULT'")
	}

	// Testing Router for POST Request for wrong formated channel names
	req, err = http.NewRequest("POST", "http://127.0.0.1/DEFAULT", nil)
	if err != nil {
		t.Error(err)
	}

	if router.Match(req, &match) {
		t.Error("Method 'POST' is not allowed for wrong formated channel ' nameDEFAULT'")
	}

	// Testing Router for DELETE Request for wrong formated channel names
	req, err = http.NewRequest("DELETE", "http://127.0.0.1/DEFAULT", nil)
	if err != nil {
		t.Error(err)
	}

	if router.Match(req, &match) {
		t.Error("Method 'DELETE' is not allowed for wrong formated channel ' nameDEFAULT'")
	}
}

func TestConnection(t *testing.T) {
	es := setupEventSource(t, nil)
	defer es.closeEventSource()

	conn, resp := es.joinChannel(t, "default")
	defer conn.Close()

	if !strings.Contains(string(resp), "HTTP/1.1 200 OK\n") {
		t.Error("Response has no HTTP status")
	}

	if !strings.Contains(string(resp), "Content-Type: text/event-stream\n") {
		t.Error("Response header does not contain 'Content-Type: text/event-stream'")
	}

	if !strings.Contains(string(resp), "Cache-Control: no-cache\n") {
		t.Error("Response header does not contain 'Cache-Control: no-cache'")
	}

	if !strings.Contains(string(resp), "Connection: keep-alive\n") {
		t.Error("Response header does not contain 'Connection: keep-alive'")
	}

	if !strings.Contains(string(resp), "Access-Control-Allow-Origin: 127.0.0.1\n") {
		t.Error("Response header does not contain 'Access-Control-Allow-Origin: 127.0.0.1'")
	}

	if !strings.Contains(string(resp), "Access-Control-Allow-Method: GET\n") {
		t.Error("Response header does not contain 'Access-Control-Allow-Method: GET'")
	}
}

func TestAuthToken(t *testing.T) {
	es := setupEventSource(t,
		&Settings{
			AuthToken: "secrect",
		})
	defer es.closeEventSource()

	conn, _ := es.joinChannel(t, "default")
	defer conn.Close()

	req, err := http.NewRequest("DELETE", es.testServer.URL+"/default", nil)
	if err != nil {
		t.Error("Creating DELETE request failed with", err)
	}
	req.Header.Add("Auth-Token", "secrect")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error("Unable to send DELETE request")
	}

	if resp.StatusCode != 200 {
		t.Error("DELETE request of channel failed with status code", resp.StatusCode)
	}
}

func TestSendMessage(t *testing.T) {
	es := setupEventSource(t, nil)
	defer es.closeEventSource()

	conn, _ := es.joinChannel(t, "default")
	defer conn.Close()

	// Test EventMessage in different modes
	for _, mode := range messageModes() {
		messageStream := buildMessageData(mode)
		var expectedMessage bytes.Buffer

		if mode != ModeNoid {
			expectedMessage.WriteString("id: 1\n")
		}

		if mode != ModeNoevent {
			expectedMessage.WriteString("event: foo\n")
		}

		if mode != ModeNodata {
			expectedMessage.WriteString("data: bar\n")
		}
		expectedMessage.WriteString("\n")

		es.eventSource.SendMessage(messageStream, "default")
		expectResponse(t, conn, string(expectedMessage.Bytes()))
	}
}

func TestSendMessageViaHTTPPost(t *testing.T) {
	es := setupEventSource(t, nil)
	defer es.closeEventSource()

	conn, _ := es.joinChannel(t, "default")
	defer conn.Close()

	// Test EventMessage in different modes
	for _, mode := range messageModes() {
		messageStream := buildMessageData(mode)
		var expectedMessage bytes.Buffer

		if mode != ModeNoid {
			expectedMessage.WriteString("id: 1\n")
		}

		if mode != ModeNoevent {
			expectedMessage.WriteString("event: foo\n")
		}

		if mode != ModeNodata {
			expectedMessage.WriteString("data: bar\n")
		}
		expectedMessage.WriteString("\n")

		resp, err := http.Post(es.testServer.URL+"/default", "application/json", messageStream)
		if err != nil {
			t.Error("POST event failed with", err)
		}

		if resp.StatusCode != 201 {
			t.Error("POST event failed with status code", resp.StatusCode)
		}

		expectResponse(t, conn, string(expectedMessage.Bytes()))
	}
}

func TestChannelExists(t *testing.T) {
	es := setupEventSource(t, nil)
	defer es.closeEventSource()

	conn, _ := es.joinChannel(t, "default")
	defer conn.Close()

	if !es.eventSource.ChannelExists("default") {
		t.Error("Channel 'default' should exist")
	}

	if es.eventSource.ChannelExists("my-channel") {
		t.Error("Channel 'my-channel' should not exist")
	}
}

func TestConsumerCount(t *testing.T) {
	es := setupEventSource(t, nil)
	defer es.closeEventSource()

	conn, _ := es.joinChannel(t, "default")
	defer conn.Close()

	if es.eventSource.ConsumerCount("default") > 1 {
		t.Error("ConsumerCount for channel 'default' is invalid")
	}
}

func TestConsumerCountAll(t *testing.T) {
	es := setupEventSource(t, nil)
	defer es.closeEventSource()

	conn, _ := es.joinChannel(t, "default")
	defer conn.Close()

	if es.eventSource.ConsumerCountAll() > 1 {
		t.Error("ConsumerCountAll is invalid")
	}
}

func TestChannels(t *testing.T) {
	es := setupEventSource(t, nil)
	defer es.closeEventSource()

	conn, _ := es.joinChannel(t, "default")
	defer conn.Close()

	if es.eventSource.Channels()[0] != "default" {
		t.Error("Returned channel names are invalid")
	}
}

func TestChannelClose(t *testing.T) {
	es := setupEventSource(t, nil)
	defer es.closeEventSource()

	conn, _ := es.joinChannel(t, "default")
	defer conn.Close()

	if !es.eventSource.ChannelExists("default") {
		t.Error("Channel 'default' should exist")
	}

	es.eventSource.Close("default")
	time.Sleep(100 * time.Millisecond)

	if es.eventSource.ChannelExists("default") {
		t.Error("Channel 'default' should not exist")
	}
}

func TestChannelCloseViaHTTPDelete(t *testing.T) {
	es := setupEventSource(t, nil)
	defer es.closeEventSource()

	conn, _ := es.joinChannel(t, "default")
	defer conn.Close()

	if !es.eventSource.ChannelExists("default") {
		t.Error("Channel 'default' should exist")
	}

	req, err := http.NewRequest("DELETE", es.testServer.URL+"/default", nil)
	if err != nil {
		t.Error("Creating DELETE request failed with", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error("Unable to send DELETE request")
	}

	if resp.StatusCode != 200 {
		t.Error("DELETE request of channel failed with status code", resp.StatusCode)
	}

	if len(es.eventSource.Channels()) != 0 {
		t.Error("Channel 'default' should be closed")
	}
}

func TestChannelCloseAll(t *testing.T) {
	es := setupEventSource(t, nil)
	defer es.closeEventSource()

	conn, _ := es.joinChannel(t, "default")
	defer conn.Close()

	if len(es.eventSource.Channels()) == 0 {
		t.Error("At least one channel should exist")
	}

	es.eventSource.CloseAll()
	time.Sleep(100 * time.Millisecond)

	if len(es.eventSource.Channels()) != 0 {
		t.Error("All channels should be closed")
	}
}

func TestChannelCloseAllViaHTTPDelete(t *testing.T) {
	es := setupEventSource(t, nil)
	defer es.closeEventSource()

	conn, _ := es.joinChannel(t, "default")
	defer conn.Close()

	if !es.eventSource.ChannelExists("default") {
		t.Error("Channel 'default' should exist")
	}

	req, err := http.NewRequest("DELETE", es.testServer.URL+"/all", nil)
	if err != nil {
		t.Error("Creating DELETE request failed with", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error("Unable to send DELETE request")
	}

	if resp.StatusCode != 200 {
		t.Error("DELETE request of all channels failed with status code", resp.StatusCode)
	}

	if len(es.eventSource.Channels()) != 0 {
		t.Error("All channels should be closed")
	}
}

func TestStats(t *testing.T) {
	es := setupEventSource(t, nil)
	defer es.closeEventSource()

	conn, _ := es.joinChannel(t, "default")
	defer conn.Close()

	// HEAD for single channel
	req, err := http.NewRequest("HEAD", es.testServer.URL+"/default", nil)
	if err != nil {
		t.Error("Creating HEAD request failed with", err)
	}
	req.Header.Add("Connection", "close")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error("Unable to send HEAD request")
	}

	if statusCode := resp.StatusCode; statusCode != 200 {
		t.Error("HEAD request for channel failed with status code", statusCode)
	}

	consumerCountHeader := resp.Header.Get("X-Consumer-Count")
	consumerCount, err := strconv.Atoi(consumerCountHeader)
	if err != nil {
		t.Error("Unable to convert to integer", err)
	}

	if consumerCount != 1 {
		t.Error("Response for X-Consumer-Count is invalid", consumerCount)
	}

	channelExistsHeader := resp.Header.Get("X-Channel-Exists")
	channelExists, err := strconv.ParseBool(channelExistsHeader)
	if err != nil {
		t.Error("Unable to convert to bool", err)
	}

	if channelExists != true {
		t.Error("Response for X-Channel-Exists is invalid", channelExists)
	}

	// HEAD for all channels
	req, err = http.NewRequest("HEAD", es.testServer.URL+"/all", nil)
	if err != nil {
		t.Error("Creating HEAD request failed with", err)
	}
	req.Header.Add("Connection", "close")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error("Unable to send HEAD request")
	}

	if statusCode := resp.StatusCode; statusCode != 200 {
		t.Error("HEAD request for channel failed with status code", statusCode)
	}

	consumerCountHeader = resp.Header.Get("X-Consumer-Count")
	consumerCount, err = strconv.Atoi(consumerCountHeader)
	if err != nil {
		t.Error("Unable to convert to integer", err)
	}

	if consumerCount != 1 {
		t.Error("Response for X-Consumer-Count is invalid", consumerCount)
	}

	if availableChannels := resp.Header.Get("X-Available-Channels"); availableChannels != "[default]" {
		t.Error("Response for X-Available-Channels is invalid", availableChannels)
	}
}

func TestRun(t *testing.T) {
	es := New(nil)
	go es.Run()
}
