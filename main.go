package main

import (
	"bytes"
	"github.com/welcomehyunseo/warlord-server/server"
)

func main() {
	addr := ":9999"
	max := 20
	favicon := ""
	desc := "Hello, World!"
	rndDist := 3

	s, err := server.NewServer(
		addr,
		max,
		favicon,
		desc,
		rndDist,
		)
	if err != nil {
		panic(err)
	}

	for cz := 20; cz >= -20; cz-- {
		for cx := 20; cx >= -20; cx-- {
			chunk := server.NewChunk()
			for z := 0; z < server.ChunkWidth; z++ {
				for x := 0; x < server.ChunkWidth; x++ {
					chunk.SetBlock(uint8(x), 0, uint8(z), server.StoneBlock)
				}
			}

			s.SetChunk(cx, 0, cz, chunk)
		}
	}

	b := bytes.NewBuffer(nil)
	b.Bytes()

	s.Render()
}
