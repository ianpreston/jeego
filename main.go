package main

import (
	"flag"
)

func main() {
	flag.Parse()
	configFilePath := flag.Arg(0)
	if configFilePath == "" {
		configFilePath = "/etc/jeego.xml"
	}
	config := LoadConfig(configFilePath)

	eso := NewESOListener(config)
	go eso.Listen()

	hl := NewHTTPListener(config)
	go hl.Listen()

	<- make(chan int)
}