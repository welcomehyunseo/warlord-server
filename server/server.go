package server

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go/types"
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"
)

const Network = "tcp"   // network type of server
const McName = "1.12.2" // minecraft version name
const ProtVer = 340     // protocol version

const CompThold = 16 // threshold for compression

const MinRndDist = 2  // minimum render distance
const MaxRndDist = 32 // maximum render distance

const CheckKeepAliveTime = time.Millisecond * 1000
const Loop3Time = time.Millisecond * 1

func findRect(
	cx, cz int, // center point
	d int, // distance from center point
) (int, int, int, int) {
	return cx + d, cz + d, cx - d, cz - d
}

func subRects(
	cx0, cz0 int, // maximum point of first rectangle
	cx1, cz1 int, // minimum point
	cx2, cz2 int, // maximum point of second rectangle
	cx3, cz3 int, //minimum point
) (
	int, int, // maximum point of subtracted rectangle
	int, int, // minimum point
) {
	l0 := []int{cx0, cx1, cx2, cx3}
	l1 := []int{cz0, cz1, cz2, cz3}
	sort.Ints(l0)
	sort.Ints(l1)
	return l0[2], l1[2], l0[1], l1[1]
}

var DifferentKeepAlivePayloadError = errors.New("the payload of keep-alive must be same as the given")
var OutOfRndDistRangeError = errors.New("it is out of maximum and minimum value of render distance")

type ChanForError chan any

type Server struct {
	addr string // address

	max    int   // maximum number of players
	online int   // number of online players
	last   int32 // last entity ID

	favicon string // base64 png image string
	desc    string // description of server

	rndDist    int // render distance
	spawnX     float64
	spawnY     float64
	spawnZ     float64
	spawnYaw   float32
	spawnPitch float32

	globalMutex *sync.RWMutex

	mutex0 *sync.RWMutex
	m0     map[ChunkPosStr]*Chunk // by string of chunk position

	mutex1 *sync.RWMutex
	m1     map[CID]*Player

	mutex2 *sync.RWMutex
	m2     map[CID]ChanForAddPlayerEvent

	mutex3 *sync.RWMutex
	m3     map[CID]ChanForRemovePlayerEvent

	mutex4 *sync.RWMutex
	m4     map[CID]ChanForUpdateLatencyEvent

	mutex5 *sync.RWMutex
	m5     map[ChunkPosStr]map[CID]types.Nil

	mutex6 *sync.RWMutex
	m6     map[CID]map[CID]types.Nil

	mutex7 *sync.RWMutex
	m7     map[CID]ChanForSpawnPlayerEvent

	mutex8 *sync.RWMutex
	m8     map[CID]ChanForDespawnEntityEvent

	mutex9 *sync.RWMutex
	m9     map[CID]ChanForRelativeMoveEvent
}

func NewServer(
	addr string,
	max int,
	favicon, desc string,
	rndDist int,
	spawnX, spawnY, spawnZ float64,
	spawnYaw, spawnPitch float32,
) (*Server, error) {
	// TODO: check addr is valid
	// TODO: check favicon is valid
	if rndDist < MinRndDist || MaxRndDist < rndDist {
		return nil, OutOfRndDistRangeError
	}

	var globalMutex sync.RWMutex

	var mutex0 sync.RWMutex
	var mutex1 sync.RWMutex
	var mutex2 sync.RWMutex
	var mutex3 sync.RWMutex
	var mutex4 sync.RWMutex
	var mutex5 sync.RWMutex
	var mutex6 sync.RWMutex
	var mutex7 sync.RWMutex
	var mutex8 sync.RWMutex
	var mutex9 sync.RWMutex

	return &Server{
		addr:        addr,
		max:         max,
		online:      0,
		last:        0,
		favicon:     favicon,
		desc:        desc,
		rndDist:     rndDist,
		spawnX:      spawnX,
		spawnY:      spawnY,
		spawnZ:      spawnZ,
		spawnYaw:    spawnYaw,
		spawnPitch:  spawnPitch,
		globalMutex: &globalMutex,
		mutex0:      &mutex0,
		m0:          make(map[ChunkPosStr]*Chunk),
		mutex1:      &mutex1,
		m1:          make(map[CID]*Player),
		mutex2:      &mutex2,
		m2:          make(map[CID]ChanForAddPlayerEvent),
		mutex3:      &mutex3,
		m3:          make(map[CID]ChanForRemovePlayerEvent),
		mutex4:      &mutex4,
		m4:          make(map[CID]ChanForUpdateLatencyEvent),
		mutex5:      &mutex5,
		m5:          make(map[ChunkPosStr]map[CID]types.Nil),
		mutex6:      &mutex6,
		m6:          make(map[CID]map[CID]types.Nil),
		mutex7:      &mutex7,
		m7:          make(map[CID]ChanForSpawnPlayerEvent),
		mutex8:      &mutex8,
		m8:          make(map[CID]ChanForDespawnEntityEvent),
		mutex9:      &mutex9,
		m9:          make(map[CID]ChanForRelativeMoveEvent),
	}, nil
}

func (s *Server) countEntity() int32 {
	x := s.last
	s.last++
	return x
}

func (s *Server) updateChunks(
	lg *Logger,
	cx0, cz0, cx1, cz1 int, // current chunk range
	cx2, cz2, cx3, cz3 int, // previous chunk range
	cx4, cz4, cx5, cz5 int, // chunk range subtracted by current and previous chunk range
	cnt *Client,
) error {
	lg.Debug(
		"It is started to update chunks.",
		NewLgElement("cx0", cx0),
		NewLgElement("cz0", cz0),
		NewLgElement("cx1", cx1),
		NewLgElement("cz1", cz1),
		NewLgElement("cx2", cx2),
		NewLgElement("cz2", cz2),
		NewLgElement("cx3", cx3),
		NewLgElement("cz3", cz3),
		NewLgElement("cx4", cx4),
		NewLgElement("cz4", cz4),
		NewLgElement("cx5", cx5),
		NewLgElement("cz5", cz5),
	)

	for i := cz0; i >= cz1; i-- {
		for j := cx0; j >= cx1; j-- {
			if cx5 <= j && j <= cx4 && cz5 <= i && i <= cz4 {
				continue
			}

			chunk := s.loadChunk(j, i)

			err := cnt.LoadChunk(
				lg,
				true,
				true,
				int32(j),
				int32(i),
				chunk,
			)
			if err != nil {
				return err
			}
		}
	}

	for i := cz2; i >= cz3; i-- {
		for j := cx2; j >= cx3; j-- {
			if cx5 <= j && j <= cx4 && cz5 <= i && i <= cz4 {
				continue
			}

			err := cnt.UnloadChunk(
				lg,
				int32(j),
				int32(i),
			)
			if err != nil {
				return err
			}
		}
	}

	lg.Debug("It is finished to update chunks.")
	return nil
}

func (s *Server) unloadChunks(
	lg *Logger,
	cx0, cz0, // max
	cx1, cz1 int, // min
	cnt *Client,
) error {
	lg.Debug(
		"It is started to unload chunks.",
	)
	for i := cz0; i >= cz1; i-- {
		for j := cx0; j >= cx1; j-- {
			err := cnt.UnloadChunk(
				lg,
				int32(j),
				int32(i),
			)
			if err != nil {
				return err
			}
		}
	}

	lg.Debug(
		"It is finished to unload chunks.",
	)
	return nil
}

func (s *Server) initChunks(
	lg *Logger,
	cx0, cz0, // max
	cx1, cz1 int, // min
	cnt *Client,
) error {
	lg.Debug(
		"It is started to init chunks.",
	)

	for i := cz0; i >= cz1; i-- {
		for j := cx0; j >= cx1; j-- {
			chunk := s.loadChunk(j, i)

			err := cnt.LoadChunk(
				lg,
				true,
				true,
				int32(j),
				int32(i),
				chunk,
			)
			if err != nil {
				return err
			}
		}
	}

	lg.Debug(
		"It is finished to init chunks.",
	)
	return nil
}

func (s *Server) registerCidToChunkPos(
	lg *Logger,
	cid uuid.UUID,
	cx, cz int,
) {
	s.mutex5.Lock()
	defer s.mutex5.Unlock()

	lg.Debug(
		"It is started to register cid to chunk pos.",
		NewLgElement("cx", cx),
		NewLgElement("cz", cz),
	)

	chunkPosStr := toChunkPosStr(cx, cz)
	v0, has := s.m5[chunkPosStr]
	if has == false {
		v1 := make(map[uuid.UUID]types.Nil)
		s.m5[chunkPosStr] = v1
		v0 = v1
	}
	v0[cid] = types.Nil{}

	lg.Debug(
		"It is finished to register cid to chunk pos.",
	)
}

func (s *Server) initUpdatePosEvent(
	lg *Logger,
	cid uuid.UUID,
	cnt *Client,
) (
	ChanForUpdatePosEvent,
	error,
) {
	lg.Debug(
		"It is started to init UpdatePosEvent.",
	)

	dist := s.rndDist
	spawnX, spawnZ := s.spawnX, s.spawnZ
	cx0, cz0 := toChunkPos(spawnX, spawnZ)
	cx1, cz1, cx2, cz2 := findRect(
		cx0, cz0, dist,
	)
	if err := s.initChunks(
		lg,
		cx1, cz1, cx2, cz2,
		cnt,
	); err != nil {
		return nil, err
	}

	s.registerCidToChunkPos(
		lg, cid,
		cx0, cz0,
	)

	chanForEvent := make(ChanForUpdatePosEvent, 1)

	lg.Debug(
		"It is finished to init UpdatePosEvent.",
	)
	return chanForEvent, nil
}

func (s *Server) closeUpdatePosEvent(
	lg *Logger,
	chanForEvent ChanForUpdatePosEvent,
) {
	lg.Debug(
		"It is started to close UpdatePosEvent.",
	)

	close(chanForEvent)

	lg.Debug(
		"It is finished to close UpdatePosEvent.",
	)
}

func (s *Server) updateCidByChunkPos(
	lg *Logger,
	cid CID,
	cx0, cz0 int,
	cx1, cz1 int,
) {
	s.mutex5.Lock()
	defer s.mutex5.Unlock()

	lg.Debug(
		"It is started to update cid by chunk pos.",
		NewLgElement("cx0", cx0),
		NewLgElement("cz0", cz0),
		NewLgElement("cx1", cx1),
		NewLgElement("cx1", cx1),
	)

	chunkPosStr := toChunkPosStr(cx0, cz0)
	v0, has := s.m5[chunkPosStr]
	if has == false {
		v1 := make(map[uuid.UUID]types.Nil)
		s.m5[chunkPosStr] = v1
		v0 = v1
	}
	v0[cid] = types.Nil{}

	prevChunkPosStr := toChunkPosStr(cx1, cz1)
	v1 := s.m5[prevChunkPosStr]
	delete(v1, cid)

	lg.Debug(
		"It is finished to update cid by chunk pos.",
	)
}

func (s *Server) handleUpdatePosEvent(
	chanForEvent ChanForUpdatePosEvent,
	cnt *Client,
	player *Player,
	chanForError ChanForError,
	ctx context.Context,
) {
	lg := NewLogger(
		NewLgElement("handler", "UpdatePosEvent"),
		NewLgElement("client", cnt),
		NewLgElement("player", player),
	)
	lg.Debug(
		"The handler for UpdatePosEvent was started.",
	)

	defer func() {

		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	cid := cnt.GetCID()
	dist := s.rndDist

	stop := false
	for {
		select {
		case event := <-chanForEvent:
			lg.Debug(
				"The event was received by the channel.",
				NewLgElement("event", event),
			)
			x, y, z :=
				event.GetX(), event.GetY(), event.GetZ()
			player.UpdatePos(x, y, z)
			prevX := player.GetPrevX()
			//prevY := player.GetPrevY()
			prevZ := player.GetPrevZ()
			//ground := packet.GetOnGround()

			cx0, cz0 := toChunkPos(x, z)
			cx1, cz1 := toChunkPos(prevX, prevZ)
			if cx0 == cx1 && cz0 == cz1 {
				break
			}

			cx2, cz2, cx3, cz3 := findRect(cx0, cz0, dist)
			cx4, cz4, cx5, cz5 := findRect(cx1, cz1, dist)
			cx6, cz6, cx7, cz7 := subRects(
				cx2, cz2, cx3, cz3,
				cx4, cz4, cx5, cz5,
			)

			if err := s.updateChunks(
				lg,
				cx2, cz2, cx3, cz3,
				cx4, cz4, cx5, cz5,
				cx6, cz6, cx7, cz7,
				cnt,
			); err != nil {
				panic(err)
			}

			s.updateCidByChunkPos(
				lg,
				cid,
				cx0, cz0,
				cx1, cz1,
			)

			lg.Debug(
				"It is finished to process the event.",
			)
		case <-ctx.Done():
			stop = true
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for UpdatePosEvent was ended.")
}

func (s *Server) initConfirmKeepAliveEvent(
	lg *Logger,
) ChanForConfirmKeepAliveEvent {
	lg.Debug(
		"It is started to init ConfirmKeepAliveEvent.",
	)

	chanForEvent := make(ChanForConfirmKeepAliveEvent, 1)

	lg.Debug(
		"It is finished to init ConfirmKeepAliveEvent.",
	)
	return chanForEvent
}

func (s *Server) closeConfirmKeepAliveEvent(
	lg *Logger,
	chanForEvent ChanForConfirmKeepAliveEvent,
) ChanForConfirmKeepAliveEvent {
	lg.Debug(
		"It is started to close ConfirmKeepAliveEvent.",
	)

	close(chanForEvent)

	lg.Debug(
		"It is started to close ConfirmKeepAliveEvent.",
	)
	return chanForEvent
}

func (s *Server) handleConfirmKeepAliveEvent(
	chanForEvent ChanForConfirmKeepAliveEvent,
	uid uuid.UUID,
	cnt *Client,
	chanForError ChanForError,
	ctx context.Context,
) {
	lg := NewLogger(
		NewLgElement("handler", "ConfirmKeepAliveEvent"),
		NewLgElement("uid", uid),
		NewLgElement("client", cnt),
	)
	lg.Debug(
		"The handler for ConfirmKeepAliveEvent was started.",
	)

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	start := time.Time{}
	var payload0 int64

	// TODO: update start

	stop := false
	for {
		select {
		case <-time.After(CheckKeepAliveTime):
			if start.IsZero() == false {
				break
			}
			payload0 = rand.Int63()
			if err := cnt.CheckKeepAlive(lg, payload0); err != nil {
				panic(err)
			}
			start = time.Now()
		case event := <-chanForEvent:
			lg.Debug(
				"The event was received by the channel.",
				NewLgElement("event", event),
			)

			payload1 := event.GetPayload()
			if payload1 != payload0 {
				panic(DifferentKeepAlivePayloadError)
			}
			end := time.Now()
			latency := end.Sub(start).Milliseconds()

			s.broadcastUpdateLatencyEvent(lg, uid, int32(latency))

			start = time.Time{}
			lg.Debug(
				"It is finished to process the event.",
			)
		case <-ctx.Done():
			stop = true
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for ConfirmKeepAliveEvent was ended")
}

func (s *Server) addAllPlayers(
	lg *Logger,
	cnt *Client,
) error {
	s.mutex1.RLock()
	defer s.mutex1.RUnlock()

	lg.Debug(
		"It is started to add all players.",
	)

	for _, player := range s.m1 {
		uid, username := player.GetUid(), player.GetUsername()

		if err := cnt.AddPlayer(lg, uid, username); err != nil {
			return err
		}
	}

	lg.Debug(
		"It is finished to add all players.",
	)
	return nil
}

func (s *Server) broadcastAddPlayerEvent(
	lg *Logger,
	uid uuid.UUID,
	username string,
) {
	s.mutex2.Lock()
	defer s.mutex2.Unlock()

	lg.Debug(
		"It is started to broadcast AddPlayerEvent.",
	)

	event, finish := NewAddPlayerEvent(uid, username)
	for _, chanForEvent := range s.m2 {
		chanForEvent <- event
		<-finish
	}

	lg.Debug(
		"It is finished to broadcast AddPlayerEvent.",
	)
}

func (s *Server) initAddPlayerEvent(
	lg *Logger,
	cid uuid.UUID,
) (
	ChanForAddPlayerEvent,
	error,
) {
	s.mutex2.Lock()
	defer s.mutex2.Unlock()

	lg.Debug(
		"It is started to init AddPlayerEvent.",
	)

	chanForEvent := make(ChanForAddPlayerEvent, 1)
	s.m2[cid] = chanForEvent

	lg.Debug(
		"It is finished to init AddPlayerEvent.",
	)

	return chanForEvent, nil
}

func (s *Server) closeAddPlayerEvent(
	lg *Logger,
	cid uuid.UUID,
	chanForEvent ChanForAddPlayerEvent,
) {
	s.mutex2.Lock()
	defer s.mutex2.Unlock()

	lg.Debug(
		"It is started to close PlayerAddEvent.",
	)

	close(chanForEvent)
	delete(s.m2, cid)

	lg.Debug(
		"It is finished to close PlayerAddEvent.",
	)
}

func (s *Server) handleAddPlayerEvent(
	chanForEvent ChanForAddPlayerEvent,
	player *Player,
	cnt *Client,
	chanForError ChanForError,
	ctx context.Context,
) {
	lg := NewLogger(
		NewLgElement("handler", "AddPlayerEvent"),
		NewLgElement("player", player),
		NewLgElement("client", cnt),
	)
	lg.Debug(
		"The handler for AddPlayerEvent was started.",
	)

	defer func() {

		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event := <-chanForEvent:
			lg.Debug(
				"The event was received by the channel.",
				NewLgElement("event", event),
			)

			uid, username := event.GetUUID(), event.GetUsername()
			finish := event.GetFinish()
			if err := cnt.AddPlayer(lg, uid, username); err != nil {
				finish <- false
				panic(err)
			}

			finish <- true

			lg.Debug(
				"It is finished to process the event.",
			)
		case <-ctx.Done():
			stop = true
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for AddPlayerEvent was ended")
}

func (s *Server) broadcastRemovePlayerEvent(
	lg *Logger,
	uid uuid.UUID,
) {
	s.mutex3.RLock()
	defer s.mutex3.RUnlock()

	lg.Debug(
		"It is started to broadcast RemovePlayerEvent.",
	)

	event := NewRemovePlayerEvent(uid)
	for _, chanForEvent := range s.m3 {
		chanForEvent <- event
	}

	lg.Debug(
		"It is finished to broadcast RemovePlayerEvent.",
	)
}

func (s *Server) initRemovePlayerEvent(
	lg *Logger,
	cid uuid.UUID,
) (
	ChanForRemovePlayerEvent,
	error,
) {
	s.mutex3.Lock()
	defer s.mutex3.Unlock()

	lg.Debug(
		"It is started to init RemovePlayerEvent.",
	)

	chanFoRemoveEvent := make(ChanForRemovePlayerEvent, 1)
	s.m3[cid] = chanFoRemoveEvent

	lg.Debug(
		"It is finished to init RemovePlayerEvent.",
	)

	return chanFoRemoveEvent, nil
}

func (s *Server) closeRemovePlayerEvent(
	lg *Logger,
	cid uuid.UUID,
	chanForRemoveEvent ChanForRemovePlayerEvent,
) {
	s.mutex3.Lock()
	defer s.mutex3.Unlock()

	lg.Debug(
		"It is started to close RemovePlayerEvent.",
	)

	close(chanForRemoveEvent)
	delete(s.m3, cid)

	lg.Debug(
		"It is finished to close RemovePlayerEvent.",
	)
}

func (s *Server) handleRemovePlayerEvent(
	chanForEvent ChanForRemovePlayerEvent,
	player *Player,
	cnt *Client,
	chanForError ChanForError,
	ctx context.Context,
) {
	lg := NewLogger(
		NewLgElement("handler", "RemovePlayerEvent"),
		NewLgElement("player", player),
		NewLgElement("client", cnt),
	)
	lg.Debug(
		"The handler for RemovePlayerEvent was started.",
	)

	defer func() {

		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event := <-chanForEvent:
			lg.Debug(
				"The event was received by the channel.",
				NewLgElement("event", event),
			)

			uid := event.GetUUID()
			if err := cnt.RemovePlayer(lg, uid); err != nil {
				panic(err)
			}

			lg.Debug(
				"It is finished to process the event.",
			)
		case <-ctx.Done():
			stop = true
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for RemovePlayerEvent was ended")
}

func (s *Server) broadcastUpdateLatencyEvent(
	lg *Logger,
	uid uuid.UUID,
	latency int32,
) {
	lg.Debug(
		"It is started to broadcast UpdateLatencyEvent.",
		NewLgElement("latency", latency),
	)

	s.mutex4.RLock()
	defer s.mutex4.RUnlock()

	event := NewUpdateLatencyEvent(uid, latency)
	for _, chanForEvent := range s.m4 {
		chanForEvent <- event
	}

	lg.Debug(
		"It is finished to broadcast UpdateLatencyEvent.",
	)
}

func (s *Server) initUpdateLatencyEvent(
	lg *Logger,
	cid uuid.UUID,
) ChanForUpdateLatencyEvent {
	s.mutex4.Lock()
	defer s.mutex4.Unlock()

	lg.Debug(
		"It is started to init UpdateLatencyEvent.",
	)

	chanForEvent := make(ChanForUpdateLatencyEvent, 1)
	s.m4[cid] = chanForEvent

	lg.Debug(
		"It is finished to init UpdateLatencyEvent.",
	)
	return chanForEvent
}

func (s *Server) closeUpdateLatencyEvent(
	lg *Logger,
	cid uuid.UUID,
	chanForEvent ChanForUpdateLatencyEvent,
) {
	s.mutex4.Lock()
	defer s.mutex4.Unlock()

	lg.Debug(
		"It is started to close UpdateLatencyEvent.",
	)

	close(chanForEvent)
	delete(s.m4, cid)

	lg.Debug(
		"It is finished to close UpdateLatencyEvent.",
	)
}

func (s *Server) handleUpdateLatencyEvent(
	uid uuid.UUID,
	chanForEvent ChanForUpdateLatencyEvent,
	cnt *Client,
	chanForError ChanForError,
	ctx context.Context,
) {
	lg := NewLogger(
		NewLgElement("handler", "UpdateLatencyEvent"),
		NewLgElement("uid", uid),
		NewLgElement("client", cnt),
	)
	lg.Debug(
		"The handler for UpdateLatencyEvent was started.",
	)

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event := <-chanForEvent:
			lg.Debug(
				"The event was received by the channel.",
				NewLgElement("event", event),
			)
			uid, latency := event.GetUUID(), event.GetLatency()
			if err := cnt.UpdateLatency(lg, uid, latency); err != nil {
				panic(err)
			}

			lg.Debug(
				"It is finished to process the event.",
			)
		case <-ctx.Done():
			stop = true
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for UpdateLatencyEvent was ended")
}

func (s *Server) initSpawnPlayerEvent(
	lg *Logger,
	cid uuid.UUID,
) ChanForSpawnPlayerEvent {
	s.mutex7.Lock()
	defer s.mutex7.Unlock()

	lg.Debug(
		"It is started to init SpawnPlayerEvent.",
	)

	chanForEvent := make(ChanForSpawnPlayerEvent, 1)
	s.m7[cid] = chanForEvent

	lg.Debug(
		"It is started to init SpawnPlayerEvent.",
	)
	return chanForEvent
}

func (s *Server) closeSpawnPlayerEvent(
	lg *Logger,
	cid uuid.UUID,
	chanForEvent ChanForSpawnPlayerEvent,
) {
	s.mutex7.Lock()
	defer s.mutex7.Unlock()

	lg.Debug(
		"It is started to close SpawnPlayerEvent.",
	)

	close(chanForEvent)
	delete(s.m7, cid)

	lg.Debug(
		"It is finished to close SpawnPlayerEvent.",
	)
}

func (s *Server) handleSpawnPlayerEvent(
	chanForEvent ChanForSpawnPlayerEvent,
	player *Player,
	cnt *Client,
	chanForError ChanForError,
	ctx context.Context,
) {
	lg := NewLogger(
		NewLgElement("handler", "SpawnPlayerEvent"),
		NewLgElement("player", player),
		NewLgElement("client", cnt),
	)
	lg.Debug(
		"The handler for SpawnPlayerEvent was started.",
	)

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event := <-chanForEvent:
			lg.Debug(
				"The event was received by the channel.",
				NewLgElement("event", event),
			)

			eid1, uid1 := event.GetEID(), event.GetUUID()
			x1, y1, z1 := event.GetX(), event.GetY(), event.GetZ()
			yaw1, pitch1 := event.GetYaw(), event.GetPitch()

			if err := cnt.SpawnPlayer(
				lg,
				eid1, uid1,
				x1, y1, z1,
				yaw1, pitch1,
			); err != nil {
				panic(err)
			}

			lg.Debug(
				"It is finished to process the event.",
			)
		case <-ctx.Done():
			stop = true
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for SpawnPlayerEvent was ended")
}

func (s *Server) initDespawnEntityEvent(
	lg *Logger,
	cid uuid.UUID,
) ChanForDespawnEntityEvent {
	s.mutex8.Lock()
	defer s.mutex8.Unlock()

	lg.Debug(
		"It is started to init DespawnEntityEvent.",
	)

	chanForEvent := make(ChanForDespawnEntityEvent, 1)
	s.m8[cid] = chanForEvent

	lg.Debug(
		"It is started to init DespawnEntityEvent.",
	)
	return chanForEvent
}

func (s *Server) closeDespawnEntityEvent(
	lg *Logger,
	cid uuid.UUID,
	chanForEvent ChanForDespawnEntityEvent,
) {
	s.mutex8.Lock()
	defer s.mutex8.Unlock()

	lg.Debug(
		"It is started to close DespawnEntityEvent.",
	)

	close(chanForEvent)
	delete(s.m8, cid)

	lg.Debug(
		"It is finished to close DespawnEntityEvent.",
	)
}

func (s *Server) handleDespawnEntityEvent(
	chanForEvent ChanForDespawnEntityEvent,
	uid uuid.UUID,
	cnt *Client,
	chanForError ChanForError,
	ctx context.Context,
) {
	lg := NewLogger(
		NewLgElement("handler", "DespawnEntityEvent"),
		NewLgElement("uid", uid),
		NewLgElement("client", cnt),
	)
	lg.Debug(
		"The handler for DespawnEntityEvent was started.",
	)

	defer func() {

		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event := <-chanForEvent:
			lg.Debug(
				"The event was received by the channel.",
				NewLgElement("event", event),
			)

			eid := event.GetEID()
			if err := cnt.DespawnEntity(
				lg, eid,
			); err != nil {
				panic(err)
			}

			lg.Debug(
				"It is finished to process the event.",
			)
		case <-ctx.Done():
			stop = true
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for DespawnEntityEvent was ended")
}

func (s *Server) initRelativeMoveEvent(
	lg *Logger,
	cid uuid.UUID,
) ChanForRelativeMoveEvent {
	s.mutex9.Lock()
	defer s.mutex9.Unlock()

	lg.Debug(
		"It is started to init RelativeMoveEvent.",
	)

	chanForEvent := make(ChanForRelativeMoveEvent, 1)
	s.m9[cid] = chanForEvent

	lg.Debug(
		"It is started to init RelativeMoveEvent.",
	)
	return chanForEvent
}

func (s *Server) closeRelativeMoveEvent(
	lg *Logger,
	cid uuid.UUID,
	chanForEvent ChanForRelativeMoveEvent,
) {
	s.mutex9.Lock()
	defer s.mutex9.Unlock()

	lg.Debug(
		"It is started to close RelativeMoveEvent.",
	)

	close(chanForEvent)
	delete(s.m9, cid)

	lg.Debug(
		"It is finished to close RelativeMoveEvent.",
	)
}

func (s *Server) handleRelativeMoveEvent(
	chanForEvent ChanForRelativeMoveEvent,
	uid uuid.UUID,
	cnt *Client,
	chanForError ChanForError,
	ctx context.Context,
) {
	lg := NewLogger(
		NewLgElement("handler", "RelativeMoveEvent"),
		NewLgElement("uid", uid),
		NewLgElement("client", cnt),
	)
	lg.Debug(
		"The handler for RelativeMoveEvent was started.",
	)

	defer func() {

		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event := <-chanForEvent:
			lg.Debug(
				"The event was received by the channel.",
				NewLgElement("event", event),
			)

			eid := event.GetEID()
			deltaX, deltaY, deltaZ :=
				event.GetDeltaX(), event.GetDeltaY(), event.GetDeltaZ()
			ground := event.GetGround()
			if err := cnt.RelativeMove(
				lg,
				eid,
				deltaX, deltaY, deltaZ,
				ground,
			); err != nil {
				panic(err)
			}

			lg.Debug(
				"It is finished to process the event.",
			)
		case <-ctx.Done():
			stop = true
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for RelativeMoveEvent was ended")
}

func (s *Server) initConnection(
	lg *Logger,
	cid uuid.UUID,
	eid int32, uid uuid.UUID,
	username string,
	cnt *Client,
	chanForError ChanForError,
	ctx context.Context,
) (
	ChanForUpdatePosEvent,
	ChanForConfirmKeepAliveEvent,
	ChanForAddPlayerEvent,
	ChanForRemovePlayerEvent,
	ChanForUpdateLatencyEvent,
	ChanForSpawnPlayerEvent,
	ChanForDespawnEntityEvent,
	ChanForRelativeMoveEvent,
	error,
) {
	s.globalMutex.Lock()
	defer s.globalMutex.Unlock()

	lg.Debug(
		"It is started to init Connection.",
	)

	spawnX, spawnY, spawnZ :=
		s.spawnX, s.spawnY, s.spawnZ
	spawnYaw, spawnPitch :=
		s.spawnYaw, s.spawnPitch

	player := NewPlayer(
		eid,
		uid,
		username,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	)
	s.addPlayer(cid, player)

	if err := cnt.Init(
		lg, eid,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	chanForUpdatePosEvent, err := s.initUpdatePosEvent(
		lg, cid, cnt,
	)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	go s.handleUpdatePosEvent(
		chanForUpdatePosEvent,
		cnt,
		player,
		chanForError,
		ctx,
	)

	chanForConfirmKeepAliveEvent := s.initConfirmKeepAliveEvent(
		lg,
	)
	go s.handleConfirmKeepAliveEvent(
		chanForConfirmKeepAliveEvent,
		uid,
		cnt,
		chanForError,
		ctx,
	)

	if err := s.addAllPlayers(lg, cnt); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	s.broadcastAddPlayerEvent(lg, uid, username)
	chanForAddPlayerEvent, err := s.initAddPlayerEvent(
		lg, cid,
	)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	go s.handleAddPlayerEvent(
		chanForAddPlayerEvent,
		player,
		cnt,
		chanForError,
		ctx,
	)

	chanForRemovePlayerEvent, err := s.initRemovePlayerEvent(
		lg, cid,
	)
	go s.handleRemovePlayerEvent(
		chanForRemovePlayerEvent,
		player,
		cnt,
		chanForError,
		ctx,
	)

	chanForUpdateLatencyEvent := s.initUpdateLatencyEvent(
		lg, cid,
	)
	go s.handleUpdateLatencyEvent(
		uid,
		chanForUpdateLatencyEvent,
		cnt,
		chanForError,
		ctx,
	)

	chanForSpawnPlayerEvent := s.initSpawnPlayerEvent(
		lg, cid,
	)
	go s.handleSpawnPlayerEvent(
		chanForSpawnPlayerEvent,
		player,
		cnt,
		chanForError,
		ctx,
	)

	chanForDespawnEntityEvent := s.initDespawnEntityEvent(
		lg, cid,
	)
	go s.handleDespawnEntityEvent(
		chanForDespawnEntityEvent,
		uid,
		cnt,
		chanForError,
		ctx,
	)

	chanForRelativeMoveEvent := s.initRelativeMoveEvent(
		lg, cid,
	)
	go s.handleRelativeMoveEvent(
		chanForRelativeMoveEvent,
		uid,
		cnt,
		chanForError,
		ctx,
	)

	lg.Debug(
		"It is finished to init Connection.",
	)
	return chanForUpdatePosEvent,
		chanForConfirmKeepAliveEvent,
		chanForAddPlayerEvent,
		chanForRemovePlayerEvent,
		chanForUpdateLatencyEvent,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForRelativeMoveEvent,
		nil
}

func (s *Server) closeConnection(
	lg *Logger,
	uid uuid.UUID, cid uuid.UUID,
	chanForUpdatePosEvent ChanForUpdatePosEvent,
	chanForConfirmKeepAliveEvent ChanForConfirmKeepAliveEvent,
	chanForAddPlayerEvent ChanForAddPlayerEvent,
	chanForRemovePlayerEvent ChanForRemovePlayerEvent,
	chanForUpdateLatencyEvent ChanForUpdateLatencyEvent,
	chanForSpawnPlayerEvent ChanForSpawnPlayerEvent,
	chanForDespawnEntityEvent ChanForDespawnEntityEvent,
	chanForRelativeMoveEvent ChanForRelativeMoveEvent,
) {
	s.globalMutex.Lock()
	defer s.globalMutex.Unlock()

	s.removePlayer(cid)
	s.closeUpdatePosEvent(lg, chanForUpdatePosEvent)
	s.closeConfirmKeepAliveEvent(lg, chanForConfirmKeepAliveEvent)
	s.closeAddPlayerEvent(lg, cid, chanForAddPlayerEvent)
	s.broadcastRemovePlayerEvent(lg, uid)
	s.closeRemovePlayerEvent(lg, cid, chanForRemovePlayerEvent)
	s.closeUpdateLatencyEvent(lg, cid, chanForUpdateLatencyEvent)
	s.closeSpawnPlayerEvent(lg, cid, chanForSpawnPlayerEvent)
	s.closeDespawnEntityEvent(lg, cid, chanForDespawnEntityEvent)
	s.closeRelativeMoveEvent(lg, cid, chanForRelativeMoveEvent)
}

func (s *Server) handleConnection(
	conn net.Conn,
) {
	addr := conn.RemoteAddr()
	lg := NewLogger(
		NewLgElement("addr", addr),
		NewLgElement("handler", "Connection"),
	)

	lg.Debug("The handler for connection was started.")

	defer func() {
		// TODO: send the Disconnect packet to the connection

		if err := recover(); err != nil {
			lg.Error(err)
		}
	}()

	cid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	cnt := NewClient(cid, conn)

	defer func() {
		cnt.Close(lg)
	}()

	state := HandshakingState

	for {
		next, err := cnt.Loop0(lg, state)
		if err != nil {
			panic(err)
		}
		state = next
		break
	}

	if state == StatusState {
		for {
			finish, err := cnt.Loop1(
				lg,
				state,
				s.max,
				s.online,
				s.desc,
				s.favicon,
			)
			if err != nil {
				panic(err)
			}
			if finish == false {
				continue
			}
			return
		}
	}

	uid, username := func() (
		uuid.UUID,
		string,
	) {
		for {
			finish, uid, username, err := cnt.Loop2(lg, state)
			if err != nil {
				panic(err)
			}
			if finish == false {
				continue
			}

			return uid, username
		}
	}()

	eid := s.countEntity()
	lg.Info(
		"The player successfully logged in.",
		NewLgElement("eid", eid),
		NewLgElement("uid", uid),
		NewLgElement("username", username),
	)

	chanForError := make(ChanForError, 1)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	chanForUpdatePosEvent,
		chanForConfirmKeepAliveEvent,
		chanForAddPlayerEvent,
		chanForRemovePlayerEvent,
		chanForUpdateLatencyEvent,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForRelativeMoveEvent,
		err :=
		s.initConnection(
			lg,
			cid,
			eid, uid,
			username,
			cnt,
			chanForError,
			ctx,
		)
	if err != nil {
		panic(err)
	}
	defer s.closeConnection(
		lg,
		uid, cid,
		chanForUpdatePosEvent,
		chanForConfirmKeepAliveEvent,
		chanForAddPlayerEvent,
		chanForRemovePlayerEvent,
		chanForUpdateLatencyEvent,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForRelativeMoveEvent,
	)

	stop := false
	for {
		select {
		case <-time.After(Loop3Time):
			finish, err := cnt.Loop3(
				lg,
				chanForUpdatePosEvent,
				chanForConfirmKeepAliveEvent,
				state,
			)
			if err != nil {
				panic(err)
			}
			if finish == false {
				continue
			}
		case <-chanForError:
			stop = true
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for connection was finished.")
}

func (s *Server) Render() {
	lg := NewLogger(
		NewLgElement("context", "server-renderer"),
	)

	addr := s.addr
	network := Network

	lg.Info(
		"It is started to render.",
		NewLgElement("addr", addr),
		NewLgElement("network", network),
	)

	ln, err := net.Listen(network, addr)

	if err != nil {
		panic(err)
	}
	defer func() {
		lg.Info(
			"It is finished to render.",
		)
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		lg.Info(
			"The server accepted a new connection.",
			NewLgElement("addr", conn.RemoteAddr()),
		)

		go s.handleConnection(conn)
	}

}

func (s *Server) GetOnline() int {
	return s.online
}

func (s *Server) loadChunk(
	cx, cz int,
) *Chunk {
	s.mutex0.RLock()
	defer s.mutex0.RUnlock()

	key := toChunkPosStr(cx, cz)
	chunk, has := s.m0[key]
	if has == false {
		chunk = NewChunk()
	}
	return chunk
}

func (s *Server) AddChunk(
	cx, cz int,
	chunk *Chunk,
) {
	s.mutex0.Lock()
	defer s.mutex0.Unlock()

	key := toChunkPosStr(cx, cz)
	s.m0[key] = chunk
}

func (s *Server) addPlayer(
	cid uuid.UUID,
	player *Player,
) {
	s.mutex1.Lock()
	defer s.mutex1.Unlock()

	s.m1[cid] = player
}

func (s *Server) removePlayer(
	cid uuid.UUID,
) {
	s.mutex1.Lock()
	defer s.mutex1.Unlock()

	delete(s.m1, cid)
}
