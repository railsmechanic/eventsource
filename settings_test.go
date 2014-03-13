// Copyright 2014 Matthias Kalb, Railsmechanic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eventsource

import (
	"testing"
	"time"
)

func TestDefaultSettings(t *testing.T) {
	ds := &Settings{}

	if timeout := ds.GetTimeout(); timeout != 2*time.Second {
		t.Error("Expected 2 seconds, got", timeout)
	}

	if authToken := ds.GetAuthToken(); authToken != "" {
		t.Error("Expected empty AuthToken, got ", authToken)
	}

	if host := ds.GetHost(); host != "127.0.0.1" {
		t.Error("Expected 127.0.0.1, got", host)
	}

	if port := ds.GetPort(); port != 8080 {
		t.Error("Expected 8080, got", port)
	}

	if corsAllowOrigin := ds.GetCorsAllowOrigin(); corsAllowOrigin != "127.0.0.1" {
		t.Error("Expected 127.0.0.1, got", corsAllowOrigin)
	}

	if corsAllowMethod := ds.GetCorsAllowMethod(); corsAllowMethod != "GET" {
		t.Error("Expected GET, got", corsAllowMethod)
	}
}

func TestCustomSettings(t *testing.T) {
	cs := &Settings{
		Timeout:         3 * time.Second,
		AuthToken:       "TOKEN",
		Host:            "192.168.1.1",
		Port:            3000,
		CorsAllowOrigin: "*",
		CorsAllowMethod: []string{"GET", "POST", "DELETE"},
	}

	if timeout := cs.GetTimeout(); timeout != 3*time.Second {
		t.Error("Expected 3 seconds, got", timeout)
	}

	if authToken := cs.GetAuthToken(); authToken != "TOKEN" {
		t.Error("AuthToken should be 'TOKEN', got ", authToken)
	}

	if host := cs.GetHost(); host != "192.168.1.1" {
		t.Error("Expected 192.168.1.1, got", host)
	}

	if port := cs.GetPort(); port != 3000 {
		t.Error("Expected 3000, got", port)
	}

	if corsAllowOrigin := cs.GetCorsAllowOrigin(); corsAllowOrigin != "*" {
		t.Error("Expected '*', got", corsAllowOrigin)
	}

	if corsAllowMethod := cs.GetCorsAllowMethod(); corsAllowMethod != "GET, POST, DELETE" {
		t.Error("Expected 'GET, POST, DELETE', got", corsAllowMethod)
	}
}
