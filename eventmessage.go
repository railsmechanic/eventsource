// Copyright 2014 Matthias Kalb, Railsmechanic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eventsource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// EventMessage stores information of a message.
type eventMessage struct {
	Id      uint   `json:"id"`
	Event   string `json:"event"`
	Data    string `json:"data"`
	Channel string `json:"-"`
}

// NewEventMessage builds and returns a new eventMessage based on the given JSON data stream.
func newEventMessage(messageStream io.Reader, channel string) (*eventMessage, error) {
	var em eventMessage
	dec := json.NewDecoder(messageStream)
	for {
		if err := dec.Decode(&em); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}

	if channel == "" {
		em.Channel = "default"
	} else {
		em.Channel = channel
	}

	return &em, nil
}

// Message formats a []byte message which is finally sent to the consumers of a channel.
// Empty fields or fields that does not match the standard are removed.
func (em *eventMessage) Message() []byte {
	var messageData bytes.Buffer

	if em.Id > 0 {
		messageData.WriteString(fmt.Sprintf("id: %d\n", em.Id))
	}

	if len(em.Event) > 0 {
		messageData.WriteString(fmt.Sprintf("event: %s\n", strings.Replace(em.Event, "\n", "", -1)))
	}

	if len(em.Data) > 0 {
		lines := strings.Split(em.Data, "\n")
		for _, line := range lines {
			messageData.WriteString(fmt.Sprintf("data: %s\n", line))
		}
	}

	messageData.WriteString("\n")
	return messageData.Bytes()
}
