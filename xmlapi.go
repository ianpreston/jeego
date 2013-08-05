package main

import (
	"encoding/xml"
	"net/http"
	"io/ioutil"
	"fmt"
)

type Command interface {
	Evaluate(es *EventSocket) error
}

type Say struct {
	Message string
}

func (s Say) Evaluate(es *EventSocket) error {
	es.SendExecuteArg("speak", s.Message)
	return nil
}

type Read struct {
	Digits int 
	Action string
}

func (r Read) Evaluate(es *EventSocket) error {
	es.SendExecuteArg("read", fmt.Sprintf("%v %v conference/8000/conf-pin.wav digits 10000 #", r.Digits, r.Digits))
	es.SendExecuteArg("phrase", "spell,${digits}")
	err := es.XmlApiRequest(r.Action)
	if err != nil {
		return err
	}

	return nil
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

func MakeXmlApiRequest(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("remote host returned non-200 status code for url: %s", url)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func ParseXmlApiResponse(xmlSrc string) ([]Command, error) {
	// Parse XML response into an XmlResponse object
	var r XmlResponse
	err := xml.Unmarshal([]byte(xmlSrc), &r)
	if err != nil {
		return nil, err
	}

	// Iterate over the XmlCommands in the XmlResponse, and create
	// Command objects
	commands := make([]Command, len(r.Commands))
	for i := 0; i < len(r.Commands); i ++ {
		cmd := r.Commands[i]

		switch cmd.XMLName.Local {
		case "Say":
			c := &Say{ cmd.Message }
			commands[i] = c
		case "Read":
			c := &Read{ cmd.Digits, cmd.Action }
			commands[i] = c
		}
	}

	return commands, nil
}