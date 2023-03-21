package server

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/welcomehyunseo/warlord-server/server/item"
	"github.com/welcomehyunseo/warlord-server/server/nbt"
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
	ChForSISEvent ChanForSpawnItemStandEvent,
	CHForSERMEvent ChanForSetEntityRelativeMoveEvent,
	CHForSELEvent ChanForSetEntityLookEvent,
	CHForSEAEvent ChanForSetEntityActionsEvent,
	CHForDEEvent ChanForDespawnEntityEvent,
	CHForLCEvent ChanForLoadChunkEvent,
	CHForUCEvent ChanForUnloadChunkEvent,
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
		uid, username,
		CHForSPEvent,
		ChForSISEvent,
		CHForSERMEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUCEvent,
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

	eid := dim.eid
	uid, username :=
		dim.uid, dim.username

	CHForCWEvent := dim.CHForCWEvent

	CHForSPEvent,
		ChForSISEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUCEvent,
		err := prevWld.FinishPlayer(
		eid,
		cnt,
	)
	if err != nil {
		return err
	}

	dim.world = world

	if err := world.InitPlayer(
		eid,
		uid, username,
		CHForSPEvent,
		ChForSISEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUCEvent,
		CHForCWEvent,
		cnt,
	); err != nil {
		return err
	}

	return nil
}

func (dim *Dimension) EnterChatText(
	s string,
	cnt *Client,
) error {
	dim.RLock()
	defer dim.RUnlock()

	if s[0] != 47 { // ascii table: 47 2F 057 &#47; /
		// TODO: process normal chat msg

		return nil
	}

	arr := strings.Split(s[1:], " ")
	switch arr[0] {
	default:
		return errors.New("not implemented")
	case "spawn":
		dim.world.InitItemStand(
			item.NewStickItem(1, &nbt.ItemNbt{}),
		)
		break
	case "give":
		// give <int16:id> <uint8:amt>
		// give item with non-nbt
		args := arr[1:]
		if len(args) != 2 {
			return errors.New("it is insufficient length of arguments to give item with non-nbt")
		}

		v0, err := strconv.ParseInt(
			args[0], 10, 16,
		)
		if err != nil {
			return errors.New("argument of 0th index must be int16 to give item with non-nbt")
		}
		id := int16(v0)

		v1, err := strconv.ParseUint(
			args[1], 10, 8,
		)
		if err != nil {
			return errors.New("argument of 1th index must be int8 to give item with non-nbt")
		}
		amt := uint8(v1)

		var it item.Item
		switch id {
		case item.WoodenSwordItemID:
			it = item.NewWoodenSwordItem(
				amt, &nbt.ItemNbt{},
			)
			break
		case item.StickItemID:
			it = item.NewStickItem(
				amt, &nbt.ItemNbt{},
			)
			break
		}
		if it == nil {
			return errors.New("there is unregistered id to give item with non-nbt")
		}

		return errors.New("not implemented")
		break
	case "smob":
		form := "/smob <int32:id> <float64:x> <float64:y> <float64:z> <float32:yaw> <float32:pitch> <float32:head-pitch> <int16:velocity-x> <int16:velocity-y> <int16:velocity-z> <metadata:corresponding-parameters>"

		args := arr[1:]
		if len(args) < 10 {
			return fmt.Errorf(
				"it is insufficient length of arguments in form \"%s\" to summon mobile entity",
				form,
			)
		}

		id, err := func() (int32, error) {
			v, err := strconv.ParseInt(
				args[0], 10, 32,
			)
			if err != nil {
				return 0, fmt.Errorf(
					"argument %s must be int32 in form \"%s\" to summon mobile entity",
					"id",
					form,
				)
			}
			args = args[1:]

			return int32(v), nil
		}()
		if err != nil {
			return err
		}

		x, err := func() (float64, error) {
			v, err := strconv.ParseFloat(
				args[0], 64,
			)
			if err != nil {
				return 0, fmt.Errorf(
					"argument %s must be float64 in form \"%s\" to summon mobile entity",
					"x",
					form,
				)
			}
			args = args[1:]

			return float64(v), nil
		}()
		if err != nil {
			return err
		}

		y, err := func() (float64, error) {
			v, err := strconv.ParseFloat(
				args[0], 64,
			)
			if err != nil {
				return 0, fmt.Errorf(
					"argument %s must be float64 in form \"%s\" to summon mobile entity",
					"y",
					form,
				)
			}
			args = args[1:]

			return float64(v), nil
		}()
		if err != nil {
			return err
		}

		z, err := func() (float64, error) {
			v, err := strconv.ParseFloat(
				args[0], 64,
			)
			if err != nil {
				return 0, fmt.Errorf(
					"argument %s must be float64 in form \"%s\" to summon mobile entity",
					"z",
					form,
				)
			}
			args = args[1:]

			return float64(v), nil
		}()
		if err != nil {
			return err
		}

		yaw, err := func() (float32, error) {
			v, err := strconv.ParseFloat(
				args[0], 32,
			)
			if err != nil {
				return 0, fmt.Errorf(
					"argument %s must be float32 in form \"%s\" to summon mobile entity",
					"yaw",
					form,
				)
			}
			args = args[1:]

			return float32(v), nil
		}()
		if err != nil {
			return err
		}

		pitch, err := func() (float32, error) {
			v, err := strconv.ParseFloat(
				args[0], 32,
			)
			if err != nil {
				return 0, fmt.Errorf(
					"argument %s must be float32 in form \"%s\" to summon mobile entity",
					"pitch",
					form,
				)
			}
			args = args[1:]

			return float32(v), nil
		}()
		if err != nil {
			return err
		}

		hdPitch, err := func() (float32, error) {
			v, err := strconv.ParseFloat(
				args[0], 32,
			)
			if err != nil {
				return 0, fmt.Errorf(
					"argument %s must be float32 in form \"%s\" to summon mobile entity",
					"head-pitch",
					form,
				)
			}
			args = args[1:]

			return float32(v), nil
		}()
		if err != nil {
			return err
		}

		vx, err := func() (int16, error) {
			v, err := strconv.ParseInt(
				args[0], 10, 16,
			)
			if err != nil {
				return 0, fmt.Errorf(
					"argument %s must be int16 in form \"%s\" to summon mobile entity",
					"velocity-x",
					form,
				)
			}
			args = args[1:]

			return int16(v), nil
		}()
		if err != nil {
			return err
		}

		vy, err := func() (int16, error) {
			v, err := strconv.ParseInt(
				args[0], 10, 16,
			)
			if err != nil {
				return 0, fmt.Errorf(
					"argument %s must be int16 in form \"%s\" to summon mobile entity",
					"velocity-y",
					form,
				)
			}
			args = args[1:]

			return int16(v), nil
		}()
		if err != nil {
			return err
		}

		vz, err := func() (int16, error) {
			v, err := strconv.ParseInt(
				args[0], 10, 16,
			)
			if err != nil {
				return 0, fmt.Errorf(
					"argument %s must be int16 in form \"%s\" to summon mobile entity",
					"velocity-z",
					form,
				)
			}
			args = args[1:]

			return int16(v), nil
		}()
		if err != nil {
			return err
		}

		fmt.Printf(
			"{ id: %d, x: %f, y: %f, z: %f, yaw: %f, pitch: %f, hdPitch: %f, vx: %d, vy: %d, vz: %d }",
			id, x, y, z, yaw, pitch, hdPitch, vx, vy, vz,
		)

		switch id {
		default:
			return errors.New("it is unregistered mob id")
		case 1:
			form := "/smob ... <int16:id> <uint8:amount>"
			if len(args) != 2 {
				return fmt.Errorf(
					"it is insufficient length of arguments in form \"%s\" to summon item entity with no-nbt",
					form,
				)
			}

			id, err := func() (int16, error) {
				v, err := strconv.ParseInt(
					args[0], 10, 16,
				)
				if err != nil {
					return 0, fmt.Errorf(
						"argument %s must be int16 in form \"%s\" to summon item entity with no-nbt",
						"id",
						form,
					)
				}
				args = args[1:]

				return int16(v), nil
			}()
			if err != nil {
				return err
			}

			amt, err := func() (uint8, error) {
				v, err := strconv.ParseUint(
					args[0], 10, 8,
				)
				if err != nil {
					return 0, fmt.Errorf(
						"argument %s must be int16 in form \"%s\" to summon item entity with no-nbt",
						"amount",
						form,
					)
				}
				args = args[1:]

				return uint8(v), nil
			}()
			if err != nil {
				return err
			}

			var it item.Item
			switch id {
			case item.WoodenSwordItemID:
				it = item.NewWoodenSwordItem(
					amt, &nbt.ItemNbt{},
				)
				break
			case item.StickItemID:
				it = item.NewStickItem(
					amt, &nbt.ItemNbt{},
				)
				break
			}

			if it == nil {
				return errors.New("it is unregistered item id for summoning item entity with no-nbt")
			}

			fmt.Printf(
				"{ id: %d, amt: %d }",
				id, amt,
			)

			return errors.New("not implemented")

			break
		}

		break
	}

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

	e := NewClickWindowEvent(
		winID,
		slot,
		btn,
		act,
		mode,
	)
	dim.CHForCWEvent <- e

	return nil
}

func (dim *Dimension) UpdatePos(
	x, y, z float64,
	ground bool,
) error {
	dim.RLock()
	defer dim.RUnlock()

	if err := dim.world.UpdatePosForPlayer(
		dim.eid,
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

func (dim *Dimension) Close() (
	ChanForAddPlayerEvent,
	ChanForUpdateLatencyEvent,
	ChanForRemovePlayerEvent,
	ChanForSpawnPlayerEvent,
	ChanForSpawnItemStandEvent,
	ChanForSetEntityRelativeMoveEvent,
	ChanForSetEntityLookEvent,
	ChanForSetEntityActionsEvent,
	ChanForDespawnEntityEvent,
	ChanForLoadChunkEvent,
	ChanForUnloadChunkEvent,
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
		ChForSISEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUCEvent :=
		world.ClosePlayer(eid)

	return CHForAPEvent,
		CHForULEvent,
		CHForRPEvent,
		CHForSPEvent,
		ChForSISEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUCEvent,
		dim.CHForCWEvent
}
