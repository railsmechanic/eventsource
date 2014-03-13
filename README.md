# EventSource (SSE)
EventSource is a powerful package for quickly setting up an EventSource server in Golang.
It can be used wherever you need to publish events/notifications.

## Features

- Support for isolated channels *(channel1 will not see events of channel2)*
- Support for global notifications across all channels *(every consumer receive this event)*
- RESTful interface for publishing events, deleting, subscribing and getting information of/to channels
- Token base authentication for publishing/deleting/getting information of channels
- Support for CORS *(Allow-Origin, Allow-Method)*
- Allows an individual configuration to set up EventSource for your needs
- Simple and easy to use interface

## Getting Started
First, get EventSource with `go get github.com/railsmechanic/eventsource`

Then create a new `.go` file in your GOPATH. We'll call it `eventserver.go` and add the code below.
~~~go
package main

import (
  "github.com/railsmechanic/eventsource"
)

func main() {
  // EventSource with default settings
  es := eventsource.New(nil)
  es.Run()
}
~~~

Then start the EventSource server with
~~~bash
$ go run eventserver.go
~~~
Now, you will have an EventSource server running on `localhost:8080`, waiting for connections.


#### Listen for events
To test the new EventSource server, just use **curl** and subscribe to a channel called `updates`.
~~~bash
$ curl http://localhost:8080/updates
~~~
You've successfully joined the channel `updates` and you're ready to receive incoming events.


#### Publish events
To publish events, **curl** is your best friend, too.
~~~bash
$ curl -H "Content-Type: application/json" -d '{"id":123, "event":"my-event", "data": "Hello World!"}' http://localhost:8080/updates
$ curl -H "Content-Type: application/json" -d '{"id":456, "event":"my-event", "data": "Hello Again!"}' http://localhost:8080/updates
~~~
Yeah, you've sent two events to the `updates` channel.


#### Received events
Your consumer from above (and each other consumer) listening on channel `updates` has received the following events:
~~~bash
$ curl http://localhost:8080/updates
id: 123
event: my-event
data: Hello World!

id: 456
event: my-event
data: Hello again!
~~~

## Available Settings
To setup EventSource with custom settings, just pass Settings to `New`
~~~go
settings := &Settings{
  Timeout: 30*time.Second,
  AuthToken: "secret",
  Host: "192.168.1.1",
  Port: 3000,
  CorsAllowOrigin: "*",
  CorsAllowMethod: []string{"GET", "POST"}
}
es := eventsource.New(settings)
es.Run()
~~~

**Timeout** *(time.Duration)* - The default timeout for consumers to be disconnected.

**AuthToken** *(string)* - Used to prevent unauthorized users to publish events, delete channels and get information on channels.

**Host** *(string)* - The hostname/ip address on which the EventSource is bind on

**Port** *(uint)* - The port on which the EventSource server will listen on

**CorsAllowOrigin** *(string)* - Allow Cross Site HTTP request e.g. from "*"

**CorsAllowMethod** *([]string)* - Explicit allow Cross Site Request Methods e.g. *"GET", "POST"*

## RESTful Interface or the Go Interface
To communicate with EventSource *(publishing, deleting, etc.)* you can either use the RESTful or the Golang interface.

#### The Go Interface
~~~go
type EventSource interface {
  Router() *mux.Router
  SendMessage(io.Reader, string)
  ChannelExists(channel string) bool
  ConsumerCount(channel string) int
  ConsumerCountAll() int
  Channels() []string
  Close(channel string)
  CloseAll()
  Run()
  Stop()
}
~~~

#### The RESTful interface
To publish events e.g. from other applications or from another host in your network, you can use the RESTful interface.

##### Subscribe to channel/listening for events (GET Request)
`GET: http://example.com/[channel] => Status: 200 OK`

~~~bash
$ curl -X GET http://example.com/[channel]
~~~


##### Publish events/messages (POST Request of Content-Type 'application/json')
`POST: http://example.com/[channel] => Status: 201 Created`

~~~bash
$ curl -X POST -H "Content-Type: application/json" -d '{"id":1, "event":"event", "data": "hello"}' http://example.com/[channel]
~~~


##### Disconnect consumers and delete channel (DELETE Request)
`DELETE: http://example.com/[channel] => Status: 200 OK`

~~~bash
$ curl -X DELETE http://example.com/[channel]
~~~


##### Get information of a channel (HEAD Request)
`HEAD: http://example.com/[channel] => Status: 200 OK`

~~~bash
$ curl -X HEAD -H "Connection: close" http://example.com/[channel]
~~~

*The requested information is returned as addional headers:*

`X-Consumer-Count` Count of consumers subscribed to this channel (integer)

`X-Channel-Exists` Channel exists (bool)

`X-Available-Channels` List of existing channels (array)


## The ALL channel
You already know how to work with individually named channels. For global tasks, EventSource offers the "special" channel name **all**.
To publish events to consumers accross all channels just *POST* your event to the special endpoint `http://example.com/all`.

~~~bash
$ curl -X POST -H "Content-Type: application/json" -d '{"id":1, "event":"event", "data": "hello"}' http://example.com/all
~~~

If you're interested in all the channels available, just *HEAD* to `http://example.com/all`.

~~~bash
$ curl -X HEAD -H "Connection: close" http://example.com/all
~~~


## Things you should know
This EventSource service is mainly implemented to met the requirements of an internal project.
Therefore it's quite possible that not all of the W3C standards are met. You have been warned!


## Special thanks to
This package is based on some concepts of [antage/eventsource](https://github.com/antage/eventsource).
Many thanks and thumbs up, you've done a great job.