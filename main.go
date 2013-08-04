package main

import (
)

func main() {
	srv := NewServer("0.0.0.0:8084")
	srv.Listen()
}