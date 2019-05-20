package data

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
