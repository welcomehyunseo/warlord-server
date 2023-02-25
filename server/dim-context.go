package server

import (
	"sync"
)

type DimContext struct {
	sync.RWMutex
	world  Overworld
	player Player
}

func NewDimContext(
	world Overworld,
	player Player,
) *DimContext {
	return &DimContext{
		world:  world,
		player: player,
	}
}

func (ctx *DimContext) Change(
	world Overworld,
	player Player,
	cnt *Client,
) error {
	ctx.Lock()
	defer ctx.Unlock()

	prevWorld := ctx.world
	prevPlayer := ctx.player

	ctx.world = world
	ctx.player = player

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
