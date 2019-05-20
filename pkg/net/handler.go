package net

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/IguteChung/flakbase/pkg/data"
	"github.com/IguteChung/flakbase/pkg/store"
	"github.com/gorilla/websocket"
)

type handler struct {
	*Config
	upgrader websocket.Upgrader
}

func (s *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check if the request can be upgraded to websocket.
	if upgradable(r.Header) {
		if err := s.serveWebsocket(w, r); err != nil {
			log.Printf("failed to serve websocket: %v", err)
		}
		return
	}
}

func (s *handler) serveWebsocket(w http.ResponseWriter, r *http.Request) error {
	// upgrade the http connection to websocket.
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade websocket: %v", err)
	}
	defer conn.Close()

	// prepare send util.
	send := func(m data.Message) error {
		log.Printf("[message sent] %+v", conn.RemoteAddr())
		return conn.WriteJSON(m.Format())
	}

	// send initial message.
	if err := send(data.InitMessage{Now: time.Now(), Host: s.Host + s.Port}); err != nil {
		return fmt.Errorf("failed to send initial message: %v", err)
	}

	// generate a handler to handle the incoming messages.
	handler := &store.Handler{}

	// iterating on receiving client messages.
	for {
		// read a request from connection.
		r, err := readMessage(conn)
		if err != nil {
			return fmt.Errorf("failed to read message: %v", err)
		}
		log.Printf("[message received] %+v: %+v", conn.RemoteAddr(), r)

		// handle the request.
		if err := handler.Handle(r); err != nil {
			return fmt.Errorf("failed to handle request %+v: %v", r, err)
		}

		// send ok message if properly handled.
		if err := send(data.OkMessage{RequestID: r.RequestID}); err != nil {
			return fmt.Errorf("failed to send ok message: %v", err)
		}
	}
}
