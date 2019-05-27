package data

import "time"

// O defines a JSON object abbreviation.
type O map[string]interface{}

// Message defines an interface which can be formatted as a JSON object for sending.
type Message interface {
	// Format converts a Message to be JSON marshallable.
	Format() O
}

// InitMessage defines the response message when connection is established.
type InitMessage struct {
	Now  time.Time
	Host string
}

// Format formats a message into response.
func (i InitMessage) Format() O {
	return O{
		"d": O{
			"t": "h",
			"d": O{
				"ts": i.Now.Unix() * 1000,
				"v":  "5",
				"h":  i.Host,
				"s":  "",
			},
		},
		"t": "c",
	}
}

// IdleMessage defines the response message when idle event received.
type IdleMessage struct{}

// Format formats a message into response.
func (i IdleMessage) Format() O {
	return O{
		"d": O{
			"t": "o",
			"d": nil,
		},
		"t": "c",
	}
}

// OkMessage defines the response message when request is handled.
type OkMessage struct {
	RequestID int64
	NoIndex   bool
}

// Format formats a message into response.
func (m OkMessage) Format() O {
	d := O{}
	if m.NoIndex {
		d["w"] = []string{"no_index"}
	}
	return O{
		"d": O{
			"r": m.RequestID,
			"b": O{
				"s": "ok",
				"d": d,
			},
		},
		"t": "d",
	}
}

// ListenMessage defines the response message when listen event received.
type ListenMessage struct {
	Ref     string
	QueryID int64
	Data    interface{}
}

// Format formats a message into response.
func (m ListenMessage) Format() O {
	return O{
		"d": O{
			"a": "d",
			"b": O{
				"p": m.Ref,
				"d": m.Data,
				"t": m.QueryID,
			},
		},
		"t": "d",
	}
}
