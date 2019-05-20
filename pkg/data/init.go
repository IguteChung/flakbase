package data

import "time"

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
