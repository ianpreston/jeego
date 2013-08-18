package main

import (
	"net"
	"fmt"
)

type Server struct {
	listener net.Listener
	config *Config
}

func NewServer(config *Config) *Server {
	listener, err := net.Listen("tcp", config.BindTo)
	if err != nil {
		return nil;
	}

	return &Server{ listener, config }
}

func (srv *Server) Listen() {
	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection")
			conn.Close()
			continue
		}
		
		go srv.Handle(conn)
	}
}

func (srv *Server) Handle(conn net.Conn) {
	fmt.Println("Connection accepted")

	es := NewEventSocket(conn, srv)
	es.Handle()
}