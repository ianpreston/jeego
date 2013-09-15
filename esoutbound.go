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

type ESOutbound struct {
	conn net.Conn
	reader *bufio.Reader
	eso *ESOListener
	uuid string
	fromDid string
	toDid string
	headers map[string]string
}

func NewESOutbound(conn net.Conn, eso *ESOListener) *ESOutbound {
	reader := bufio.NewReader(conn)
	headers := make(map[string]string)

	return &ESOutbound { conn, reader, eso, "", "", "", headers }
}

func (es *ESOutbound) Handle() {
	// Initiate the Event Socket Outbound connection and answer the call
	es.Answer()

	// Send initial setup commands for this channel
	es.Setup()

	// Print debug info
	fmt.Println("Channel UUID: " + es.uuid)
	fmt.Println("From DID    : " + es.fromDid)
	fmt.Println("To DID      : " + es.toDid)

	// Determine, via the routes from the config file, the root URL for
	// the 'to' DID
	rr := es.eso.config.RouteRuleForDID(es.toDid)
	if rr == nil {
		fmt.Println("There is no route rule configured for this DID")
		es.SendExecute("hangup")
		es.conn.Close()
		return
	}

	// Make an XML API request to this URL
	err := es.XmlApiRequest(rr.URL, nil)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())

		es.SendExecute("hangup")
		es.conn.Close()
		return
	}

	es.conn.Close()
}

func (es *ESOutbound) Answer() {
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

func (es *ESOutbound) Hangup() {
	fmt.Fprintf(es.conn, "sendmsg\ncall-command: execute\nexecute-app-name: hangup\n\n")
}

func (es *ESOutbound) Setup() {
	es.SendExecuteArg("set", "tts_engine=flite")
	es.SendExecuteArg("set", "tts_voice=kal")
}

func (es *ESOutbound) SendExecute(appName string) {
	fmt.Fprintf(es.conn, "sendmsg\ncall-command: execute\nexecute-app-name: %s\n\n", appName)
	es.ParseResponse()
}

func (es *ESOutbound) SendExecuteArg(appName string, appArg string) {
	fmt.Fprintf(es.conn, "sendmsg\ncall-command: execute\nexecute-app-name: %s\nexecute-app-arg: %s\n\n", appName, appArg)
	es.ParseResponse()
}

func (es *ESOutbound) SendApi(appName string, appArg string) string {
	fmt.Fprintf(es.conn, "api %s %s %s\n\n", appName, es.uuid, appArg)
	r, _ := es.ParseResponse()
	return r
}

func (es *ESOutbound) ParseResponse() (string, error) {
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

func (es *ESOutbound) ReadHeaders() error {
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
	}

	es.uuid = es.headers["Channel-Unique-ID"]
	es.fromDid = es.headers["variable_sip_from_user"]
	es.toDid = es.headers["variable_sip_to_user"]

	// TODO On an outbound call, 'variable_jeego_outbound_number' will be
	// the # we are actually calling from.

	return nil
}

func (es *ESOutbound) XmlApiRequest(rootUrl string, additionalRequestParams url.Values) error {
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