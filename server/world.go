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

		cnt *Client,
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
	) error
	UpdatePlayerLatency(
		uid UID,
		latency int32,
	) error
	UpdatePlayerPos(
		eid EID,
		deltaX, deltaY, deltaZ int16,
		ground bool,
	) error
	UpdatePlayerLook(
		eid EID,
		yaw, pitch float32,
		ground bool,
	) error
	UpdatePlayerSneaking(
		eid EID,
		sneaking bool,
	) error
	UpdatePlayerSprinting(
		eid EID,
		sprinting bool,
	) error
	UpdatePlayerChunk(
		eid EID, uid UID,
		x, y, z float64,
		prevX, prevY, prevZ float64,
		yaw, pitch float32,
		sneaking, sprinting bool,
	) error
	ClosePlayer(
		eid EID,
	) (
		ChanForSpawnPlayerEvent,
		ChanForDespawnEntityEvent,
		ChanForSetEntityRelativePosEvent,
		ChanForSetEntityLookEvent,
		ChanForSetEntityMetadataEvent,
		ChanForLoadChunkEvent,
		ChanForUnloadChunkEvent,

		ChanForAddPlayerEvent,
		ChanForUpdateLatencyEvent,
		ChanForRemovePlayerEvent,
	)

	MakeFlat()
}

type overworld struct {
	sync.RWMutex

	rndDist int32

	spawnX, spawnY, spawnZ float64
	spawnYaw, spawnPitch   float32

	playerList                 map[EID]*PlayerListItem
	chansForAddPlayerEvent     map[EID]ChanForAddPlayerEvent
	chansForUpdateLatencyEvent map[EID]ChanForUpdateLatencyEvent
	chansForRemovePlayerEvent  map[EID]ChanForRemovePlayerEvent

	chunks map[ChunkPosStr]*Chunk

	players                           map[EID]Player
	chansForSpawnPlayerEvent          map[EID]ChanForSpawnPlayerEvent
	chansForDespawnEntityEvent        map[EID]ChanForDespawnEntityEvent
	chansForSetEntityLookEvent        map[EID]ChanForSetEntityLookEvent
	chansForSetEntityRelativePosEvent map[EID]ChanForSetEntityRelativePosEvent
	chansForSetEntityMetadataEvent    map[EID]ChanForSetEntityMetadataEvent
	chansForLoadChunkEvent            map[EID]ChanForLoadChunkEvent
	chansForUnloadChunkEvent          map[EID]ChanForUnloadChunkEvent

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

		playerList:                 make(map[EID]*PlayerListItem),
		chansForAddPlayerEvent:     make(map[EID]ChanForAddPlayerEvent),
		chansForUpdateLatencyEvent: make(map[EID]ChanForUpdateLatencyEvent),
		chansForRemovePlayerEvent:  make(map[EID]ChanForRemovePlayerEvent),

		chunks: make(map[ChunkPosStr]*Chunk),

		players:                           make(map[EID]Player),
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

	cnt *Client,
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
) error {
	w.Lock()
	defer w.Unlock()

	eid := player.GetEid()

	uid, username :=
		player.GetUid(), player.GetUsername()
	for eid1, item := range w.playerList {
		uid1, username1 :=
			item.GetUID(),
			item.GetUsername()
		if err := cnt.AddPlayer(
			uid1, username1,
		); err != nil {
			return err
		}

		event := NewAddPlayerEvent(
			uid, username,
		)
		chanForEvent := w.chansForAddPlayerEvent[eid1]
		chanForEvent <- event
		event.Wait()
	}
	if err := cnt.AddPlayer(
		uid, username,
	); err != nil {
		return err
	}
	item := NewPlayerListItem(
		uid,
		username,
	)
	w.playerList[eid] = item
	w.chansForAddPlayerEvent[eid] =
		chanForAddPlayerEvent
	w.chansForUpdateLatencyEvent[eid] =
		chanForUpdateLatencyEvent
	w.chansForRemovePlayerEvent[eid] =
		chanForRemovePlayerEvent

	if err := cnt.Respawn(
		-1,
		2,
		0,
		"default",
	); err != nil {
		return err
	}
	if err := cnt.Teleport(
		0, 0, 0,
		0, 0,
	); err != nil {
		return err
	}

	spawnX, spawnY, spawnZ :=
		w.spawnX, w.spawnY, w.spawnZ
	spawnYaw, spawnPitch :=
		w.spawnYaw, w.spawnPitch

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

	player.SetPos(
		spawnX, spawnY, spawnZ,
		false,
	)
	player.SetLook(
		spawnYaw, spawnPitch,
		false,
	)

	w.players[eid] = player
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

	sneaking, sprinting :=
		player.IsSneaking(), player.IsSprinting()
	spawnPlayerEvent :=
		NewSpawnPlayerEvent(
			eid, uid,
			spawnX, spawnY, spawnZ,
			spawnYaw, spawnPitch,
			sneaking, sprinting,
		)

	w.connsBetweenPlayers[eid] = make(map[EID]types.Nil)

	dist := w.rndDist
	cx, cz := toChunkPos(
		spawnX, spawnZ,
	)
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
					player1.GetEid(),
					player1.GetUid()
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

	event := NewUpdateLatencyEvent(
		uid, latency,
	)
	for eid1, _ := range w.playerList {
		chanForEvent :=
			w.chansForUpdateLatencyEvent[eid1]
		chanForEvent <- event
	}

	return nil
}

func (w *overworld) UpdatePlayerPos(
	eid EID,
	deltaX, deltaY, deltaZ int16,
	ground bool,
) error {
	w.RLock()
	defer w.RUnlock()

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

func (w *overworld) UpdatePlayerLook(
	eid EID,
	yaw, pitch float32,
	ground bool,
) error {
	w.RLock()
	defer w.RUnlock()

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

func (w *overworld) UpdatePlayerSneaking(
	eid EID,
	sneaking bool,
) error {
	w.RLock()
	defer w.RUnlock()

	player := w.players[eid]
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

func (w *overworld) UpdatePlayerSprinting(
	eid EID,
	sprinting bool,
) error {
	w.RLock()
	defer w.RUnlock()

	player := w.players[eid]
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

func (w *overworld) UpdatePlayerChunk(
	eid EID, uid UID,
	x, y, z float64,
	prevX, prevY, prevZ float64,
	yaw, pitch float32,
	sneaking, sprinting bool,
) error {
	w.Lock()
	defer w.Unlock()

	cx, cz := toChunkPos(
		x, z,
	)
	prevCx, prevCz := toChunkPos(
		prevX, prevZ,
	)
	if cx == prevCx && cz == prevCz {
		return nil
	}

	dist := w.rndDist
	maxCx, maxCz, minCx, minCz :=
		findRect(cx, cz, dist)
	maxPrevCx, maxPrevCz, minPrevCx, minPrevCz :=
		findRect(prevCx, prevCz, dist)
	maxSubCx, maxSubCz, minSubCx, minSubCz :=
		subRects(
			maxCx, maxCz, minCx, minCz,
			maxPrevCx, maxPrevCz, minPrevCx, minPrevCz,
		)
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
					player1.GetEid(), player1.GetUid()
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

func (w *overworld) ClosePlayer(
	eid EID,
) (
	ChanForSpawnPlayerEvent,
	ChanForDespawnEntityEvent,
	ChanForSetEntityRelativePosEvent,
	ChanForSetEntityLookEvent,
	ChanForSetEntityMetadataEvent,
	ChanForLoadChunkEvent,
	ChanForUnloadChunkEvent,

	ChanForAddPlayerEvent,
	ChanForUpdateLatencyEvent,
	ChanForRemovePlayerEvent,
) {
	w.Lock()
	defer w.Unlock()

	player := w.players[eid]

	x, z := player.GetX(), player.GetZ()
	cx, cz := toChunkPos(x, z)
	chunkPosStr := toChunkPosStr(cx, cz)
	delete(w.playersByChunkPos[chunkPosStr], eid)

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

	delete(w.chansForSpawnPlayerEvent, eid)
	delete(w.chansForDespawnEntityEvent, eid)
	delete(w.chansForSetEntityRelativePosEvent, eid)
	delete(w.chansForSetEntityLookEvent, eid)
	delete(w.chansForSetEntityMetadataEvent, eid)
	delete(w.chansForLoadChunkEvent, eid)
	delete(w.chansForUnloadChunkEvent, eid)

	delete(w.players, eid)

	chanForAddPlayerEvent := w.chansForAddPlayerEvent[eid]
	chanForUpdateLatencyEvent := w.chansForUpdateLatencyEvent[eid]
	chanForRemovePlayerEvent := w.chansForRemovePlayerEvent[eid]

	uid := player.GetUid()
	removePlayerEvent :=
		NewRemovePlayerEvent(
			uid,
		)
	for eid1, _ := range w.playerList {
		chanForEvent := w.chansForRemovePlayerEvent[eid1]
		chanForEvent <- removePlayerEvent
	}

	delete(w.chansForAddPlayerEvent, eid)
	delete(w.chansForUpdateLatencyEvent, eid)
	delete(w.chansForRemovePlayerEvent, eid)

	delete(w.playerList, eid)

	return chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForSetEntityRelativePosEvent,
		chanForSetEntityLookEvent,
		chanForSetEntityMetadataEvent,
		chanForLoadChunkEvent,
		chanForUnloadChunkEvent,
		chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent
}

func (w *overworld) MakeFlat() {
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
	eid EID,
	deltaX, deltaY, deltaZ int16,
	ground bool,
) error {
	w.RLock()
	defer w.RUnlock()

	if err := w.overworld.UpdatePlayerPos(
		eid,
		deltaX, deltaY, deltaZ,
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
