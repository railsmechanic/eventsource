// Copyright 2014 Matthias Kalb, Railsmechanic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eventsource

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Consumer stores information of a connected consumer.
type consumer struct {
	connection net.Conn
	es         *eventSource
	inbox      chan *eventMessage
	channel    string
	expired    bool
}

// NewConsumer builds and returns a new consumer based on the given attributes.
// A goroutine is started for handling incoming messages.
func newConsumer(resp http.ResponseWriter, req *http.Request, es *eventSource, channel string) (*consumer, error) {
	connection, _, err := resp.(http.Hijacker).Hijack()
	if err != nil {
		return nil, err
	}

	cr := &consumer{
		connection: connection,
		es:         es,
		inbox:      make(chan *eventMessage),
		channel:    channel,
		expired:    false,
	}

	if err := cr.setupConnection(); err != nil {
		return nil, err
	}

	go cr.inboxDispatcher()

	return cr, nil
}

// SetupConnection is responsible to setup a usable connection to a consumer.
// If an unexpected error (timeout,...) occurs, the connection gets closed.
func (cr *consumer) setupConnection() error {
	headers := [][]byte{
		[]byte("HTTP/1.1 200 OK"),
		[]byte("Content-Type: text/event-stream"),
		[]byte("Cache-Control: no-cache"),
		[]byte("Connection: keep-alive"),
		[]byte(fmt.Sprintf("Access-Control-Allow-Origin: %s", cr.es.settings.GetCorsAllowOrigin())),
		[]byte(fmt.Sprintf("Access-Control-Allow-Method: %s", cr.es.settings.GetCorsAllowMethod())),
	}

	headersData := append(bytes.Join(headers, []byte("\n")), []byte("\n\n")...)

	if _, err := cr.connection.Write(headersData); err != nil {
		cr.connection.Close()
		return err
	}

	return nil
}

// InboxDispatcher processes incoming eventMessages.
// It disconnects timed out consumers and initiates the removal from the consumer pool.
func (cr *consumer) inboxDispatcher() {
	for message := range cr.inbox {
		cr.connection.SetWriteDeadline(time.Now().Add(cr.es.settings.GetTimeout()))
		if _, err := cr.connection.Write(message.Message()); err != nil {
			if netErr, ok := err.(net.Error); !ok || netErr.Timeout() {
				cr.expired = true
				cr.connection.Close()
				cr.es.expireConsumer <- cr
				return
			}
		}
	}
	cr.connection.Close()
}
