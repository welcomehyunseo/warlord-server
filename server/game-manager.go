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

type GameManager struct {
	sync.RWMutex

	lobby *Lobby

	games   []*Game     // room by index
	indices map[EID]int // index by eid

	chansForChangeDimEvent map[EID]ChanForChangeDimEvent
}

func NewGameManager(
	lobby *Lobby,
	games ...*Game,
) *GameManager {
	return &GameManager{
		lobby: lobby,

		games:   games,
		indices: make(map[EID]int),

		chansForChangeDimEvent: make(map[EID]ChanForChangeDimEvent),
	}
}

func (m *GameManager) InitPlayer(
	eid EID,
	chanForChangeDimEvent ChanForChangeDimEvent,
) error {
	m.Lock()
	defer m.Unlock()

	m.chansForChangeDimEvent[eid] =
		chanForChangeDimEvent

	return nil
}

func (m *GameManager) Join(
	player Player,
	index int,
) error {
	m.Lock()
	defer m.Unlock()

	eid := player.GetEID()

	if _, has := m.indices[eid]; has == true {
		return errors.New("player is already joined to room")
	}

	length := len(m.games)
	if length-1 < index {
		return errors.New("room for that index does not existed to join")
	}

	game := m.games[index]

	m.indices[eid] = index

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
		m.chansForChangeDimEvent[eid]
	chanForChangeDimEvent <- changeDimEvent

	return nil
}

func (m *GameManager) Leave(
	player Player,
) error {
	m.Lock()
	defer m.Unlock()

	eid := player.GetEID()

	if _, has := m.indices[eid]; has == false {
		return errors.New("room is not existed to leave")
	}

	delete(m.indices, eid)

	uid, username :=
		player.GetUID(),
		player.GetUsername()

	guest := NewGuest(
		eid, uid, username,
	)
	lobby := m.lobby
	changeDimEvent :=
		NewChangeDimEvent(
			lobby,
			guest,
		)
	chanForChangeDimEvent :=
		m.chansForChangeDimEvent[eid]
	chanForChangeDimEvent <- changeDimEvent

	return nil
}

func (m *GameManager) ClosePlayer(
	eid EID,
) ChanForChangeDimEvent {
	m.Lock()
	defer m.Unlock()

	chanForChangeDimEvent :=
		m.chansForChangeDimEvent[eid]

	delete(m.chansForChangeDimEvent, eid)

	if _, has := m.indices[eid]; has == true {
		delete(m.indices, eid)
	}

	return chanForChangeDimEvent
}
