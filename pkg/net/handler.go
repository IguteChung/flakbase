package net

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/IguteChung/flakbase/pkg/data"
	"github.com/IguteChung/flakbase/pkg/store"
	"github.com/gorilla/websocket"
)

type handler struct {
	*Config
	upgrader  websocket.Upgrader
	datastore store.Handler
}

func (s *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check if the request can be upgraded to websocket.
	ctx := r.Context()
	if upgradable(r.Header) {
		if err := s.serveWebsocket(ctx, w, r); err != nil {
			log.Printf("failed to serve websocket: %v", err)
		}
		return
	}
}

func (s *handler) serveWebsocket(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// upgrade the http connection to websocket.
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade websocket: %v", err)
	}
	defer conn.Close()

	// prepare send util and lock to avoid concurrent write.
	mux := sync.Mutex{}
	send := func(m data.Message) error {
		mux.Lock()
		defer mux.Unlock()
		log.Printf("[message sent] %+v", conn.RemoteAddr())
		return conn.WriteJSON(m.Format())
	}

	// send initial message.
	if err := send(data.InitMessage{Now: time.Now(), Host: s.Host + s.Port}); err != nil {
		return fmt.Errorf("failed to send initial message: %v", err)
	}

	// generate listen channel.
	ch := make(store.ListenChannel)
	go func() {
		for msg := range ch {
			if err := send(msg); err != nil {
				log.Printf("failed to send listen message: %v", err)
			}
		}
	}()

	// iterating on receiving client messages.
	for {
		// read a request from connection.
		r, err := readMessage(conn)
		if err != nil {
			return fmt.Errorf("failed to read message: %v", err)
		}
		log.Printf("[message received] %+v: %+v", conn.RemoteAddr(), r)

		// handle the request by request type.
		result := &store.ListenResult{}
		switch r.Type {
		case data.TypeSet:
			err = s.datastore.HandleSet(ctx, r.Ref, r.Data)
		case data.TypeUpdate:
			err = s.datastore.HandleUpdate(ctx, r.Ref, r.Data)
		case data.TypeListen:
			result, err = s.datastore.HandleListen(ctx, r.Ref, r.Query, ch)
			defer s.datastore.HandleUnlisten(ctx, r.Ref, r.Query, ch)
		case data.TypeUnlisten:
			err = s.datastore.HandleUnlisten(ctx, r.Ref, r.Query, ch)
		}
		if err != nil {
			return fmt.Errorf("failed to handle request %+v: %v", r, err)
		}

		// send ok message if properly handled.
		if err := send(data.OkMessage{RequestID: r.RequestID, NoIndex: result.NoIndex}); err != nil {
			return fmt.Errorf("failed to send ok message: %v", err)
		}
	}
}
