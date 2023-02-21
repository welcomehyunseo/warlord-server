package server

import (
	"sync"
)

type PlayerListItem struct {
	uid      UID
	username string
}

func NewPlayerListItem(
	uid UID,
	username string,
) *PlayerListItem {
	return &PlayerListItem{
		uid,
		username,
	}
}

func (i *PlayerListItem) GetUID() UID {
	return i.uid
}

func (i *PlayerListItem) GetUsername() string {
	return i.username
}

type PlayerList struct {
	sync.RWMutex

	items                      map[UID]*PlayerListItem
	chansForAddPlayerEvent     map[UID]ChanForAddPlayerEvent
	chansForUpdateLatencyEvent map[UID]ChanForUpdateLatencyEvent
	chansForRemovePlayerEvent  map[UID]ChanForRemovePlayerEvent
}

func NewPlayerList() *PlayerList {
	return &PlayerList{
		items:                      make(map[UID]*PlayerListItem),
		chansForAddPlayerEvent:     make(map[UID]ChanForAddPlayerEvent),
		chansForUpdateLatencyEvent: make(map[UID]ChanForUpdateLatencyEvent),
		chansForRemovePlayerEvent:  make(map[UID]ChanForRemovePlayerEvent),
	}
}

func (l *PlayerList) InitPlayer(
	lg *Logger,
	uid UID, username string,
	cnt *Client,
	chanForAddPlayerEvent ChanForAddPlayerEvent,
	chanForUpdateLatencyEvent ChanForUpdateLatencyEvent,
	chanForRemovePlayerEvent ChanForRemovePlayerEvent,
) error {
	l.Lock()
	defer l.Unlock()

	lg.Debug(
		"it is started to init player in PlayerList",
		NewLgElement("uid", uid),
		NewLgElement("username", username),
		NewLgElement("cnt", cnt),
	)
	defer func() {
		lg.Debug("it is finished to init player in PlayerList")
	}()

	for eid1, item := range l.items {
		uid1, username1 :=
			item.GetUID(),
			item.GetUsername()
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

	item := NewPlayerListItem(
		uid,
		username,
	)
	l.items[uid] = item
	l.chansForAddPlayerEvent[uid] = chanForAddPlayerEvent
	l.chansForUpdateLatencyEvent[uid] = chanForUpdateLatencyEvent
	l.chansForRemovePlayerEvent[uid] = chanForRemovePlayerEvent

	return nil
}

func (l *PlayerList) UpdateLatency(
	lg *Logger,
	uid UID,
	latency int32,
	cnt *Client,
) error {
	l.RLock()
	defer l.RUnlock()

	lg.Debug(
		"it is started to update latency in PlayerList",
		NewLgElement("uid", uid),
		NewLgElement("latency", latency),
	)
	defer func() {
		lg.Debug("it is finished to update latency in PlayerList")
	}()

	event := NewUpdateLatencyEvent(
		uid, latency,
	)
	for uid1, _ := range l.items {
		if uid == uid1 {
			continue
		}
		chanForEvent := l.chansForUpdateLatencyEvent[uid1]
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
	uid UID,
) (
	ChanForAddPlayerEvent,
	ChanForUpdateLatencyEvent,
	ChanForRemovePlayerEvent,
) {
	l.Lock()
	defer l.Unlock()

	lg.Debug(
		"it is started to close player in PlayerList",
		NewLgElement("uid", uid),
	)
	defer func() {
		lg.Debug("it is finished to close player in PlayerList")
	}()

	chanForAddPlayerEvent := l.chansForAddPlayerEvent[uid]
	chanForUpdateLatencyEvent := l.chansForUpdateLatencyEvent[uid]
	chanForRemovePlayerEvent := l.chansForRemovePlayerEvent[uid]

	delete(l.chansForAddPlayerEvent, uid)
	delete(l.chansForUpdateLatencyEvent, uid)
	delete(l.chansForRemovePlayerEvent, uid)
	delete(l.items, uid)

	event := NewRemovePlayerEvent(
		uid,
	)
	for uid1, _ := range l.items {
		chanForEvent := l.chansForRemovePlayerEvent[uid1]
		chanForEvent <- event
	}

	return chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent
}
