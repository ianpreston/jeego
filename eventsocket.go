package main

import (
	"net"
	"io"
	"bufio"
	"strings"
	"strconv"
	"fmt"
)

type EventSocket struct {
	conn net.Conn
	reader *bufio.Reader
	config *Config
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