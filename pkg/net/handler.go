package net

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

	// serve restful api.
	// TODO: enable cors now.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := s.serveRestful(ctx, w, r); err != nil {
		// not handled error.
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
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
	if err := send(data.InitMessage{Now: time.Now(), Host: s.Host}); err != nil {
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
		} else if r == nil {
			continue
		}
		log.Printf("[message received] %+v: %+v", conn.RemoteAddr(), r)

		// handle the request asynchronously.
		go func() {
			// handle the request by request type.
			var err error
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
			case data.TypeIdle:
				// send idle message.
				if err := send(data.IdleMessage{}); err != nil {
					log.Printf("failed to send idle message: %v", err)
				}
				return
			}
			if err != nil {
				// request not properly handled.
				log.Printf("failed to handle request %+v: %v", r, err)
			}

			// send ok message.
			if err := send(data.OkMessage{RequestID: r.RequestID, NoIndex: result.NoIndex}); err != nil {
				log.Printf("failed to send ok message: %v", err)
			}
		}()

		// unlisten if connection broke.
		if r.Type == data.TypeListen {
			defer s.datastore.HandleUnlisten(ctx, r.Ref, r.Query, ch)
		}

	}
}

func (s *handler) serveRestful(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// check url is valid.
	u := r.URL.Path
	if !strings.HasSuffix(u, ".json") {
		return fmt.Errorf("invalid path %s, should end with .json", u)
	}

	// truncate .json to get path ref.
	ref := u[:len(u)-len(".json")]

	// handle the request by method.
	switch r.Method {
	case http.MethodGet:
		query, err := ParseQuery(r.URL.Query())
		if err != nil {
			return fmt.Errorf("invalid query: %v", err)
		}

		// get the data from store.
		data, err := s.datastore.HandleGet(ctx, ref, *query)
		if err != nil {
			return fmt.Errorf("failed to handle get %s: %v", ref, err)
		}

		// marshal the data to bytes for response.
		bytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal response: %v", err)
		}
		w.Write(bytes)
	case http.MethodPut, http.MethodPatch:
		// decode the json body for set or update.
		var data interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return fmt.Errorf("failed to unmarshal body: %v", err)
		}

		// call set or update according to method.
		if r.Method == http.MethodPut {
			if err := s.datastore.HandleSet(ctx, ref, data); err != nil {
				return fmt.Errorf("failed to handle set %s: %v", ref, err)
			}
		} else {
			if err := s.datastore.HandleUpdate(ctx, ref, data); err != nil {
				return fmt.Errorf("failed to handle update %s: %v", ref, err)
			}
		}
	case http.MethodDelete:
		if err := s.datastore.HandleSet(ctx, ref, nil); err != nil {
			return fmt.Errorf("failed to handle remove %s: %v", ref, err)
		}
	default:
		return fmt.Errorf("not supported method: %s", r.Method)
	}

	return nil
}
