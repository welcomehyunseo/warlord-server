package server

import "sync"

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

func (pli *PlayerListItem) GetUID() UID {
	return pli.uid
}

func (pli *PlayerListItem) GetUsername() string {
	return pli.username
}

type PlayerList struct {
	*sync.RWMutex

	items map[EID]*PlayerListItem

	chansForAPEvent map[EID]ChanForAddPlayerEvent
	chansForULEvent map[EID]ChanForUpdateLatencyEvent
	chansForRPEvent map[EID]ChanForRemovePlayerEvent
}

func NewPlayerList() *PlayerList {
	return &PlayerList{
		new(sync.RWMutex),

		make(map[EID]*PlayerListItem),

		make(map[EID]ChanForAddPlayerEvent),
		make(map[EID]ChanForUpdateLatencyEvent),
		make(map[EID]ChanForRemovePlayerEvent),
	}
}

func (pl *PlayerList) Init(
	eid EID,
	uid UID, username string,
	chanForAPEvent ChanForAddPlayerEvent,
	chanForULEvent ChanForUpdateLatencyEvent,
	chanForRPEvent ChanForRemovePlayerEvent,
	cnt *Client,
) error {
	pl.Lock()
	defer pl.Unlock()

	item := NewPlayerListItem(
		uid, username,
	)

	for eid1, item1 := range pl.items {
		uid1, username1 :=
			item1.GetUID(),
			item1.GetUsername()
		if err := cnt.AddPlayer(
			uid1, username1,
		); err != nil {
			return err
		}

		APEvent := NewAddPlayerEvent(
			uid, username,
		)
		pl.chansForAPEvent[eid1] <- APEvent
		APEvent.Wait()
	}
	if err := cnt.AddPlayer(
		uid, username,
	); err != nil {
		return err
	}

	pl.items[eid] = item
	pl.chansForAPEvent[eid] = chanForAPEvent
	pl.chansForULEvent[eid] = chanForULEvent
	pl.chansForRPEvent[eid] = chanForRPEvent

	return nil
}

func (pl *PlayerList) UpdateLatency(
	eid EID,
	latency int32,
) error {
	pl.RLock()
	defer pl.RUnlock()

	item := pl.items[eid]
	uid := item.GetUID()
	ULEvent := NewUpdateLatencyEvent(
		uid, latency,
	)
	for eid, _ := range pl.items {
		pl.chansForULEvent[eid] <- ULEvent
	}

	return nil
}

func (pl *PlayerList) Finish(
	eid EID,
	cnt *Client,
) (
	ChanForAddPlayerEvent,
	ChanForUpdateLatencyEvent,
	ChanForRemovePlayerEvent,
	error,
) {
	pl.Lock()
	defer pl.Unlock()

	item := pl.items[eid]
	chanForAPEvent := pl.chansForAPEvent[eid]
	chanForULEvent := pl.chansForULEvent[eid]
	chanForRPEvent := pl.chansForRPEvent[eid]

	uid := item.GetUID()
	for eid1, item1 := range pl.items {
		uid1 := item1.GetUID()
		RPEvent1 :=
			NewRemovePlayerEvent(
				uid1,
			)
		chanForRPEvent <- RPEvent1
		RPEvent1.Wait()

		RPEvent :=
			NewRemovePlayerEvent(
				uid,
			)

		pl.chansForRPEvent[eid1] <- RPEvent
		RPEvent.Wait()
	}
	if err := cnt.RemovePlayer(
		uid,
	); err != nil {
		return nil, nil, nil, err
	}

	delete(pl.items, eid)
	delete(pl.chansForAPEvent, eid)
	delete(pl.chansForULEvent, eid)
	delete(pl.chansForRPEvent, eid)

	return chanForAPEvent,
		chanForULEvent,
		chanForRPEvent,
		nil
}

func (pl *PlayerList) Close(
	eid EID,
) (
	ChanForAddPlayerEvent,
	ChanForUpdateLatencyEvent,
	ChanForRemovePlayerEvent,
) {
	item := pl.items[eid]
	chanForAPEvent := pl.chansForAPEvent[eid]
	chanForULEvent := pl.chansForULEvent[eid]
	chanForRPEvent := pl.chansForRPEvent[eid]

	uid := item.GetUID()
	for eid1, _ := range pl.items {
		RPEvent := NewRemovePlayerEvent(
			uid,
		)
		pl.chansForRPEvent[eid1] <- RPEvent
	}

	delete(pl.items, eid)
	delete(pl.chansForAPEvent, eid)
	delete(pl.chansForULEvent, eid)
	delete(pl.chansForRPEvent, eid)

	return chanForAPEvent,
		chanForULEvent,
		chanForRPEvent
}
