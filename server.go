package main

import (
	"net"
	"fmt"
)

type ESOListener struct {
	listener net.Listener
	config *Config
}

func NewESOListener(config *Config) *ESOListener {
	listener, err := net.Listen("tcp", config.BindTo)
	if err != nil {
		return nil;
	}

	return &ESOListener{ listener, config }
}

func (eso *ESOListener) Listen() {
	for {
		conn, err := eso.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection")
			conn.Close()
			continue
		}
		
		go eso.Handle(conn)
	}
}

func (eso *ESOListener) Handle(conn net.Conn) {
	fmt.Println("Connection accepted")

	es := NewEventSocket(conn, eso)
	es.Handle()
}