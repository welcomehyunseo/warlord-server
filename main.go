package main

import (
	"github.com/welcomehyunseo/warlord-server/server"
)

func main() {
	lc := server.NewLoggerConfigurator()
	lc.SetLogLevel(server.DebugLevel)
	lc.EnableReport()
	lc.SetFilter("server-renderer")
	//lc.SetFilter("client-handler")
	//lc.SetFilter("confirm-keep-alive-event-handler")
	//lc.SetFilter("set-entity-relative-pos-event-handler")
	//lc.SetFilter("update-look-event-handler")

	addr := ":9999"
	max := 20
	favicon, desc := "", "Warlord Server for Dev"
	spawnX, spawnY, spawnZ :=
		float64(8), float64(70), float64(8)
	spawnYaw, spawnPitch :=
		float32(0), float32(0)

	s := server.NewServer(
		addr,
		max,
		favicon,
		desc,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	)

	//for cz := int32(10); cz >= -10; cz-- {
	//	for cx := int32(10); cx >= -10; cx-- {
	//		chunk := server.NewChunk(cx, cz)
	//		part := server.NewChunkPart()
	//		for z := 0; z < server.ChunkPartWidth; z++ {
	//			for x := 0; x < server.ChunkPartWidth; x++ {
	//				part.SetBlock(uint8(x), 0, uint8(z), server.StoneBlock)
	//			}
	//		}
	//
	//		chunk.SetChunkPart(4, part)
	//		s.AddChunk(cx, cz, chunk)
	//	}
	//}

	rndDist := int32(4)
	playerList := server.NewPlayerList()
	world := server.NewOverworld(rndDist)
	s.Render(
		playerList,
		world,
	)

}
