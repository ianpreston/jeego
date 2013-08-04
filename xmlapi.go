package main

import (
	"encoding/xml"
)

type Command interface {
	Evaluate(es *EventSocket)
}

type Say struct {
	Message string
}

func (s *Say) Evaluate(es *EventSocket) {
	es.SendExecuteArg("speak", s.Message)
}

type Read struct {
	Digits int 
	Action string
}

func (r *Read) Evaluate(es *EventSocket) {
}


type XmlResponse struct {
	Commands []XmlCommand `xml:",any"`
}

type XmlCommand struct {
	// The name of the command, such as 'Say' or 'Read'
	XMLName xml.Name
	
	// 'Say' fields
	Message string `xml:"message,attr"`

	// 'Read' fields
	Digits int `xml:"digits,attr"`
	Action string `xml:"action,attr"`
}

func ParseXmlApiResponse(xmlSrc string) []Command {
	// Parse XML response into an XmlResponse object
	var r XmlResponse
	err := xml.Unmarshal([]byte(xmlSrc), &r)
	if err != nil {
		return nil
	}

	// Iterate over the XmlCommands in the XmlResponse, and create
	// Command objects
	commands := make([]Command, len(r.Commands))
	for i := 0; i < len(r.Commands); i ++ {
		cmd := r.Commands[i]
		cmdName := cmd.XMLName.Local

		switch cmdName {
		case "Say":
			c := &Say{ cmd.Message }
			commands[i] = c
		case "Read":
			c := &Read{ cmd.Digits, cmd.Action }
			commands[i] = c
		}
	}

	return commands
}
