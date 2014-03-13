// Copyright 2014 Matthias Kalb, Railsmechanic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eventsource

import (
	"strings"
	"time"
)

// Default settings.
const (
	defaultTimeout         = 2 * time.Second
	defaultAuthToken       = ""
	defaultHost            = "127.0.0.1"
	defaultPort            = 8080
	defaultCorsAllowOrigin = "127.0.0.1"
	defaultCorsAllowMethod = "GET"
)

// Settings stores all essential settings.
type Settings struct {
	Timeout         time.Duration
	AuthToken       string
	Host            string
	Port            uint
	CorsAllowOrigin string
	CorsAllowMethod []string
}

// GetTimeout returns the timeout for consumers.
func (s *Settings) GetTimeout() time.Duration {
	if s == nil || s.Timeout <= 0*time.Second {
		return defaultTimeout
	}
	return s.Timeout
}

// GetAuthToken returns the authenticatoin token.
func (s *Settings) GetAuthToken() string {
	if s == nil || len(s.AuthToken) <= 0 {
		return defaultAuthToken
	}
	return strings.TrimSpace(s.AuthToken)
}

// GetHost returns the hostname/ip address on which the service should listen on.
func (s *Settings) GetHost() string {
	if s == nil || s.Host == "" {
		return defaultHost
	}
	return s.Host
}

// GetPort returns the port on which the service should listen on.
func (s *Settings) GetPort() uint {
	if s == nil || s.Port == 0 {
		return defaultPort
	}
	return s.Port
}

// GetCorsAllowOrigin returns the Access-Control-Allow-Origin.
func (s *Settings) GetCorsAllowOrigin() string {
	if s == nil || s.CorsAllowOrigin == "" {
		return defaultCorsAllowOrigin
	}
	return s.CorsAllowOrigin
}

// GetCorsAllowMethod returns the Access-Control-Allow-Method.
func (s *Settings) GetCorsAllowMethod() string {
	if s == nil || len(s.CorsAllowMethod) == 0 {
		return defaultCorsAllowMethod
	}
	return strings.Join(s.CorsAllowMethod, ", ")
}
