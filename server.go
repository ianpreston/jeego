package main

import (
	"net"
	"fmt"
)

type Server struct {
	listener net.Listener
}

func NewServer(bindTo string) *Server {
	listener, err := net.Listen("tcp", bindTo)
	if err != nil {
		return nil;
	}

	return &Server{ listener }
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

	es := NewEventSocket(conn)
	es.Handle()
}