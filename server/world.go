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
) ChunkPosStr {
	return fmt.Sprintf("%d/%d", cx, cz)
}

func findRect(
	cx, cz int32,
	d int32,
) (
	maxCx int32, maxCz int32,
	minCx int32, minCz int32,
) {
	maxCx, maxCz, minCx, minCz =
		cx+d, cz+d, cx-d, cz-d
	return
}

func subRects(
	maxCx0, maxCz0, minCx0, minCz0 int32,
	maxCx1, maxCz1, minCx1, minCz1 int32,
) (
	maxSubCx int32, maxSubCz int32,
	minSubCx int32, minSubCz int32,
) {
	l0 := []int{int(maxCx0), int(minCx0), int(maxCx1), int(minCx1)}
	l1 := []int{int(maxCz0), int(minCz0), int(maxCz1), int(minCz1)}
	sort.Ints(l0)
	sort.Ints(l1)
	maxSubCx, maxSubCz, minSubCx, minSubCz =
		int32(l0[2]), int32(l1[2]), int32(l0[1]), int32(l1[1])
	return
}

func toChunkPos(
	x, z float64,
) (
	cx int32, cz int32,
) {
	if x < 0 {
		x = x - 16
	}
	if z < 0 {
		z = z - 16
	}

	cx, cz = int32(x)/16, int32(z)/16
	return
}

type Overworld struct {
	sync.RWMutex

	rndDist int32

	chunks map[ChunkPosStr]*Chunk

	players                    map[EID]*Player
	chansForSpawnPlayerEvent   map[EID]ChanForSpawnPlayerEvent
	chansForDespawnEntityEvent map[EID]ChanForDespawnEntityEvent

	chansForSetEntityLookEvent        map[EID]ChanForSetEntityLookEvent
	chansForSetEntityRelativePosEvent map[EID]ChanForSetEntityRelativePosEvent
	chansForSetEntityMetadataEvent    map[EID]ChanForSetEntityMetadataEvent

	connsBetweenPlayers map[EID]map[EID]types.Nil
	playersByChunkPos   map[ChunkPosStr]map[EID]types.Nil
}

func NewOverworld(
	rndDist int32,
) *Overworld {
	return &Overworld{
		rndDist: rndDist,

		chunks: make(map[ChunkPosStr]*Chunk),

		players:                    make(map[EID]*Player),
		chansForSpawnPlayerEvent:   make(map[EID]ChanForSpawnPlayerEvent),
		chansForDespawnEntityEvent: make(map[EID]ChanForDespawnEntityEvent),

		chansForSetEntityLookEvent:        make(map[EID]ChanForSetEntityLookEvent),
		chansForSetEntityRelativePosEvent: make(map[EID]ChanForSetEntityRelativePosEvent),
		chansForSetEntityMetadataEvent:    make(map[EID]ChanForSetEntityMetadataEvent),

		connsBetweenPlayers: make(map[EID]map[EID]types.Nil),
		playersByChunkPos:   make(map[ChunkPosStr]map[EID]types.Nil),
	}
}

func (w *Overworld) InitPlayer(
	lg *Logger,
	player *Player,
	cnt *Client,
	chanForSpawnPlayerEvent ChanForSpawnPlayerEvent,
	chanForDespawnEntityEvent ChanForDespawnEntityEvent,
	chanForSetEntityRelativePosEvent ChanForSetEntityRelativePosEvent,
	chanForSetEntityLookEvent ChanForSetEntityLookEvent,
	chanForSetEntityMetadataEvent ChanForSetEntityMetadataEvent,
) error {
	w.Lock()
	defer w.Unlock()

	lg.Debug(
		"it is started to init player in Overworld",
		NewLgElement("player", player),
		NewLgElement("cnt", cnt),
	)
	defer func() {
		lg.Debug("it is finished to init player in Overworld")
	}()

	eid := player.GetEid()
	w.players[eid] = player
	w.chansForSpawnPlayerEvent[eid] = chanForSpawnPlayerEvent
	w.chansForDespawnEntityEvent[eid] = chanForDespawnEntityEvent
	w.chansForSetEntityRelativePosEvent[eid] = chanForSetEntityRelativePosEvent
	w.chansForSetEntityLookEvent[eid] = chanForSetEntityLookEvent
	w.chansForSetEntityMetadataEvent[eid] = chanForSetEntityMetadataEvent

	uid := player.GetUid()
	x, y, z :=
		player.GetX(), player.GetY(), player.GetZ()
	yaw, pitch :=
		player.GetYaw(), player.GetPitch()
	sneaking, sprinting :=
		player.IsSneaking(), player.IsSprinting()
	spawnPlayerEvent := NewSpawnPlayerEvent(
		eid, uid,
		x, y, z,
		yaw, pitch,
		sneaking, sprinting,
	)

	w.connsBetweenPlayers[eid] = make(map[EID]types.Nil)

	dist := w.rndDist
	cx, cz := toChunkPos(x, z)
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
			if err := cnt.LoadChunk(
				lg,
				overworld, init,
				cx, cz,
				chunk,
			); err != nil {
				return err
			}

			a, has := w.playersByChunkPos[chunkPosStr]
			if has == false {
				continue
			}
			for eid1, _ := range a {
				player1 := w.players[eid1]
				eid1, uid1 :=
					player1.GetEid(), player1.GetUid()
				x1, y1, z1 :=
					player1.GetX(), player1.GetY(), player1.GetZ()
				yaw1, pitch1 :=
					player1.GetYaw(), player1.GetPitch()
				sneaking1, sprinting1 :=
					player1.IsSneaking(), player1.IsSprinting()
				if err := cnt.SpawnPlayer(
					lg,
					eid1, uid1,
					x1, y1, z1,
					yaw1, pitch1,
					sneaking1, sprinting1,
				); err != nil {
					return err
				}

				chanForSpawnPlayerEvent1 :=
					w.chansForSpawnPlayerEvent[eid1]
				chanForSpawnPlayerEvent1 <- spawnPlayerEvent

				w.connsBetweenPlayers[eid][eid1] = types.Nil{}
				w.connsBetweenPlayers[eid1][eid] = types.Nil{}
			}
		}
	}

	chunkPosStr := toChunkPosStr(cx, cz)
	a, has := w.playersByChunkPos[chunkPosStr]
	if has == false {
		b := make(map[EID]types.Nil)
		w.playersByChunkPos[chunkPosStr] = b
		a = b
	}
	a[eid] = types.Nil{}

	return nil
}

func (w *Overworld) UpdatePlayerPos(
	lg *Logger,
	player *Player,
	x, y, z float64,
	ground bool,
) error {
	w.RLock()
	defer w.RUnlock()

	lg.Debug(
		"it is started to update player pos in Overworld",
		NewLgElement("player", player),
		NewLgElement("x", x),
		NewLgElement("y", y),
		NewLgElement("z", z),
		NewLgElement("ground", ground),
	)
	defer func() {
		lg.Debug("it is finished to update player pos in Overworld")
	}()

	// TODO: ground
	eid := player.GetEid()
	player.UpdatePos(
		x, y, z,
	)
	deltaX, deltaY, deltaZ :=
		player.GetDeltaX(),
		player.GetDeltaY(),
		player.GetDeltaZ()

	setEntityRelativePosEvent :=
		NewSetEntityRelativePosEvent(
			eid,
			deltaX, deltaY, deltaZ,
			ground,
		)
	if deltaX == 0 && deltaY == 0 && deltaZ == 0 {
		return nil
	}
	a := w.connsBetweenPlayers[eid]
	for eid1, _ := range a {
		chanForSetEntityRelativePosEvent1 :=
			w.chansForSetEntityRelativePosEvent[eid1]
		chanForSetEntityRelativePosEvent1 <- setEntityRelativePosEvent
	}

	return nil
}

func (w *Overworld) UpdatePlayerLook(
	lg *Logger,
	player *Player,
	yaw, pitch float32,
	ground bool,
) error {
	w.RLock()
	defer w.RUnlock()

	lg.Debug(
		"it is started to update player look in Overworld",
		NewLgElement("player", player),
		NewLgElement("yaw", yaw),
		NewLgElement("pitch", pitch),
		NewLgElement("ground", ground),
	)
	defer func() {
		lg.Debug("it is finished to update player look in Overworld")
	}()

	eid := player.GetEid()
	player.UpdateLook(
		yaw, pitch,
	)

	setEntityLookEvent :=
		NewSetEntityLookEvent(
			eid,
			yaw, pitch,
			ground,
		)
	a := w.connsBetweenPlayers[eid]
	for eid1, _ := range a {
		chanForSetEntityLookEvent1 :=
			w.chansForSetEntityLookEvent[eid1]
		chanForSetEntityLookEvent1 <- setEntityLookEvent
	}

	return nil
}

func (w *Overworld) UpdatePlayerSneaking(
	lg *Logger,
	player *Player,
	sneaking bool,
) error {
	w.RLock()
	defer w.RUnlock()

	lg.Debug(
		"it is started to update player sneaking in Overworld",
		NewLgElement("player", player),
		NewLgElement("sneaking", sneaking),
	)
	defer func() {
		lg.Debug("it is finished to update player sneaking in Overworld")
	}()

	eid := player.GetEid()

	if sneaking == true {
		player.StartSneaking()
	} else {
		player.StopSneaking()
	}

	sprinting := player.IsSprinting()
	metadata := NewEntityMetadata()
	if err := metadata.SetActions(
		sneaking, sprinting,
	); err != nil {
		return err
	}

	setEntityMetadataEvent :=
		NewSetEntityMetadataEvent(
			eid,
			metadata,
		)
	a := w.connsBetweenPlayers[eid]
	for eid1, _ := range a {
		chanForSetEntityMetadataEvent1 :=
			w.chansForSetEntityMetadataEvent[eid1]
		chanForSetEntityMetadataEvent1 <- setEntityMetadataEvent
	}

	return nil
}

func (w *Overworld) UpdatePlayerSprinting(
	lg *Logger,
	player *Player,
	sprinting bool,
) error {
	w.RLock()
	defer w.RUnlock()

	lg.Debug(
		"it is started to update player sprinting in Overworld",
		NewLgElement("player", player),
		NewLgElement("sprinting", sprinting),
	)
	defer func() {
		lg.Debug("it is finished to update player sprinting in Overworld")
	}()

	eid := player.GetEid()

	if sprinting == true {
		player.StartSprinting()
	} else {
		player.StopSprinting()
	}

	sneaking := player.IsSneaking()
	metadata := NewEntityMetadata()
	if err := metadata.SetActions(
		sneaking, sprinting,
	); err != nil {
		return err
	}

	setEntityMetadataEvent :=
		NewSetEntityMetadataEvent(
			eid,
			metadata,
		)
	a := w.connsBetweenPlayers[eid]
	for eid1, _ := range a {
		chanForSetEntityMetadataEvent1 :=
			w.chansForSetEntityMetadataEvent[eid1]
		chanForSetEntityMetadataEvent1 <- setEntityMetadataEvent
	}

	return nil
}

func (w *Overworld) UpdatePlayerChunk(
	lg *Logger,
	player *Player,
	cnt *Client,
) error {
	w.Lock()
	defer w.Unlock()

	lg.Debug(
		"it is started to update player chunk in Overworld",
		NewLgElement("player", player),
	)
	defer func() {
		lg.Debug("it is finished to update player chunk in Overworld")
	}()

	x, z := player.GetX(), player.GetZ()
	cx, cz := toChunkPos(x, z)
	prevX, prevZ := player.GetPrevX(), player.GetPrevZ()
	prevCx, prevCz := toChunkPos(prevX, prevZ)
	if cx == prevCx && cz == prevCz {
		return nil
	}

	dist := w.rndDist
	maxCx, maxCz, minCx, minCz :=
		findRect(cx, cz, dist)
	maxPrevCx, maxPrevCz, minPrevCx, minPrevCz :=
		findRect(prevCx, prevCz, dist)
	maxSubCx, maxSubCz, minSubCx, minSubCz := subRects(
		maxCx, maxCz, minCx, minCz,
		maxPrevCx, maxPrevCz, minPrevCx, minPrevCz,
	)
	eid, uid := player.GetEid(), player.GetUid()
	y := player.GetY()
	yaw, pitch :=
		player.GetYaw(), player.GetPitch()
	sneaking, sprinting :=
		player.IsSneaking(), player.IsSprinting()
	spawnPlayerEvent := NewSpawnPlayerEvent(
		eid, uid,
		x, y, z,
		yaw, pitch,
		sneaking, sprinting,
	)
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
			if err := cnt.LoadChunk(
				lg,
				overworld, init,
				cx, cz,
				chunk,
			); err != nil {
				return err
			}

			a, has := w.playersByChunkPos[chunkPosStr]
			if has == false {
				continue
			}
			for eid1, _ := range a {
				player1 := w.players[eid1]
				eid1, uid1 :=
					player1.GetEid(), player1.GetUid()
				x1, y1, z1 :=
					player1.GetX(), player1.GetY(), player1.GetZ()
				yaw1, pitch1 :=
					player1.GetYaw(), player1.GetPitch()
				sneaking1, sprinting1 :=
					player1.IsSneaking(), player1.IsSprinting()
				if err := cnt.SpawnPlayer(
					lg,
					eid1, uid1,
					x1, y1, z1,
					yaw1, pitch1,
					sneaking1, sprinting1,
				); err != nil {
					return err
				}

				chanForSpawnPlayerEvent1 :=
					w.chansForSpawnPlayerEvent[eid1]
				chanForSpawnPlayerEvent1 <- spawnPlayerEvent

				w.connsBetweenPlayers[eid][eid1] = types.Nil{}
				w.connsBetweenPlayers[eid1][eid] = types.Nil{}
			}
		}
	}

	despawnEntityEvent := NewDespawnEntityEvent(
		eid,
	)
	for cz := maxPrevCz; cz >= minPrevCz; cz-- {
		for cx := maxPrevCx; cx >= minPrevCx; cx-- {
			if minSubCx <= cx && cx <= maxSubCx &&
				minSubCz <= cz && cz <= maxSubCz {
				continue
			}

			chunkPosStr := toChunkPosStr(cx, cz)

			if err := cnt.UnloadChunk(
				lg,
				cx, cz,
			); err != nil {
				return err
			}
			a, has := w.playersByChunkPos[chunkPosStr]
			if has == false {
				continue
			}
			for eid1, _ := range a {
				if err := cnt.DespawnEntity(
					lg,
					eid1,
				); err != nil {
					return err
				}

				chanForDespawnEntityEvent1 :=
					w.chansForDespawnEntityEvent[eid1]
				chanForDespawnEntityEvent1 <- despawnEntityEvent

				delete(w.connsBetweenPlayers[eid], eid1)
				delete(w.connsBetweenPlayers[eid1], eid)
			}

		}
	}

	chunkPrevPosStr := toChunkPosStr(prevCx, prevCz)
	delete(w.playersByChunkPos[chunkPrevPosStr], eid)

	chunkPosStr := toChunkPosStr(cx, cz)
	a, has := w.playersByChunkPos[chunkPosStr]
	if has == false {
		b := make(map[EID]types.Nil)
		w.playersByChunkPos[chunkPosStr] = b
		a = b
	}
	a[eid] = types.Nil{}

	return nil
}

func (w *Overworld) ClosePlayer(
	lg *Logger,
	player *Player,
) (
	ChanForSpawnPlayerEvent,
	ChanForDespawnEntityEvent,
	ChanForSetEntityRelativePosEvent,
	ChanForSetEntityLookEvent,
	ChanForSetEntityMetadataEvent,
) {
	w.Lock()
	defer w.Unlock()

	lg.Debug(
		"it is started to close player in Overworld",
		NewLgElement("player", player),
	)
	defer func() {
		lg.Debug("it is finished to close player in Overworld")
	}()

	eid := player.GetEid()

	x, z := player.GetX(), player.GetZ()
	cx, cz := toChunkPos(x, z)
	chunkPosStr := toChunkPosStr(cx, cz)
	delete(w.playersByChunkPos[chunkPosStr], eid)

	despawnEntityEvent := NewDespawnEntityEvent(
		eid,
	)
	a := w.connsBetweenPlayers[eid]
	for eid1, _ := range a {
		chanForDespawnEntityEvent1 :=
			w.chansForDespawnEntityEvent[eid1]
		chanForDespawnEntityEvent1 <- despawnEntityEvent

		delete(w.connsBetweenPlayers[eid1], eid)
	}
	delete(w.connsBetweenPlayers, eid)

	chanForSpawnPlayerEvent := w.chansForSpawnPlayerEvent[eid]
	chanForDespawnEntityEvent := w.chansForDespawnEntityEvent[eid]
	chanForSetEntityRelativePosEvent := w.chansForSetEntityRelativePosEvent[eid]
	chanForSetEntityLookEvent := w.chansForSetEntityLookEvent[eid]
	chanForSetEntityMetadataEvent := w.chansForSetEntityMetadataEvent[eid]
	delete(w.chansForSpawnPlayerEvent, eid)
	delete(w.chansForDespawnEntityEvent, eid)
	delete(w.chansForSetEntityRelativePosEvent, eid)
	delete(w.chansForSetEntityLookEvent, eid)
	delete(w.chansForSetEntityMetadataEvent, eid)

	delete(w.players, eid)

	return chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForSetEntityRelativePosEvent,
		chanForSetEntityLookEvent,
		chanForSetEntityMetadataEvent
}

func (w *Overworld) MakeFlat() {
	w.Lock()
	defer w.Unlock()

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
