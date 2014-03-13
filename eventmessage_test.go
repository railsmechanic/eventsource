// Copyright 2014 Matthias Kalb, Railsmechanic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eventsource

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// Helper function to build EventMessages
func buildEventMessage(messageType, channel string) (*eventMessage, error) {
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

	return newEventMessage(messageStream, channel)
}

// Available test modes
func messageModes() []string {
	return []string{ModeAll, ModeNoid, ModeNoevent, ModeNodata}
}

func TestBuildEventMessage(t *testing.T) {

	// Test EventMessage in different modes
	for _, mode := range messageModes() {
		if _, err := buildEventMessage(mode, "my-channel"); err != nil {
			t.Error("Unable build EventMessage from JSON data in mode", mode)
		}
	}

	// Test EventMessage to build without a channel name
	if _, err := buildEventMessage("all", ""); err != nil {
		t.Error("Unable build EventMessage from JSON data without channel name")
	}
}

func TestContentOfEventMessage(t *testing.T) {

	// Test EventMessage in different modes
	for _, mode := range messageModes() {
		em, _ := buildEventMessage(mode, "my-channel")
		switch mode {
		case ModeAll:
			if em.Id != 1 {
				t.Error("Expected 1 got", em.Id)
			}

			if em.Event != "foo" {
				t.Error("Expected 'foo' got", em.Event)
			}

			if em.Data != "bar" {
				t.Error("Expected 'bar' got", em.Data)
			}

			if em.Channel != "my-channel" {
				t.Error("Expected 'my-channel' got", em.Channel)
			}

		case ModeNoid:
			if em.Id != 0 {
				t.Error("Expected 0 got", em.Id)
			}

			if em.Event != "foo" {
				t.Error("Expected 'foo' got", em.Event)
			}

			if em.Data != "bar" {
				t.Error("Expected 'bar' got", em.Data)
			}

			if em.Channel != "my-channel" {
				t.Error("Expected 'my-channel' got", em.Channel)
			}

		case ModeNoevent:
			if em.Id != 1 {
				t.Error("Expected 1 got", em.Id)
			}

			if em.Event != "" {
				t.Error("Expected '' got", em.Event)
			}

			if em.Data != "bar" {
				t.Error("Expected 'bar' got", em.Data)
			}

			if em.Channel != "my-channel" {
				t.Error("Expected 'my-channel' got", em.Channel)
			}

		case ModeNodata:
			if em.Id != 1 {
				t.Error("Expected 1 got", em.Id)
			}

			if em.Event != "foo" {
				t.Error("Expected 'foo' got", em.Event)
			}

			if em.Data != "" {
				t.Error("Expected '' got", em.Data)
			}

			if em.Channel != "my-channel" {
				t.Error("Expected 'my-channel' got", em.Channel)
			}
		}
	}

	// Test EventMessage to use channel name 'default' when its omited
	if em, _ := buildEventMessage("all", ""); em.Channel != "default" {
		t.Error("Expected 'default' on empty channel argument, got", em.Channel)
	}
}

func TestByteMesssage(t *testing.T) {

	for _, mode := range messageModes() {
		em, _ := buildEventMessage(mode, "my-channel")

		var messageData bytes.Buffer

		if mode != ModeNoid {
			messageData.WriteString("id: 1\n")
		}

		if mode != ModeNoevent {
			messageData.WriteString("event: foo\n")
		}

		if mode != ModeNodata {
			messageData.WriteString("data: bar\n")
		}
		messageData.WriteString("\n")

		if !bytes.Equal(em.Message(), messageData.Bytes()) {
			t.Error("Byte Message is malformed in mode", mode)
		}
	}
}
