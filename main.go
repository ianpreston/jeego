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

	srv := NewServer(config)
	srv.Listen()
}