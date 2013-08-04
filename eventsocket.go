package main

import (
	"fmt"
	"net"
	"bufio"
	"strings"
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

	// Send initial setup commands for this channel
	es.Setup()
	
	// Testing: print debug info and run a test XML API Response
	fmt.Println("Channel UUID: " + es.uuid)
	fmt.Println("Called ID   : " + es.callerId)
	xmlSrc := `
		<Response>
			<Say message="Hello, world! This is built with Jeego, and FreeSWITCH." />
			<Say message="Testing one two three four five" />
			<Read digits="2" action="http://example.com/" />
		</Response>
	`
	es.EvaluateXmlApiResponse(xmlSrc)

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

func (es *EventSocket) Setup() {
	es.SendExecuteArg("set", "tts_engine=flite")
	es.SendExecuteArg("set", "tts_voice=kal")
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

func (es *EventSocket) EvaluateXmlApiResponse(xmlSrc string) {
	commands := ParseXmlApiResponse(xmlSrc)

	for i := 0; i < len(commands); i ++ {
		commands[i].Evaluate(es)
	}
}