package data

// O defines a JSON object abbreviation.
type O map[string]interface{}

// Message defines an interface which can be formatted as a JSON object for sending.
type Message interface {
	// Format converts a Message to be JSON marshallable.
	Format() O
}
