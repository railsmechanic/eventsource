// Copyright 2014 Matthias Kalb, Railsmechanic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package railsmechanic/eventsource implements Server Sent Events (SSE) for Go.

This implementation of an EventSource service offers multiple, isolated channels and
a simple RESTful interface. Publishing events to a channel or the deleting of a channel
can be secured by an easy to use AUTH token.

The main features are:

  - Support for multiple, isolated channels
  - Support for sending global notifications to all consumers
  - Publishing new events just by 'posting' JSON data to a channel endpoint
  - Subscribing to a channel simply by calling the endpoint via a GET request
  - Deleting of ab entire channel by sending a DELETE request to a channel endpoint
  - Simple, token based authentifaction for publishing/deleting (to) channels
  - Support for CORS
  - Easy to use interface with multiple information methods

Let's start with a basic example:

  func main() {
    es := eventsource.New(nil)

    // Run the service on '127.0.0.1:8080'
    es.Run()
  }

Just with this few instructions we launched an EventSource service on '127.0.0.1:8080'.
*/
package eventsource
