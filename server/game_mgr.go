package server

//
//import (
//	"errors"
//	"sync"
//)
//
//type Game struct {
//	sync.RWMutex
//
//	warlords map[EID]*Warlord
//
//	greenRoom   *GreenRoom
//	battlefield *Battlefield
//
//	chansForChangeDimEvent map[EID]ChanForChangeDimEvent
//}
//
//func NewGame(
//	rndDist int32,
//	spawnX, spawnY, spawnZ float64,
//	spawnYaw, spawnPitch float32,
//) *Game {
//	greenRoom := NewGreenRoom(
//		rndDist,
//		spawnX, spawnY, spawnZ,
//		spawnYaw, spawnPitch,
//	)
//	battlefield := NewBattlefield(
//		rndDist,
//		spawnX, spawnY, spawnZ,
//		spawnYaw, spawnPitch,
//	)
//	greenRoom.MakeFlat(GrassBlock)
//	battlefield.MakeFlat(GrassBlock)
//
//	return &Game{
//		warlords: make(map[EID]*Warlord),
//
//		greenRoom:   greenRoom,
//		battlefield: battlefield,
//
//		chansForChangeDimEvent: make(map[EID]ChanForChangeDimEvent),
//	}
//}
//
//func (g *Game) Join(
//	player Player,
//	chanForChangeDimEvent ChanForChangeDimEvent,
//) error {
//	g.Lock()
//	defer g.Unlock()
//
//	eid := player.GetEID()
//	uid, username :=
//		player.GetUID(),
//		player.GetUsername()
//	warlord := NewWarlord(
//		eid, uid, username,
//	)
//	g.warlords[eid] = warlord
//
//	greenRoom := g.greenRoom
//	changeDimEvent :=
//		NewChangeDimEvent(
//			greenRoom,
//			warlord,
//		)
//	g.chansForChangeDimEvent[eid] = chanForChangeDimEvent
//	chanForChangeDimEvent <- changeDimEvent
//
//	return nil
//}
//
//func (g *Game) Leave(
//	lobby *Lobby,
//	player Player,
//) {
//	g.Lock()
//	defer g.Unlock()
//
//	eid := player.GetEID()
//	uid, username :=
//		player.GetUID(),
//		player.GetUsername()
//	delete(g.warlords, eid)
//	guest := NewGuest(
//		eid, uid, username,
//	)
//
//	chanForChangeDimEvent :=
//		g.chansForChangeDimEvent[eid]
//	delete(g.chansForChangeDimEvent, eid)
//	changeDimEvent :=
//		NewChangeDimEvent(
//			lobby,
//			guest,
//		)
//	chanForChangeDimEvent <- changeDimEvent
//}
//
//type GameMgr struct {
//	sync.RWMutex
//
//	lobby *Lobby
//
//	games   []*Game     // room by index
//	indices map[EID]int // index by eid
//
//	chansForChangeDimEvent map[EID]ChanForChangeDimEvent
//}
//
//func NewGameMgr(
//	lobby *Lobby,
//	games ...*Game,
//) *GameMgr {
//	return &GameMgr{
//		lobby: lobby,
//
//		games:   games,
//		indices: make(map[EID]int),
//
//		chansForChangeDimEvent: make(map[EID]ChanForChangeDimEvent),
//	}
//}
//
//func (mgr *GameMgr) Init(
//	eid EID,
//	chanForChangeDimEvent ChanForChangeDimEvent,
//) error {
//	mgr.Lock()
//	defer mgr.Unlock()
//
//	mgr.chansForChangeDimEvent[eid] =
//		chanForChangeDimEvent
//
//	return nil
//}
//
//func (mgr *GameMgr) Join(
//	player Player,
//	index int,
//) error {
//	mgr.Lock()
//	defer mgr.Unlock()
//
//	eid := player.GetEID()
//
//	if _, has := mgr.indices[eid]; has == true {
//		return errors.New("player is already joined to room")
//	}
//
//	length := len(mgr.games)
//	if length-1 < index {
//		return errors.New("room for that index does not existed to join")
//	}
//
//	game := mgr.games[index]
//
//	mgr.indices[eid] = index
//
//	chanForChangeDimEvent := mgr.chansForChangeDimEvent[eid]
//	if err := game.Join(
//		player,
//		chanForChangeDimEvent,
//	); err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (mgr *GameMgr) Leave(
//	player Player,
//) error {
//	mgr.Lock()
//	defer mgr.Unlock()
//
//	eid := player.GetEID()
//
//	index, has := mgr.indices[eid]
//	if has == false {
//		return errors.New("room is not existed to leave")
//	}
//
//	delete(mgr.indices, eid)
//
//	lobby := mgr.lobby
//	game := mgr.games[index]
//	game.Leave(
//		lobby,
//		player,
//	)
//
//	return nil
//}
//
//func (mgr *GameMgr) Close(
//	player Player,
//) ChanForChangeDimEvent {
//	mgr.Lock()
//	defer mgr.Unlock()
//
//	eid := player.GetEID()
//
//	chanForChangeDimEvent :=
//		mgr.chansForChangeDimEvent[eid]
//
//	delete(mgr.chansForChangeDimEvent, eid)
//
//	if index, has := mgr.indices[eid]; has == true {
//		delete(mgr.indices, eid)
//
//		lobby := mgr.lobby
//		game := mgr.games[index]
//		game.Leave(
//			lobby,
//			player,
//		)
//	}
//
//	return chanForChangeDimEvent
//}
