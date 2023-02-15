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
	m1     map[EID]*Player

	mutex2 *sync.RWMutex
	m2     map[EID]types.Nil // player list
	m3     map[EID]ChanForAddPlayerEvent
	m4     map[EID]ChanForRemovePlayerEvent
	m5     map[EID]ChanForUpdateLatencyEvent

	mutex6 *sync.RWMutex
	m6     map[EID]map[EID]types.Nil // interconnections between player and players in visible range, bidirectional way
	mutex7 *sync.RWMutex
	m7     map[ChunkPosStr]map[EID]types.Nil // players by chunk pos
	m15    map[EID]ChanForSpawnMobEvent
	m8     map[EID]ChanForSpawnPlayerEvent
	m9     map[EID]ChanForDespawnEntityEvent
	m10    map[EID]ChanForSetEntityLookEvent
	m11    map[EID]ChanForSetEntityRelativePosEvent
	m12    map[EID]ChanForSetEntityActionsEvent
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
	var mutex2 sync.RWMutex
	var mutex6 sync.RWMutex
	var mutex7 sync.RWMutex

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
		m1:     make(map[EID]*Player),

		mutex2: &mutex2,
		m2:     make(map[EID]types.Nil),
		m3:     make(map[EID]ChanForAddPlayerEvent),
		m4:     make(map[EID]ChanForRemovePlayerEvent),
		m5:     make(map[EID]ChanForUpdateLatencyEvent),

		mutex6: &mutex6,
		m6:     make(map[EID]map[EID]types.Nil),
		mutex7: &mutex7,
		m7:     make(map[ChunkPosStr]map[EID]types.Nil),
		m8:     make(map[EID]ChanForSpawnPlayerEvent),
		m9:     make(map[EID]ChanForDespawnEntityEvent),
		m10:    make(map[EID]ChanForSetEntityLookEvent),
		m11:    make(map[EID]ChanForSetEntityRelativePosEvent),
		m12:    make(map[EID]ChanForSetEntityActionsEvent),
	}, nil
}

func (s *Server) countEID() int32 {
	eid := s.last
	s.last++
	return eid
}

func (s *Server) initPlayerList(
	lg *Logger,
	player *Player,
	cnt *Client,
) error {
	s.mutex2.Lock()
	defer s.mutex2.Unlock()

	lg.Debug(
		"It is started to init player report.",
		NewLgElement("player", player),
		NewLgElement("cnt", cnt),
	)

	eid0 := player.GetEid()
	s.m2[eid0] = types.Nil{}

	uid0, username0 := player.GetUid(), player.GetUsername()
	for eid1, _ := range s.m2 {

		player1 := s.m1[eid1]
		uid1, username1 :=
			player1.GetUid(),
			player1.GetUsername()
		if err := cnt.AddPlayer(
			lg, uid1, username1,
		); err != nil {
			return err
		}

		if eid0 == eid1 {
			continue
		}
		event0 := NewAddPlayerEvent(uid0, username0)
		ch := s.m3[eid1]
		ch <- event0
		event0.Wait()
	}

	lg.Debug(
		"It is finished to init player report.",
	)
	return nil
}

func (s *Server) updateLatencyOfPlayerList(
	lg *Logger,
	uid uuid.UUID,
	latency int32,
) {
	s.mutex2.RLock()
	defer s.mutex2.RUnlock()

	lg.Debug(
		"It is started to update latency.",
		NewLgElement("uid", uid),
		NewLgElement("latency", latency),
	)

	event0 := NewUpdateLatencyEvent(uid, latency)
	for key, _ := range s.m2 {
		ch1 := s.m5[key]
		ch1 <- event0
	}

	lg.Debug(
		"It is finished to update latency.",
	)
}

func (s *Server) closePlayerList(
	lg *Logger,
	uid0 uuid.UUID, eid0 EID,
) {
	s.mutex2.Lock()
	defer s.mutex2.Unlock()

	lg.Debug(
		"It is started to reportChan player report.",
	)

	delete(s.m2, eid0)

	event0 := NewRemovePlayerEvent(uid0)
	for key, _ := range s.m2 {
		ch := s.m4[key]
		ch <- event0
	}

	lg.Debug(
		"It is finished to reportChan player report.",
	)
}

func (s *Server) handleAddPlayerEvent(
	chanForEvent ChanForAddPlayerEvent,
	player *Player,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"add-player-event-handler",
		NewLgElement("player", player),
		NewLgElement("client", cnt),
	)
	defer lg.Close()
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

func (s *Server) handleRemovePlayerEvent(
	chanForEvent ChanForRemovePlayerEvent,
	player *Player,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"remove-player-event-handler",
		NewLgElement("player", player),
		NewLgElement("client", cnt),
	)
	defer lg.Close()
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

func (s *Server) handleUpdateLatencyEvent(
	chanForEvent ChanForUpdateLatencyEvent,
	uid UID,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"update-latency-event-handler",
		NewLgElement("uid", uid),
		NewLgElement("client", cnt),
	)
	defer lg.Close()
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

func (s *Server) initChunks(
	lg *Logger,
	ch0 ChanForSpawnPlayerEvent,
	cnt *Client, p0 *Player,
	cx, cz int,
) error {
	s.mutex6.Lock()
	defer s.mutex6.Unlock()
	s.mutex7.Lock()
	defer s.mutex7.Unlock()

	lg.Debug(
		"It is started to init chunks.",
		NewLgElement("p0", p0),
		NewLgElement("cnt", cnt),
		NewLgElement("cx", cx),
		NewLgElement("cz", cz),
	)

	eid0, uid0 := p0.GetEid(), p0.GetUid()
	x0, y0, z0 :=
		p0.GetX(), p0.GetY(), p0.GetZ()
	yaw0, pitch0 :=
		p0.GetYaw(), p0.GetPitch()
	event0 := NewSpawnPlayerEvent(
		eid0, uid0,
		x0, y0, z0,
		yaw0, pitch0,
	)

	// init coupling
	s.m6[eid0] = make(map[EID]types.Nil)

	dist := s.rndDist
	maxCx, maxCz, minCx, minCz := findRect(
		cx, cz, dist,
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

			m, has := s.m7[key]
			if has == false {
				continue
			}
			for eid1, _ := range m {
				chan1 := s.m8[eid1]
				chan1 <- event0

				player1 := s.m1[eid1]
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
				ch0 <- event1

				s.m6[eid0][eid1] = types.Nil{}
				s.m6[eid1][eid0] = types.Nil{}
			}
		}
	}

	// init client ID by chunk pos
	chunkPosStr := toChunkPosStr(cx, cz)
	m, has := s.m7[chunkPosStr]
	if has == false {
		newMap := make(map[EID]types.Nil)
		s.m7[chunkPosStr] = newMap
		m = newMap
	}
	m[eid0] = types.Nil{}

	lg.Debug(
		"It is finished to init chunks.",
	)

	return nil
}

func (s *Server) updateChunks(
	lg *Logger,
	cnt *Client,
	player *Player,
	currCx, currCz int,
	prevCx, prevCz int,
) error {
	s.mutex6.Lock()
	defer s.mutex6.Unlock()
	s.mutex7.Lock()
	defer s.mutex7.Unlock()

	lg.Debug(
		"It is started to update chunks.",
		NewLgElement("cnt", cnt),
		NewLgElement("currCx", currCx),
		NewLgElement("currCz", currCz),
		NewLgElement("prevCx", prevCx),
		NewLgElement("prevCz", prevCz),
	)

	eid0 := player.GetEid()

	prevChunkPosStr := toChunkPosStr(prevCx, prevCz)
	m0 := s.m7[prevChunkPosStr]
	delete(m0, eid0)

	currChunkPosStr := toChunkPosStr(currCx, currCz)
	m1, has1 := s.m7[currChunkPosStr]
	if has1 == false {
		newMap := make(map[EID]types.Nil)
		s.m7[currChunkPosStr] = newMap
		m1 = newMap
	}
	m1[eid0] = types.Nil{}

	dist := s.rndDist
	maxCurrCx, maxCurrCz, minCurrCx, minCurrCz :=
		findRect(currCx, currCz, dist)
	maxPrevCx, maxPrevCz, minPrevCx, minPrevCz :=
		findRect(prevCx, prevCz, dist)
	maxSubCx, maxSubCz, minSubCx, minSubCz := subRects(
		maxCurrCx, maxCurrCz, minCurrCx, minCurrCz,
		maxPrevCx, maxPrevCz, minPrevCx, minPrevCz,
	)

	p0 := s.m1[eid0]
	eid, uid := p0.GetEid(), p0.GetUid()
	x, y, z := p0.GetX(), p0.GetY(), p0.GetZ()
	yaw, pitch := p0.GetYaw(), p0.GetPitch()
	ch0 := s.m8[eid0]
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

			m, has := s.m7[key]
			if has == false {
				continue
			}
			for eid1, _ := range m {
				ch1 := s.m8[eid1]
				ch1 <- event0

				p1 := s.m1[eid1]
				uid1 := p1.GetUid()
				x1, y1, z1 :=
					p1.GetX(), p1.GetY(), p1.GetZ()
				yaw1, pitch1 :=
					p1.GetYaw(), p1.GetPitch()
				event1 := NewSpawnPlayerEvent(
					eid1, uid1,
					x1, y1, z1,
					yaw1, pitch1,
				)
				ch0 <- event1

				s.m6[eid0][eid1] = types.Nil{}
				s.m6[eid1][eid0] = types.Nil{}
			}
		}
	}

	ch1 := s.m9[eid0]
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
			m, has := s.m7[key]
			if has == false {
				continue
			}
			for eid1, _ := range m {
				ch2 := s.m9[eid1]
				ch2 <- event1

				event2 := NewDespawnEntityEvent(
					eid1,
				)
				ch1 <- event2

				delete(s.m6[eid0], eid1)
				delete(s.m6[eid1], eid0)
			}
		}
	}

	lg.Debug("It is finished to update chunks.")
	return nil
}

func (s *Server) closeChunks(
	lg *Logger,
	player *Player,
) {
	s.mutex6.Lock()
	defer s.mutex6.Unlock()
	s.mutex7.Lock()
	defer s.mutex7.Unlock()

	lg.Debug(
		"It is started to reportChan chunks.",
		NewLgElement("player", player),
	)

	eid0 := player.GetEid()
	x0, z0 := player.GetX(), player.GetZ()
	cx0, cz0 := toChunkPos(x0, z0)

	// reportChan client ID by chunk pos
	chunkPosStr := toChunkPosStr(cx0, cz0)
	m := s.m7[chunkPosStr]
	delete(m, eid0)

	// reportChan interactions
	event0 := NewDespawnEntityEvent(eid0)
	for eid1, _ := range s.m6[eid0] {
		m := s.m6[eid1]
		delete(m, eid0)

		ch1 := s.m9[eid1]
		ch1 <- event0
	}
	delete(s.m6, eid0)

	lg.Debug(
		"It is finished to reportChan chunks.",
	)

	return
}

func (s *Server) handleUpdateChunkPosEvent(
	chanForEvent ChanForUpdateChunkPosEvent,
	uid UID,
	cnt *Client,
	player *Player,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"update-chunk-pos-event-handler",
		NewLgElement("uid", uid),
		NewLgElement("client", cnt),
		NewLgElement("player", player),
	)
	defer lg.Close()
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
				player,
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

func (s *Server) handleSpawnPlayerEvent(
	chanForEvent ChanForSpawnPlayerEvent,
	player *Player,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"spawn-player-event-handler",
		NewLgElement("player", player),
		NewLgElement("client", cnt),
	)
	defer lg.Close()
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

func (s *Server) handleDespawnEntityEvent(
	chanForEvent ChanForDespawnEntityEvent,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"despawn-entity-event-handler",
		NewLgElement("client", cnt),
	)
	defer lg.Close()
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

func (s *Server) broadcastSetEntityLookEvent(
	lg *Logger,
	eid int32,
	yaw, pitch float32,
	ground bool,
) {
	s.mutex6.RLock()
	defer s.mutex6.RUnlock()

	lg.Debug(
		"It is started to broadcast SetEntityLookEvent.",
		NewLgElement("eid", eid),
		NewLgElement("yaw", yaw),
		NewLgElement("pitch", pitch),
		NewLgElement("ground", ground),
	)

	event := NewSetEntityLookEvent(
		eid,
		yaw, pitch,
		ground,
	)

	m := s.m6[eid]
	for eid1, _ := range m {
		ch1 := s.m10[eid1]
		ch1 <- event
	}

	lg.Debug(
		"It is finished to broadcast SetEntityLookEvent.",
	)
}

func (s *Server) handleSetEntityLookEvent(
	chanForEvent ChanForSetEntityLookEvent,
	uid UID,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"set-entity-look-event-handler",
		NewLgElement("uid", uid),
		NewLgElement("client", cnt),
	)
	defer lg.Close()
	lg.Debug(
		"The handler for SetEntityLookEvent was started.",
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
			yaw, pitch := event.GetYaw(), event.GetPitch()
			ground := event.GetGround()
			if err := cnt.SetEntityLook(
				lg,
				eid,
				yaw, pitch,
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

	lg.Debug("The handler for SetEntityLookEvent was ended")
}

func (s *Server) handleUpdateLookEvent(
	chanForEvent ChanForUpdateLookEvent,
	cnt *Client,
	player *Player,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"update-look-event-handler",
		NewLgElement("cnt", cnt),
		NewLgElement("player", player),
	)
	defer lg.Close()
	lg.Debug(
		"It is started to handle UpdateLookEvent.",
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

			eid := player.GetEid()
			yaw, pitch := event.GetYaw(), event.GetPitch()
			ground := event.GetGround()

			player.UpdateLook(yaw, pitch)

			s.broadcastSetEntityLookEvent(
				lg,
				eid,
				yaw, pitch,
				ground,
			)

			lg.Debug(
				"It is finished to process the event.",
			)
		}

		if stop == true {
			break
		}
	}

	lg.Debug("It is finished to handle UpdateLookEvent.")
}

func (s *Server) broadcastSetEntityRelativePosEvent(
	lg *Logger,
	eid int32,
	x, y, z float64,
	prevX, prevY, prevZ float64,
) {
	s.mutex6.RLock()
	defer s.mutex6.RUnlock()

	lg.Debug(
		"It is started to broadcast SetEntityRelativePosEvent.",
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

	event := NewSetEntityRelativePosEvent(
		eid,
		deltaX, deltaY, deltaZ,
		true, // TODO
	)

	m := s.m6[eid]
	for eid1, _ := range m {
		ch1 := s.m11[eid1]
		ch1 <- event
	}

	lg.Debug(
		"It is finished to broadcast SetEntityRelativePosEvent.",
	)
}

func (s *Server) handleSetEntityRelativePosEvent(
	chanForEvent ChanForSetEntityRelativePosEvent,
	uid UID,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"set-entity-relative-pos-event-handler",
		NewLgElement("uid", uid),
		NewLgElement("client", cnt),
	)
	defer lg.Close()
	lg.Debug(
		"It is started to handle SetEntityRelativePosEvent.",
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
			if err := cnt.SetEntityRelativePos(
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

	lg.Debug("It is finished to handle SetEntityRelativePosEvent.")
}

func (s *Server) handleUpdatePosEvent(
	chan0 ChanForUpdatePosEvent,
	chan1 ChanForUpdateChunkPosEvent,
	player *Player,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"update-pos-event-handler",
		NewLgElement("player", player),
	)
	defer lg.Close()
	lg.Debug(
		"The handler for UpdatePosEvent was started.",
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

			s.broadcastSetEntityRelativePosEvent(
				lg,
				eid,
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

func (s *Server) broadcastSetEntityActionsEvent(
	lg *Logger,
	eid int32,
	sneaking bool,
	sprinting bool,
) {
	s.mutex6.RLock()
	defer s.mutex6.RUnlock()

	lg.Debug(
		"It is started to broadcast SetEntityActionsEvent.",
		NewLgElement("eid", eid),
		NewLgElement("sneaking", sneaking),
		NewLgElement("sprinting", sprinting),
	)

	event := NewSetEntityActionsEvent(
		eid,
		sneaking, sprinting,
	)
	m := s.m6[eid]
	for eid1, _ := range m {
		ch1 := s.m12[eid1]
		ch1 <- event
	}

	lg.Debug(
		"It is finished to broadcast SetEntityActionsEvent.",
	)
}

func (s *Server) handleSetEntityActionsEvent(
	chanForEvent ChanForSetEntityActionsEvent,
	cnt *Client,
	player *Player,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"set-entity-actions-event-handler",
		NewLgElement("cnt", cnt),
		NewLgElement("player", player),
	)
	defer lg.Close()
	lg.Debug("It is started to handle SetEntityActionsEvent.")

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
			sneaking, sprinting :=
				event.IsSneaking(), event.IsSprinting()
			if err := cnt.SetEntityActions(
				lg,
				eid,
				sneaking, sprinting,
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

	lg.Debug("It is finished to handle SetEntityActionsEvent.")
}

func (s *Server) handleStartSneakingEvent(
	chanForEvent ChanForStartSneakingEvent,
	cnt *Client,
	player *Player,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"start-sneaking-event-handler",
		NewLgElement("cnt", cnt),
		NewLgElement("player", player),
	)
	defer lg.Close()
	lg.Debug("It is started to handle StartSneakingEvent.")

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	eid := player.GetEid()

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

			player.StartSneaking()
			sneaking, sprinting :=
				player.IsSneaking(), player.IsSprinting()
			s.broadcastSetEntityActionsEvent(
				lg,
				eid,
				sneaking, sprinting,
			)

			lg.Debug(
				"It is finished to process the event.",
			)
		}

		if stop == true {
			break
		}
	}

	lg.Debug("It is finished to handle StartSneakingEvent.")
}

func (s *Server) handleStopSneakingEvent(
	chanForEvent ChanForStopSneakingEvent,
	cnt *Client,
	player *Player,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"stop-sneaking-event-handler",
		NewLgElement("cnt", cnt),
		NewLgElement("player", player),
	)
	defer lg.Close()
	lg.Debug("It is started to handle StopSneakingEvent.")

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	eid := player.GetEid()

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

			player.StopSneaking()
			sneaking, sprinting :=
				player.IsSneaking(), player.IsSprinting()
			s.broadcastSetEntityActionsEvent(
				lg,
				eid,
				sneaking, sprinting,
			)

			lg.Debug(
				"It is finished to process the event.",
			)
		}

		if stop == true {
			break
		}
	}

	lg.Debug("It is finished to handle StopSneakingEvent.")
}

func (s *Server) handleStartSprintingEvent(
	chanForEvent ChanForStartSprintingEvent,
	cnt *Client,
	player *Player,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"start-sprinting-event-handler",
		NewLgElement("cnt", cnt),
		NewLgElement("player", player),
	)
	defer lg.Close()
	lg.Debug("It is started to handle StartSprintingEvent.")

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	eid := player.GetEid()

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

			player.StartSprinting()
			sneaking, sprinting :=
				player.IsSneaking(), player.IsSprinting()
			s.broadcastSetEntityActionsEvent(
				lg,
				eid,
				sneaking, sprinting,
			)

			lg.Debug(
				"It is finished to process the event.",
			)
		}

		if stop == true {
			break
		}
	}

	lg.Debug("It is finished to handle StartSprintingEvent.")
}

func (s *Server) handleStopSprintingEvent(
	chanForEvent ChanForStopSprintingEvent,
	cnt *Client,
	player *Player,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"stop-sneaking-event-handler",
		NewLgElement("cnt", cnt),
		NewLgElement("player", player),
	)
	defer lg.Close()
	lg.Debug("It is started to handle StopSprintingEvent.")

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	eid := player.GetEid()

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

			player.StopSprinting()
			sneaking, sprinting :=
				player.IsSneaking(), player.IsSprinting()
			s.broadcastSetEntityActionsEvent(
				lg,
				eid,
				sneaking, sprinting,
			)

			lg.Debug(
				"It is finished to process the event.",
			)
		}

		if stop == true {
			break
		}
	}

	lg.Debug("It is finished to handle StopSprintingEvent.")
}

func (s *Server) handleConfirmKeepAliveEvent(
	chanForEvent ChanForConfirmKeepAliveEvent,
	uid UID,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"confirm-keep-alive-event-handler",
		NewLgElement("uid", uid),
		NewLgElement("client", cnt),
	)
	defer lg.Close()
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

			s.updateLatencyOfPlayerList(lg, uid, int32(latency))

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

func (s *Server) initConnection(
	lg *Logger,
	eid int32, uid UID,
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
	ChanForSetEntityLookEvent,
	ChanForUpdateLookEvent,
	ChanForSetEntityRelativePosEvent,
	ChanForUpdatePosEvent,
	ChanForSetEntityActionsEvent,
	ChanForStartSneakingEvent,
	ChanForStopSneakingEvent,
	ChanForStartSprintingEvent,
	ChanForStopSprintingEvent,
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
	s.addPlayer(eid, player)

	if err := cnt.Init(
		lg, eid,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	chanForAddPlayerEvent := make(ChanForAddPlayerEvent, 1)
	s.m3[eid] = chanForAddPlayerEvent
	go s.handleAddPlayerEvent(
		chanForAddPlayerEvent,
		player,
		cnt,
		chanForError,
	)

	chanForRemovePlayerEvent := make(ChanForRemovePlayerEvent, 1)
	s.m4[eid] = chanForRemovePlayerEvent
	go s.handleRemovePlayerEvent(
		chanForRemovePlayerEvent,
		player,
		cnt,
		chanForError,
	)

	chanForUpdateLatencyEvent := make(ChanForUpdateLatencyEvent, 1)
	s.m5[eid] = chanForUpdateLatencyEvent
	go s.handleUpdateLatencyEvent(
		chanForUpdateLatencyEvent,
		uid,
		cnt,
		chanForError,
	)

	if err := s.initPlayerList(lg, player, cnt); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	chanForSpawnPlayerEvent :=
		make(ChanForSpawnPlayerEvent, 1)
	s.m8[eid] = chanForSpawnPlayerEvent
	go s.handleSpawnPlayerEvent(
		chanForSpawnPlayerEvent,
		player,
		cnt,
		chanForError,
	)

	chanForDespawnEntityEvent :=
		make(ChanForDespawnEntityEvent, 1)
	s.m9[eid] = chanForDespawnEntityEvent
	go s.handleDespawnEntityEvent(
		chanForDespawnEntityEvent,
		cnt,
		chanForError,
	)

	chanForUpdateChunkPosEvent :=
		make(ChanForUpdateChunkPosEvent, 1)
	go s.handleUpdateChunkPosEvent(
		chanForUpdateChunkPosEvent,
		uid,
		cnt,
		player,
		chanForError,
	)

	chanForSetEntityLookEvent := make(ChanForSetEntityLookEvent, 1)
	s.m10[eid] = chanForSetEntityLookEvent
	go s.handleSetEntityLookEvent(
		chanForSetEntityLookEvent,
		uid,
		cnt,
		chanForError,
	)

	chanForUpdateLookEvent := make(ChanForUpdateLookEvent, 1)
	go s.handleUpdateLookEvent(
		chanForUpdateLookEvent,
		cnt,
		player,
		chanForError,
	)

	chanForSetEntityRelativePosEvent :=
		make(ChanForSetEntityRelativePosEvent, 1)
	s.m11[eid] = chanForSetEntityRelativePosEvent
	go s.handleSetEntityRelativePosEvent(
		chanForSetEntityRelativePosEvent,
		uid,
		cnt,
		chanForError,
	)

	chanForUpdatePosEvent := make(ChanForUpdatePosEvent, 1)
	go s.handleUpdatePosEvent(
		chanForUpdatePosEvent,
		chanForUpdateChunkPosEvent,
		player,
		chanForError,
	)

	chanForSetEntityActionsEvent := make(ChanForSetEntityActionsEvent, 1)
	s.m12[eid] = chanForSetEntityActionsEvent
	go s.handleSetEntityActionsEvent(
		chanForSetEntityActionsEvent,
		cnt,
		player,
		chanForError,
	)

	chanForStartSneakingEvent := make(ChanForStartSneakingEvent, 1)
	go s.handleStartSneakingEvent(
		chanForStartSneakingEvent,
		cnt,
		player,
		chanForError,
	)

	chanForStopSneakingEvent := make(ChanForStopSneakingEvent, 1)
	go s.handleStopSneakingEvent(
		chanForStopSneakingEvent,
		cnt,
		player,
		chanForError,
	)

	chanForStartSprintingEvent := make(ChanForStartSprintingEvent, 1)
	go s.handleStartSprintingEvent(
		chanForStartSprintingEvent,
		cnt,
		player,
		chanForError,
	)

	chanForStopSprintingEvent := make(ChanForStopSprintingEvent, 1)
	go s.handleStopSprintingEvent(
		chanForStopSprintingEvent,
		cnt,
		player,
		chanForError,
	)

	if err := s.initChunks(
		lg,
		chanForSpawnPlayerEvent,
		cnt, player,
		cx, cz,
	); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

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
		chanForSetEntityLookEvent,
		chanForUpdateLookEvent,
		chanForSetEntityRelativePosEvent,
		chanForUpdatePosEvent,
		chanForSetEntityActionsEvent,
		chanForStartSneakingEvent,
		chanForStopSneakingEvent,
		chanForStartSprintingEvent,
		chanForStopSprintingEvent,
		chanForConfirmKeepAliveEvent,
		nil
}

func (s *Server) closeConnection(
	lg *Logger,
	player *Player,
	chanForAddPlayerEvent ChanForAddPlayerEvent,
	chanForRemovePlayerEvent ChanForRemovePlayerEvent,
	chanForUpdateLatencyEvent ChanForUpdateLatencyEvent,
	chanForSpawnPlayerEvent ChanForSpawnPlayerEvent,
	chanForDespawnEntityEvent ChanForDespawnEntityEvent,
	chanForUpdateChunkPosEvent ChanForUpdateChunkPosEvent,
	chanForSetEntityLookEvent ChanForSetEntityLookEvent,
	chanForUpdateLookEvent ChanForUpdateLookEvent,
	chanForSetEntityRelativePosEvent ChanForSetEntityRelativePosEvent,
	chanForUpdatePosEvent ChanForUpdatePosEvent,
	chanForSetEntityActionsEvent ChanForSetEntityActionsEvent,
	chanForStartSneakingEvent ChanForStartSneakingEvent,
	chanForStopSneakingEvent ChanForStopSneakingEvent,
	chanForStartSprintingEvent ChanForStartSprintingEvent,
	chanForStopSprintingEvent ChanForStopSprintingEvent,
	chanForConfirmKeepAliveEvent ChanForConfirmKeepAliveEvent,
) {
	s.globalMutex.Lock()
	defer s.globalMutex.Unlock()

	eid, uid := player.GetEid(), player.GetUid()

	close(chanForConfirmKeepAliveEvent)

	s.closeChunks(
		lg,
		player,
	)

	close(chanForStopSprintingEvent)

	close(chanForStartSprintingEvent)

	close(chanForStopSneakingEvent)

	close(chanForStartSneakingEvent)

	delete(s.m12, eid)
	close(chanForSetEntityActionsEvent)

	close(chanForUpdatePosEvent)

	delete(s.m11, eid)
	close(chanForSetEntityRelativePosEvent)

	close(chanForUpdateLookEvent)

	delete(s.m10, eid)
	close(chanForSetEntityLookEvent)

	close(chanForUpdateChunkPosEvent)

	delete(s.m9, eid)
	close(chanForDespawnEntityEvent)

	delete(s.m8, eid)
	close(chanForSpawnPlayerEvent)

	s.closePlayerList(
		lg, uid, eid,
	)

	delete(s.m5, eid)
	close(chanForUpdateLatencyEvent)

	delete(s.m4, eid)
	close(chanForRemovePlayerEvent)

	delete(s.m3, eid)
	close(chanForAddPlayerEvent)

	s.removePlayer(eid)

}

func (s *Server) handleConnection(
	conn net.Conn,
) {
	addr := conn.RemoteAddr()
	lg := NewLogger(
		"connection-handler",
		NewLgElement("addr", addr),
	)
	defer lg.Close()
	lg.Debug("It is started to handle Connection.")

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
		chanForSetEntityLookEvent,
		chanForUpdateLookEvent,
		chanForSetEntityRelativePosEvent,
		chanForUpdatePosEvent,
		chanForSetEntityActionsEvent,
		chanForStartSneakingEvent,
		chanForStopSneakingEvent,
		chanForStartSprintingEvent,
		chanForStopSprintingEvent,
		chanForConfirmKeepAliveEvent,
		err :=
		s.initConnection(
			lg,
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
		player,
		chanForAddPlayerEvent,
		chanForRemovePlayerEvent,
		chanForUpdateLatencyEvent,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForUpdateChunkPosEvent,
		chanForSetEntityLookEvent,
		chanForUpdateLookEvent,
		chanForSetEntityRelativePosEvent,
		chanForUpdatePosEvent,
		chanForSetEntityActionsEvent,
		chanForStartSneakingEvent,
		chanForStopSneakingEvent,
		chanForStartSprintingEvent,
		chanForStopSprintingEvent,
		chanForConfirmKeepAliveEvent,
	)

	stop := false
	for {
		select {
		case <-time.After(Loop3Time):
			finish, err := cnt.Loop3(
				lg,
				chanForUpdatePosEvent,
				chanForUpdateLookEvent,
				chanForStartSneakingEvent,
				chanForStopSneakingEvent,
				chanForStartSprintingEvent,
				chanForStopSprintingEvent,
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

	lg.Debug("It is finished to handle Connection.")
}

func (s *Server) Render() {
	lg := NewLogger(
		"server-renderer",
	)
	defer lg.Close()
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
	eid EID,
	player *Player,
) {
	s.mutex1.Lock()
	defer s.mutex1.Unlock()

	s.m1[eid] = player
}

func (s *Server) removePlayer(
	eid EID,
) {
	s.mutex1.Lock()
	defer s.mutex1.Unlock()

	delete(s.m1, eid)
}
