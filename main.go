package main

import (
	"github.com/welcomehyunseo/warlord-server/server"
)

func main() {
	lc := server.NewLoggerConfigurator()
	lc.SetLogLevel(server.DebugLevel)
	//lc.EnableReport()
	lc.SetFilter("server-renderer")
	//lc.SetFilter("client-handler")
	//lc.SetFilter("confirm-keep-alive-event-handler")
	//lc.SetFilter("set-entity-relative-pos-event-handler")
	//lc.SetFilter("update-look-event-handler")

	addr := ":9999"
	max := 20
	favicon, desc := "", "Warlord Server for Dev"
	s := server.NewServer(
		addr,
		max,
		favicon,
		desc,
	)

	rndDist := int32(5)
	spawnX, spawnY, spawnZ :=
		float64(0), float64(70), float64(0)
	spawnYaw, spawnPitch :=
		float32(0), float32(0)
	world := server.NewLobby(
		rndDist,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	)
	world.MakeFlat()
	s.Render(
		world,
	)

}
