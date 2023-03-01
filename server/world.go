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

type Overworld interface {
	InitPlayer(
		eid EID,
		uid UID, username string,
		chansForSPEvent ChanForSpawnPlayerEvent,
		chansForDEEvent ChanForDespawnEntityEvent,
		chansForSERPEvent ChanForSetEntityRelativePosEvent,
		chansForSELEvent ChanForSetEntityLookEvent,
		chansForSEMEvent ChanForSetEntityMetadataEvent,
		chansForLCEvent ChanForLoadChunkEvent,
		chansForUnCEvent ChanForUnloadChunkEvent,
		chansForUpCEvent ChanForUpdateChunkEvent,
		cnt *Client,
	) error
	UpdatePlayerPos(
		eid EID,
		x, y, z float64,
		ground bool,
	) error
	UpdatePlayerChunk(
		eid EID,
		prevCx, prevCz int32,
		currCx, currCz int32,
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
	FinishPlayer(
		eid EID,
		cnt *Client,
	) (
		ChanForSpawnPlayerEvent,
		ChanForDespawnEntityEvent,
		ChanForSetEntityRelativePosEvent,
		ChanForSetEntityLookEvent,
		ChanForSetEntityMetadataEvent,
		ChanForLoadChunkEvent,
		ChanForUnloadChunkEvent,
		ChanForUpdateChunkEvent,
		error,
	)
	Close(
		eid EID,
	) (
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
	*sync.RWMutex

	rndDist int32

	spawnX, spawnY, spawnZ float64
	spawnYaw, spawnPitch   float32

	chunks map[ChunkPosStr]*Chunk

	conns map[EID]map[EID]types.Nil         // connsBetweenPlayers
	PBCP  map[ChunkPosStr]map[EID]types.Nil // playersByChunkPos

	players           map[EID]*Player
	chansForSPEvent   map[EID]ChanForSpawnPlayerEvent
	chansForDEEvent   map[EID]ChanForDespawnEntityEvent
	chansForSELEvent  map[EID]ChanForSetEntityLookEvent
	chansForSERPEvent map[EID]ChanForSetEntityRelativePosEvent
	chansForSEMEvent  map[EID]ChanForSetEntityMetadataEvent
	chansForLCEvent   map[EID]ChanForLoadChunkEvent
	chansForUnCEvent  map[EID]ChanForUnloadChunkEvent
	chansForUpCEvent  map[EID]ChanForUpdateChunkEvent
}

func newOverworld(
	rndDist int32,
	spawnX, spawnY, spawnZ float64,
	spawnYaw, spawnPitch float32,
) *overworld {
	return &overworld{
		new(sync.RWMutex),

		rndDist,

		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,

		make(map[ChunkPosStr]*Chunk),

		make(map[EID]map[EID]types.Nil),
		make(map[ChunkPosStr]map[EID]types.Nil),

		make(map[EID]*Player),
		make(map[EID]ChanForSpawnPlayerEvent),
		make(map[EID]ChanForDespawnEntityEvent),
		make(map[EID]ChanForSetEntityLookEvent),
		make(map[EID]ChanForSetEntityRelativePosEvent),
		make(map[EID]ChanForSetEntityMetadataEvent),
		make(map[EID]ChanForLoadChunkEvent),
		make(map[EID]ChanForUnloadChunkEvent),
		make(map[EID]ChanForUpdateChunkEvent),
	}
}

func (w *overworld) InitPlayer(
	eid EID,
	uid UID, username string,
	chanForSPEvent ChanForSpawnPlayerEvent,
	chanForDEEvent ChanForDespawnEntityEvent,
	chanForSERPEvent ChanForSetEntityRelativePosEvent,
	chanForSELEvent ChanForSetEntityLookEvent,
	chanForSEMEvent ChanForSetEntityMetadataEvent,
	chanForLCEvent ChanForLoadChunkEvent,
	chanForUnCEvent ChanForUnloadChunkEvent,
	chanForUpCEvent ChanForUpdateChunkEvent,
	cnt *Client,
) error {
	w.Lock()
	defer w.Unlock()

	dist := w.rndDist
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

	p := NewPlayer(
		eid,
		uid, username,
	)

	if err := p.UpdatePos(
		spawnX, spawnY, spawnZ,
		false,
	); err != nil {
		return err
	}
	if err := p.UpdateLook(
		spawnYaw, spawnPitch,
		false,
	); err != nil {
		return err
	}

	w.conns[eid] = make(map[EID]types.Nil)

	sneaking, sprinting :=
		p.IsSneaking(),
		p.IsSprinting()
	SPEvent :=
		NewSpawnPlayerEvent(
			eid, uid,
			spawnX, spawnY, spawnZ,
			spawnYaw, spawnPitch,
			sneaking, sprinting,
		)

	overworld, init := true, true

	cx, cz := toChunkPos(
		spawnX, spawnZ,
	)
	maxCx, maxCz, minCx, minCz :=
		findRect(
			cx, cz, dist,
		)
	for cz := maxCz; cz >= minCz; cz-- {
		for cx := maxCx; cx >= minCx; cx-- {
			CPStr := toChunkPosStr(cx, cz)

			chunk, has := w.chunks[CPStr]
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

			d, has := w.PBCP[CPStr]
			if has == false {
				continue
			}
			for eid1, _ := range d {
				p1 := w.players[eid1]
				eid1, uid1 :=
					p1.GetEID(),
					p1.GetUID()
				x1, y1, z1 :=
					p1.GetX(),
					p1.GetY(),
					p1.GetZ()
				yaw1, pitch1 :=
					p1.GetYaw(),
					p1.GetPitch()
				sneaking1, sprinting1 :=
					p1.IsSneaking(),
					p1.IsSprinting()
				if err := cnt.SpawnPlayer(
					eid1, uid1,
					x1, y1, z1,
					yaw1, pitch1,
					sneaking1, sprinting1,
				); err != nil {
					return err
				}

				w.chansForSPEvent[eid1] <- SPEvent

				w.conns[eid][eid1] = types.Nil{}
				w.conns[eid1][eid] = types.Nil{}
			}
		}
	}

	CPStr := toChunkPosStr(cx, cz)
	eids, has := w.PBCP[CPStr]
	if has == false {
		a := make(map[EID]types.Nil)
		w.PBCP[CPStr] = a
		eids = a
	}
	eids[eid] = types.Nil{}

	w.players[eid] = p
	w.chansForSPEvent[eid] = chanForSPEvent
	w.chansForDEEvent[eid] = chanForDEEvent
	w.chansForSERPEvent[eid] = chanForSERPEvent
	w.chansForSELEvent[eid] = chanForSELEvent
	w.chansForSEMEvent[eid] = chanForSEMEvent
	w.chansForLCEvent[eid] = chanForLCEvent
	w.chansForUnCEvent[eid] = chanForUnCEvent
	w.chansForUpCEvent[eid] = chanForUpCEvent

	return nil
}

func (w *overworld) UpdatePlayerPos(
	eid EID,
	x, y, z float64,
	ground bool,
) error {
	w.RLock()
	defer w.RUnlock()

	p := w.players[eid]
	prevX, prevY, prevZ :=
		p.GetX(),
		p.GetY(),
		p.GetZ()
	deltaX, deltaY, deltaZ :=
		int16(((x*32)-(prevX*32))*128),
		int16(((y*32)-(prevY*32))*128),
		int16(((z*32)-(prevZ*32))*128)
	if deltaX == 0 && deltaY == 0 && deltaZ == 0 {
		return nil
	}

	SERPEvent :=
		NewSetEntityRelativePosEvent(
			eid,
			deltaX, deltaY, deltaZ,
			ground,
		)

	for eid, _ := range w.conns[eid] {
		w.chansForSERPEvent[eid] <- SERPEvent
	}

	if err := p.UpdatePos(
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

	UpCEvent := NewUpdateChunkEvent(
		prevCx, prevCz,
		currCx, currCz,
	)
	w.chansForUpCEvent[eid] <- UpCEvent

	return nil
}

func (w *overworld) UpdatePlayerChunk(
	eid EID,
	prevCx, prevCz int32,
	currCx, currCz int32,
) error {
	w.Lock()
	defer w.Unlock()

	chanForSPEvent := w.chansForSPEvent[eid]
	chanForDEEvent := w.chansForDEEvent[eid]
	chanForLCEvent := w.chansForLCEvent[eid]
	chanForUnCEvent := w.chansForUnCEvent[eid]

	p := w.players[eid]
	uid := p.GetUID()
	x, y, z :=
		p.GetX(),
		p.GetY(),
		p.GetZ()
	yaw, pitch :=
		p.GetYaw(),
		p.GetPitch()
	sneaking, sprinting :=
		p.IsSneaking(),
		p.IsSprinting()
	SPEvent :=
		NewSpawnPlayerEvent(
			eid, uid,
			x, y, z,
			yaw, pitch,
			sneaking, sprinting,
		)
	DEEvent :=
		NewDespawnEntityEvent(
			eid,
		)

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
	overworld, init := true, true
	for cz := maxCz; cz >= minCz; cz-- {
		for cx := maxCx; cx >= minCx; cx-- {
			if minSubCx <= cx && cx <= maxSubCx &&
				minSubCz <= cz && cz <= maxSubCz {
				continue
			}

			CPStr := toChunkPosStr(cx, cz)

			chunk, has := w.chunks[CPStr]
			if has == false {
				chunk = NewChunk()
			}
			LCEvent :=
				NewLoadChunkEvent(
					overworld, init,
					cx, cz,
					chunk,
				)
			chanForLCEvent <- LCEvent

			a, has := w.PBCP[CPStr]
			if has == false {
				continue
			}
			for eid1, _ := range a {
				p1 := w.players[eid1]
				eid1, uid1 :=
					p1.GetEID(), p1.GetUID()
				x1, y1, z1 :=
					p1.GetX(), p1.GetY(), p1.GetZ()
				yaw1, pitch1 :=
					p1.GetYaw(), p1.GetPitch()
				sneaking1, sprinting1 :=
					p1.IsSneaking(), p1.IsSprinting()
				SPEvent1 := NewSpawnPlayerEvent(
					eid1, uid1,
					x1, y1, z1,
					yaw1, pitch1,
					sneaking1, sprinting1,
				)
				chanForSPEvent <- SPEvent1

				w.chansForSPEvent[eid1] <- SPEvent

				w.conns[eid][eid1] = types.Nil{}
				w.conns[eid1][eid] = types.Nil{}
			}
		}
	}

	for cz := maxPrevCz; cz >= minPrevCz; cz-- {
		for cx := maxPrevCx; cx >= minPrevCx; cx-- {
			if minSubCx <= cx && cx <= maxSubCx &&
				minSubCz <= cz && cz <= maxSubCz {
				continue
			}

			CPStr := toChunkPosStr(cx, cz)

			UnCEvent := NewUnloadChunkEvent(
				cx, cz,
			)
			chanForUnCEvent <- UnCEvent

			a, has := w.PBCP[CPStr]
			if has == false {
				continue
			}
			for eid1, _ := range a {
				DEEvent1 :=
					NewDespawnEntityEvent(
						eid1,
					)
				chanForDEEvent <- DEEvent1

				w.chansForDEEvent[eid1] <- DEEvent

				delete(w.conns[eid], eid1)
				delete(w.conns[eid1], eid)
			}

		}
	}

	prevCPStr := toChunkPosStr(
		prevCx, prevCz,
	)
	delete(w.PBCP[prevCPStr], eid)

	currCPStr := toChunkPosStr(
		currCx, currCz,
	)
	eids, has := w.PBCP[currCPStr]
	if has == false {
		c := make(map[EID]types.Nil)
		w.PBCP[currCPStr] = c
		eids = c
	}
	eids[eid] = types.Nil{}

	return nil
}

func (w *overworld) UpdatePlayerLook(
	eid EID,
	yaw, pitch float32,
	ground bool,
) error {
	w.RLock()
	defer w.RUnlock()

	p := w.players[eid]
	SELEvent :=
		NewSetEntityLookEvent(
			eid,
			yaw, pitch,
			ground,
		)
	for eid, _ := range w.conns[eid] {
		w.chansForSELEvent[eid] <- SELEvent
	}

	if err := p.UpdateLook(
		yaw, pitch,
		ground,
	); err != nil {
		return err
	}

	return nil
}

func (w *overworld) UpdatePlayerSneaking(
	eid EID,
	sneaking bool,
) error {
	w.RLock()
	defer w.RUnlock()

	p := w.players[eid]
	sprinting := p.IsSprinting()
	metadata := NewEntityMetadata()
	if err := metadata.SetActions(
		sneaking, sprinting,
	); err != nil {
		return err
	}

	SEMEvent :=
		NewSetEntityMetadataEvent(
			eid,
			metadata,
		)
	for eid, _ := range w.conns[eid] {
		w.chansForSEMEvent[eid] <- SEMEvent
	}

	if err := p.UpdateSneaking(
		sneaking,
	); err != nil {
		return err
	}

	return nil
}

func (w *overworld) UpdatePlayerSprinting(
	eid EID,
	sprinting bool,
) error {
	w.RLock()
	defer w.RUnlock()

	p := w.players[eid]
	sneaking := p.IsSneaking()
	metadata := NewEntityMetadata()
	if err := metadata.SetActions(
		sneaking, sprinting,
	); err != nil {
		return err
	}

	SEMEvent :=
		NewSetEntityMetadataEvent(
			eid,
			metadata,
		)
	for eid, _ := range w.conns[eid] {
		w.chansForSEMEvent[eid] <- SEMEvent
	}

	if err := p.UpdateSprinting(
		sprinting,
	); err != nil {
		return err
	}

	return nil
}

func (w *overworld) FinishPlayer(
	eid EID,
	cnt *Client,
) (
	ChanForSpawnPlayerEvent,
	ChanForDespawnEntityEvent,
	ChanForSetEntityRelativePosEvent,
	ChanForSetEntityLookEvent,
	ChanForSetEntityMetadataEvent,
	ChanForLoadChunkEvent,
	ChanForUnloadChunkEvent,
	ChanForUpdateChunkEvent,
	error,
) {
	w.Lock()
	defer w.Unlock()

	p := w.players[eid]
	chanForSPEvent := w.chansForSPEvent[eid]
	chanForDEEvent := w.chansForDEEvent[eid]
	chanForSERPEvent := w.chansForSERPEvent[eid]
	chanForSELEvent := w.chansForSELEvent[eid]
	chanForSEMEvent := w.chansForSEMEvent[eid]
	chanForLCEvent := w.chansForLCEvent[eid]
	chanForUnCEvent := w.chansForUnCEvent[eid]
	chanForUpCEvent := w.chansForUpCEvent[eid]

	DEEvent :=
		NewDespawnEntityEvent(
			eid,
		)
	for eid1, _ := range w.conns[eid] {
		if err := cnt.DespawnEntity(
			eid1,
		); err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, err
		}

		w.chansForDEEvent[eid1] <- DEEvent

		delete(w.conns[eid1], eid)
	}
	delete(w.conns, eid)

	x, z := p.GetX(), p.GetZ()
	cx, cz := toChunkPos(x, z)
	CPStr := toChunkPosStr(cx, cz)
	delete(w.PBCP[CPStr], eid)

	delete(w.players, eid)
	delete(w.chansForSPEvent, eid)
	delete(w.chansForDEEvent, eid)
	delete(w.chansForSERPEvent, eid)
	delete(w.chansForSELEvent, eid)
	delete(w.chansForSEMEvent, eid)
	delete(w.chansForLCEvent, eid)
	delete(w.chansForUnCEvent, eid)
	delete(w.chansForUpCEvent, eid)

	return chanForSPEvent,
		chanForDEEvent,
		chanForSERPEvent,
		chanForSELEvent,
		chanForSEMEvent,
		chanForLCEvent,
		chanForUnCEvent,
		chanForUpCEvent,
		nil
}

func (w *overworld) Close(
	eid EID,
) (
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

	p := w.players[eid]
	chanForSPEvent := w.chansForSPEvent[eid]
	chanForDEEvent := w.chansForDEEvent[eid]
	chanForSERPEvent := w.chansForSERPEvent[eid]
	chanForSELEvent := w.chansForSELEvent[eid]
	chanForSEMEvent := w.chansForSEMEvent[eid]
	chanForLCEvent := w.chansForLCEvent[eid]
	chanForUnCEvent := w.chansForUnCEvent[eid]
	chanForUpCEvent := w.chansForUpCEvent[eid]

	DEEvent :=
		NewDespawnEntityEvent(
			eid,
		)
	for eid1, _ := range w.conns[eid] {
		w.chansForDEEvent[eid1] <- DEEvent

		delete(w.conns[eid1], eid)
	}
	delete(w.conns, eid)

	x, z := p.GetX(), p.GetZ()
	cx, cz := toChunkPos(x, z)
	CPStr := toChunkPosStr(cx, cz)
	delete(w.PBCP[CPStr], eid)

	delete(w.players, eid)
	delete(w.chansForSPEvent, eid)
	delete(w.chansForDEEvent, eid)
	delete(w.chansForSERPEvent, eid)
	delete(w.chansForSELEvent, eid)
	delete(w.chansForSEMEvent, eid)
	delete(w.chansForLCEvent, eid)
	delete(w.chansForUnCEvent, eid)
	delete(w.chansForUpCEvent, eid)

	return chanForSPEvent,
		chanForDEEvent,
		chanForSERPEvent,
		chanForSELEvent,
		chanForSEMEvent,
		chanForLCEvent,
		chanForUnCEvent,
		chanForUpCEvent
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

type WaitingRoom struct {
	*overworld
}

func NewWaitingRoom(
	rndDist int32,
	spawnX, spawnY, spawnZ float64,
	spawnYaw, spawnPitch float32,
) *WaitingRoom {
	return &WaitingRoom{
		newOverworld(
			rndDist,
			spawnX, spawnY, spawnZ,
			spawnYaw, spawnPitch,
		),
	}
}
