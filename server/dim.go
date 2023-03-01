package server

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type Dimension struct {
	*sync.RWMutex

	space *Space

	eid      EID
	uid      UID
	username string
}

func NewDimension(
	space *Space,
	eid EID,
	uid UID, username string,
	index int,
	chanForAPEvent ChanForAddPlayerEvent,
	chanForULEvent ChanForUpdateLatencyEvent,
	chanForRPEvent ChanForRemovePlayerEvent,
	chanForSPEvent ChanForSpawnPlayerEvent,
	chanForDEEvent ChanForDespawnEntityEvent,
	chanForSERPEvent ChanForSetEntityRelativePosEvent,
	chanForSELEvent ChanForSetEntityLookEvent,
	chanForSEMEvent ChanForSetEntityMetadataEvent,
	chanForLCEvent ChanForLoadChunkEvent,
	chanForUnCEvent ChanForUnloadChunkEvent,
	chanForUpCEvent ChanForUpdateChunkEvent,
	cnt *Client,
) (
	*Dimension,
	error,
) {

	if err := space.Init(
		eid,
		uid, username,
		index,
		chanForAPEvent,
		chanForULEvent,
		chanForRPEvent,
		chanForSPEvent,
		chanForDEEvent,
		chanForSERPEvent,
		chanForSELEvent,
		chanForSEMEvent,
		chanForLCEvent,
		chanForUnCEvent,
		chanForUpCEvent,
		cnt,
	); err != nil {
		return nil, err
	}

	return &Dimension{
		new(sync.RWMutex),
		space,
		eid,
		uid, username,
	}, nil
}

func (dim *Dimension) Change(
	space *Space,
	index int,
	cnt *Client,
) error {
	dim.Lock()
	defer dim.Unlock()

	prevSpace := dim.space

	eid := dim.eid
	uid, username :=
		dim.uid, dim.username

	chanForAPEvent,
		chanForULEvent,
		chanForRPEvent,
		chanForSPEvent,
		chanForDEEvent,
		chanForSERPEvent,
		chanForSELEvent,
		chanForSEMEvent,
		chanForLCEvent,
		chanForUnCEvent,
		chanForUpCEvent,
		err :=
		prevSpace.Finish(
			eid,
			cnt,
		)
	if err != nil {
		return err
	}

	if err := space.Init(
		eid,
		uid, username,
		index,
		chanForAPEvent,
		chanForULEvent,
		chanForRPEvent,
		chanForSPEvent,
		chanForDEEvent,
		chanForSERPEvent,
		chanForSELEvent,
		chanForSEMEvent,
		chanForLCEvent,
		chanForUnCEvent,
		chanForUpCEvent,
		cnt,
	); err != nil {
		return err
	}

	dim.space = space

	return nil
}

func (dim *Dimension) ChangeWorld(
	index int,
	cnt *Client,
) error {
	dim.RLock()
	defer dim.RUnlock()

	space := dim.space

	eid := dim.eid
	uid, username :=
		dim.uid, dim.username

	if err := space.Change(
		eid,
		uid, username,
		index,
		cnt,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dimension) EnterChatText(
	headCmdMgr *HeadCmdMgr,
	worldCmdMgr *WorldCmdMgr,
	text string,
	cnt *Client,
) error {
	dim.RLock()
	defer dim.RUnlock()

	space := dim.space

	eid := dim.eid
	uid, username :=
		dim.uid, dim.username

	dec := text[0]

	if dec != 47 { // ascii table: 47 2F 057 &#47; /
		// TODO: process normal chat msg

		return nil
	}

	text = text[1:]
	chars := strings.Split(text, " ") // space
	args, cmd, err := headCmdMgr.Distribute(
		chars,
	)
	if err != nil {
		return err
	}

	switch cmd {
	case HeadCmdForWorld:
		args, cmd, err := worldCmdMgr.Distribute(
			args,
		)
		if err != nil {
			return err
		}
		switch cmd {
		case WorldCmdToChange:
			length := len(args)
			if length != 1 {
				return errors.New("it is invalid length of arguments to change world")
			}

			indexStr := args[0]
			fmt.Println(indexStr)
			index, err := strconv.Atoi(
				indexStr,
			)
			if err != nil {
				return err
			}

			numOfWorlds :=
				space.GetNumberOfWorlds()
			if index < 0 || numOfWorlds <= index {
				return fmt.Errorf(
					"it is invalid index %d to init in space",
					index,
				)
			}

			if err := space.Change(
				eid,
				uid, username,
				index,
				cnt,
			); err != nil {
				return err
			}

			break
		case WorldCmdToTeleport:
			return errors.New("not implemented")
			break
		}
		break
		//case  HeadCmdForGame:
		//	break
	}

	return nil
}

func (dim *Dimension) UpdateLatency(
	latency int32,
) error {
	dim.RLock()
	defer dim.RUnlock()

	space := dim.space

	eid := dim.eid

	if err := space.UpdatePlayerLatency(
		eid,
		latency,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dimension) UpdatePlayerPos(
	x, y, z float64,
	ground bool,
) error {
	dim.RLock()
	defer dim.RUnlock()

	space := dim.space

	eid := dim.eid

	if err := space.UpdatePlayerPos(
		eid,
		x, y, z,
		ground,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dimension) UpdatePlayerChunk(
	prevCx, prevCz int32,
	currCx, currCz int32,
) error {
	dim.RLock()
	defer dim.RUnlock()

	space := dim.space

	eid := dim.eid

	if err := space.UpdatePlayerChunk(
		eid,
		prevCx, prevCz,
		currCx, currCz,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dimension) UpdatePlayerLook(
	yaw, pitch float32,
	ground bool,
) error {
	dim.RLock()
	defer dim.RUnlock()

	space := dim.space

	eid := dim.eid

	if err := space.UpdatePlayerLook(
		eid,
		yaw, pitch,
		ground,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dimension) UpdatePlayerSneaking(
	flag bool,
) error {
	dim.RLock()
	defer dim.RUnlock()

	space := dim.space

	eid := dim.eid

	if err := space.UpdatePlayerSneaking(
		eid,
		flag,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dimension) UpdatePlayerSprinting(
	flag bool,
) error {
	dim.RLock()
	defer dim.RUnlock()

	space := dim.space

	eid := dim.eid

	if err := space.UpdatePlayerSprinting(
		eid,
		flag,
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

	space := dim.space

	eid := dim.eid

	chanForAPEvent,
		chanForULEvent,
		chanForRPEvent,
		chanForSPEvent,
		chanForDEEvent,
		chanForSERPEvent,
		chanForSELEvent,
		chanForSEMEvent,
		chanForLCEvent,
		chanForUnCEvent,
		chanForUpCEvent :=
		space.Close(
			eid,
		)

	return chanForAPEvent,
		chanForULEvent,
		chanForRPEvent,
		chanForSPEvent,
		chanForDEEvent,
		chanForSERPEvent,
		chanForSELEvent,
		chanForSEMEvent,
		chanForLCEvent,
		chanForUnCEvent,
		chanForUpCEvent
}
