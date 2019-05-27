package net

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/IguteChung/flakbase/pkg/data"
	"github.com/gorilla/websocket"
)

// upgradable checks whether the header in http request can be upgraded to
// websocket connection.
func upgradable(header http.Header) bool {
	if header == nil {
		return false
	}
	upgrade, ok := header["Upgrade"]
	if !ok {
		return false
	}
	for _, s := range upgrade {
		if s == "websocket" {
			return true
		}
	}
	return false
}

// readMessage reads a request from connection, automatically concats the splitted
// chunks if the request payload too large.
func readMessage(conn *websocket.Conn) (*data.Request, error) {
	// read message from client.
	_, bytes, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %v", err)
	}

	// handle the message if map or number.
	var o interface{}
	if err := json.Unmarshal(bytes, &o); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message %s: %v", string(bytes), err)
	}

	// finally return a request model by bytes.
	toReq := func(bytes []byte) (*data.Request, error) {
		var r *data.Request
		if err := json.Unmarshal(bytes, &r); err != nil {
			log.Printf("failed to unmarshal to request: %v", err)
			return nil, nil
		}
		return r, nil
	}

	switch v := o.(type) {
	case map[string]interface{}:
		// if the message is a json map, directly return.
		return toReq(bytes)
	case float64:
		if v == 0 {
			return nil, nil
		}
		// try read message with given count.
		var buffer []byte
		for i := 0; i < int(v); i++ {
			// append the bytes into buffer
			_, bytes, err := conn.ReadMessage()
			if err != nil {
				return nil, fmt.Errorf("failed to read message: %v", err)
			}
			buffer = append(buffer, bytes...)
		}

		// unmarshal the collected buffer.
		return toReq(buffer)
	default:
		return nil, fmt.Errorf("invalid message: %v", o)
	}
}
