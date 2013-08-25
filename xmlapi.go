package main

import (
	"net/http"
	"net/url"
	"encoding/xml"
	"io/ioutil"
	"os"
	"fmt"
)

/*
 * XML API Manager
 */
type XMLAPI struct {
	es *EventSocket
	rootUrl string
	additionalRequestParams url.Values

	commands []Command
}

func NewXMLAPI(es *EventSocket, rootUrl string, additionalRequestParams url.Values) (*XMLAPI, error) {
	x := &XMLAPI{ es, rootUrl, additionalRequestParams, nil }

	xmlSrc, err := x.MakeRequest()
	if err != nil {
		return nil, err
	}

	x.commands, err = x.ParseResponse(xmlSrc)
	if err != nil {
		return nil, err
	}

	return x, nil
}

func (x *XMLAPI) MakeRequest() (string, error) {
	// Merge the provided request URL, the default paramaters, and the specified additional
	// paramaters to create the request URL w/ querystring
	params := x.DefaultRequestParams()
	if x.additionalRequestParams != nil {
		for k, v := range x.additionalRequestParams {
			params[k] = v
		}
	}
	requestUrl := fmt.Sprintf("%s?%s", x.rootUrl, params.Encode())
	fmt.Println(requestUrl)

	// Send the request
	res, err := http.Get(requestUrl)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("remote host returned non-200 status code for url: %s", x.rootUrl)
	}

	// Read & return the response
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (x *XMLAPI) DefaultRequestParams() url.Values {
	return url.Values{ "fromDid": {x.es.fromDid}, "toDid": {x.es.toDid} }
}

func (x *XMLAPI) ParseResponse(xmlSrc string) ([]Command, error) {
	// Parse XML response into an XmlResponse object
	r := &XmlResponse{}
	err := xml.Unmarshal([]byte(xmlSrc), r)
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
		case "Bridge":
			c := &Bridge{ cmd.Did }
			commands[i] = c
		default:
			return nil, fmt.Errorf("Invalid XML API command: %s", cmd.XMLName.Local)
		}
	}

	return commands, nil
}

func (x *XMLAPI) EvaluateAll() error {
	for i := 0; i < len(x.commands); i ++ {
		err := x.commands[i].Evaluate(x.es)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
 * XML API Commands
 */
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
	cwd, _ := os.Getwd()
	es.SendExecuteArg("read", fmt.Sprintf("%v %v %v/empty.wav digits 10000 #", r.Digits, r.Digits, cwd))
	digits := es.SendApi("uuid_getvar", "digits")

	err := es.XmlApiRequest(r.Action, url.Values{ "digits": {digits} })
	if err != nil {
		return err
	}

	return nil
}

type Bridge struct {
	Did string
}

func (b Bridge) Evaluate(es *EventSocket) error {
	bridgeSipUrl := fmt.Sprintf("sofia/gateway/%s/%s@%s", es.eso.config.SofiaGatewayName, b.Did, es.eso.config.SofiaGatewayHost)
	es.SendExecuteArg("bridge", bridgeSipUrl)
	return nil
}

/*
 * Data structures for deserializing XML
 */
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

	// 'Bridge' fields
	Did string `xml:"did,attr"`
}