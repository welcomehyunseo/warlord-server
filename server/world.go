package server

import (
	"fmt"
	"go/types"
	"sort"
	"sync"
)

type ChunkPosStr = string

func toChunkPosStr(
	cx, cz int32,
) string {
	return fmt.Sprintf("%d/%d", cx, cz)
}

func findRect(
	cx, cz int32,
	d int,
) (
	int32, int32, int32, int32,
) {
	maxCx, maxCz, minCx, minCz :=
		cx+int32(d), cz+int32(d), cx-int32(d), cz-int32(d)
	return maxCx, maxCz, minCx, minCz
}

func subRects(
	maxCx0, maxCz0, minCx0, minCz0 int32,
	maxCx1, maxCz1, minCx1, minCz1 int32,
) (
	int32, int32, int32, int32,
) {
	l0 := []int{int(maxCx0), int(minCx0), int(maxCx1), int(minCx1)}
	l1 := []int{int(maxCz0), int(minCz0), int(maxCz1), int(minCz1)}
	sort.Ints(l0)
	sort.Ints(l1)
	maxSubCx, maxSubCz, minSubCx, minSubCz :=
		l0[2], l1[2], l0[1], l1[1]
	return int32(maxSubCx), int32(maxSubCz), int32(minSubCx), int32(minSubCz)
}

func toChunkPos(
	x, z float64,
) (
	int32,
	int32,
) {
	if x < 0 {
		x = x - 16
	}
	if z < 0 {
		z = z - 16
	}

	cx, cz := int32(x)/16, int32(z)/16

	return cx, cz
}

type Overworld struct {
	sync.RWMutex

	rndDist int

	spawnX     float64
	spawnY     float64
	spawnZ     float64
	spawnYaw   float32
	spawnPitch float32

	players map[EID]*Player
	clients map[EID]*Client

	chansForAddPlayerEvent     map[EID]ChanForAddPlayerEvent
	chansForUpdateLatencyEvent map[EID]ChanForUpdateLatencyEvent
	chansForRemovePlayerEvent  map[EID]ChanForRemovePlayerEvent

	playersByChunkPos   map[ChunkPosStr]map[EID]types.Nil
	connsBetweenPlayers map[EID]map[EID]types.Nil

	chunks map[ChunkPosStr]*Chunk

	chansForSpawnPlayerEvent   map[EID]ChanForSpawnPlayerEvent   // by player
	chansForDespawnEntityEvent map[EID]ChanForDespawnEntityEvent // by player

	chansForLoadChunkEvent   map[EID]ChanForLoadChunkEvent   // by player
	chansForUnloadChunkEvent map[EID]ChanForUnloadChunkEvent // by player

	chansForSetEntityLookEvent        map[EID]ChanForSetEntityLookEvent
	chansForSetEntityRelativePosEvent map[EID]ChanForSetEntityRelativePosEvent
}

func NewOverworld(
	rndDist int,
	spawnX, spawnY, spawnZ float64,
	spawnYaw, spawnPitch float32,
) *Overworld {
	return &Overworld{
		rndDist: rndDist,

		spawnX:     spawnX,
		spawnY:     spawnY,
		spawnZ:     spawnZ,
		spawnYaw:   spawnYaw,
		spawnPitch: spawnPitch,

		players: make(map[EID]*Player),
		clients: make(map[EID]*Client),

		chansForAddPlayerEvent:     make(map[EID]ChanForAddPlayerEvent),
		chansForUpdateLatencyEvent: make(map[EID]ChanForUpdateLatencyEvent),
		chansForRemovePlayerEvent:  make(map[EID]ChanForRemovePlayerEvent),

		playersByChunkPos:   make(map[ChunkPosStr]map[EID]types.Nil),
		connsBetweenPlayers: make(map[EID]map[EID]types.Nil),

		chunks: make(map[ChunkPosStr]*Chunk),

		chansForSpawnPlayerEvent:   make(map[EID]ChanForSpawnPlayerEvent),
		chansForDespawnEntityEvent: make(map[EID]ChanForDespawnEntityEvent),

		chansForLoadChunkEvent:   make(map[EID]ChanForLoadChunkEvent),
		chansForUnloadChunkEvent: make(map[EID]ChanForUnloadChunkEvent),

		chansForSetEntityLookEvent:        make(map[EID]ChanForSetEntityLookEvent),
		chansForSetEntityRelativePosEvent: make(map[EID]ChanForSetEntityRelativePosEvent),
	}
}

func (w *Overworld) InitPlayerList(
	eid EID, uid UID,
	username string,
	chanForAddPlayerEvent ChanForAddPlayerEvent,
	chanForUpdateLatencyEvent ChanForUpdateLatencyEvent,
	chanForRemovePlayerEvent ChanForRemovePlayerEvent,
) {
	w.Lock()
	defer w.Unlock()

	w.chansForAddPlayerEvent[eid] = chanForAddPlayerEvent
	w.chansForUpdateLatencyEvent[eid] = chanForUpdateLatencyEvent
	w.chansForRemovePlayerEvent[eid] = chanForRemovePlayerEvent

	for _, player := range w.players {
		uid := player.GetUid()
		username := player.GetUsername()
		addPlayerEvent := NewAddPlayerEvent(
			uid, username,
		)
		chanForAddPlayerEvent <- addPlayerEvent
		addPlayerEvent.Wait()
	}

	spawnX, spawnY, spawnZ :=
		w.spawnX, w.spawnY, w.spawnZ
	spawnYaw, spawnPitch :=
		w.spawnYaw, w.spawnPitch
	player := NewPlayer(
		eid,
		uid, username,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	)
	w.players[eid] = player

	for eid1, _ := range w.players {
		addPlayerEvent := NewAddPlayerEvent(
			uid, username,
		)
		w.chansForAddPlayerEvent[eid1] <- addPlayerEvent
		addPlayerEvent.Wait()
	}
}

func (w *Overworld) ClosePlayerList(
	eid EID,
) {
	w.Lock()
	defer w.Unlock()

	player := w.players[eid]

	delete(w.players, eid)

	delete(w.chansForAddPlayerEvent, eid)
	delete(w.chansForUpdateLatencyEvent, eid)
	delete(w.chansForRemovePlayerEvent, eid)

	uid := player.GetUid()
	removePlayerEvent := NewRemovePlayerEvent(uid)
	for eid1, _ := range w.players {
		w.chansForRemovePlayerEvent[eid1] <- removePlayerEvent
	}

}

func (w *Overworld) InitPlayer(
	eid EID, uid UID,
	username string,
	cnt *Client,
	chanForAddPlayerEvent ChanForAddPlayerEvent,
	chanForUpdateLatencyEvent ChanForUpdateLatencyEvent,
	chanForRemovePlayerEvent ChanForRemovePlayerEvent,
	chanForSpawnPlayerEvent ChanForSpawnPlayerEvent,
	chanForDespawnEntityEvent ChanForDespawnEntityEvent,
	chanForLoadChunkEvent ChanForLoadChunkEvent,
	chanForUnloadChunkEvent ChanForUnloadChunkEvent,
	chanForSetEntityLookEvent ChanForSetEntityLookEvent,
	chanForSetEntityRelativePosEvent ChanForSetEntityRelativePosEvent,
) {
	w.Lock()
	defer w.Unlock()

	w.chansForAddPlayerEvent[eid] = chanForAddPlayerEvent
	w.chansForUpdateLatencyEvent[eid] = chanForUpdateLatencyEvent
	w.chansForRemovePlayerEvent[eid] = chanForRemovePlayerEvent

	w.clients[eid] = cnt

	spawnX, spawnY, spawnZ :=
		w.spawnX, w.spawnY, w.spawnZ
	spawnYaw, spawnPitch :=
		w.spawnYaw, w.spawnPitch
	player := NewPlayer(
		eid,
		uid, username,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	)
	w.players[eid] = player

	for eid1, player1 := range w.players {
		uid1, username1 :=
			player1.GetUid(), player1.GetUsername()
		addPlayerEvent1 := NewAddPlayerEvent(
			uid1, username1,
		)
		chanForAddPlayerEvent <- addPlayerEvent1
		addPlayerEvent1.Wait()

		if eid == eid1 {
			continue
		}
		addPlayerEvent := NewAddPlayerEvent(
			uid, username,
		)
		w.chansForAddPlayerEvent[eid1] <- addPlayerEvent
		addPlayerEvent.Wait()
	}

	w.chansForSpawnPlayerEvent[eid] = chanForSpawnPlayerEvent
	w.chansForDespawnEntityEvent[eid] = chanForDespawnEntityEvent
	w.chansForLoadChunkEvent[eid] = chanForLoadChunkEvent
	w.chansForUnloadChunkEvent[eid] = chanForUnloadChunkEvent
	w.chansForSetEntityLookEvent[eid] = chanForSetEntityLookEvent
	w.chansForSetEntityRelativePosEvent[eid] = chanForSetEntityRelativePosEvent

	w.connsBetweenPlayers[eid] = make(map[EID]types.Nil)

	//spawnPlayerEvent := NewSpawnPlayerEvent(
	//	eid,
	//	uid,
	//	spawnX, spawnY, spawnZ,
	//	spawnYaw, spawnPitch,
	//)

	dist := w.rndDist
	cx, cz := player.GetCx(), player.GetCz()
	maxCx, maxCz, minCx, minCz := findRect(
		cx, cz, dist,
	)
	overworld, init := true, true
	for cz := maxCz; cz >= minCz; cz-- {
		for cx := maxCx; cx >= minCx; cx-- {
			chunkPosStr := toChunkPosStr(cx, cz)

			chunk, has := w.chunks[chunkPosStr]
			if has == false {
				chunk = NewChunk()
			}
			bitmask, data := chunk.GenerateData(init, overworld)
			packet := NewSendChunkDataPacket(
				cx, cz,
				init,
				bitmask,
				data,
			)
			if err := cnt.WriteWithComp(packet); err != nil {
				panic(err)
			}
			//chanForLoadChunkEvent <- NewLoadChunkEvent(
			//	overworld, init,
			//	cx, cz,
			//	chunk,
			//)

			a, has := w.playersByChunkPos[chunkPosStr]
			if has == false {
				continue
			}
			for eid1, _ := range a {
				//w.chansForSpawnPlayerEvent[eid1] <- spawnPlayerEvent
				cnt1 := w.clients[eid1]
				packet := NewSpawnPlayerPacket(
					eid, uid,
					spawnX, spawnY, spawnZ,
					spawnYaw, spawnPitch,
				)
				if err := cnt1.WriteWithComp(packet); err != nil {
					panic(err)
				}

				player1 := w.players[eid1]
				eid1, uid1 :=
					player1.GetEid(), player1.GetUid()
				x1, y1, z1 :=
					player1.GetX(), player1.GetY(), player1.GetZ()
				yaw1, pitch1 :=
					player1.GetYaw(), player1.GetPitch()
				//chanForSpawnPlayerEvent <- NewSpawnPlayerEvent(
				//	eid1, uid1,
				//	x1, y1, z1,
				//	yaw1, pitch1,
				//)
				packet1 := NewSpawnPlayerPacket(
					eid1, uid1,
					x1, y1, z1,
					yaw1, pitch1,
				)
				if err := cnt.WriteWithComp(packet1); err != nil {
					panic(err)
				}

				w.connsBetweenPlayers[eid][eid1] = types.Nil{}
				w.connsBetweenPlayers[eid1][eid] = types.Nil{}
			}
		}
	}

	chunkPosStr := toChunkPosStr(cx, cz)
	a, has := w.playersByChunkPos[chunkPosStr]
	if has == false {
		b := make(map[EID]types.Nil)
		a = b
		w.playersByChunkPos[chunkPosStr] = b
	}
	a[eid] = types.Nil{}

}

func (w *Overworld) UpdateLatency(
	eid EID,
	latency int32,
) {
	w.RLock()
	defer w.RUnlock()

	//player := w.players[eid]
	//uid := player.GetUid()
	//updateLatencyEvent := NewUpdateLatencyEvent(
	//	uid, latency,
	//)
	//for eid1, _ := range w.players {
	//	w.chansForUpdateLatencyEvent[eid1] <- updateLatencyEvent
	//}
}

func (w *Overworld) UpdatePlayerLook(
	eid EID,
	yaw, pitch float32,
) {
	w.RLock()
	defer w.RUnlock()

	player := w.players[eid]
	player.UpdateLook(yaw, pitch)

	setEntityLookEvent := NewSetEntityLookEvent(
		eid,
		yaw, pitch,
		true,
	)
	for eid1, _ := range w.connsBetweenPlayers[eid] {
		w.chansForSetEntityLookEvent[eid1] <- setEntityLookEvent
	}
}

func (w *Overworld) UpdatePlayerPos(
	eid EID,
	x, y, z float64,
) {
	w.RLock()
	defer w.RUnlock()

	player := w.players[eid]
	player.UpdatePos(x, y, z)

	deltaX, deltaY, deltaZ :=
		player.GetDeltaX(), player.GetDeltaY(), player.GetDeltaZ()
	//setEntityRelativePosEvent := NewSetEntityRelativePosEvent(
	//	eid,
	//	deltaX, deltaY, deltaZ,
	//	true,
	//)
	packet := NewSetEntityRelativePosPacket(
		eid,
		deltaX, deltaY, deltaZ,
		true,
	)
	for eid1, _ := range w.connsBetweenPlayers[eid] {
		//w.chansForSetEntityRelativePosEvent[eid1] <- setEntityRelativePosEvent
		cnt1 := w.clients[eid1]
		if err := cnt1.WriteWithComp(packet); err != nil {
			panic(err)
		}
	}
}

func (w *Overworld) UpdatePlayerChunk(
	eid EID,
) {
	w.Lock()
	defer w.Unlock()

	player := w.players[eid] // TODO: panic({0x969140, 0xbf41d0})id memory address or nil pointer dereference [recovered]
	isChunkPosChanged := player.IsChunkPosChanged()
	if isChunkPosChanged == false {
		return
	}

	prevCx, prevCz :=
		player.GetPrevCx(), player.GetPrevCz()
	chunkPrevPosStr := toChunkPosStr(prevCx, prevCz)
	delete(w.playersByChunkPos[chunkPrevPosStr], eid)

	cx, cz :=
		player.GetCx(), player.GetCz()
	chunkPosStr := toChunkPosStr(cx, cz)
	a, has := w.playersByChunkPos[chunkPosStr]
	if has == false {
		b := make(map[EID]types.Nil)
		a = b
		w.playersByChunkPos[chunkPosStr] = b
	}
	a[eid] = types.Nil{}

	dist := w.rndDist
	maxCx, maxCz, minCx, minCz :=
		findRect(cx, cz, dist)
	maxPrevCx, maxPrevCz, minPrevCx, minPrevCz :=
		findRect(prevCx, prevCz, dist)
	maxSubCx, maxSubCz, minSubCx, minSubCz := subRects(
		maxCx, maxCz, minCx, minCz,
		maxPrevCx, maxPrevCz, minPrevCx, minPrevCz,
	)
	uid := player.GetUid()
	x, y, z :=
		player.GetX(), player.GetY(), player.GetZ()
	yaw, pitch := player.GetYaw(), player.GetPitch()
	spawnPlayerEvent := NewSpawnPlayerEvent(
		eid, uid,
		x, y, z,
		yaw, pitch,
	)
	chanForSpawnPlayerEvent := w.chansForSpawnPlayerEvent[eid]
	chanForLoadChunkEvent := w.chansForLoadChunkEvent[eid]
	overworld, init := true, true
	for cz := maxCz; cz >= minCz; cz-- {
		for cx := maxCx; cx >= minCx; cx-- {
			if minSubCx <= cx && cx <= maxSubCx &&
				minSubCz <= cz && cz <= maxSubCz {
				continue
			}

			chunkPosStr := toChunkPosStr(cx, cz)
			chunk, has := w.chunks[chunkPosStr]
			if has == false {
				chunk = NewChunk()
			}
			chanForLoadChunkEvent <- NewLoadChunkEvent(
				overworld, init,
				cx, cz,
				chunk,
			)

			a, has := w.playersByChunkPos[chunkPosStr]
			if has == false {
				continue
			}
			for eid1, _ := range a {
				w.connsBetweenPlayers[eid][eid1] = types.Nil{}
				w.connsBetweenPlayers[eid1][eid] = types.Nil{}

				w.chansForSpawnPlayerEvent[eid1] <- spawnPlayerEvent
				player1 := w.players[eid1]
				eid1, uid1 :=
					player1.GetEid(), player1.GetUid()
				x1, y1, z1 :=
					player1.GetX(), player1.GetY(), player1.GetZ()
				yaw1, pitch1 :=
					player1.GetYaw(), player1.GetPitch()
				chanForSpawnPlayerEvent <- NewSpawnPlayerEvent(
					eid1, uid1,
					x1, y1, z1,
					yaw1, pitch1,
				)
			}
		}
	}

	chanForUnloadChunkEvent := w.chansForUnloadChunkEvent[eid]
	chanForDespawnEntityEvent := w.chansForDespawnEntityEvent[eid]
	despawnPlayerEvent := NewDespawnEntityEvent(
		eid,
	)
	for cz := maxPrevCz; cz >= minPrevCz; cz-- {
		for cx := maxPrevCx; cx >= minPrevCx; cx-- {
			if minSubCx <= cx && cx <= maxSubCx &&
				minSubCz <= cz && cz <= maxSubCz {
				continue
			}

			chunkPosStr := toChunkPosStr(cx, cz)

			chanForUnloadChunkEvent <- NewUnloadChunkEvent(
				cx, cz,
			)

			a, has := w.playersByChunkPos[chunkPosStr]
			if has == false {
				continue
			}
			for eid1, _ := range a {
				delete(w.connsBetweenPlayers[eid], eid1)
				delete(w.connsBetweenPlayers[eid1], eid)

				w.chansForDespawnEntityEvent[eid1] <- despawnPlayerEvent
				chanForDespawnEntityEvent <- NewDespawnEntityEvent(
					eid1,
				)
			}
		}
	}
}

func (w *Overworld) ClosePlayer(
	eid EID,
) {
	w.Lock()
	defer w.Unlock()

	player := w.players[eid]
	cx, cz := player.GetCx(), player.GetCz()
	chunkPosStr := toChunkPosStr(cx, cz)
	delete(w.playersByChunkPos[chunkPosStr], eid)

	event := NewDespawnEntityEvent(eid)
	for eid1, _ := range w.connsBetweenPlayers[eid] {
		delete(w.connsBetweenPlayers[eid1], eid)

		w.chansForDespawnEntityEvent[eid1] <- event
	}
	delete(w.connsBetweenPlayers, eid)

	delete(w.chansForSpawnPlayerEvent, eid)
	delete(w.chansForDespawnEntityEvent, eid)
	delete(w.chansForLoadChunkEvent, eid)
	delete(w.chansForUnloadChunkEvent, eid)
	delete(w.chansForSetEntityLookEvent, eid)
	delete(w.chansForSetEntityRelativePosEvent, eid)

	delete(w.players, eid)

	delete(w.chansForAddPlayerEvent, eid)
	delete(w.chansForUpdateLatencyEvent, eid)
	delete(w.chansForRemovePlayerEvent, eid)

	uid := player.GetUid()
	removePlayerEvent := NewRemovePlayerEvent(uid)
	for eid1, _ := range w.players {
		w.chansForRemovePlayerEvent[eid1] <- removePlayerEvent
	}
}

//func (w *Overworld) UpdateMob() {
//w.Lock()
//	defer w.Unlock()
//}

func (w *Overworld) MakeFlat() {
	for cz := int32(10); cz >= -10; cz-- {
		for cx := int32(10); cx >= -10; cx-- {
			chunk := NewChunk()
			part := NewChunkPart()
			for z := 0; z < ChunkPartWidth; z++ {
				for x := 0; x < ChunkPartWidth; x++ {
					part.SetBlock(uint8(x), 0, uint8(z), StoneBlock)
				}
			}

			chunk.SetChunkPart(4, part)
			chunkPosStr := toChunkPosStr(cx, cz)
			w.chunks[chunkPosStr] = chunk
		}
	}
}

func (w *Overworld) GetSpawnX() float64 {
	return w.spawnX
}

func (w *Overworld) GetSpawnY() float64 {
	return w.spawnY
}

func (w *Overworld) GetSpawnZ() float64 {
	return w.spawnZ
}

func (w *Overworld) GetSpawnYaw() float32 {
	return w.spawnYaw
}

func (w *Overworld) GetSpawnPitch() float32 {
	return w.spawnPitch
}
