package server

import (
	"sync"
)

type Dim struct {
	sync.RWMutex

	gameMgr *GameMgr
	world   Overworld
	player  Player
}

func NewDim(
	gameMgr *GameMgr,
	world Overworld,
	player Player,
) *Dim {
	return &Dim{
		gameMgr: gameMgr,
		world:   world,
		player:  player,
	}
}

func (dim *Dim) Init(
	chanForChangeDimEvent ChanForChangeDimEvent,
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
	dim.Lock()
	defer dim.Unlock()

	gameMgr := dim.gameMgr
	world := dim.world
	player := dim.player

	eid := player.GetEID()
	if err := gameMgr.Init(
		eid,
		chanForChangeDimEvent,
	); err != nil {
		return err
	}

	if err := world.InitPlayer(
		player,
		chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForSetEntityRelativePosEvent,
		chanForSetEntityLookEvent,
		chanForSetEntityMetadataEvent,
		chanForLoadChunkEvent,
		chanForUnloadChunkEvent,
		chanForUpdateChunkEvent,
		cnt,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dim) Change(
	world Overworld,
	player Player,
	cnt *Client,
) error {
	dim.Lock()
	defer dim.Unlock()

	prevWorld := dim.world
	prevPlayer := dim.player

	dim.world = world
	dim.player = player

	//eid := player.GetEID()
	chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForSetEntityLookEvent,
		chanForSetEntityRelativePosEvent,
		chanForSetEntityMetadataEvent,
		chanForLoadChunkEvent,
		chanForUnloadChunkEvent,

		chanForUpdateChunkEvent :=
		prevWorld.ClosePlayer(
			prevPlayer,
		)

	if err := world.InitPlayer(
		player,

		chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForSetEntityLookEvent,
		chanForSetEntityRelativePosEvent,
		chanForSetEntityMetadataEvent,
		chanForLoadChunkEvent,
		chanForUnloadChunkEvent,

		chanForUpdateChunkEvent,

		cnt,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dim) EnterChatMessage(
	text string,
) error {
	dim.RLock()
	defer dim.RUnlock()

	gameMgr := dim.gameMgr
	player := dim.player

	if text == "join" {
		if err := gameMgr.Join(
			player,
			0,
		); err != nil {
			return err
		}
	} else if text == "leave" {
		if err := gameMgr.Leave(
			player,
		); err != nil {
			return err
		}
	}

	return nil
}

func (dim *Dim) UpdatePlayerLatency(
	latency int32,
) error {
	dim.RLock()
	defer dim.RUnlock()

	world := dim.world
	player := dim.player

	uid := player.GetUID()
	if err := world.UpdatePlayerLatency(
		uid, latency,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dim) UpdatePlayerPos(
	x, y, z float64,
	ground bool,
) error {
	dim.RLock()
	defer dim.RUnlock()

	world := dim.world
	player := dim.player

	if err := world.UpdatePlayerPos(
		player,
		x, y, z,
		ground,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dim) UpdatePlayerChunk(
	prevCx, prevCz int32,
	currCx, currCz int32,
) error {
	dim.RLock()
	defer dim.RUnlock()

	world := dim.world
	player := dim.player

	if err := world.UpdatePlayerChunk(
		player,
		prevCx, prevCz,
		currCx, currCz,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dim) UpdatePlayerLook(
	yaw, pitch float32,
	ground bool,
) error {
	dim.RLock()
	defer dim.RUnlock()

	world := dim.world
	player := dim.player

	if err := world.UpdatePlayerLook(
		player,
		yaw, pitch,
		ground,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dim) UpdatePlayerSneaking(
	sneaking bool,
) error {
	dim.RLock()
	defer dim.RUnlock()

	world := dim.world
	player := dim.player

	if err := world.UpdatePlayerSneaking(
		player,
		sneaking,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dim) UpdatePlayerSprinting(
	sprinting bool,
) error {
	dim.RLock()
	defer dim.RUnlock()

	world := dim.world
	player := dim.player

	if err := world.UpdatePlayerSprinting(
		player,
		sprinting,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dim) Close() (
	ChanForChangeDimEvent,
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
	dim.Lock()
	defer dim.Unlock()

	gameMgr := dim.gameMgr
	world := dim.world
	player := dim.player

	eid := player.GetEID()
	chanForChangeDimEvent :=
		gameMgr.Close(eid)

	chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForSetEntityRelativePosEvent,
		chanForSetEntityLookEvent,
		chanForSetEntityMetadataEvent,
		chanForLoadChunkEvent,
		chanForUnloadChunkEvent,
		chanForUpdateChunkEvent :=
		world.ClosePlayer(player)

	return chanForChangeDimEvent,
		chanForAddPlayerEvent,
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
