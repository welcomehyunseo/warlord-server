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
	//lc.SetFilter("play-state-handler")
	//lc.SetFilter("add-player-event-handler")
	//lc.SetFilter("spawn-player-event-handler")

	rndDist := 4
	spawnX, spawnY, spawnZ :=
		float64(0), float64(70), float64(0)
	spawnYaw, spawnPitch :=
		float32(0), float32(0)
	world := server.NewOverworld(
		rndDist,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	)
	world.MakeFlat()

	addr := ":9999"
	max := 20
	favicon, desc := "", "Warlord Server for Dev"
	srv := server.NewServer(
		addr,
		max,
		favicon,
		desc,
		world,
	)
	srv.Render()

}
