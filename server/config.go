package server

type Config struct {
	HTTP HTTP
}

type HTTP struct {
	Port uint64
}
