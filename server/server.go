package server

import (
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

var DifferentKeepAlivePayloadError = errors.New("the payload of keep-alive must be same as the given")
var OutOfRndDistRangeError = errors.New("it is out of maximum and minimum value of render distance")

type ChanForError chan any

func findRect(
	cx, cz int,
	d int,
) (
	int, int, int, int,
) {
	maxCx, maxCz, minCx, minCz :=
		cx+d, cz+d, cx-d, cz-d
	return maxCx, maxCz, minCx, minCz
}

func subRects(
	maxCx0, maxCz0, minCx0, minCz0 int,
	maxCx1, maxCz1, minCx1, minCz1 int,
) (
	int, int, int, int,
) {
	l0 := []int{maxCx0, minCx0, maxCx1, minCx1}
	l1 := []int{maxCz0, minCz0, maxCz1, minCz1}
	sort.Ints(l0)
	sort.Ints(l1)
	maxSubCx, maxSubCz, minSubCx, minSubCz :=
		l0[2], l1[2], l0[1], l1[1]
	return maxSubCx, maxSubCz, minSubCx, minSubCz
}

type Server struct {
	addr string // address

	max    int   // maximum number of players
	online int   // number of online players
	last   int32 // last entity ID

	favicon string // base64 png image string
	text    string // description of server

	rndDist    int // render distance
	spawnX     float64
	spawnY     float64
	spawnZ     float64
	spawnYaw   float32
	spawnPitch float32

	globalMutex *sync.RWMutex

	mutex0 *sync.RWMutex
	m0     map[ChunkPosStr]*Chunk

	mutex1 *sync.RWMutex
	m1     map[CID]*Player // init first, close at the very end

	mutex8 *sync.RWMutex
	m8     map[CID]map[CID]types.Nil // for coupling
	mutex5 *sync.RWMutex
	m5     map[ChunkPosStr]map[CID]types.Nil
	m6     map[CID]ChanForSpawnPlayerEvent
	m7     map[CID]ChanForDespawnEntityEvent

	m9 map[CID]ChanForRelativeMoveEvent

	// TODO: player list map & mutex
	m2 map[CID]ChanForAddPlayerEvent
	m3 map[CID]ChanForRemovePlayerEvent
	m4 map[CID]ChanForUpdateLatencyEvent
}

func NewServer(
	addr string,
	max int,
	favicon, text string,
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
	var mutex8 sync.RWMutex
	var mutex5 sync.RWMutex

	return &Server{
		addr:       addr,
		max:        max,
		online:     0,
		last:       0,
		favicon:    favicon,
		text:       text,
		rndDist:    rndDist,
		spawnX:     spawnX,
		spawnY:     spawnY,
		spawnZ:     spawnZ,
		spawnYaw:   spawnYaw,
		spawnPitch: spawnPitch,

		globalMutex: &globalMutex,

		mutex0: &mutex0,
		m0:     make(map[ChunkPosStr]*Chunk),

		mutex1: &mutex1,
		m1:     make(map[CID]*Player),

		mutex8: &mutex8,
		m8:     make(map[CID]map[CID]types.Nil),
		mutex5: &mutex5,
		m5:     make(map[ChunkPosStr]map[CID]types.Nil),
		m6:     make(map[CID]ChanForSpawnPlayerEvent),
		m7:     make(map[CID]ChanForDespawnEntityEvent),

		m9: make(map[CID]ChanForRelativeMoveEvent),

		m2: make(map[CID]ChanForAddPlayerEvent),
		m3: make(map[CID]ChanForRemovePlayerEvent),
		m4: make(map[CID]ChanForUpdateLatencyEvent),
	}, nil
}

func (s *Server) countEID() int32 {
	eid := s.last
	s.last++
	return eid
}

func (s *Server) handleConfirmKeepAliveEvent(
	chanForEvent ChanForConfirmKeepAliveEvent,
	uid uuid.UUID,
	cnt *Client,
	chanForError ChanForError,
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
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
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
		}

		if stop == true {
			break
		}
	}

	lg.Debug(
		"The handler for ConfirmKeepAliveEvent was ended",
	)
}

func (s *Server) initChunks(
	lg *Logger,
	chan0 ChanForSpawnPlayerEvent,
	cnt *Client, player0 *Player,
	cx, cz int,
) error {
	s.mutex8.Lock()
	defer s.mutex8.Unlock()
	s.mutex5.Lock()
	defer s.mutex5.Unlock()

	lg.Debug(
		"It is started to init chunks.",
		NewLgElement("player0", player0),
		NewLgElement("cnt", cnt),
		NewLgElement("cx", cx),
		NewLgElement("cz", cz),
	)

	cid0 := cnt.GetCID()

	// init coupling
	s.m8[cid0] = make(map[CID]types.Nil)
	dist := s.rndDist
	maxCx, maxCz, minCx, minCz := findRect(
		cx, cz, dist,
	)
	eid0, uid0 := player0.GetEid(), player0.GetUid()
	x0, y0, z0 :=
		player0.GetX(), player0.GetY(), player0.GetZ()
	yaw0, pitch0 :=
		player0.GetYaw(), player0.GetPitch()
	event0 := NewSpawnPlayerEvent(
		eid0, uid0,
		x0, y0, z0,
		yaw0, pitch0,
	)
	for cz := maxCz; cz >= minCz; cz-- {
		for cx := maxCx; cx >= minCx; cx-- {
			key := toChunkPosStr(cx, cz)

			chunk, has := s.m0[key]
			if has == false {
				chunk = NewChunk()
			}

			if err := cnt.LoadChunk(
				lg,
				true,
				true,
				int32(cx),
				int32(cz),
				chunk,
			); err != nil {
				return err
			}

			m, has := s.m5[key]
			if has == false {
				continue
			}
			for cid1, _ := range m {
				chan1 := s.m6[cid1]
				chan1 <- event0

				player1 := s.m1[cid1]
				eid1, uid1 :=
					player1.GetEid(), player1.GetUid()
				x1, y1, z1 :=
					player1.GetX(), player1.GetY(), player1.GetZ()
				yaw1, pitch1 :=
					player1.GetYaw(), player1.GetPitch()
				event1 := NewSpawnPlayerEvent(
					eid1, uid1,
					x1, y1, z1,
					yaw1, pitch1,
				)
				chan0 <- event1

				s.m8[cid0][cid1] = types.Nil{}
				s.m8[cid1][cid0] = types.Nil{}
			}
		}
	}

	// init client ID by chunk pos
	chunkPosStr := toChunkPosStr(cx, cz)
	m0, has := s.m5[chunkPosStr]
	if has == false {
		m1 := make(map[uuid.UUID]types.Nil)
		s.m5[chunkPosStr] = m1
		m0 = m1
	}
	m0[cid0] = types.Nil{}

	lg.Debug(
		"It is finished to init chunks.",
	)

	return nil
}

func (s *Server) updateChunks(
	lg *Logger,
	cnt *Client,
	currCx, currCz int,
	prevCx, prevCz int,
) error {
	s.mutex8.Lock()
	defer s.mutex8.Unlock()
	s.mutex5.RLock()
	defer s.mutex5.RUnlock()

	lg.Debug(
		"It is started to update chunks.",
		NewLgElement("currCx", currCx),
		NewLgElement("currCz", currCz),
		NewLgElement("prevCx", prevCx),
		NewLgElement("prevCz", prevCz),
	)

	dist := s.rndDist
	maxCurrCx, maxCurrCz, minCurrCx, minCurrCz :=
		findRect(currCx, currCz, dist)
	maxPrevCx, maxPrevCz, minPrevCx, minPrevCz :=
		findRect(prevCx, prevCz, dist)
	maxSubCx, maxSubCz, minSubCx, minSubCz := subRects(
		maxCurrCx, maxCurrCz, minCurrCx, minCurrCz,
		maxPrevCx, maxPrevCz, minPrevCx, minPrevCz,
	)

	cid0 := cnt.GetCID()
	p := s.m1[cid0]
	eid, uid := p.GetEid(), p.GetUid()
	x, y, z := p.GetX(), p.GetY(), p.GetZ()
	yaw, pitch := p.GetYaw(), p.GetPitch()
	chan0 := s.m6[cid0]
	event0 := NewSpawnPlayerEvent(
		eid, uid,
		x, y, z,
		yaw, pitch,
	)

	for cz := maxCurrCz; cz >= minCurrCz; cz-- {
		for cx := maxCurrCx; cx >= minCurrCx; cx-- {
			if minSubCx <= cx && cx <= maxSubCx &&
				minSubCz <= cz && cz <= maxSubCz {
				continue
			}

			key := toChunkPosStr(cx, cz)
			chunk, has := s.m0[key]
			if has == false {
				chunk = NewChunk()
			}

			if err := cnt.LoadChunk(
				lg,
				true,
				true,
				int32(cx),
				int32(cz),
				chunk,
			); err != nil {
				return err
			}

			m, has := s.m5[key]
			if has == false {
				continue
			}
			for cid1, _ := range m {
				chan1 := s.m6[cid1]
				chan1 <- event0

				p1 := s.m1[cid1]
				eid1, uid1 :=
					p1.GetEid(), p1.GetUid()
				x1, y1, z1 :=
					p1.GetX(), p1.GetY(), p1.GetZ()
				yaw1, pitch1 :=
					p1.GetYaw(), p1.GetPitch()
				event1 := NewSpawnPlayerEvent(
					eid1, uid1,
					x1, y1, z1,
					yaw1, pitch1,
				)
				chan0 <- event1

				s.m8[cid0][cid1] = types.Nil{}
				s.m8[cid1][cid0] = types.Nil{}
			}
		}
	}

	chan1 := s.m7[cid0]
	event1 := NewDespawnEntityEvent(
		eid,
	)
	for cz := maxPrevCz; cz >= minPrevCz; cz-- {
		for cx := maxPrevCx; cx >= minPrevCx; cx-- {
			if minSubCx <= cx && cx <= maxSubCx &&
				minSubCz <= cz && cz <= maxSubCz {
				continue
			}

			if err := cnt.UnloadChunk(
				lg,
				int32(cx),
				int32(cz),
			); err != nil {
				return err
			}

			key := toChunkPosStr(cx, cz)
			m, has := s.m5[key]
			if has == false {
				continue
			}
			for cid1, _ := range m {
				chan2 := s.m7[cid1]
				chan2 <- event1

				p1 := s.m1[cid1]
				eid1 := p1.GetEid()
				event2 := NewDespawnEntityEvent(
					eid1,
				)
				chan1 <- event2

				delete(s.m8[cid0], cid1)
				delete(s.m8[cid1], cid0)
			}
		}
	}

	lg.Debug("It is finished to update chunks.")
	return nil
}

func (s *Server) closeChunks(
	lg *Logger,
	cid0 CID, player *Player,
) {
	s.mutex8.Lock()
	defer s.mutex8.Unlock()
	s.mutex5.Lock()
	defer s.mutex5.Unlock()

	lg.Debug(
		"It is started to close chunks.",
		NewLgElement("cid0", cid0),
		NewLgElement("player", player),
	)

	x, z := player.GetX(), player.GetZ()
	cx, cz := toChunkPos(x, z)

	// close client ID by chunk pos
	chunkPosStr := toChunkPosStr(cx, cz)
	m := s.m5[chunkPosStr]
	delete(m, cid0)

	// close interactions
	eid0 := player.GetEid()
	event0 := NewDespawnEntityEvent(eid0)
	for cid1, _ := range s.m8[cid0] {
		m := s.m8[cid1]
		delete(m, cid0)

		chan1 := s.m7[cid1]
		chan1 <- event0
	}
	delete(s.m8, cid0)

	lg.Debug(
		"It is finished to close chunks.",
	)

	return
}

func (s *Server) handleUpdateChunkPosEvent(
	chanForEvent ChanForUpdateChunkPosEvent,
	uid uuid.UUID,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		NewLgElement("handler", "UpdateChunkPosEvent"),
		NewLgElement("uid", uid),
		NewLgElement("client", cnt),
	)
	lg.Debug(
		"It is started to handle UpdateChunkPosEvent.",
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
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			lg.Debug(
				"The event was received by the channel.",
				NewLgElement("event", event),
			)

			currCx, currCz := event.GetCurrCx(), event.GetCurrCz()
			prevCx, prevCz := event.GetPrevCx(), event.GetPrevCz()

			if err := s.updateChunks(
				lg,
				cnt,
				currCx, currCz,
				prevCx, prevCz,
			); err != nil {
				panic(err)
			}

			lg.Debug(
				"It is finished to process the event.",
			)
		}

		if stop == true {
			break
		}
	}

	lg.Debug("It is finished to handle UpdateChunkPosEvent.")
}

func (s *Server) broadcastSpawnEntityEvent(
	lg *Logger,
	cid CID,
	eid int32, uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
) {
	lg.Debug(
		"It is started to broadcast SpawnPlayerEvent.",
	)

	chanForEvent := s.m6[cid]
	event := NewSpawnPlayerEvent(
		eid, uid,
		x, y, z,
		yaw, pitch,
	)
	chanForEvent <- event

	lg.Debug(
		"It is finished to broadcast SpawnPlayerEvent.",
	)
}

func (s *Server) handleSpawnPlayerEvent(
	chanForEvent ChanForSpawnPlayerEvent,
	player *Player,
	cnt *Client,
	chanForError ChanForError,
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
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
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
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for SpawnPlayerEvent was ended")
}

func (s *Server) broadcastDespawnEntityEvent(
	lg *Logger,
	cid CID,
	eid int32,
) {
	s.mutex8.RLock()
	defer s.mutex8.RUnlock()

	lg.Debug(
		"It is started to broadcast DespawnEntityEvent.",
	)

	chanForEvent := s.m7[cid]
	event := NewDespawnEntityEvent(
		eid,
	)
	chanForEvent <- event

	lg.Debug(
		"It is finished to broadcast DespawnEntityEvent.",
	)
}

func (s *Server) handleDespawnEntityEvent(
	chanForEvent ChanForDespawnEntityEvent,
	uid uuid.UUID,
	cnt *Client,
	chanForError ChanForError,
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
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
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
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for DespawnEntityEvent was ended")
}

func (s *Server) broadcastRelativeMoveEvent(
	lg *Logger,
	cid CID, eid int32,
	x, y, z float64,
	prevX, prevY, prevZ float64,
) {
	s.mutex8.RLock()
	defer s.mutex8.RUnlock()

	lg.Debug(
		"It is started to broadcast RelativeMoveEvent.",
		NewLgElement("cid", cid),
		NewLgElement("eid", eid),
		NewLgElement("x", x),
		NewLgElement("y", y),
		NewLgElement("z", z),
		NewLgElement("prevX", prevX),
		NewLgElement("prevY", prevY),
		NewLgElement("prevZ", prevZ),
	)

	deltaX, deltaY, deltaZ :=
		int16(((x*32)-(prevX*32))*128),
		int16(((y*32)-(prevY*32))*128),
		int16(((z*32)-(prevZ*32))*128)

	event := NewRelativeMoveEvent(
		eid,
		deltaX, deltaY, deltaZ,
		true, // TODO
	)

	m := s.m8[cid]
	for cid, _ := range m {
		ch := s.m9[cid]
		ch <- event
	}

	lg.Debug(
		"It is finished to broadcast RelativeMoveEvent.",
	)
}

func (s *Server) handleRelativeMoveEvent(
	chanForEvent ChanForRelativeMoveEvent,
	uid uuid.UUID,
	cnt *Client,
	chanForError ChanForError,
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
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
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
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for RelativeMoveEvent was ended")
}

func (s *Server) handleUpdatePosEvent(
	chan0 ChanForUpdatePosEvent,
	chan1 ChanForUpdateChunkPosEvent,
	cnt *Client,
	player *Player,
	chanForError ChanForError,
) {
	lg := NewLogger(
		NewLgElement("handler", "UpdatePosEvent"),
		NewLgElement("cnt", cnt),
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

	stop := false
	for {
		select {
		case event, ok := <-chan0:
			if ok == false {
				stop = true
				break
			}
			lg.Debug(
				"The event was received by the channel.",
				NewLgElement("event", event),
			)

			eid := player.GetEid()
			x, y, z :=
				event.GetX(), event.GetY(), event.GetZ()
			player.UpdatePos(x, y, z)
			prevX := player.GetPrevX()
			prevY := player.GetPrevY()
			prevZ := player.GetPrevZ()

			s.broadcastRelativeMoveEvent(
				lg,
				cid, eid,
				x, y, z,
				prevX, prevY, prevZ,
			)

			currCx, currCz := toChunkPos(x, z)
			prevCx, prevCz := toChunkPos(prevX, prevZ)
			if currCx != prevCx || currCz != prevCz {
				event := NewUpdateChunkPosEvent(
					currCx, currCz,
					prevCx, prevCz,
				)
				chan1 <- event
			}

			lg.Debug(
				"It is finished to process the event.",
			)
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for UpdatePosEvent was ended.")
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
		uid, username :=
			player.GetUid(),
			player.GetUsername()

		if err := cnt.AddPlayer(
			lg, uid, username,
		); err != nil {
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
	lg.Debug(
		"It is started to broadcast AddPlayerEvent.",
	)

	event := NewAddPlayerEvent(uid, username)
	for _, chanForEvent := range s.m2 {
		chanForEvent <- event
		event.Wait()
	}

	lg.Debug(
		"It is finished to broadcast AddPlayerEvent.",
	)
}

func (s *Server) handleAddPlayerEvent(
	chanForEvent ChanForAddPlayerEvent,
	player *Player,
	cnt *Client,
	chanForError ChanForError,
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
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			lg.Debug(
				"The event was received by the channel.",
				NewLgElement("event", event),
			)

			uid, username := event.GetUUID(), event.GetUsername()
			if err := cnt.AddPlayer(lg, uid, username); err != nil {
				event.Fail()
				panic(err)
			}

			event.Done()

			lg.Debug(
				"It is finished to process the event.",
			)
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
	lg.Debug(
		"It is started to broadcast RemovePlayerEvent.",
		NewLgElement("uid", uid),
	)

	event := NewRemovePlayerEvent(uid)
	for _, chanForEvent := range s.m3 {
		chanForEvent <- event
	}

	lg.Debug(
		"It is finished to broadcast RemovePlayerEvent.",
	)
}

func (s *Server) handleRemovePlayerEvent(
	chanForEvent ChanForRemovePlayerEvent,
	player *Player,
	cnt *Client,
	chanForError ChanForError,
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
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
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

	event := NewUpdateLatencyEvent(uid, latency)
	for _, chanForEvent := range s.m4 {
		chanForEvent <- event
	}

	lg.Debug(
		"It is finished to broadcast UpdateLatencyEvent.",
	)
}

func (s *Server) handleUpdateLatencyEvent(
	uid uuid.UUID,
	chanForEvent ChanForUpdateLatencyEvent,
	cnt *Client,
	chanForError ChanForError,
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
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
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
		}

		if stop == true {
			break
		}
	}

	lg.Debug("The handler for UpdateLatencyEvent was ended")
}

func (s *Server) initConnection(
	lg *Logger,
	cid uuid.UUID,
	eid int32, uid uuid.UUID,
	username string,
	cnt *Client,
	chanForError ChanForError,
) (
	*Player,
	ChanForAddPlayerEvent,
	ChanForRemovePlayerEvent,
	ChanForUpdateLatencyEvent,
	ChanForSpawnPlayerEvent,
	ChanForDespawnEntityEvent,
	ChanForUpdateChunkPosEvent,
	ChanForRelativeMoveEvent,
	ChanForUpdatePosEvent,
	ChanForConfirmKeepAliveEvent,
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
	cx, cz := toChunkPos(spawnX, spawnZ)

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
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	if err := s.addAllPlayers(lg, cnt); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	s.broadcastAddPlayerEvent(lg, uid, username)
	chanForAddPlayerEvent := make(ChanForAddPlayerEvent, 1)
	s.m2[cid] = chanForAddPlayerEvent
	go s.handleAddPlayerEvent(
		chanForAddPlayerEvent,
		player,
		cnt,
		chanForError,
	)

	chanForRemovePlayerEvent := make(ChanForRemovePlayerEvent, 1)
	s.m3[cid] = chanForRemovePlayerEvent
	go s.handleRemovePlayerEvent(
		chanForRemovePlayerEvent,
		player,
		cnt,
		chanForError,
	)

	chanForUpdateLatencyEvent := make(ChanForUpdateLatencyEvent, 1)
	s.m4[cid] = chanForUpdateLatencyEvent
	go s.handleUpdateLatencyEvent(
		uid,
		chanForUpdateLatencyEvent,
		cnt,
		chanForError,
	)

	chanForSpawnPlayerEvent :=
		make(ChanForSpawnPlayerEvent, 1)
	s.m6[cid] = chanForSpawnPlayerEvent
	go s.handleSpawnPlayerEvent(
		chanForSpawnPlayerEvent,
		player,
		cnt,
		chanForError,
	)

	chanForDespawnEntityEvent :=
		make(ChanForDespawnEntityEvent, 1)
	s.m7[cid] = chanForDespawnEntityEvent
	go s.handleDespawnEntityEvent(
		chanForDespawnEntityEvent,
		uid,
		cnt,
		chanForError,
	)

	chanForUpdateChunkPosEvent :=
		make(ChanForUpdateChunkPosEvent, 1)
	if err := s.initChunks(
		lg,
		chanForSpawnPlayerEvent,
		cnt, player,
		cx, cz,
	); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	go s.handleUpdateChunkPosEvent(
		chanForUpdateChunkPosEvent,
		uid,
		cnt,
		chanForError,
	)

	chanForRelativeMoveEvent := make(ChanForRelativeMoveEvent, 1)
	s.m9[cid] = chanForRelativeMoveEvent
	go s.handleRelativeMoveEvent(
		chanForRelativeMoveEvent,
		uid,
		cnt,
		chanForError,
	)

	chanForUpdatePosEvent := make(ChanForUpdatePosEvent, 1)
	go s.handleUpdatePosEvent(
		chanForUpdatePosEvent,
		chanForUpdateChunkPosEvent,
		cnt,
		player,
		chanForError,
	)

	chanForConfirmKeepAliveEvent := make(ChanForConfirmKeepAliveEvent, 1)
	go s.handleConfirmKeepAliveEvent(
		chanForConfirmKeepAliveEvent,
		uid,
		cnt,
		chanForError,
	)

	lg.Debug(
		"It is finished to init Connection.",
	)
	return player,
		chanForAddPlayerEvent,
		chanForRemovePlayerEvent,
		chanForUpdateLatencyEvent,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForUpdateChunkPosEvent,
		chanForRelativeMoveEvent,
		chanForUpdatePosEvent,
		chanForConfirmKeepAliveEvent,
		nil
}

func (s *Server) closeConnection(
	lg *Logger,
	player *Player, cid uuid.UUID,
	chanForAddPlayerEvent ChanForAddPlayerEvent,
	chanForRemovePlayerEvent ChanForRemovePlayerEvent,
	chanForUpdateLatencyEvent ChanForUpdateLatencyEvent,
	chanForSpawnPlayerEvent ChanForSpawnPlayerEvent,
	chanForDespawnEntityEvent ChanForDespawnEntityEvent,
	chanForUpdateChunkPosEvent ChanForUpdateChunkPosEvent,
	chanForRelativeMoveEvent ChanForRelativeMoveEvent,
	chanForUpdatePosEvent ChanForUpdatePosEvent,
	chanForConfirmKeepAliveEvent ChanForConfirmKeepAliveEvent,
) {
	s.globalMutex.Lock()
	defer s.globalMutex.Unlock()

	uid := player.GetUid()

	close(chanForConfirmKeepAliveEvent)

	close(chanForUpdatePosEvent)

	delete(s.m9, cid)
	close(chanForRelativeMoveEvent)

	s.closeChunks(
		lg,
		cid, player,
	)
	close(chanForUpdateChunkPosEvent)

	delete(s.m7, cid)
	close(chanForDespawnEntityEvent)

	delete(s.m6, cid)
	close(chanForSpawnPlayerEvent)

	delete(s.m4, cid)
	close(chanForUpdateLatencyEvent)

	delete(s.m3, cid)
	close(chanForRemovePlayerEvent)
	s.broadcastRemovePlayerEvent(lg, uid)

	delete(s.m2, cid)
	close(chanForAddPlayerEvent)

	s.removePlayer(cid)
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
				s.text,
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

	eid := s.countEID()
	lg.Info(
		"The player successfully logged in.",
		NewLgElement("eid", eid),
		NewLgElement("uid", uid),
		NewLgElement("username", username),
	)

	chanForError := make(ChanForError, 1)

	//ctx := context.Background()
	//ctx, cancel := context.WithCancel(ctx)
	//defer func() {
	//	cancel()
	//}()

	player,
		chanForAddPlayerEvent,
		chanForRemovePlayerEvent,
		chanForUpdateLatencyEvent,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForUpdateChunkPosEvent,
		chanForRelativeMoveEvent,
		chanForUpdatePosEvent,
		chanForConfirmKeepAliveEvent,
		err :=
		s.initConnection(
			lg,
			cid,
			eid, uid,
			username,
			cnt,
			chanForError,
		)
	if err != nil {
		panic(err)
	}
	defer s.closeConnection(
		lg,
		player, cid,
		chanForAddPlayerEvent,
		chanForRemovePlayerEvent,
		chanForUpdateLatencyEvent,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForUpdateChunkPosEvent,
		chanForRelativeMoveEvent,
		chanForUpdatePosEvent,
		chanForConfirmKeepAliveEvent,
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

// unused
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
