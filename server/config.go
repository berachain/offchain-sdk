package server

// Config represents the config object for the server.
type Config struct {
	HTTP HTTP
}

// HTTP represents the http config object for the http server.
type HTTP struct {
	Port uint64
}
