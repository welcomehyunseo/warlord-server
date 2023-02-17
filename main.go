package main

import (
	"github.com/welcomehyunseo/warlord-server/server"
)

func main() {
	lc := server.NewLoggerConfigurator()
	lc.SetLogLevel(server.DebugLevel)
	//lc.EnableReport()
	//lc.SetFilter("server-renderer")
	//lc.SetFilter("client-handler")
	//lc.SetFilter("load-chunk-event-handler")
	lc.SetFilter("play-state-handler")

	addr := ":9999"
	max := 20
	favicon, desc := "", "Warlord Server for Dev"
	spawnX, spawnY, spawnZ :=
		float64(0), float64(70), float64(0)
	spawnYaw, spawnPitch :=
		float32(0), float32(0)

	rndDist := 4
	world := server.NewOverworld(
		rndDist,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	)
	world.MakeFlat()

	server := server.NewServer(
		addr,
		max,
		favicon,
		desc,
		world,
	)

	server.Render()

}
