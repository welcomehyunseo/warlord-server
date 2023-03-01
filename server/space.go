package server

import (
	"sync"
)

type Space struct {
	*sync.RWMutex

	WIBEID map[EID]int // world index by EID

	pl     *PlayerList
	worlds []Overworld
}

func NewSpace() *Space {
	return &Space{
		new(sync.RWMutex),

		make(map[EID]int),

		NewPlayerList(),
		[]Overworld{},
	}
}

func (s *Space) GetNumberOfWorlds() int {
	return len(s.worlds)
}

func (s *Space) AddWorld(
	world Overworld,
) error {
	s.Lock()
	defer s.Unlock()

	s.worlds = append(s.worlds, world)

	return nil
}

func (s *Space) Init(
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
) error {
	s.Lock()
	defer s.Unlock()

	pl := s.pl
	if err := pl.Init(
		eid,
		uid, username,
		chanForAPEvent,
		chanForULEvent,
		chanForRPEvent,
		cnt,
	); err != nil {
		return err
	}

	world := s.worlds[index]
	if err := world.InitPlayer(
		eid,
		uid, username,
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

	s.WIBEID[eid] = index

	return nil
}

func (s *Space) Change(
	eid EID,
	uid UID, username string,
	index int,
	cnt *Client,
) error {
	s.Lock()
	defer s.Unlock()

	prevIndex := s.WIBEID[eid]
	prevWorld := s.worlds[prevIndex]

	chanForSPEvent,
		chanForDEEvent,
		chanForSERPEvent,
		chanForSELEvent,
		chanForSEMEvent,
		chanForLCEvent,
		chanForUnCEvent,
		chanForUpCEvent,
		err :=
		prevWorld.FinishPlayer(
			eid,
			cnt,
		)
	if err != nil {
		return err
	}

	world := s.worlds[index]
	if err := world.InitPlayer(
		eid,
		uid, username,
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

	s.WIBEID[eid] = index

	return nil
}

func (s *Space) UpdatePlayerLatency(
	eid EID,
	latency int32,
) error {
	s.RLock()
	defer s.RUnlock()

	pl := s.pl

	if err := pl.UpdateLatency(
		eid,
		latency,
	); err != nil {
		return err
	}

	return nil
}

func (s *Space) UpdatePlayerPos(
	eid EID,
	x, y, z float64,
	ground bool,
) error {
	s.RLock()
	defer s.RUnlock()

	index := s.WIBEID[eid]
	world := s.worlds[index]
	if err := world.UpdatePlayerPos(
		eid,
		x, y, z,
		ground,
	); err != nil {
		return err
	}

	return nil
}

func (s *Space) UpdatePlayerChunk(
	eid EID,
	prevCx, prevCz int32,
	currCx, currCz int32,
) error {
	s.RLock()
	defer s.RUnlock()

	index := s.WIBEID[eid]
	world := s.worlds[index]
	if err := world.UpdatePlayerChunk(
		eid,
		prevCx, prevCz,
		currCx, currCz,
	); err != nil {
		return err
	}

	return nil
}

func (s *Space) UpdatePlayerLook(
	eid EID,
	yaw, pitch float32,
	ground bool,
) error {
	s.RLock()
	defer s.RUnlock()

	index := s.WIBEID[eid]
	world := s.worlds[index]
	if err := world.UpdatePlayerLook(
		eid,
		yaw, pitch,
		ground,
	); err != nil {
		return err
	}

	return nil
}

func (s *Space) UpdatePlayerSneaking(
	eid EID,
	flag bool,
) error {
	s.RLock()
	defer s.RUnlock()

	index := s.WIBEID[eid]
	world := s.worlds[index]
	if err := world.UpdatePlayerSneaking(
		eid,
		flag,
	); err != nil {
		return err
	}

	return nil
}

func (s *Space) UpdatePlayerSprinting(
	eid EID,
	flag bool,
) error {
	s.RLock()
	defer s.RUnlock()

	index := s.WIBEID[eid]
	world := s.worlds[index]
	if err := world.UpdatePlayerSprinting(
		eid,
		flag,
	); err != nil {
		return err
	}

	return nil
}

func (s *Space) Finish(
	eid EID,
	cnt *Client,
) (
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
	error,
) {
	s.Lock()
	defer s.Unlock()

	index := s.WIBEID[eid]

	pl := s.pl
	world := s.worlds[index]

	chanForAPEvent,
		chanForULEvent,
		chanForRPEvent,
		err := pl.Finish(
		eid,
		cnt,
	)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	chanForSPEvent,
		chanForDEEvent,
		chanForSERPEvent,
		chanForSELEvent,
		chanForSEMEvent,
		chanForLCEvent,
		chanForUnCEvent,
		chanForUpCEvent,
		err := world.FinishPlayer(
		eid,
		cnt,
	)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	delete(s.WIBEID, eid)

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
		chanForUpCEvent,
		nil
}

func (s *Space) Close(
	eid EID,
) (
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
	s.Lock()
	defer s.Unlock()

	index := s.WIBEID[eid]

	pl := s.pl
	world := s.worlds[index]

	chanForAPEvent,
		chanForULEvent,
		chanForRPEvent :=
		pl.Close(
			eid,
		)

	chanForSPEvent,
		chanForDEEvent,
		chanForSERPEvent,
		chanForSELEvent,
		chanForSEMEvent,
		chanForLCEvent,
		chanForUnCEvent,
		chanForUpCEvent :=
		world.Close(
			eid,
		)

	delete(s.WIBEID, eid)

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
