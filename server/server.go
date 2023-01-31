package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
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
	cx, cz int, // player pos
	d int, // positive
) (int, int, int, int) {
	return cx + d, cz + d, cx - d, cz - d
}

var DifferentKeepAlivePayloadError = errors.New("the payload of keep-alive must be same as the given")
var OutOfRndDistRangeError = errors.New("it is out of maximum and minimum value of render distance")

type ChanForError chan any

type ChanForUpdatePosEvent chan *UpdatePosEvent
type ChanForConfirmKeepAliveEvent chan *ConfirmKeepAliveEvent
type ChanForAddPlayerEvent chan *AddPlayerEvent
type ChanForRemovePlayerEvent chan *RemovePlayerEvent
type ChanForUpdateLatencyEvent chan *UpdateLatencyEvent

type UpdatePosEvent struct {
	x float64
	y float64
	z float64
}

func NewUpdatePosEvent(
	x, y, z float64,
) *UpdatePosEvent {
	return &UpdatePosEvent{
		x: x,
		y: y,
		z: z,
	}
}

func (e *UpdatePosEvent) GetX() float64 {
	return e.x
}

func (e *UpdatePosEvent) GetY() float64 {
	return e.y
}

func (e *UpdatePosEvent) GetZ() float64 {
	return e.z
}

func (e *UpdatePosEvent) String() string {
	return fmt.Sprintf(
		"{ x: %f, y: %f, z: %f }",
		e.x, e.y, e.z,
	)
}

type ConfirmKeepAliveEvent struct {
	payload int64
}

func NewConfirmKeepAliveEvent(
	payload int64,
) *ConfirmKeepAliveEvent {
	return &ConfirmKeepAliveEvent{
		payload: payload,
	}
}

func (e *ConfirmKeepAliveEvent) GetPayload() int64 {
	return e.payload
}

func (e *ConfirmKeepAliveEvent) String() string {
	return fmt.Sprintf(
		"{ payload: %d }", e.payload,
	)
}

type AddPlayerEvent struct {
	uid      uuid.UUID
	username string
}

func NewAddPlayerEvent(
	uid uuid.UUID,
	username string,
) *AddPlayerEvent {
	return &AddPlayerEvent{
		uid:      uid,
		username: username,
	}
}

func (p *AddPlayerEvent) GetUUID() uuid.UUID {
	return p.uid
}

func (p *AddPlayerEvent) GetUsername() string {
	return p.username
}

func (p *AddPlayerEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v, username: %s } ",
		p.uid, p.username,
	)
}

type RemovePlayerEvent struct {
	uid uuid.UUID
}

func NewRemovePlayerEvent(
	uid uuid.UUID,
) *RemovePlayerEvent {
	return &RemovePlayerEvent{
		uid: uid,
	}
}

func (p *RemovePlayerEvent) GetUUID() uuid.UUID {
	return p.uid
}

func (p *RemovePlayerEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v } ",
		p.uid,
	)
}

type UpdateLatencyEvent struct {
	uid     uuid.UUID
	latency int32
}

func NewUpdateLatencyEvent(
	uid uuid.UUID,
	latency int32,
) *UpdateLatencyEvent {
	return &UpdateLatencyEvent{
		uid:     uid,
		latency: latency,
	}
}

func (p *UpdateLatencyEvent) GetUUID() uuid.UUID {
	return p.uid
}

func (p *UpdateLatencyEvent) GetLatency() int32 {
	return p.latency
}

func (p *UpdateLatencyEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v, latency: %d } ",
		p.uid, p.latency,
	)
}

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

	mutex0 *sync.RWMutex
	m0     map[ChunkPosStr]*Chunk // by string of chunk position

	mutex1 *sync.RWMutex
	m1     map[uuid.UUID]*Player // by player uid

	mutex2 *sync.RWMutex
	m2     map[uuid.UUID]ChanForAddPlayerEvent     // by player uid
	m3     map[uuid.UUID]ChanForRemovePlayerEvent  // by player uid
	m4     map[uuid.UUID]ChanForUpdateLatencyEvent // by player id
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

	var mutex0 sync.RWMutex
	var mutex1 sync.RWMutex
	var mutex2 sync.RWMutex

	return &Server{
		addr:       addr,
		max:        max,
		online:     0,
		last:       0,
		favicon:    favicon,
		desc:       desc,
		rndDist:    rndDist,
		spawnX:     spawnX,
		spawnY:     spawnY,
		spawnZ:     spawnZ,
		spawnYaw:   spawnYaw,
		spawnPitch: spawnPitch,
		mutex0:     &mutex0,
		m0:         make(map[ChunkPosStr]*Chunk),
		mutex1:     &mutex1,
		m1:         make(map[uuid.UUID]*Player),
		mutex2:     &mutex2,
		m2:         make(map[uuid.UUID]ChanForAddPlayerEvent),
		m3:         make(map[uuid.UUID]ChanForRemovePlayerEvent),
		m4:         make(map[uuid.UUID]ChanForUpdateLatencyEvent),
	}, nil
}

func (s *Server) countEntity() int32 {
	x := s.last
	s.last++
	return x
}

func (s *Server) updateChunks(
	lg *Logger,
	cx0, cz0, cx1, cz1, // current chunk range
	cx2, cz2, cx3, cz3 int, // previous chunk range
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
	)

	l0 := []int{cx0, cx1, cx2, cx3}
	l1 := []int{cz0, cz1, cz2, cz3}
	sort.Ints(l0)
	sort.Ints(l1)

	cx4, cz4, cx5, cz5 := l0[2], l1[2], l0[1], l1[1]
	lg.Debug(
		"It is completed to find the rectangle that is overlapped.",
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
	)
	lg.Debug(
		"The handler for UpdatePosEvent was started.",
	)

	defer func() {
		close(chanForEvent)

		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	dist := s.rndDist

	if err := func() error {
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
			return err
		}
		return nil
	}(); err != nil {
		panic(err)
	}

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

			cx0, cz0 := toChunkPos(x, z)
			cx1, cz1 := toChunkPos(prevX, prevZ)
			if cx0 != cx1 || cz0 != cz1 {
				cx2, cz2, cx3, cz3 := findRect(cx0, cz0, dist)
				cx4, cz4, cx5, cz5 := findRect(cx1, cz1, dist)
				if err := s.updateChunks(
					lg,
					cx2, cz2, cx3, cz3,
					cx4, cz4, cx5, cz5,
					cnt,
				); err != nil {
					panic(err)
				}
			}

			//ground := packet.GetOnGround()

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

	lg.Debug("The handler for UpdatePosEvent was ended")
}

func (s *Server) handlePlayerListEvent(
	player *Player,
	cnt *Client,
	chanForError ChanForError,
	ctx context.Context,
) {
	lg := NewLogger(
		NewLgElement("handler", "PlayerListEvent"),
		NewLgElement("player", player),
		NewLgElement("client", cnt),
	)
	lg.Debug(
		"The handler for PlayerListEvent was started.",
	)

	defer func() {

		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	uid, username := player.GetUid(), player.GetUsername()

	chanForAddEvent, chanForRemoveEvent := func() (
		ChanForAddPlayerEvent,
		ChanForRemovePlayerEvent,
	) {
		mutex := s.mutex2
		mutex.Lock()
		defer mutex.Unlock()

		event := NewAddPlayerEvent(uid, username)
		for _, chanForEvent := range s.m2 {
			chanForEvent <- event
		}

		for _, player := range s.m1 {
			uid, username := player.GetUid(), player.GetUsername()

			if err := cnt.AddPlayer(lg, uid, username); err != nil {
				panic(err)
			}
		}

		chanForAddEvent := make(ChanForAddPlayerEvent, 1)
		s.m2[uid] = chanForAddEvent

		chanForRemoveEvent := make(ChanForRemovePlayerEvent, 1)
		s.m3[uid] = chanForRemoveEvent

		return chanForAddEvent, chanForRemoveEvent
	}()

	defer func() {
		mutex := s.mutex2
		mutex.Lock()
		defer mutex.Unlock()

		event := NewRemovePlayerEvent(uid)
		for _, chanForEvent := range s.m3 {
			chanForEvent <- event
		}

		close(chanForAddEvent)
		delete(s.m2, uid)
		close(chanForRemoveEvent)
		delete(s.m3, uid)
	}()

	stop := false
	for {
		select {
		case event := <-chanForAddEvent:
			lg.Debug(
				"AddPlayerEvent was received by the channel.",
				NewLgElement("event", event),
			)

			uid, username := event.GetUUID(), event.GetUsername()
			if err := cnt.AddPlayer(lg, uid, username); err != nil {
				panic(err)
			}

			lg.Debug(
				"It is finished to process the event.",
			)
		case event := <-chanForRemoveEvent:
			lg.Debug(
				"RemovePlayerEvent was received by the channel.",
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

	lg.Debug("The handler for PlayerListEvent was ended")
}

func (s *Server) broadcastUpdateLatencyEvent(
	lg *Logger,
	uid uuid.UUID,
	latency int32,
) {
	lg.Debug(
		"It is started to broadcast UpdateLatencyEvent.",
		NewLgElement("uid", uid),
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
	cnt *Client,
	chanForError ChanForError,
	ctx context.Context,
) {
	lg := NewLogger(
		NewLgElement("handler", "UpdateLatencyEvent"),
		NewLgElement("client", cnt),
	)
	lg.Debug(
		"The handler for UpdateLatencyEvent was started.",
	)

	chanForEvent := make(ChanForUpdateLatencyEvent, 1)
	s.m4[uid] = chanForEvent

	defer func() {
		close(chanForEvent)
		delete(s.m4, uid)

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
		close(chanForEvent)

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

	spawnX, spawnY, spawnZ :=
		s.spawnX, s.spawnY, s.spawnZ
	spawnYaw, spawnPitch :=
		s.spawnYaw, s.spawnPitch

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
	player := NewPlayer(
		eid,
		uid,
		username,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	)
	s.addPlayer(player)
	defer func() {
		s.removePlayer(uid)
	}()

	lg.Info(
		"The player successfully logged in.",
		NewLgElement("eid", eid),
		NewLgElement("uid", uid),
		NewLgElement("username", username),
	)

	if err := cnt.Init(
		lg, eid,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	); err != nil {
		panic(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	chanForError := make(ChanForError, 1)

	chanForEvent0 := make(ChanForUpdatePosEvent, 1)
	go s.handleUpdatePosEvent(
		chanForEvent0,
		cnt,
		player,
		chanForError,
		ctx,
	)

	go s.handlePlayerListEvent(
		player,
		cnt,
		chanForError,
		ctx,
	)

	go s.handleUpdateLatencyEvent(
		uid,
		cnt,
		chanForError,
		ctx,
	)

	chanForEvent1 := make(ChanForConfirmKeepAliveEvent, 1)
	go s.handleConfirmKeepAliveEvent(
		chanForEvent1,
		uid,
		cnt,
		chanForError,
		ctx,
	)

	stop := false
	for {
		select {
		case <-time.After(Loop3Time):
			finish, err := cnt.Loop3(
				lg,
				chanForEvent0,
				chanForEvent1,
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

func (s *Server) loadPlayer(
	uid uuid.UUID,
) (*Player, bool) {
	s.mutex1.RLock()
	defer s.mutex1.RUnlock()

	player, has := s.m1[uid]
	return player, has
}

func (s *Server) addPlayer(
	player *Player,
) {
	s.mutex1.Lock()
	defer s.mutex1.Unlock()

	key := player.GetUid()
	s.m1[key] = player
}

func (s *Server) removePlayer(
	uid uuid.UUID,
) {
	s.mutex1.Lock()
	defer s.mutex1.Unlock()

	delete(s.m1, uid)
}
