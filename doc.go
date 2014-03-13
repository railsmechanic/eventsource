// Copyright 2014 Matthias Kalb, Railsmechanic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package railsmechanic/eventsource implements EventSource Server-Sent Events (SSE) for Go.

This implementation of an EventSource service offers isolated channels and
a simple RESTful interface. It also offers a simple, token based authentication strategy.
So publishing events, deleting channels or getting informations of a channel can be protected from unauthorized access.

Main features are:

  - Support for isolated channels *(channel1 will not see events of channel2)*
  - Support for global notifications across all channels *(every consumer receive this event)*
  - RESTful interface for publishing events, deleting, subscribing and getting information of/to channels
  - Token base authentication for publishing/deleting/getting information of channels
  - Support for CORS *(Allow-Origin, Allow-Method)*
  - Allows an individual configuration to set up EventSource for your needs
  - Simple and easy to use interface

Getting started:

  package main

  import (
    "github.com/railsmechanic/eventsource"
  )

  func main() {
    // EventSource with default settings
    es := eventsource.New(nil)
    es.Run()
  }

Launched an EventSource service on '127.0.0.1:8080', just with few instructions.
*/
package eventsource
