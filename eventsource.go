// Copyright 2014 Matthias Kalb, Railsmechanic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eventsource

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"runtime"
	"sort"
	"strings"
)

const (
	globalChannel = "all"
)

// Interface of EventSource
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

// EventSource stores information required by the event source service.
type eventSource struct {
	messageRouter   chan *eventMessage
	expireConsumer  chan *consumer
	addConsumer     chan *consumer
	closeChannel    chan string
	stopApplication chan bool
	settings        *Settings
	consumers       map[string][]*consumer
}

// New builds and returns a configured EventSource instance.
// The instance is configured with default settings if no settings are given.
// It starts a goroutine, which is the 'main hub' of the EventSource service.
func New(settings *Settings) EventSource {
	if settings == nil {
		settings = &Settings{}
	}

	es := &eventSource{
		messageRouter:   make(chan *eventMessage),
		expireConsumer:  make(chan *consumer),
		addConsumer:     make(chan *consumer),
		closeChannel:    make(chan string),
		stopApplication: make(chan bool),
		settings:        settings,
		consumers:       make(map[string][]*consumer),
	}

	go es.actionDispatcher()

	return es
}

// Router returns a router that can be used to integrate EventSource in already existing servers
func (es *eventSource) Router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/{channel:[a-z0-9-_]+}", es.subscribeHandler).Methods("GET")
	router.HandleFunc("/{channel:[a-z0-9-_]+}", es.publishHandler).Methods("POST")
	router.HandleFunc("/{channel:[a-z0-9-_]+}", es.closeHandler).Methods("DELETE")
	router.HandleFunc("/{channel:[a-z0-9-_]+}", es.informationHandler).Methods("HEAD")
	router.NotFoundHandler = http.HandlerFunc(channelNotFoundHandler)
	return router
}

// SendMessage sends a message to the consumers of a channel.
// It is also used for sending messages to 'all' consumers.
func (es *eventSource) SendMessage(messageStream io.Reader, channel string) {
	em, err := newEventMessage(messageStream, channel)
	if err != nil {
		log.Printf("[E] Unable to create event message for channel '%s'. %s", channel, err)
		return
	}
	es.messageRouter <- em
}

// ChannelExists checks whether a channel exits.
func (es *eventSource) ChannelExists(channel string) bool {
	_, ok := es.consumers[channel]
	return ok
}

// ConsumerCount returns the amount of consumers subscribed to a channel.
func (es *eventSource) ConsumerCount(channel string) int {
	if consumers, ok := es.consumers[channel]; ok {
		return len(consumers)
	}
	return 0
}

// ConsumerCountAll returns the overall amount of consumers.
func (es *eventSource) ConsumerCountAll() int {
	var consumerCount int
	for _, consumers := range es.consumers {
		consumerCount += len(consumers)
	}
	return consumerCount
}

// Channel returns all available channels.
func (es *eventSource) Channels() []string {
	channels := make([]string, 0)
	for channel, _ := range es.consumers {
		channels = append(channels, channel)
	}
	sort.Strings(channels)
	return channels
}

// Close closes a single, specified channel
// Consumers gets disconnected.
func (es *eventSource) Close(channel string) {
	es.closeChannel <- channel
}

// CloseAll closes all available channels
// Consumers gets disconnected.
func (es *eventSource) CloseAll() {
	es.closeChannel <- globalChannel
}

// Run starts the EventSource service
func (es *eventSource) Run() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	router := es.Router()
	log.Printf("[I] Starting EventSource service on %s:%d\n", es.settings.GetHost(), es.settings.GetPort())
	log.Fatal("[E]", http.ListenAndServe(fmt.Sprintf("%s:%d", es.settings.GetHost(), es.settings.GetPort()), router))
}

// Stop stops the EventSource service
func (es *eventSource) Stop() {
	es.stopApplication <- true
}

// SubscribeHandler handels new, incoming connections of consumers.
// Allowed request type: [GET]
//
// Subscriptions to channel 'all' are rejected, because this is an reserved channel name.
func (es *eventSource) subscribeHandler(rw http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	if channel := params["channel"]; len(channel) > 0 {
		if channel == globalChannel {
			log.Printf("[E] Subscribing consumer on %s to global notification channel 'all' rejected\n", req.RemoteAddr)
			http.Error(rw, "Error: Channel 'all' is reserved for global notifications. Please choose another channel name.", http.StatusBadRequest)
			return
		}

		cr, err := newConsumer(rw, req, es, channel)
		if err != nil {
			log.Printf("[E] Subscribing consumer on %s to channel '%s' failed\n", req.RemoteAddr, channel)
			http.Error(rw, fmt.Sprintf("[E] Unable to connect to channel '%s'.", channel), http.StatusInternalServerError)
			return
		}
		es.addConsumer <- cr
	}
}

// PublishHandler is responsible for publishing messages to channels.
// Allowed request type: [POST]
//
// The Content-Type of this handler need to be 'application/json'.
// If an Auth-Token is set up, only authenticated users can publish messages to channels.
func (es *eventSource) publishHandler(rw http.ResponseWriter, req *http.Request) {
	if !es.Authenticated(req) {
		log.Printf("[E] Authentication of %s failed. Publishing to channel rejected\n", req.RemoteAddr)
		http.Error(rw, "Error: Authentication failed. Publishing to channel rejected.", http.StatusForbidden)
		return
	}

	if !validContentType(req.Header.Get("Content-Type")) {
		log.Printf("[E] Invalid Content-Type sent by %s. Expecting application/json\n", req.RemoteAddr)
		http.Error(rw, "Error: Invalid Content-Type. Expecting application/json.", http.StatusBadRequest)
		return
	}

	params := mux.Vars(req)
	if channel := params["channel"]; len(channel) > 0 {
		es.SendMessage(req.Body, channel)
		defer req.Body.Close()
	}
	rw.WriteHeader(http.StatusCreated)
}

// CloseHandler is responsible for the closing channels
// Allowed request type: [DELETE]
//
// Consumers are disconnected.
// If an Auth-Token is set up, only authenticated users can delete a channel.
func (es *eventSource) closeHandler(rw http.ResponseWriter, req *http.Request) {
	if !es.Authenticated(req) {
		log.Printf("[E] Authentication of %s failed. Closing of channel rejected\n", req.RemoteAddr)
		http.Error(rw, "Error: Authentication failed. Closing of channel rejected.", http.StatusForbidden)
		return
	}

	params := mux.Vars(req)
	if channel := params["channel"]; len(channel) > 0 {
		es.Close(channel)
	}
	rw.WriteHeader(http.StatusOK)
}

// InformationHandler is responsible for the closing channels
// Allowed request type: [HEAD]
//
// If an Auth-Token is set up, only authenticated users can view information of channels.
func (es *eventSource) informationHandler(rw http.ResponseWriter, req *http.Request) {
	if !es.Authenticated(req) {
		log.Printf("[E] Authentication of %s failed. Gettings stats for channel rejected\n", req.RemoteAddr)
		http.Error(rw, "Error: Authentication failed. Gettings stats for channel rejected.", http.StatusForbidden)
		return
	}

	params := mux.Vars(req)
	if channel := params["channel"]; len(channel) > 0 {

		if channel == globalChannel {
			rw.Header().Add("X-Consumer-Count", fmt.Sprint(es.ConsumerCountAll()))
			rw.Header().Add("X-Available-Channels", fmt.Sprintf("[%s]", strings.Join(es.Channels(), ",")))
		} else {
			rw.Header().Add("X-Consumer-Count", fmt.Sprint(es.ConsumerCount(channel)))
			rw.Header().Add("X-Channel-Exists", fmt.Sprint(es.ChannelExists(channel)))
		}

	}
	rw.WriteHeader(http.StatusOK)
}

// ChannelNotFoundHandler is responsible for unknown channels.
// When a consumer wants to connect to an unknown endpoint, an error message is returned.
func channelNotFoundHandler(rw http.ResponseWriter, req *http.Request) {
	log.Printf("[E] Consumer %s tries to join invalid channel", req.RemoteAddr)
	http.Error(rw, "Error: Invalid channel name.", http.StatusNotFound)
}

// Authenticated validates the user submitted AUTH Token.
func (es eventSource) Authenticated(req *http.Request) bool {
	authToken := strings.TrimSpace(req.Header.Get("Auth-Token"))
	if len(es.settings.GetAuthToken()) == 0 && len(authToken) == 0 {
		return true
	}
	return len(es.settings.GetAuthToken()) > 0 && authToken == es.settings.GetAuthToken()
}

// ValidContentType validates the submitted Content-Type.
func validContentType(contentType string) bool {
	if strings.Contains(strings.ToLower(contentType), "application/json") {
		return true
	}
	return false
}

// ActionDispatcher is the central hub of the EventSource service.
func (es *eventSource) actionDispatcher() {
	for {
		select {

		// em.messageRouter is responsible for delivering messages to consumers of channels.
		case em := <-es.messageRouter:
			switch em.Channel {
			default:
				if channelConsumers, ok := es.consumers[em.Channel]; ok {
					for _, channelConsumer := range channelConsumers {
						if cr := channelConsumer; !cr.expired {
							select {
							case cr.inbox <- em:
							default:
							}
						}
					}
				}
			case globalChannel:
				log.Println("[I] Sending global notification to all consumers")
				for _, channelConsumers := range es.consumers {
					for _, channelConsumer := range channelConsumers {
						if cr := channelConsumer; !cr.expired {
							select {
							case cr.inbox <- em:
							default:
							}
						}
					}
				}
			}

		// em.closeChannel is responsible for closing seleted or all channels.
		case channel := <-es.closeChannel:
			switch channel {
			default:
				if channelConsumers, ok := es.consumers[channel]; ok {
					log.Printf("[I] Closing channel '%s' and disconnecting consumers\n", channel)
					for _, channelConsumer := range channelConsumers {
						close(channelConsumer.inbox)
					}
					delete(es.consumers, channel)
				}
			case globalChannel:
				log.Println("[I] Closing all channels and disconnecting consumers")
				for channelName, channelConsumers := range es.consumers {
					for _, channelConsumer := range channelConsumers {
						close(channelConsumer.inbox)
					}
					delete(es.consumers, channelName)
				}
			}

		// em.stopApplication is responsible for shutting down the service properly.
		case <-es.stopApplication:
			log.Println("[I] Halting EventSource server")
			es.closeChannel <- globalChannel
			close(es.messageRouter)
			close(es.addConsumer)
			close(es.expireConsumer)
			close(es.closeChannel)
			close(es.stopApplication)
			return

		// em.addConsumer is responsible for adding consumers to channels.
		case cr := <-es.addConsumer:
			log.Printf("[I] Consumer %s joined channel '%s'\n", cr.connection.RemoteAddr(), cr.channel)
			es.consumers[cr.channel] = append(es.consumers[cr.channel], cr)

		// em.expireConsumer is responsible disconnecting and removing staled consumers.
		case expiredConsumer := <-es.expireConsumer:
			log.Printf("[I] Consumer %s expired and gets removed from channel '%s'\n", expiredConsumer.connection.RemoteAddr(), expiredConsumer.channel)
			if consumers, ok := es.consumers[expiredConsumer.channel]; ok {
				consumerSlice := make([]*consumer, 0)

				for _, cr := range consumers {
					if cr != expiredConsumer {
						consumerSlice = append(consumerSlice, cr)
					}
				}

				es.consumers[expiredConsumer.channel] = consumerSlice
				close(expiredConsumer.inbox)
			}
		}
	}
}
