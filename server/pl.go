package server

import (
	"github.com/google/uuid"
	"sync"
)

type PlayerListItem struct {
	uid      uuid.UUID
	username string
}

func NewPlayerListItem(
	uid uuid.UUID,
	username string,
) *PlayerListItem {
	return &PlayerListItem{
		uid,
		username,
	}
}

func (pli *PlayerListItem) GetUID() uuid.UUID {
	return pli.uid
}

func (pli *PlayerListItem) GetUsername() string {
	return pli.username
}

type PlayerList struct {
	*sync.RWMutex

	items map[int32]*PlayerListItem

	CHsForAPEvent map[int32]ChanForAddPlayerEvent
	CHsForULEvent map[int32]ChanForUpdateLatencyEvent
	CHsForRPEvent map[int32]ChanForRemovePlayerEvent
}

func NewPlayerList() *PlayerList {
	return &PlayerList{
		new(sync.RWMutex),

		make(map[int32]*PlayerListItem),

		make(map[int32]ChanForAddPlayerEvent),
		make(map[int32]ChanForUpdateLatencyEvent),
		make(map[int32]ChanForRemovePlayerEvent),
	}
}

func (pl *PlayerList) Init(
	eid int32,
	uid uuid.UUID, username string,
	CHForAPEvent ChanForAddPlayerEvent,
	CHForULEvent ChanForUpdateLatencyEvent,
	CHForRPEvent ChanForRemovePlayerEvent,
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
		pl.CHsForAPEvent[eid1] <- APEvent
		APEvent.Wait()
	}
	if err := cnt.AddPlayer(
		uid, username,
	); err != nil {
		return err
	}

	pl.items[eid] = item
	pl.CHsForAPEvent[eid] = CHForAPEvent
	pl.CHsForULEvent[eid] = CHForULEvent
	pl.CHsForRPEvent[eid] = CHForRPEvent

	return nil
}

func (pl *PlayerList) UpdateLatency(
	id int32,
	ms int32,
) error {
	pl.RLock()
	defer pl.RUnlock()

	item := pl.items[id]

	uid := item.GetUID()
	ULEvent := NewUpdateLatencyEvent(
		uid, ms,
	)
	for id, _ := range pl.items {
		pl.CHsForULEvent[id] <- ULEvent
	}

	return nil
}

func (pl *PlayerList) close(
	eid int32,
) (
	ChanForAddPlayerEvent,
	ChanForUpdateLatencyEvent,
	ChanForRemovePlayerEvent,
) {
	item := pl.items[eid]

	CHForAPEvent := pl.CHsForAPEvent[eid]
	CHForULEvent := pl.CHsForULEvent[eid]
	CHForRPEvent := pl.CHsForRPEvent[eid]

	uid := item.GetUID()
	for eid1, _ := range pl.items {
		if eid1 == eid {
			continue
		}
		RPEvent := NewRemovePlayerEvent(
			uid,
		)
		pl.CHsForRPEvent[eid1] <- RPEvent
		RPEvent.Wait()
	}

	delete(pl.items, eid)
	delete(pl.CHsForAPEvent, eid)
	delete(pl.CHsForULEvent, eid)
	delete(pl.CHsForRPEvent, eid)

	return CHForAPEvent,
		CHForULEvent,
		CHForRPEvent
}

func (pl *PlayerList) Finish(
	eid int32,
	cnt *Client,
) (
	ChanForAddPlayerEvent,
	ChanForUpdateLatencyEvent,
	ChanForRemovePlayerEvent,
	error,
) {
	pl.Lock()
	defer pl.Unlock()

	for _, item1 := range pl.items {
		uid1 := item1.GetUID()
		if err := cnt.RemovePlayer(
			uid1,
		); err != nil {
			return nil, nil, nil, err
		}
	}

	CHForAPEvent,
		CHForULEvent,
		CHForRPEvent :=
		pl.close(eid)

	return CHForAPEvent,
		CHForULEvent,
		CHForRPEvent,
		nil
}

func (pl *PlayerList) Close(
	eid int32,
) (
	ChanForAddPlayerEvent,
	ChanForUpdateLatencyEvent,
	ChanForRemovePlayerEvent,
) {
	pl.Lock()
	defer pl.Unlock()

	return pl.close(eid)
}
