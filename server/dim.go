package server

import (
	"github.com/google/uuid"
	"sync"
)

type Dimension struct {
	*sync.RWMutex

	pl    *PlayerList
	world Overworld

	eid      int32
	uid      uuid.UUID
	username string

	CHForCWEvent ChanForClickWindowEvent
}

func NewDimension(
	pl *PlayerList,
	world Overworld,
	eid int32,
	uid uuid.UUID, username string,
	CHForAPEvent ChanForAddPlayerEvent,
	CHForULEvent ChanForUpdateLatencyEvent,
	CHForRPEvent ChanForRemovePlayerEvent,
	CHForSPEvent ChanForSpawnPlayerEvent,
	CHForSERPEvent ChanForSetEntityRelativeMoveEvent,
	CHForSELEvent ChanForSetEntityLookEvent,
	CHForSEAEvent ChanForSetEntityActionsEvent,
	CHForDEEvent ChanForDespawnEntityEvent,
	CHForLCEvent ChanForLoadChunkEvent,
	CHForUnCEvent ChanForUnloadChunkEvent,
	CHForUpCEvent ChanForUpdateChunkEvent,
	CHForCWEvent ChanForClickWindowEvent,
	cnt *Client,
) (
	*Dimension,
	error,
) {

	if err := pl.Init(
		eid,
		uid, username,
		CHForAPEvent,
		CHForULEvent,
		CHForRPEvent,
		cnt,
	); err != nil {
		return nil, err
	}

	if err := world.InitPlayer(
		eid,
		uid,
		CHForSPEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUnCEvent,
		CHForUpCEvent,
		CHForCWEvent,
		cnt,
	); err != nil {
		return nil, err
	}

	return &Dimension{
		new(sync.RWMutex),

		pl,
		world,

		eid,
		uid, username,

		CHForCWEvent,
	}, nil
}

func (dim *Dimension) ChangePlayerList(
	pl *PlayerList,
	cnt *Client,
) error {
	dim.RLock()
	defer dim.RUnlock()

	prevPl := dim.pl
	dim.pl = pl

	eid := dim.eid
	uid, username :=
		dim.uid, dim.username

	CHForAPEvent,
		CHForULEvent,
		CHForRPEvent,
		err := prevPl.Finish(
		eid,
		cnt,
	)
	if err != nil {
		return err
	}

	if err := pl.Init(
		eid,
		uid, username,
		CHForAPEvent,
		CHForULEvent,
		CHForRPEvent,
		cnt,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dimension) ChangeWorld(
	world Overworld,
	cnt *Client,
) error {
	dim.RLock()
	defer dim.RUnlock()

	prevWld := dim.world
	dim.world = world

	eid := dim.eid
	//uid, username :=
	//	dim.uid, dim.username
	uid := dim.uid

	CHForCWEvent := dim.CHForCWEvent

	CHForSPEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUnCEvent,
		CHForUpCEvent,
		err := prevWld.FinishPlayer(
		eid,
		cnt,
	)
	if err != nil {
		return err
	}

	if err := world.InitPlayer(
		eid,
		uid,
		CHForSPEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUnCEvent,
		CHForUpCEvent,
		CHForCWEvent,
		cnt,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dimension) EnterChatText(
	text string,
	cnt *Client,
) error {
	dim.RLock()
	defer dim.RUnlock()

	return nil
}

func (dim *Dimension) UpdateLatency(
	ms int32,
) error {
	dim.RLock()
	defer dim.RUnlock()

	pl := dim.pl

	eid := dim.eid

	if err := pl.UpdateLatency(
		eid,
		ms,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dimension) ClickWindow(
	winID int8,
	slot int16,
	btn int8,
	act int16,
	mode int32,
) error {
	dim.RLock()
	defer dim.RUnlock()

	dim.CHForCWEvent <- NewClickWindowEvent(
		winID,
		slot,
		btn,
		act,
		mode,
	)

	return nil
}

func (dim *Dimension) UpdatePos(
	x, y, z float64,
	ground bool,
) error {
	dim.RLock()
	defer dim.RUnlock()

	world := dim.world

	eid := dim.eid

	if err := world.UpdatePosForPlayer(
		eid,
		x, y, z,
		ground,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dimension) UpdateLook(
	yaw, pitch float32,
	ground bool,
) error {
	dim.RLock()
	defer dim.RUnlock()

	world := dim.world

	eid := dim.eid

	if err := world.UpdateLookForPlayer(
		eid,
		yaw, pitch,
		ground,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dimension) UpdateChunk(
	prevCx, prevCz int32,
	currCx, currCz int32,
) error {
	dim.RLock()
	defer dim.RUnlock()

	world := dim.world

	eid := dim.eid

	if err := world.UpdateChunkForPlayer(
		eid,
		prevCx, prevCz,
		currCx, currCz,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dimension) Close() (
	ChanForAddPlayerEvent,
	ChanForUpdateLatencyEvent,
	ChanForRemovePlayerEvent,
	ChanForSpawnPlayerEvent,
	ChanForSetEntityRelativeMoveEvent,
	ChanForSetEntityLookEvent,
	ChanForSetEntityActionsEvent,
	ChanForDespawnEntityEvent,
	ChanForLoadChunkEvent,
	ChanForUnloadChunkEvent,
	ChanForUpdateChunkEvent,
	ChanForClickWindowEvent,
) {
	dim.Lock()
	defer dim.Unlock()

	pl := dim.pl
	world := dim.world

	eid := dim.eid

	CHForAPEvent,
		CHForULEvent,
		CHForRPEvent :=
		pl.Close(eid)

	CHForSPEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUnCEvent,
		CHForUpCEvent :=
		world.ClosePlayer(eid)

	return CHForAPEvent,
		CHForULEvent,
		CHForRPEvent,
		CHForSPEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUnCEvent,
		CHForUpCEvent,
		dim.CHForCWEvent
}
