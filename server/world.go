package server

import (
	"go/types"
	"sort"
	"sync"
)

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

	playersByChunkPos   map[int32]map[int32]map[EID]types.Nil
	connsBetweenPlayers map[EID]map[EID]types.Nil

	chunks map[int32]map[int32]*Chunk

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

		players: make(map[EID]*Player),

		spawnX:     spawnX,
		spawnY:     spawnY,
		spawnZ:     spawnZ,
		spawnYaw:   spawnYaw,
		spawnPitch: spawnPitch,

		playersByChunkPos:   make(map[int32]map[int32]map[EID]types.Nil),
		connsBetweenPlayers: make(map[EID]map[EID]types.Nil),

		chunks: make(map[int32]map[int32]*Chunk),

		chansForSpawnPlayerEvent:   make(map[EID]ChanForSpawnPlayerEvent),
		chansForDespawnEntityEvent: make(map[EID]ChanForDespawnEntityEvent),

		chansForLoadChunkEvent:   make(map[EID]ChanForLoadChunkEvent),
		chansForUnloadChunkEvent: make(map[EID]ChanForUnloadChunkEvent),

		chansForSetEntityLookEvent:        make(map[EID]ChanForSetEntityLookEvent),
		chansForSetEntityRelativePosEvent: make(map[EID]ChanForSetEntityRelativePosEvent),
	}
}

func (w *Overworld) getChunk(
	cx, cz int32,
) *Chunk {
	w.RLock()
	defer w.RUnlock()
	m := w.chunks[cx]
	if m == nil {
		return nil
	}
	return m[cz]
}

func (w *Overworld) setChunk(
	cx, cz int32,
	chunk *Chunk,
) {
	w.Lock()
	defer w.Unlock()
	m := w.chunks[cx]
	if m == nil {
		a := make(map[int32]*Chunk)
		m = a
		w.chunks[cx] = a
	}
	m[cz] = chunk
}

func (w *Overworld) InitPlayer(
	eid EID, uid UID,
	username string,
	chanForSpawnPlayerEvent ChanForSpawnPlayerEvent,
	chanForDespawnEntityEvent ChanForDespawnEntityEvent,
	chanForLoadChunkEvent ChanForLoadChunkEvent,
	chanForUnloadChunkEvent ChanForUnloadChunkEvent,
	chanForSetEntityLookEvent ChanForSetEntityLookEvent,
	chanForSetEntityRelativePosEvent ChanForSetEntityRelativePosEvent,
) error {
	w.Lock()
	defer w.Unlock()

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
	w.chansForSpawnPlayerEvent[eid] = chanForSpawnPlayerEvent
	w.chansForDespawnEntityEvent[eid] = chanForDespawnEntityEvent
	w.chansForLoadChunkEvent[eid] = chanForLoadChunkEvent
	w.chansForUnloadChunkEvent[eid] = chanForUnloadChunkEvent
	w.chansForSetEntityLookEvent[eid] = chanForSetEntityLookEvent
	w.chansForSetEntityRelativePosEvent[eid] = chanForSetEntityRelativePosEvent

	w.connsBetweenPlayers[eid] = make(map[EID]types.Nil)

	x, y, z := player.GetX(), player.GetY(), player.GetZ()
	yaw, pitch := player.GetYaw(), player.GetPitch()
	spawnPlayerEvent := NewSpawnPlayerEvent(
		eid,
		uid,
		x, y, z,
		yaw, pitch,
	)

	dist := w.rndDist
	cx, cz := toChunkPos(x, z)
	maxCx, maxCz, minCx, minCz := findRect(
		cx, cz, dist,
	)
	overworld, init := true, true
	for cz := maxCz; cz >= minCz; cz-- {
		for cx := maxCx; cx >= minCx; cx-- {

			var chunk *Chunk
			m0, has := w.chunks[cx]
			if has == false {
				chunk = NewChunk()
			} else {
				chunk1, has := m0[cz]
				if has == false {
					chunk = NewChunk()
				} else {
					chunk = chunk1
				}
			}
			chanForLoadChunkEvent <- NewLoadChunkEvent(
				overworld, init,
				cx, cz,
				chunk,
			)

			m1, has := w.playersByChunkPos[cx]
			if has == false {
				continue
			}
			m2, has := m1[cz]
			if has == false {
				continue
			}
			for eid1, _ := range m2 {
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

	m0, has := w.playersByChunkPos[cx]
	if has == false {
		a := make(map[int32]map[EID]types.Nil)
		m0 = a
		w.playersByChunkPos[cx] = a
	}
	m1, has := m0[cz]
	if has == false {
		a := make(map[EID]types.Nil)
		m1 = a
		m0[cz] = a
	}
	m1[eid] = types.Nil{}

	return nil
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
	setEntityRelativePosEvent := NewSetEntityRelativePosEvent(
		eid,
		deltaX, deltaY, deltaZ,
		true,
	)
	for eid1, _ := range w.connsBetweenPlayers[eid] {
		w.chansForSetEntityRelativePosEvent[eid1] <- setEntityRelativePosEvent
	}
}

func (w *Overworld) UpdatePlayerChunk(
	eid EID,
) {
	w.Lock()
	defer w.Unlock()

	player := w.players[eid]
	isChunkPosChanged := player.IsChunkPosChanged()
	if isChunkPosChanged == false {
		return
	}

	prevCx, prevCz :=
		player.GetPrevCx(), player.GetPrevCz()
	delete(w.playersByChunkPos[prevCx][prevCz], eid)

	cx, cz :=
		player.GetCx(), player.GetCz()
	m0, has := w.playersByChunkPos[cx]
	if has == false {
		a := make(map[int32]map[EID]types.Nil)
		m0 = a
		w.playersByChunkPos[cx] = a
	}
	m1, has := m0[cz]
	if has == false {
		a := make(map[EID]types.Nil)
		m1 = a
		m0[cz] = a
	}
	m1[eid] = types.Nil{}

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

			var chunk *Chunk
			m0, has := w.chunks[cx]
			if has == false {
				chunk = NewChunk()
			} else {
				chunk1, has := m0[cz]
				if has == false {
					chunk = NewChunk()
				} else {
					chunk = chunk1
				}
			}
			chanForLoadChunkEvent <- NewLoadChunkEvent(
				overworld, init,
				cx, cz,
				chunk,
			)

			m1, has := w.playersByChunkPos[cx]
			if has == false {
				continue
			}
			m2, has := m1[cz]
			if has == false {
				continue
			}
			for eid1, _ := range m2 {
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

			chanForUnloadChunkEvent <- NewUnloadChunkEvent(
				cx, cz,
			)

			m1, has := w.playersByChunkPos[cx]
			if has == false {
				continue
			}
			m2, has := m1[cz]
			if has == false {
				continue
			}
			for eid1, _ := range m2 {
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
	x, z := player.GetX(), player.GetZ()
	cx, cz := toChunkPos(x, z)
	delete(w.playersByChunkPos[cx][cz], eid)

	event := NewDespawnEntityEvent(eid)
	m := w.connsBetweenPlayers[eid]
	for eid1, _ := range m {
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
}

//func (w *Overworld) UpdateMob() {
//w.Lock()
//	defer w.Unlock()
//}

func (w *Overworld) MakeFlat() {
	for cz := 10; cz >= -10; cz-- {
		for cx := 10; cx >= -10; cx-- {
			chunk := NewChunk()
			part := NewChunkPart()
			for z := 0; z < ChunkPartWidth; z++ {
				for x := 0; x < ChunkPartWidth; x++ {
					part.SetBlock(uint8(x), 0, uint8(z), StoneBlock)
				}
			}

			chunk.SetChunkPart(4, part)
			w.setChunk(int32(cx), int32(cz), chunk)
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
