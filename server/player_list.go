package server

import (
	"sync"
)

type PlayerList struct {
	sync.RWMutex

	players                    map[EID]*Player
	chansForAddPlayerEvent     map[EID]ChanForAddPlayerEvent
	chansForUpdateLatencyEvent map[EID]ChanForUpdateLatencyEvent
	chansForRemovePlayerEvent  map[EID]ChanForRemovePlayerEvent
}

func NewPlayerList() *PlayerList {
	return &PlayerList{
		players:                    make(map[EID]*Player),
		chansForAddPlayerEvent:     make(map[EID]ChanForAddPlayerEvent),
		chansForUpdateLatencyEvent: make(map[EID]ChanForUpdateLatencyEvent),
		chansForRemovePlayerEvent:  make(map[EID]ChanForRemovePlayerEvent),
	}
}

func (l *PlayerList) InitPlayer(
	lg *Logger,
	player *Player,
	cnt *Client,
	chanForAddPlayerEvent ChanForAddPlayerEvent,
	chanForUpdateLatencyEvent ChanForUpdateLatencyEvent,
	chanForRemovePlayerEvent ChanForRemovePlayerEvent,
) error {
	l.Lock()
	defer l.Unlock()

	lg.Debug(
		"it is started to init player in PlayerList",
		NewLgElement("player", player),
		NewLgElement("cnt", cnt),
	)
	defer func() {
		lg.Debug("it is finished to init player in PlayerList")
	}()

	uid, username :=
		player.GetUid(), player.GetUsername()
	for eid1, player1 := range l.players {
		uid1, username1 :=
			player1.GetUid(),
			player1.GetUsername()
		if err := cnt.AddPlayer(
			lg, uid1, username1,
		); err != nil {
			return err
		}

		event := NewAddPlayerEvent(
			uid, username,
		)
		chanForEvent := l.chansForAddPlayerEvent[eid1]
		chanForEvent <- event
		event.Wait()
	}

	if err := cnt.AddPlayer(
		lg, uid, username,
	); err != nil {
		return err
	}

	eid := player.GetEid()
	l.players[eid] = player
	l.chansForAddPlayerEvent[eid] = chanForAddPlayerEvent
	l.chansForUpdateLatencyEvent[eid] = chanForUpdateLatencyEvent
	l.chansForRemovePlayerEvent[eid] = chanForRemovePlayerEvent

	return nil
}

func (l *PlayerList) UpdateLatency(
	lg *Logger,
	player *Player,
	latency int32,
	cnt *Client,
) error {
	l.RLock()
	defer l.RUnlock()

	lg.Debug(
		"it is started to update latency in PlayerList",
		NewLgElement("player", player),
		NewLgElement("latency", latency),
		NewLgElement("cnt", cnt),
	)
	defer func() {
		lg.Debug("it is finished to update latency in PlayerList")
	}()

	eid := player.GetEid()
	uid := player.GetUid()
	event := NewUpdateLatencyEvent(
		uid, latency,
	)
	for eid1, _ := range l.players {
		if eid == eid1 {
			continue
		}
		chanForEvent := l.chansForUpdateLatencyEvent[eid1]
		chanForEvent <- event
	}

	if err := cnt.UpdateLatency(
		lg,
		uid, latency,
	); err != nil {
		return err
	}

	return nil
}

func (l *PlayerList) ClosePlayer(
	lg *Logger,
	player *Player,
) (
	ChanForAddPlayerEvent,
	ChanForUpdateLatencyEvent,
	ChanForRemovePlayerEvent,
) {
	l.Lock()
	defer l.Unlock()

	lg.Debug(
		"it is started to close player in PlayerList",
		NewLgElement("player", player),
	)
	defer func() {
		lg.Debug("it is finished to close player in PlayerList")
	}()

	eid := player.GetEid()
	chanForAddPlayerEvent := l.chansForAddPlayerEvent[eid]
	chanForUpdateLatencyEvent := l.chansForUpdateLatencyEvent[eid]
	chanForRemovePlayerEvent := l.chansForRemovePlayerEvent[eid]

	delete(l.chansForAddPlayerEvent, eid)
	delete(l.chansForUpdateLatencyEvent, eid)
	delete(l.chansForRemovePlayerEvent, eid)
	delete(l.players, eid)

	uid := player.GetUid()
	event := NewRemovePlayerEvent(
		uid,
	)
	for eid1, _ := range l.players {
		chanForEvent := l.chansForRemovePlayerEvent[eid1]
		chanForEvent <- event
	}

	return chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent
}
