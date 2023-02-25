package server

import (
	"errors"
	"sync"
)

type Game struct {
	sync.RWMutex

	greenRoom   *GreenRoom
	battlefield *Battlefield
}

func NewGame(
	rndDist int32,
	spawnX, spawnY, spawnZ float64,
	spawnYaw, spawnPitch float32,
) *Game {
	greenRoom := NewGreenRoom(
		rndDist,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	)
	battlefield := NewBattlefield(
		rndDist,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	)
	greenRoom.MakeFlat(GrassBlock)
	greenRoom.MakeFlat(GrassBlock)

	return &Game{
		greenRoom:   greenRoom,
		battlefield: battlefield,
	}
}

func (g *Game) GetGreenRoom() *GreenRoom {
	return g.greenRoom
}

type GameMgr struct {
	sync.RWMutex

	lobby *Lobby

	games   []*Game     // room by index
	indices map[EID]int // index by eid

	chansForChangeDimEvent map[EID]ChanForChangeDimEvent
}

func NewGameMgr(
	lobby *Lobby,
	games ...*Game,
) *GameMgr {
	return &GameMgr{
		lobby: lobby,

		games:   games,
		indices: make(map[EID]int),

		chansForChangeDimEvent: make(map[EID]ChanForChangeDimEvent),
	}
}

func (mgr *GameMgr) Init(
	eid EID,
	chanForChangeDimEvent ChanForChangeDimEvent,
) error {
	mgr.Lock()
	defer mgr.Unlock()

	mgr.chansForChangeDimEvent[eid] =
		chanForChangeDimEvent

	return nil
}

func (mgr *GameMgr) Join(
	player Player,
	index int,
) error {
	mgr.Lock()
	defer mgr.Unlock()

	eid := player.GetEID()

	if _, has := mgr.indices[eid]; has == true {
		return errors.New("player is already joined to room")
	}

	length := len(mgr.games)
	if length-1 < index {
		return errors.New("room for that index does not existed to join")
	}

	game := mgr.games[index]

	mgr.indices[eid] = index

	uid, username :=
		player.GetUID(),
		player.GetUsername()

	greenRoom := game.GetGreenRoom()
	warlord := NewWarlord(
		eid, uid, username,
	)
	changeDimEvent :=
		NewChangeDimEvent(
			greenRoom,
			warlord,
		)
	chanForChangeDimEvent :=
		mgr.chansForChangeDimEvent[eid]
	chanForChangeDimEvent <- changeDimEvent

	return nil
}

func (mgr *GameMgr) Leave(
	player Player,
) error {
	mgr.Lock()
	defer mgr.Unlock()

	eid := player.GetEID()

	if _, has := mgr.indices[eid]; has == false {
		return errors.New("room is not existed to leave")
	}

	delete(mgr.indices, eid)

	uid, username :=
		player.GetUID(),
		player.GetUsername()

	guest := NewGuest(
		eid, uid, username,
	)
	lobby := mgr.lobby
	changeDimEvent :=
		NewChangeDimEvent(
			lobby,
			guest,
		)
	chanForChangeDimEvent :=
		mgr.chansForChangeDimEvent[eid]
	chanForChangeDimEvent <- changeDimEvent

	return nil
}

func (mgr *GameMgr) Close(
	eid EID,
) ChanForChangeDimEvent {
	mgr.Lock()
	defer mgr.Unlock()

	chanForChangeDimEvent :=
		mgr.chansForChangeDimEvent[eid]

	delete(mgr.chansForChangeDimEvent, eid)

	if _, has := mgr.indices[eid]; has == true {
		delete(mgr.indices, eid)
	}

	return chanForChangeDimEvent
}
