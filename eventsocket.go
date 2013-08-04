package main

import (
	"fmt"
	"net"
	"bufio"
	"strings"
	"time"
)

type EventSocket struct {
	conn net.Conn
	reader *bufio.Reader
	uuid string
	callerId string
	headers map[string]string
}

func NewEventSocket(conn net.Conn) *EventSocket {
	reader := bufio.NewReader(conn)
	headers := make(map[string]string)

	return &EventSocket { conn, reader, "", "", headers }
}

func (es *EventSocket) Handle() {
	// Initiate the Event Socket Outbound connection and answer the call
	es.Answer()

	// Read in headers from FreeSWITCH. This will also populate the 'uuid' and 'callerId'
	// properties
	es.ReadHeaders()
	
	// Testing: print debug info, play a tone for 1.5 seconds, hang up
	fmt.Println("Channel UUID: " + es.uuid)
	fmt.Println("Called ID   : " + es.callerId)
	es.SendExecuteArg("playback", "{loops=-1}tone_stream://%%(251,0,1004)")
	time.Sleep(time.Millisecond * 1500)
	es.SendExecute("break")
	time.Sleep(time.Millisecond * 1000)
	es.SendExecute("hangup")

	es.conn.Close()
}

func (es *EventSocket) ReadHeaders() {
	for {
		line, err := es.reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from socket")
			es.conn.Close()
			return
		}

		tokens := strings.Split(line, ": ")
		if len(tokens) < 2 {
			// If the string could not be split, this is probably a blank line, the
			// end of the headers
			break
		}

		key := strings.Trim(tokens[0], "\n")
		value := strings.Trim(tokens[1], "\n")

		es.headers[key] = value
	}

	es.uuid = es.headers["Channel-Unique-ID"]
	es.callerId = es.headers["Caller-Caller-ID-Number"]
}

func (es *EventSocket) Answer() {
	fmt.Fprintf(es.conn, "connect\n\n")
	fmt.Fprintf(es.conn, "sendmsg\ncall-command: execute\nexecute-app-name: answer\n\n")
}

func (es *EventSocket) SendExecute(appName string) {
	fmt.Fprintf(es.conn, "sendmsg\ncall-command: execute\nexecute-app-name: %s\n\n", appName)
	es.EatResponse()
}

func (es *EventSocket) SendExecuteArg(appName string, appArg string) {
	fmt.Fprintf(es.conn, "sendmsg\ncall-command: execute\nexecute-app-name: %s\nexecute-app-arg: %s\n\n", appName, appArg)
	es.EatResponse()
}

func (es *EventSocket) EatResponse() {
	for {
		line, err := es.reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from socket")
			es.conn.Close()
			return
		}

		tokens := strings.Split(line, ": ")
		if len(tokens) < 2 {
			return
		}
	}
}