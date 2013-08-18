package main

import (
	"fmt"
	"net"
	"net/url"
	"io"
	"bufio"
	"strings"
	"strconv"
)

type EventSocket struct {
	conn net.Conn
	reader *bufio.Reader
	uuid string
	fromDid string
	toDid string
	headers map[string]string
}

func NewEventSocket(conn net.Conn) *EventSocket {
	reader := bufio.NewReader(conn)
	headers := make(map[string]string)

	return &EventSocket { conn, reader, "", "", "", headers }
}

func (es *EventSocket) Handle() {
	// Initiate the Event Socket Outbound connection and answer the call
	es.Answer()

	// Send initial setup commands for this channel
	es.Setup()
	
	// Testing: print debug info and run a test XML API Response
	fmt.Println("Channel UUID: " + es.uuid)
	fmt.Println("From DID    : " + es.fromDid)
	fmt.Println("To DID      : " + es.toDid)
	err := es.XmlApiRequest("http://ian-preston.com/jeego/example.xml", nil)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())

		es.SendExecute("hangup")
		es.conn.Close()
		return
	}
	es.conn.Close()
}

func (es *EventSocket) Answer() {
	// Send the 'connect' command to initiate the Event Socket Outbound session
	fmt.Fprintf(es.conn, "connect\n\n")

	// Read in headers from FreeSWITCH. This will also populate the 'uuid', 'fromDid' and 'toDid'
	// properties
	err := es.ReadHeaders()
	if err != nil {
		fmt.Printf("Failed to read headers with error: %s\n", err.Error())
		es.conn.Close()
		return
	}

	es.SendExecute("answer")
}

func (es *EventSocket) Hangup() {
	fmt.Fprintf(es.conn, "sendmsg\ncall-command: execute\nexecute-app-name: hangup\n\n")
}

func (es *EventSocket) Setup() {
	es.SendExecuteArg("set", "tts_engine=flite")
	es.SendExecuteArg("set", "tts_voice=kal")
}

func (es *EventSocket) SendExecute(appName string) {
	fmt.Fprintf(es.conn, "sendmsg\ncall-command: execute\nexecute-app-name: %s\n\n", appName)
	es.ParseResponse()
}

func (es *EventSocket) SendExecuteArg(appName string, appArg string) {
	fmt.Fprintf(es.conn, "sendmsg\ncall-command: execute\nexecute-app-name: %s\nexecute-app-arg: %s\n\n", appName, appArg)
	es.ParseResponse()
}

func (es *EventSocket) SendApi(appName string, appArg string) string {
	fmt.Fprintf(es.conn, "api %s %s %s\n\n", appName, es.uuid, appArg)
	r, _ := es.ParseResponse()
	return r
}

func (es *EventSocket) ParseResponse() (string, error) {
	/**
	 * FreeSWITCH seems to return responses in one of two formats. The first, for 'execute' commands
	 * looks like this:
	 *
	 * Content-Type: command/reply
	 * Reply-Text: +OK did something
	 *
	 * The second, for 'api' commands, looks like thids:
	 *
	 * Content-Type: api/response
	 * Content-Length: 13
	 *
	 * did something
	 *
	 * This method will determine which type of response it is and return the relevant data,
	 * in this example, that would be either '+OK did something' or 'did something', respectively
	 */
	headers := make(map[string]string)

	// Read in headers
	for {
		line, err := es.reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		tokens := strings.Split(line, ": ")
		if len(tokens) < 2 {
			break
		}

		key := strings.Trim(tokens[0], "\n")
		value := strings.Trim(tokens[1], "\n")

		headers[key] = value
	}

	// Handle 'command/reply' responses
	if headers["Content-Type"] == "command/reply" {
		return headers["Reply-Text"], nil
	}

	// Handle 'api/response' responses
	if headers["Content-Type"] == "api/response" {
		numBytes, err := strconv.Atoi(headers["Content-Length"])
		if err != nil {
			return "", err
		}

		buffer := make([]byte, numBytes)
		_, err = io.ReadFull(es.reader, buffer)
		if err != nil {
			return "", err
		}

		return string(buffer), nil
	}

	return "", fmt.Errorf("Event Socket Outbound response was in an unrecognized format")
}

func (es *EventSocket) ReadHeaders() error {
	for {
		line, err := es.reader.ReadString('\n')
		if err != nil {
			return err
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

		fmt.Println(line)
	}

	es.uuid = es.headers["Channel-Unique-ID"]
	es.fromDid = es.headers["Caller-Caller-ID-Number"]
	es.toDid = es.headers["variable_sip_to_user"]

	return nil
}

func (es *EventSocket) XmlApiRequest(rootUrl string, additionalRequestParams url.Values) error {
	x, err := NewXMLAPI(es, rootUrl, additionalRequestParams)
	if err != nil {
		return err
	}

	err = x.EvaluateAll()
	if err != nil {
		return err
	}

	return nil
}