package main

import "github.com/welcomehyunseo/warlord-server/server"

func main() {
	s := server.NewServer()
	s.Render()
}
