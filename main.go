package main

import (
	"github.com/welcomehyunseo/warlord-server/server"
)

func main() {
	addr := ":9999"
	max := 20
	favicon, desc := "", "Warlord Server for Dev"
	rndDist := 4
	spawnX, spawnY, spawnZ :=
		float64(0), float64(70), float64(0)
	spawnYaw, spawnPitch :=
		float32(0), float32(0)

	s, err := server.NewServer(
		addr,
		max,
		favicon,
		desc,
		rndDist,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	)
	if err != nil {
		panic(err)
	}

	for cz := 10; cz >= -10; cz-- {
		for cx := 10; cx >= -10; cx-- {
			chunk := server.NewChunk()
			part := server.NewChunkPart()
			for z := 0; z < server.ChunkPartWidth; z++ {
				for x := 0; x < server.ChunkPartWidth; x++ {
					part.SetBlock(uint8(x), 0, uint8(z), server.StoneBlock)
				}
			}

			chunk.SetChunkPart(4, part)
			s.AddChunk(cx, cz, chunk)
		}
	}
	s.Render()
}
