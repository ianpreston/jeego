package main

import (
	"net"
	"bufio"
	"fmt"
)

type ESInbound struct {
	EventSocket
}

func NewESInbound(config *Config) (*ESInbound, error) {
	conn, err := net.Dial("tcp", config.InboundDialTo)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)

	return &ESInbound{ EventSocket{ conn, reader, config } }, nil
}

func (es *ESInbound) Setup() error {
	// Read the 'Content-Type: auth/request' line
	es.reader.ReadString('\n')
	es.reader.ReadString('\n')

	// Authenticate
	es.Auth(es.config.InboundPassword)

	return nil
}

func (es *ESInbound) Close() {
	es.conn.Close()
}

func (es *ESInbound) Originate(from string, to string) {
	es.SendOriginate(from, to)
}

func (es *ESInbound) SendApi(appName string, appArg string) string {
	fmt.Fprintf(es.conn, "api %s %s\n\n", appName, appArg)
	r, _ := es.ParseResponse()
	return r
}

func (es *ESInbound) Auth(password string) string {
	fmt.Fprintf(es.conn, "auth %s\n\n", password)
	r, _ := es.ParseResponse()
	return r
}

func (es *ESInbound) SendOriginate(from string, to string) {
	prefix := fmt.Sprintf("{jeego_outbound_number=%s,execute_on_answer='socket 127.0.0.1:8084 full'}", from)
	body := fmt.Sprintf("sofia/gateway/callcentric.com/%s@callcentric.com &park()", to)
	appArg := prefix + body

	es.SendApi("originate", appArg)
}