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

type PlayerListItem struct {
	uid      UID
	username string
}

func NewPlayerListItem(
	uid UID,
	username string,
) *PlayerListItem {
	return &PlayerListItem{
		uid,
		username,
	}
}

func (pli *PlayerListItem) GetUID() UID {
	return pli.uid
}

func (pli *PlayerListItem) GetUsername() string {
	return pli.username
}

type Overworld interface {
	InitPlayer(
		player Player,

		chanForAddPlayerEvent ChanForAddPlayerEvent,
		chanForUpdateLatencyEvent ChanForUpdateLatencyEvent,
		chanForRemovePlayerEvent ChanForRemovePlayerEvent,
		chanForSpawnPlayerEvent ChanForSpawnPlayerEvent,
		chanForDespawnEntityEvent ChanForDespawnEntityEvent,
		chanForSetEntityRelativePosEvent ChanForSetEntityRelativePosEvent,
		chanForSetEntityLookEvent ChanForSetEntityLookEvent,
		chanForSetEntityMetadataEvent ChanForSetEntityMetadataEvent,
		chanForLoadChunkEvent ChanForLoadChunkEvent,
		chanForUnloadChunkEvent ChanForUnloadChunkEvent,

		chanForUpdateChunkEvent ChanForUpdateChunkEvent,

		cnt *Client,
	) error
	UpdatePlayerLatency(
		uid UID,
		latency int32,
	) error
	UpdatePlayerPos(
		player Player,
		x, y, z float64,
		ground bool,
	) error
	UpdatePlayerChunk(
		player Player,
		prevCx, prevCz int32,
		currCx, currCz int32,
	) error
	UpdatePlayerLook(
		player Player,
		yaw, pitch float32,
		ground bool,
	) error
	UpdatePlayerSneaking(
		player Player,
		sneaking bool,
	) error
	UpdatePlayerSprinting(
		player Player,
		sprinting bool,
	) error
	ClosePlayer(
		player Player,
	) (
		ChanForAddPlayerEvent,
		ChanForUpdateLatencyEvent,
		ChanForRemovePlayerEvent,
		ChanForSpawnPlayerEvent,
		ChanForDespawnEntityEvent,
		ChanForSetEntityRelativePosEvent,
		ChanForSetEntityLookEvent,
		ChanForSetEntityMetadataEvent,
		ChanForLoadChunkEvent,
		ChanForUnloadChunkEvent,

		ChanForUpdateChunkEvent,
	)

	MakeFlat(
		block *Block,
	)
}

type overworld struct {
	sync.RWMutex

	rndDist int32

	spawnX, spawnY, spawnZ float64
	spawnYaw, spawnPitch   float32

	chunks map[ChunkPosStr]*Chunk

	players map[EID]Player

	chansForAddPlayerEvent            map[EID]ChanForAddPlayerEvent
	chansForUpdateLatencyEvent        map[EID]ChanForUpdateLatencyEvent
	chansForRemovePlayerEvent         map[EID]ChanForRemovePlayerEvent
	chansForSpawnPlayerEvent          map[EID]ChanForSpawnPlayerEvent
	chansForDespawnEntityEvent        map[EID]ChanForDespawnEntityEvent
	chansForSetEntityLookEvent        map[EID]ChanForSetEntityLookEvent
	chansForSetEntityRelativePosEvent map[EID]ChanForSetEntityRelativePosEvent
	chansForSetEntityMetadataEvent    map[EID]ChanForSetEntityMetadataEvent
	chansForLoadChunkEvent            map[EID]ChanForLoadChunkEvent
	chansForUnloadChunkEvent          map[EID]ChanForUnloadChunkEvent

	chansForUpdateChunkEvent map[EID]ChanForUpdateChunkEvent

	connsBetweenPlayers map[EID]map[EID]types.Nil
	playersByChunkPos   map[ChunkPosStr]map[EID]types.Nil
}

func newOverworld(
	rndDist int32,
	spawnX, spawnY, spawnZ float64,
	spawnYaw, spawnPitch float32,
) *overworld {
	return &overworld{
		rndDist: rndDist,

		spawnX: spawnX, spawnY: spawnY, spawnZ: spawnZ,
		spawnYaw: spawnYaw, spawnPitch: spawnPitch,

		players:                    make(map[EID]Player),
		chansForAddPlayerEvent:     make(map[EID]ChanForAddPlayerEvent),
		chansForUpdateLatencyEvent: make(map[EID]ChanForUpdateLatencyEvent),
		chansForRemovePlayerEvent:  make(map[EID]ChanForRemovePlayerEvent),

		chunks: make(map[ChunkPosStr]*Chunk),

		chansForUpdateChunkEvent: make(map[EID]ChanForUpdateChunkEvent),

		chansForSpawnPlayerEvent:          make(map[EID]ChanForSpawnPlayerEvent),
		chansForDespawnEntityEvent:        make(map[EID]ChanForDespawnEntityEvent),
		chansForSetEntityLookEvent:        make(map[EID]ChanForSetEntityLookEvent),
		chansForSetEntityRelativePosEvent: make(map[EID]ChanForSetEntityRelativePosEvent),
		chansForSetEntityMetadataEvent:    make(map[EID]ChanForSetEntityMetadataEvent),
		chansForLoadChunkEvent:            make(map[EID]ChanForLoadChunkEvent),
		chansForUnloadChunkEvent:          make(map[EID]ChanForUnloadChunkEvent),

		connsBetweenPlayers: make(map[EID]map[EID]types.Nil),
		playersByChunkPos:   make(map[ChunkPosStr]map[EID]types.Nil),
	}
}

func (w *overworld) InitPlayer(
	player Player,

	chanForAddPlayerEvent ChanForAddPlayerEvent,
	chanForUpdateLatencyEvent ChanForUpdateLatencyEvent,
	chanForRemovePlayerEvent ChanForRemovePlayerEvent,
	chanForSpawnPlayerEvent ChanForSpawnPlayerEvent,
	chanForDespawnEntityEvent ChanForDespawnEntityEvent,
	chanForSetEntityRelativePosEvent ChanForSetEntityRelativePosEvent,
	chanForSetEntityLookEvent ChanForSetEntityLookEvent,
	chanForSetEntityMetadataEvent ChanForSetEntityMetadataEvent,
	chanForLoadChunkEvent ChanForLoadChunkEvent,
	chanForUnloadChunkEvent ChanForUnloadChunkEvent,

	chanForUpdateChunkEvent ChanForUpdateChunkEvent,

	cnt *Client,
) error {
	w.Lock()
	defer w.Unlock()

	spawnX, spawnY, spawnZ :=
		w.spawnX, w.spawnY, w.spawnZ
	spawnYaw, spawnPitch :=
		w.spawnYaw, w.spawnPitch

	if err := cnt.Respawn(
		-1,
		2,
		0,
		"default",
	); err != nil {
		return err
	}
	if err := cnt.Teleport(
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	); err != nil {
		return err
	}

	if err := cnt.Respawn(
		0,
		2,
		0,
		"default",
	); err != nil {
		return err
	}
	if err := cnt.Teleport(
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	); err != nil {
		return err
	}

	if err := player.UpdatePos(
		spawnX, spawnY, spawnZ,
		false,
	); err != nil {
		return err
	}
	if err := player.UpdateLook(
		spawnYaw, spawnPitch,
		false,
	); err != nil {
		return err
	}

	eid := player.GetEID()

	w.players[eid] = player

	w.chansForAddPlayerEvent[eid] =
		chanForAddPlayerEvent
	w.chansForUpdateLatencyEvent[eid] =
		chanForUpdateLatencyEvent
	w.chansForRemovePlayerEvent[eid] =
		chanForRemovePlayerEvent

	uid, username :=
		player.GetUID(),
		player.GetUsername()
	for eid1, player1 := range w.players {
		uid1, username1 :=
			player1.GetUID(),
			player1.GetUsername()
		if err := cnt.AddPlayer(
			uid1, username1,
		); err != nil {
			return err
		}

		addPlayerEvent := NewAddPlayerEvent(
			uid, username,
		)
		chanForEvent1 :=
			w.chansForAddPlayerEvent[eid1]
		chanForEvent1 <- addPlayerEvent
		addPlayerEvent.Wait()
	}
	if err := cnt.AddPlayer(
		uid, username,
	); err != nil {
		return err
	}

	w.chansForSpawnPlayerEvent[eid] =
		chanForSpawnPlayerEvent
	w.chansForDespawnEntityEvent[eid] =
		chanForDespawnEntityEvent
	w.chansForSetEntityRelativePosEvent[eid] =
		chanForSetEntityRelativePosEvent
	w.chansForSetEntityLookEvent[eid] =
		chanForSetEntityLookEvent
	w.chansForSetEntityMetadataEvent[eid] =
		chanForSetEntityMetadataEvent
	w.chansForLoadChunkEvent[eid] =
		chanForLoadChunkEvent
	w.chansForUnloadChunkEvent[eid] =
		chanForUnloadChunkEvent
	w.chansForUpdateChunkEvent[eid] =
		chanForUpdateChunkEvent

	w.connsBetweenPlayers[eid] = make(map[EID]types.Nil)

	sneaking, sprinting :=
		player.IsSneaking(), player.IsSprinting()
	spawnPlayerEvent :=
		NewSpawnPlayerEvent(
			eid, uid,
			spawnX, spawnY, spawnZ,
			spawnYaw, spawnPitch,
			sneaking, sprinting,
		)

	dist := w.rndDist
	cx, cz := toChunkPos(
		spawnX, spawnZ,
	)
	maxCx, maxCz, minCx, minCz :=
		findRect(
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
					player1.GetEID(),
					player1.GetUID()
				x1, y1, z1 :=
					player1.GetX(),
					player1.GetY(),
					player1.GetZ()
				yaw1, pitch1 :=
					player1.GetYaw(),
					player1.GetPitch()
				sneaking1, sprinting1 :=
					player1.IsSneaking(),
					player1.IsSprinting()
				if err := cnt.SpawnPlayer(
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

func (w *overworld) UpdatePlayerLatency(
	uid UID,
	latency int32,
) error {
	w.RLock()
	defer w.RUnlock()

	updateLatencyEvent := NewUpdateLatencyEvent(
		uid, latency,
	)
	for eid, _ := range w.players {
		chanForUpdateLatencyEvent :=
			w.chansForUpdateLatencyEvent[eid]
		chanForUpdateLatencyEvent <- updateLatencyEvent
	}

	return nil
}

func (w *overworld) UpdatePlayerPos(
	player Player,
	x, y, z float64,
	ground bool,
) error {
	w.RLock()
	defer w.RUnlock()

	prevX, prevY, prevZ :=
		player.GetX(),
		player.GetY(),
		player.GetZ()
	deltaX, deltaY, deltaZ :=
		int16(((x*32)-(prevX*32))*128),
		int16(((y*32)-(prevY*32))*128),
		int16(((z*32)-(prevZ*32))*128)
	if deltaX == 0 && deltaY == 0 && deltaZ == 0 {
		return nil
	}

	eid := player.GetEID()
	setEntityRelativePosEvent :=
		NewSetEntityRelativePosEvent(
			eid,
			deltaX, deltaY, deltaZ,
			ground,
		)

	a := w.connsBetweenPlayers[eid]
	for eid1, _ := range a {
		chanForSetEntityRelativePosEvent1 :=
			w.chansForSetEntityRelativePosEvent[eid1]
		chanForSetEntityRelativePosEvent1 <- setEntityRelativePosEvent
	}

	if err := player.UpdatePos(
		x, y, z,
		ground,
	); err != nil {
		return err
	}

	currCx, currCz := toChunkPos(
		x, z,
	)
	prevCx, prevCz := toChunkPos(
		prevX, prevZ,
	)
	if currCx == prevCx && currCz == prevCz {
		return nil
	}

	updateChunkEvent := NewUpdateChunkEvent(
		currCx, currCz,
		prevCx, prevCz,
	)
	chanForUpdateChunkEvent :=
		w.chansForUpdateChunkEvent[eid]
	chanForUpdateChunkEvent <- updateChunkEvent

	return nil
}

func (w *overworld) UpdatePlayerChunk(
	player Player,
	prevCx, prevCz int32,
	currCx, currCz int32,
) error {
	w.Lock()
	defer w.Unlock()

	eid, uid :=
		player.GetEID(), player.GetUID()
	x, y, z :=
		player.GetX(),
		player.GetY(),
		player.GetZ()
	yaw, pitch :=
		player.GetYaw(),
		player.GetPitch()
	sneaking, sprinting :=
		player.IsSneaking(),
		player.IsSprinting()
	spawnPlayerEvent :=
		NewSpawnPlayerEvent(
			eid, uid,
			x, y, z,
			yaw, pitch,
			sneaking, sprinting,
		)

	overworld, init := true, true
	chanForLoadChunkEvent :=
		w.chansForLoadChunkEvent[eid]
	chanForSpawnPlayerEvent :=
		w.chansForSpawnPlayerEvent[eid]

	dist := w.rndDist
	maxCx, maxCz, minCx, minCz :=
		findRect(currCx, currCz, dist)
	maxPrevCx, maxPrevCz,
		minPrevCx, minPrevCz :=
		findRect(
			prevCx, prevCz, dist,
		)
	maxSubCx, maxSubCz,
		minSubCx, minSubCz :=
		subRects(
			maxCx, maxCz,
			minCx, minCz,
			maxPrevCx, maxPrevCz,
			minPrevCx, minPrevCz,
		)
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

			loadChunkEvent :=
				NewLoadChunkEvent(
					overworld, init,
					cx, cz,
					chunk,
				)
			chanForLoadChunkEvent <- loadChunkEvent

			a, has := w.playersByChunkPos[chunkPosStr]
			if has == false {
				continue
			}
			for eid1, _ := range a {
				player1 := w.players[eid1]
				eid1, uid1 :=
					player1.GetEID(), player1.GetUID()
				x1, y1, z1 :=
					player1.GetX(), player1.GetY(), player1.GetZ()
				yaw1, pitch1 :=
					player1.GetYaw(), player1.GetPitch()
				sneaking1, sprinting1 :=
					player1.IsSneaking(), player1.IsSprinting()
				spawnPlayerEvent1 := NewSpawnPlayerEvent(
					eid1, uid1,
					x1, y1, z1,
					yaw1, pitch1,
					sneaking1, sprinting1,
				)
				chanForSpawnPlayerEvent <- spawnPlayerEvent1

				chanForSpawnPlayerEvent1 :=
					w.chansForSpawnPlayerEvent[eid1]
				chanForSpawnPlayerEvent1 <- spawnPlayerEvent

				w.connsBetweenPlayers[eid][eid1] = types.Nil{}
				w.connsBetweenPlayers[eid1][eid] = types.Nil{}
			}
		}
	}

	despawnEntityEvent :=
		NewDespawnEntityEvent(
			eid,
		)
	chanForUnloadChunkEvent :=
		w.chansForUnloadChunkEvent[eid]
	chanForDespawnEntityEvent :=
		w.chansForDespawnEntityEvent[eid]
	for cz := maxPrevCz; cz >= minPrevCz; cz-- {
		for cx := maxPrevCx; cx >= minPrevCx; cx-- {
			if minSubCx <= cx && cx <= maxSubCx &&
				minSubCz <= cz && cz <= maxSubCz {
				continue
			}

			chunkPosStr := toChunkPosStr(cx, cz)

			unloadChunkEvent := NewUnloadChunkEvent(
				cx, cz,
			)
			chanForUnloadChunkEvent <- unloadChunkEvent

			a, has := w.playersByChunkPos[chunkPosStr]
			if has == false {
				continue
			}
			for eid1, _ := range a {
				despawnEntityEvent1 :=
					NewDespawnEntityEvent(
						eid1,
					)
				chanForDespawnEntityEvent <- despawnEntityEvent1

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

	chunkPosStr := toChunkPosStr(currCx, currCz)
	a, has := w.playersByChunkPos[chunkPosStr]
	if has == false {
		b := make(map[EID]types.Nil)
		w.playersByChunkPos[chunkPosStr] = b
		a = b
	}
	a[eid] = types.Nil{}

	return nil
}

func (w *overworld) UpdatePlayerLook(
	player Player,
	yaw, pitch float32,
	ground bool,
) error {
	w.RLock()
	defer w.RUnlock()

	eid := player.GetEID()
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

	if err := player.UpdateLook(
		yaw, pitch,
		ground,
	); err != nil {
		return err
	}

	return nil
}

func (w *overworld) UpdatePlayerSneaking(
	player Player,
	sneaking bool,
) error {
	w.RLock()
	defer w.RUnlock()

	eid := player.GetEID()
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

	if err := player.UpdateSneaking(
		sneaking,
	); err != nil {
		return err
	}

	return nil
}

func (w *overworld) UpdatePlayerSprinting(
	player Player,
	sprinting bool,
) error {
	w.RLock()
	defer w.RUnlock()

	eid := player.GetEID()
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

	if err := player.UpdateSprinting(
		sprinting,
	); err != nil {
		return err
	}

	return nil
}

func (w *overworld) ClosePlayer(
	player Player,
) (
	ChanForAddPlayerEvent,
	ChanForUpdateLatencyEvent,
	ChanForRemovePlayerEvent,
	ChanForSpawnPlayerEvent,
	ChanForDespawnEntityEvent,
	ChanForSetEntityRelativePosEvent,
	ChanForSetEntityLookEvent,
	ChanForSetEntityMetadataEvent,
	ChanForLoadChunkEvent,
	ChanForUnloadChunkEvent,

	ChanForUpdateChunkEvent,
) {
	w.Lock()
	defer w.Unlock()

	eid := player.GetEID()
	delete(w.players, eid)

	chanForAddPlayerEvent :=
		w.chansForAddPlayerEvent[eid]
	chanForUpdateLatencyEvent :=
		w.chansForUpdateLatencyEvent[eid]
	chanForRemovePlayerEvent :=
		w.chansForRemovePlayerEvent[eid]

	delete(w.chansForAddPlayerEvent, eid)
	delete(w.chansForUpdateLatencyEvent, eid)
	delete(w.chansForRemovePlayerEvent, eid)

	uid := player.GetUID()
	for eid1, player1 := range w.players {
		uid1 := player1.GetUID()
		removePlayerEvent1 :=
			NewRemovePlayerEvent(
				uid1,
			)
		chanForRemovePlayerEvent <- removePlayerEvent1
		removePlayerEvent1.Wait()

		removePlayerEvent :=
			NewRemovePlayerEvent(
				uid,
			)
		chanForRemovePlayerEvent1 :=
			w.chansForRemovePlayerEvent[eid1]
		chanForRemovePlayerEvent1 <- removePlayerEvent
		removePlayerEvent.Wait()
	}
	removePlayerEvent :=
		NewRemovePlayerEvent(
			uid,
		)
	chanForRemovePlayerEvent <- removePlayerEvent
	removePlayerEvent.Wait()

	chanForSpawnPlayerEvent :=
		w.chansForSpawnPlayerEvent[eid]
	chanForDespawnEntityEvent :=
		w.chansForDespawnEntityEvent[eid]
	chanForSetEntityRelativePosEvent :=
		w.chansForSetEntityRelativePosEvent[eid]
	chanForSetEntityLookEvent :=
		w.chansForSetEntityLookEvent[eid]
	chanForSetEntityMetadataEvent :=
		w.chansForSetEntityMetadataEvent[eid]
	chanForLoadChunkEvent :=
		w.chansForLoadChunkEvent[eid]
	chanForUnloadChunkEvent :=
		w.chansForUnloadChunkEvent[eid]

	delete(w.chansForSpawnPlayerEvent, eid)
	delete(w.chansForDespawnEntityEvent, eid)
	delete(w.chansForSetEntityRelativePosEvent, eid)
	delete(w.chansForSetEntityLookEvent, eid)
	delete(w.chansForSetEntityMetadataEvent, eid)
	delete(w.chansForLoadChunkEvent, eid)
	delete(w.chansForUnloadChunkEvent, eid)

	chanForUpdateChunkEvent :=
		w.chansForUpdateChunkEvent[eid]
	delete(w.chansForUpdateChunkEvent, eid)

	despawnEntityEvent := NewDespawnEntityEvent(
		eid,
	)
	a := w.connsBetweenPlayers[eid]
	for eid1, _ := range a {
		despawnPlayerEvent1 :=
			NewDespawnEntityEvent(
				eid1,
			)
		chanForDespawnEntityEvent <- despawnPlayerEvent1

		chanForDespawnEntityEvent1 :=
			w.chansForDespawnEntityEvent[eid1]
		chanForDespawnEntityEvent1 <- despawnEntityEvent

		delete(w.connsBetweenPlayers[eid1], eid)
	}
	delete(w.connsBetweenPlayers, eid)

	x, z := player.GetX(), player.GetZ()
	cx, cz := toChunkPos(x, z)
	chunkPosStr := toChunkPosStr(cx, cz)
	delete(w.playersByChunkPos[chunkPosStr], eid)

	return chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForSetEntityRelativePosEvent,
		chanForSetEntityLookEvent,
		chanForSetEntityMetadataEvent,
		chanForLoadChunkEvent,
		chanForUnloadChunkEvent,

		chanForUpdateChunkEvent
}

func (w *overworld) MakeFlat(
	block *Block,
) {
	w.Lock()
	defer w.Unlock()

	for cz := int32(10); cz >= -10; cz-- {
		for cx := int32(10); cx >= -10; cx-- {
			chunk := NewChunk()
			part := NewChunkPart()
			for z := 0; z < ChunkPartWidth; z++ {
				for x := 0; x < ChunkPartWidth; x++ {
					part.SetBlock(uint8(x), 0, uint8(z), block)
				}
			}

			chunk.SetChunkPart(4, part)
			chunkPosStr := toChunkPosStr(cx, cz)
			w.chunks[chunkPosStr] = chunk
		}
	}
}

type Lobby struct {
	sync.RWMutex

	*overworld
}

func NewLobby(
	rndDist int32,
	spawnX, spawnY, spawnZ float64,
	spawnYaw, spawnPitch float32,
) *Lobby {
	return &Lobby{
		overworld: newOverworld(
			rndDist,
			spawnX, spawnY, spawnZ,
			spawnYaw, spawnPitch,
		),
	}
}

func (w *Lobby) UpdatePlayerPos(
	player Player,
	x, y, z float64,
	ground bool,
) error {
	w.RLock()
	defer w.RUnlock()

	if err := w.overworld.UpdatePlayerPos(
		player,
		x, y, z,
		ground,
	); err != nil {
		return err
	}

	return nil
}

type GreenRoom struct {
	sync.RWMutex

	*overworld
}

func NewGreenRoom(
	rndDist int32,
	spawnX, spawnY, spawnZ float64,
	spawnYaw, spawnPitch float32,
) *GreenRoom {
	return &GreenRoom{
		overworld: newOverworld(
			rndDist,
			spawnX, spawnY, spawnZ,
			spawnYaw, spawnPitch,
		),
	}
}

type Battlefield struct {
	sync.RWMutex

	*overworld
}

func NewBattlefield(
	rndDist int32,
	spawnX, spawnY, spawnZ float64,
	spawnYaw, spawnPitch float32,
) *Battlefield {
	return &Battlefield{
		overworld: newOverworld(
			rndDist,
			spawnX, spawnY, spawnZ,
			spawnYaw, spawnPitch,
		),
	}
}
