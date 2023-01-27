package main

import (
	"bytes"
	"github.com/welcomehyunseo/warlord-server/server"
)

func main() {
	s := server.NewServer()

	for z := 5; z >= -5; z-- {
		for x := 5; x >= -5; x-- {
			chunk := server.NewChunk()
			for z := 0; z < server.ChunkWidth; z++ {
				for x := 0; x < server.ChunkWidth; x++ {
					chunk.SetBlock(uint8(x), 0, uint8(z), server.StoneBlock)
				}
			}

			s.SetChunk(x, 0, z, chunk)
		}
	}

	b := bytes.NewBuffer(nil)
	b.Bytes()

	s.Render()
}
