package server

// Config represents the config object for the server.
type Config struct {
	HTTP HTTP
}

// HTTP represents the http config object for the http server.
type HTTP struct {
	Host string // optional, empty corresponds to "0.0.0.0"
	Port uint64
}

// Enabled returns true if the http server is enabled (i.e. the Port is non-zero).
func (h HTTP) Enabled() bool {
	return h.Port > 0
}
