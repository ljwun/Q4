// modified from https://github.com/gin-gonic/examples/blob/master/server-sent-event/main.go
package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"
)

// New event messages are broadcast to all registered client connection channels
type ClientChan chan BidRequest

type ServerChan chan BidNotification

type RegisterChan chan RegisterInfo

type RegisterInfo struct {
	ItemID string
	C      ClientChan
}

// It keeps a list of clients those are currently attached
// and broadcasting events to those clients.
type Event struct {
	// Events are pushed to this channel by the main events-gathering routine
	Message ServerChan

	// New client connections
	NewClients RegisterChan

	// Closed client connections
	ClosedClients RegisterChan

	// Total client connections
	TotalClients map[string]map[ClientChan]struct{}

	DistributionMQ           *nats.Conn
	DistributionMessage      chan *nats.Msg
	DistributionSubscription *nats.Subscription
}

// Initialize event and Start procnteessing requests
func NewSSEServer(mqURL string) (*Event, error) {
	mq, err := nats.Connect(mqURL)
	if err != nil {
		return nil, fmt.Errorf("fail to connect to nats, err=%w", err)
	}
	ch := make(chan *nats.Msg, 64)
	sub, err := mq.ChanSubscribe("notification", ch)
	if err != nil {
		return nil, fmt.Errorf("fail to subscribe to nats, err=%w", err)
	}
	event := &Event{
		Message:                  make(ServerChan),
		NewClients:               make(RegisterChan),
		ClosedClients:            make(RegisterChan),
		TotalClients:             make(map[string]map[ClientChan]struct{}),
		DistributionSubscription: sub,
		DistributionMessage:      ch,
		DistributionMQ:           mq,
	}
	go event.listen()
	return event, nil
}

// It Listens all incoming requests from clients.
// Handles addition and removal of clients and broadcast messages to clients.
func (stream *Event) listen() {
	for {
		select {
		// Add new available client
		case data := <-stream.NewClients:
			_, ok := stream.TotalClients[data.ItemID]
			if !ok {
				stream.TotalClients[data.ItemID] = make(map[ClientChan]struct{})
			}
			stream.TotalClients[data.ItemID][data.C] = struct{}{}
			log.Printf("Client added into item {%s}. %d registered clients", data.ItemID, len(stream.TotalClients[data.ItemID]))

		// Remove closed client
		case data := <-stream.ClosedClients:
			delete(stream.TotalClients[data.ItemID], data.C)
			close(data.C)
			log.Printf("Removed client from item {%s}. %d registered clients", data.ItemID, len(stream.TotalClients[data.ItemID]))
			if len(stream.TotalClients[data.ItemID]) == 0 {
				delete(stream.TotalClients, data.ItemID)
			}

		// Publish message into nats
		case notification := <-stream.Message:
			data, err := msgpack.Marshal(notification)
			if err != nil {
				log.Printf("cannot encode notification, err=%v", err)
				continue
			}
			if err := stream.DistributionMQ.Publish("notification", data); err != nil {
				log.Printf("cannot send notification into nats, err=%v", err)
			}

		// Broadcast message to client
		case data := <-stream.DistributionMessage:
			notification := new(BidNotification)
			if err := msgpack.Unmarshal(data.Data, notification); err != nil {
				log.Printf("cannot retrieve notification from nats, err=%v", err)
				continue
			}
			for clientMessageChan := range stream.TotalClients[notification.ItemID] {
				clientMessageChan <- notification.Info
			}
		}
	}
}

func (stream *Event) serveHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize client channel
		registerInfo := RegisterInfo{
			ItemID: c.Param("itemID"),
			C:      make(ClientChan, 1),
		}

		// Send new connection to event server
		stream.NewClients <- registerInfo

		defer func() {
			// Send closed connection to event server
			stream.ClosedClients <- registerInfo
		}()

		c.Set("sseChannel", registerInfo.C)
		c.Next()

		w := c.Writer
		clientGone := w.CloseNotify()
		for {
			select {
			case <-clientGone:
				return
			case msg := <-registerInfo.C:
				c.SSEvent("message", msg)
				w.Flush()
			}
		}
	}
}

func HeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}
