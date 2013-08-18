package main

func main() {
	config := LoadConfig("config.xml")
	srv := NewServer(config)
	srv.Listen()
}