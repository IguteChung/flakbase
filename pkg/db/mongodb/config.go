package mongodb

// Config defines the
type Config struct {
	// URI endpoint for mongodb.
	URI string `json:"uri"`
	// Database name to use in mongodb.
	Database string `json:"database"`
	// CollectionsTable name for lookup Flakbase paths.
	CollectionsTable string `json:"collections_table"`
}

const (
	defaultDatabase  = "flakbase"
	defaultCollTable = "flakbase-collection"
)
