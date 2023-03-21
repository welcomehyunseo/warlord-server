package server

import (
	"context"
	"fmt"
	"go/types"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/welcomehyunseo/warlord-server/server/item"
)

func init() {
}

func toChunkPosStr(
	cx, cz int32,
) string {
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
	l0 := []int{
		int(maxCx0), int(minCx0), int(maxCx1), int(minCx1),
	}
	l1 := []int{
		int(maxCz0), int(minCz0), int(maxCz1), int(minCz1),
	}
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
	InitItemStand(
		it item.Item,
	) error

	InitPlayer(
		eid int32,
		uid uuid.UUID, username string,
		CHForSPEvent ChanForSpawnPlayerEvent,
		CHForSISEvent ChanForSpawnItemStandEvent,
		CHForSERPEvent ChanForSetEntityRelativeMoveEvent,
		CHForSELEvent ChanForSetEntityLookEvent,
		CHForSEAEvent ChanForSetEntityActionsEvent,
		CHForDEEvent ChanForDespawnEntityEvent,
		CHForLCEvent ChanForLoadChunkEvent,
		CHForUnCEvent ChanForUnloadChunkEvent,
		CHForCWEvent ChanForClickWindowEvent,
		cnt *Client,
	) error

	UpdatePosForPlayer(
		eid int32,
		x, y, z float64,
		ground bool,
	) error

	UpdateLookForPlayer(
		eid int32,
		yaw, pitch float32,
		ground bool,
	) error

	FinishPlayer(
		eid int32,
		cnt *Client,
	) (
		ChanForSpawnPlayerEvent,
		ChanForSpawnItemStandEvent,
		ChanForSetEntityRelativeMoveEvent,
		ChanForSetEntityLookEvent,
		ChanForSetEntityActionsEvent,
		ChanForDespawnEntityEvent,
		ChanForLoadChunkEvent,
		ChanForUnloadChunkEvent,
		error,
	)

	ClosePlayer(
		eid int32,
	) (
		ChanForSpawnPlayerEvent,
		ChanForSpawnItemStandEvent,
		ChanForSetEntityRelativeMoveEvent,
		ChanForSetEntityLookEvent,
		ChanForSetEntityActionsEvent,
		ChanForDespawnEntityEvent,
		ChanForLoadChunkEvent,
		ChanForUnloadChunkEvent,
	)

	MakeFlat(
		block *Block,
	)
}

type overworld struct {
	rndDist                int32
	spawnX, spawnY, spawnZ float64
	spawnYaw, spawnPitch   float32

	mForCsByCPS *sync.RWMutex
	CsByCPS     map[string]*Chunk // chunk by pos string "x/z"

	mForPConns *sync.RWMutex
	PtoP       map[int32]map[int32]types.Nil // connections between players
	PtoIS      map[int32]map[int32]types.Nil // connections between player and item stand

	mForISConns *sync.RWMutex
	IStoP       map[int32]map[int32]types.Nil // connections between item stand and player
	IStoIS      map[int32]map[int32]types.Nil // connections between item stand and item stand

	mForPsByCPS *sync.RWMutex
	PsByCPS     map[string]map[int32]types.Nil // player eids by chunk pos string

	mForISsByCPS *sync.RWMutex
	ISsByCPS     map[string]map[int32]types.Nil

	Ps              map[int32]*Player                           // players by eid
	CHsForSPEvent   map[int32]ChanForSpawnPlayerEvent           // event channel by eid for client
	CHsForSISEvent  map[int32]ChanForSpawnItemStandEvent        // event channel by eid for client
	CHsForSELEvent  map[int32]ChanForSetEntityLookEvent         // event channel by eid for client
	CHsForSERMEvent map[int32]ChanForSetEntityRelativeMoveEvent // event channel by eid for client
	CHsForSEAEvent  map[int32]ChanForSetEntityActionsEvent      // event channel by eid for client
	CHsForDEEvent   map[int32]ChanForDespawnEntityEvent         // event channel by eid for client
	CHsForLCEvent   map[int32]ChanForLoadChunkEvent             // event channel by eid for client
	CHsForUCEvent   map[int32]ChanForUnloadChunkEvent           // event channel by eid for client

	ISs map[int32]*ItemStand // item stands by eid

	WGs      map[int32]*sync.WaitGroup    // by eid for handler
	cancelFs map[int32]context.CancelFunc // by eid for handler

}

func newOverworld(
	rndDist int32,
	spawnX, spawnY, spawnZ float64,
	spawnYaw, spawnPitch float32,
) *overworld {
	return &overworld{
		rndDist,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,

		new(sync.RWMutex),
		make(map[string]*Chunk),

		new(sync.RWMutex),
		make(map[int32]map[int32]types.Nil),
		make(map[int32]map[int32]types.Nil),

		new(sync.RWMutex),
		make(map[int32]map[int32]types.Nil),
		make(map[int32]map[int32]types.Nil),

		new(sync.RWMutex),
		make(map[string]map[int32]types.Nil),

		new(sync.RWMutex),
		make(map[string]map[int32]types.Nil),

		make(map[int32]*Player),
		make(map[int32]ChanForSpawnPlayerEvent),
		make(map[int32]ChanForSpawnItemStandEvent),
		make(map[int32]ChanForSetEntityLookEvent),
		make(map[int32]ChanForSetEntityRelativeMoveEvent),
		make(map[int32]ChanForSetEntityActionsEvent),
		make(map[int32]ChanForDespawnEntityEvent),
		make(map[int32]ChanForLoadChunkEvent),
		make(map[int32]ChanForUnloadChunkEvent),

		make(map[int32]*ItemStand),

		make(map[int32]*sync.WaitGroup),
		make(map[int32]context.CancelFunc),
	}
}

func (w *overworld) InitItemStand(
	it item.Item,
) error {
	dist := w.rndDist
	x, y, z :=
		w.spawnX, w.spawnY, w.spawnZ
	yaw, pitch :=
		w.spawnYaw, w.spawnPitch

	eid := GetEIDCounter().count()
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	e0 := NewItemStand(
		eid,
		uid,
		x, y, z,
		yaw, pitch,
		it,
	)
	w.ISs[eid] = e0

	SISEvent := NewSpawnItemStandEvent(
		eid,
		uid,
		x, y, z,
		yaw, pitch,
		it,
	)

	w.IStoP[eid] = make(map[int32]types.Nil)
	w.IStoIS[eid] = make(map[int32]types.Nil)

	cx, cz := toChunkPos(
		x, z,
	)
	maxCx, maxCz, minCx, minCz :=
		findRect(
			cx, cz, dist,
		)
	for cz := maxCz; cz >= minCz; cz-- {
		for cx := maxCx; cx >= minCx; cx-- {
			CPS := toChunkPosStr(cx, cz)

			if err := func() error {
				w.mForPsByCPS.RLock()
				defer w.mForPsByCPS.RUnlock()

				a, has := w.PsByCPS[CPS]
				if has == false {
					return nil
				}
				for eid1, _ := range a {
					w.CHsForSISEvent[eid1] <- SISEvent

					func() {
						w.mForISConns.Lock()
						defer w.mForISConns.Unlock()

						w.IStoP[eid][eid1] = types.Nil{}
					}()

					func() {
						w.mForPConns.Lock()
						defer w.mForPConns.Unlock()

						w.PtoIS[eid1][eid] = types.Nil{}
					}()

				}

				return nil
			}(); err != nil {
				return err
			}

			if err := func() error {
				w.mForISsByCPS.RLock()
				defer w.mForISsByCPS.RUnlock()

				a, has := w.ISsByCPS[CPS]
				if has == false {
					return nil
				}
				for eid1, _ := range a {

					func() {
						w.mForISConns.Lock()
						defer w.mForISConns.Unlock()

						w.IStoIS[eid][eid1] = types.Nil{}
						w.IStoIS[eid1][eid] = types.Nil{}
					}()

				}

				return nil
			}(); err != nil {
				return err
			}

		}
	}

	func() {
		w.mForISsByCPS.Lock()
		defer w.mForISsByCPS.Unlock()

		CPS := toChunkPosStr(cx, cz)
		a := w.ISsByCPS
		eids, has := a[CPS]
		if has == false {
			b := make(map[int32]types.Nil)
			a[CPS] = b
			eids = b
		}
		eids[eid] = types.Nil{}
	}()

	// wg := new(sync.WaitGroup)
	// ctx := context.Background()
	// ctx, cancel := context.WithCancel(ctx)

	// go w.handlePlayerLoop(
	// 	p0,
	// 	CHForCWEvent,
	// 	ctx,
	// 	wg,
	// )

	// w.WGs[eid] = wg
	// w.cancelFs[eid] = cancel

	return nil
}

func (w *overworld) handlePlayerLoop(
	p *Player,
	CHForCWEvent ChanForClickWindowEvent,
	ctx context.Context,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer wg.Done()

	var stop bool
	for {
		select {
		case <-time.After(time.Millisecond * 1):
			break
		case e := <-CHForCWEvent:
			if e.GetWindowID() == 0 {
				if err := p.ClickInventoryWindow(
					e.GetSlotNumber(),
					e.GetButtonEnum(),
					e.GetModeEnum(),
				); err != nil {
					// TODO: send error message to client
				}
			} else {

			}

			//fmt.Println(e)
			break
		case <-ctx.Done():
			stop = true
			break
		}

		if stop == true {
			break
		}
	}
}

func (w *overworld) InitPlayer(
	eid int32,
	uid uuid.UUID, username string,
	CHForSPEvent ChanForSpawnPlayerEvent,
	CHForSISEvent ChanForSpawnItemStandEvent,
	CHForSERPEvent ChanForSetEntityRelativeMoveEvent,
	CHForSELEvent ChanForSetEntityLookEvent,
	CHForSEAEvent ChanForSetEntityActionsEvent,
	CHForDEEvent ChanForDespawnEntityEvent,
	CHForLCEvent ChanForLoadChunkEvent,
	CHForUCEvent ChanForUnloadChunkEvent,
	CHForCWEvent ChanForClickWindowEvent,
	cnt *Client,
) error {
	dist := w.rndDist
	x, y, z :=
		w.spawnX, w.spawnY, w.spawnZ
	yaw, pitch :=
		w.spawnYaw, w.spawnPitch
	if err := cnt.Respawn(
		x, y, z,
		yaw, pitch,
	); err != nil {
		return err
	}

	e0 := NewPlayer(
		eid,
		uid, username,
		x, y, z,
		yaw, pitch,
	)
	w.Ps[eid] = e0
	w.CHsForSPEvent[eid] = CHForSPEvent
	w.CHsForSISEvent[eid] = CHForSISEvent
	w.CHsForSERMEvent[eid] = CHForSERPEvent
	w.CHsForSELEvent[eid] = CHForSELEvent
	w.CHsForSEAEvent[eid] = CHForSEAEvent
	w.CHsForDEEvent[eid] = CHForDEEvent
	w.CHsForLCEvent[eid] = CHForLCEvent
	w.CHsForUCEvent[eid] = CHForUCEvent

	SPEvent := NewSpawnPlayerEvent(
		eid,
		uid,
		x, y, z,
		yaw, pitch,
	)

	w.PtoP[eid] = make(map[int32]types.Nil)
	w.PtoIS[eid] = make(map[int32]types.Nil)

	cx, cz := toChunkPos(
		x, z,
	)
	maxCx, maxCz, minCx, minCz :=
		findRect(
			cx, cz, dist,
		)
	overworld, init := true, true
	for cz := maxCz; cz >= minCz; cz-- {
		for cx := maxCx; cx >= minCx; cx-- {
			CPS := toChunkPosStr(cx, cz)

			if err := func() error {
				w.mForCsByCPS.RLock()
				defer w.mForCsByCPS.RUnlock()

				chunk, has := w.CsByCPS[CPS]
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

				return nil
			}(); err != nil {
				return err
			}

			if err := func() error {
				w.mForPsByCPS.RLock()
				defer w.mForPsByCPS.RUnlock()

				a, has := w.PsByCPS[CPS]
				if has == false {
					return nil
				}
				for eid1, _ := range a {
					e1 := w.Ps[eid1]
					x, y, z := e1.GetPosition()
					yaw, pitch := e1.GetLook()
					if err := cnt.SpawnPlayer(
						e1.GetEID(),
						e1.GetUID(),
						x, y, z,
						yaw, pitch,
					); err != nil {
						return err
					}

					w.CHsForSPEvent[eid1] <- SPEvent

					func() {
						w.mForPConns.Lock()
						defer w.mForPConns.Unlock()

						w.PtoP[eid][eid1] = types.Nil{}
						w.PtoP[eid1][eid] = types.Nil{}
					}()

				}

				return nil
			}(); err != nil {
				return err
			}

			if err := func() error {
				w.mForISsByCPS.RLock()
				defer w.mForISsByCPS.RUnlock()

				a, has := w.ISsByCPS[CPS]
				if has == false {
					return nil
				}
				for eid1, _ := range a {
					e1 := w.ISs[eid1]
					x, y, z := e1.GetPosition()
					yaw, pitch := e1.GetLook()
					if err := cnt.SpawnItemStand(
						e1.GetEID(),
						e1.GetUID(),
						x, y, z,
						yaw, pitch,
						e1.GetItem(),
					); err != nil {
						return err
					}

					func() {
						w.mForPConns.Lock()
						defer w.mForPConns.Unlock()

						w.PtoIS[eid][eid1] = types.Nil{}
					}()

					func() {
						w.mForISConns.Lock()
						defer w.mForISConns.Unlock()

						w.IStoP[eid1][eid] = types.Nil{}
					}()

				}

				return nil
			}(); err != nil {
				return err
			}

		}
	}

	func() {
		w.mForPsByCPS.Lock()
		defer w.mForPsByCPS.Unlock()

		CPS := toChunkPosStr(cx, cz)
		a := w.PsByCPS
		eids, has := a[CPS]
		if has == false {
			b := make(map[int32]types.Nil)
			a[CPS] = b
			eids = b
		}
		eids[eid] = types.Nil{}
	}()

	wg := new(sync.WaitGroup)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	go w.handlePlayerLoop(
		e0,
		CHForCWEvent,
		ctx,
		wg,
	)

	w.WGs[eid] = wg
	w.cancelFs[eid] = cancel

	return nil
}

func (w *overworld) UpdatePosForPlayer(
	eid int32,
	x, y, z float64,
	ground bool,
) error {
	e0 := w.Ps[eid]
	prevX, prevY, prevZ :=
		e0.GetPosition()
	e0.UpdatePos(
		x, y, z,
		ground,
	)
	dx, dy, dz :=
		int16(((x*32)-(prevX*32))*128),
		int16(((y*32)-(prevY*32))*128),
		int16(((z*32)-(prevZ*32))*128)
	if dx == 0 && dy == 0 && dz == 0 {
		return nil
	}

	SERPEvent :=
		NewSetEntityRelativeMoveEvent(
			eid,
			dx, dy, dz,
			ground,
		)
	func() {
		w.mForPConns.RLock()
		defer w.mForPConns.RUnlock()

		for eid1, _ := range w.PtoP[eid] {
			w.CHsForSERMEvent[eid1] <- SERPEvent
		}
	}()

	currCx, currCz := toChunkPos(
		x, z,
	)
	prevCx, prevCz := toChunkPos(
		prevX, prevZ,
	)
	if currCx == prevCx && currCz == prevCz {
		return nil
	}

	dist := w.rndDist
	CHForSPEvent := w.CHsForSPEvent[eid]
	CHForSISEvent := w.CHsForSISEvent[eid]
	CHForDEEvent := w.CHsForDEEvent[eid]
	CHForLCEvent := w.CHsForLCEvent[eid]
	CHForUCEvent := w.CHsForUCEvent[eid]

	SPEvent :=
		NewSpawnPlayerEvent(
			eid, e0.GetUID(),
			e0.GetX(), e0.GetY(), e0.GetZ(),
			e0.GetYaw(), e0.GetPitch(),
		)
	DEEvent :=
		NewDespawnEntityEvent(
			eid,
		)

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

			CPS := toChunkPosStr(cx, cz)

			if err := func() error {
				w.mForCsByCPS.RLock()
				defer w.mForCsByCPS.RUnlock()

				chunk, has := w.CsByCPS[CPS]
				if has == false {
					chunk = NewChunk()
				}
				LCEvent :=
					NewLoadChunkEvent(
						overworld, init,
						cx, cz,
						chunk,
					)
				CHForLCEvent <- LCEvent

				return nil
			}(); err != nil {
				return err
			}

			if err := func() error {
				w.mForPsByCPS.RLock()
				defer w.mForPsByCPS.RUnlock()

				a, has := w.PsByCPS[CPS]
				if has == false {
					return nil
				}
				for eid1, _ := range a {
					e1 := w.Ps[eid1]
					SPEvent1 := NewSpawnPlayerEvent(
						e1.GetEID(), e1.GetUID(),
						e1.GetX(), e1.GetY(), e1.GetZ(),
						e1.GetYaw(), e1.GetPitch(),
					)
					CHForSPEvent <- SPEvent1

					w.CHsForSPEvent[eid1] <- SPEvent

					func() {
						w.mForPConns.Lock()
						defer w.mForPConns.Unlock()

						w.PtoP[eid][eid1] = types.Nil{}
						w.PtoP[eid1][eid] = types.Nil{}
					}()
				}

				return nil
			}(); err != nil {
				return err
			}

			if err := func() error {
				w.mForISsByCPS.RLock()
				defer w.mForISsByCPS.RUnlock()

				a, has := w.ISsByCPS[CPS]
				if has == false {
					return nil
				}
				for eid1, _ := range a {
					e1 := w.ISs[eid1]
					x, y, z := e1.GetPosition()
					yaw, pitch := e1.GetLook()
					SISEvent1 := NewSpawnItemStandEvent(
						e1.GetEID(),
						e1.GetUID(),
						x, y, z,
						yaw, pitch,
						e1.GetItem(),
					)
					CHForSISEvent <- SISEvent1

					func() {
						w.mForPConns.Lock()
						defer w.mForPConns.Unlock()

						w.PtoIS[eid][eid1] = types.Nil{}
					}()

					func() {
						w.mForISConns.Lock()
						defer w.mForISConns.Unlock()

						w.IStoP[eid1][eid] = types.Nil{}
					}()

				}

				return nil
			}(); err != nil {
				return err
			}

		}
	}

	for cz := maxPrevCz; cz >= minPrevCz; cz-- {
		for cx := maxPrevCx; cx >= minPrevCx; cx-- {
			if minSubCx <= cx && cx <= maxSubCx &&
				minSubCz <= cz && cz <= maxSubCz {
				continue
			}

			CPS := toChunkPosStr(cx, cz)

			UnCEvent := NewUnloadChunkEvent(
				cx, cz,
			)
			CHForUCEvent <- UnCEvent

			if err := func() error {
				w.mForPsByCPS.RLock()
				defer w.mForPsByCPS.RUnlock()

				a, has := w.PsByCPS[CPS]
				if has == false {
					return nil
				}
				for eid1, _ := range a {
					DEEvent1 :=
						NewDespawnEntityEvent(
							eid1,
						)
					CHForDEEvent <- DEEvent1

					w.CHsForDEEvent[eid1] <- DEEvent

					func() {
						w.mForPConns.Lock()
						defer w.mForPConns.Unlock()

						delete(w.PtoP[eid], eid1)
						delete(w.PtoP[eid1], eid)
					}()
				}

				return nil
			}(); err != nil {
				return err
			}

			if err := func() error {
				w.mForISsByCPS.RLock()
				defer w.mForISsByCPS.RUnlock()

				a, has := w.ISsByCPS[CPS]
				if has == false {
					return nil
				}
				for eid1, _ := range a {
					DEEvent1 :=
						NewDespawnEntityEvent(
							eid1,
						)
					CHForDEEvent <- DEEvent1

					func() {
						w.mForPConns.Lock()
						defer w.mForPConns.Unlock()

						delete(w.PtoIS[eid], eid1)
					}()

					func() {
						w.mForISConns.Lock()
						defer w.mForISConns.Unlock()

						delete(w.IStoP[eid1], eid)
					}()

				}

				return nil
			}(); err != nil {
				return err
			}

		}
	}

	func() {
		a := w.PsByCPS

		prevCPS := toChunkPosStr(
			prevCx, prevCz,
		)
		delete(a[prevCPS], eid)

		currCPS := toChunkPosStr(
			currCx, currCz,
		)
		eids, has := a[currCPS]
		if has == false {
			b := make(map[int32]types.Nil)
			a[currCPS] = b
			eids = b
		}
		eids[eid] = types.Nil{}
	}()

	return nil
}

func (w *overworld) UpdateLookForPlayer(
	eid int32,
	yaw, pitch float32,
	ground bool,
) error {
	p0 := w.Ps[eid]
	p0.UpdateLook(
		yaw, pitch,
		ground,
	)

	SELEvent :=
		NewSetEntityLookEvent(
			eid,
			yaw, pitch,
			ground,
		)
	func() {
		w.mForPConns.RLock()
		defer w.mForPConns.RUnlock()

		for eid1, _ := range w.PtoP[eid] {
			w.CHsForSELEvent[eid1] <- SELEvent
		}
	}()

	return nil
}

func (w *overworld) closePlayer(
	eid int32,
) (
	ChanForSpawnPlayerEvent,
	ChanForSpawnItemStandEvent,
	ChanForSetEntityRelativeMoveEvent,
	ChanForSetEntityLookEvent,
	ChanForSetEntityActionsEvent,
	ChanForDespawnEntityEvent,
	ChanForLoadChunkEvent,
	ChanForUnloadChunkEvent,
) {
	p0 := w.Ps[eid]
	CHForSPEvent := w.CHsForSPEvent[eid]
	CHForSISEvent := w.CHsForSISEvent[eid]
	CHForSERPEvent := w.CHsForSERMEvent[eid]
	CHForSELEvent := w.CHsForSELEvent[eid]
	CHForSEAEvent := w.CHsForSEAEvent[eid]
	CHForDEEvent := w.CHsForDEEvent[eid]
	CHForLCEvent := w.CHsForLCEvent[eid]
	CHForUCEvent := w.CHsForUCEvent[eid]

	w.cancelFs[eid]()
	w.WGs[eid].Wait()

	DEEvent :=
		NewDespawnEntityEvent(
			eid,
		)
	func() {
		w.mForPConns.Lock()
		defer w.mForPConns.Unlock()
		w.mForISConns.Lock()
		defer w.mForISConns.Unlock()

		for eid1, _ := range w.PtoP[eid] {
			w.CHsForDEEvent[eid1] <- DEEvent

			delete(w.PtoP[eid1], eid)
		}
		delete(w.PtoP, eid)

		for eid1, _ := range w.PtoIS[eid] {
			delete(w.IStoP[eid1], eid)
		}
		delete(w.PtoIS, eid)
	}()

	cx, cz := toChunkPos(
		p0.GetX(), p0.GetZ(),
	)
	CPS := toChunkPosStr(
		cx, cz,
	)
	func() {
		w.mForPsByCPS.Lock()
		defer w.mForPsByCPS.Unlock()

		delete(w.PsByCPS[CPS], eid)
	}()

	delete(w.Ps, eid)
	delete(w.CHsForSPEvent, eid)
	delete(w.CHsForSISEvent, eid)
	delete(w.CHsForSERMEvent, eid)
	delete(w.CHsForSELEvent, eid)
	delete(w.CHsForSEAEvent, eid)
	delete(w.CHsForDEEvent, eid)
	delete(w.CHsForLCEvent, eid)
	delete(w.CHsForUCEvent, eid)

	delete(w.WGs, eid)
	delete(w.cancelFs, eid)

	return CHForSPEvent,
		CHForSISEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUCEvent
}

func (w *overworld) FinishPlayer(
	eid int32,
	cnt *Client,
) (
	ChanForSpawnPlayerEvent,
	ChanForSpawnItemStandEvent,
	ChanForSetEntityRelativeMoveEvent,
	ChanForSetEntityLookEvent,
	ChanForSetEntityActionsEvent,
	ChanForDespawnEntityEvent,
	ChanForLoadChunkEvent,
	ChanForUnloadChunkEvent,
	error,
) {
	if err := func() error {
		w.mForPConns.RLock()
		defer w.mForPConns.RUnlock()
		w.mForISConns.RLock()
		defer w.mForISConns.RUnlock()

		for eid1, _ := range w.PtoP[eid] {
			if err := cnt.DespawnEntity(
				eid1,
			); err != nil {
				return err
			}
		}

		for eid1, _ := range w.PtoIS[eid] {
			if err := cnt.DespawnEntity(
				eid1,
			); err != nil {
				return err
			}
		}

		return nil
	}(); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	CHForSPEvent,
		CHForSISEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUCEvent :=
		w.closePlayer(eid)

	return CHForSPEvent,
		CHForSISEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUCEvent,
		nil
}

func (w *overworld) ClosePlayer(
	eid int32,
) (
	ChanForSpawnPlayerEvent,
	ChanForSpawnItemStandEvent,
	ChanForSetEntityRelativeMoveEvent,
	ChanForSetEntityLookEvent,
	ChanForSetEntityActionsEvent,
	ChanForDespawnEntityEvent,
	ChanForLoadChunkEvent,
	ChanForUnloadChunkEvent,
) {
	return w.closePlayer(eid)
}

func (w *overworld) MakeFlat(
	block *Block,
) {
	w.mForCsByCPS.Lock()
	defer w.mForCsByCPS.Unlock()

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
			CPS := toChunkPosStr(cx, cz)
			w.CsByCPS[CPS] = chunk
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
