package server

type Server interface {
	ServerStats() interface{}
	Done()
}
