package config

type ListenerStruct struct {
	AddressToListen string
	EventName       string
}

type Config struct {
	Job1 ListenerStruct
	Job2 ListenerStruct
}
