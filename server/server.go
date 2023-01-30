package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"net"
	"sort"
	"time"
)

const (
	NetType = "tcp"    // network type of server
	McName  = "1.12.2" // minecraft version name
	ProtVer = 340      // protocol version

	CompThold = 16 // threshold for compression

	MinRndDist = 2  // minimum render distance
	MaxRndDist = 32 // maximum render distance

	CheckKeepAliveTime = time.Second * 10
)

func findRect(
	cx, cz int, // player pos
	d int,      // positive
) (int, int, int, int) {
	return cx + d, cz + d, cx - d, cz - d
}

var DifferentKeepAlivePayloadError = errors.New("the payload of keep-alive must be same as the given")
var OutOfRndDistRangeError = errors.New("it is out of maximum and minimum value of render distance")

type ChanForError chan any
type ChanForUpdatePlayerPosEvent chan *UpdatePlayerPosEvent

type UpdatePlayerPosEvent struct {
	x float64
	y float64
	z float64
}

func NewUpdatePlayerPosEvent(
	x, y, z float64,
) *UpdatePlayerPosEvent {
	return &UpdatePlayerPosEvent{
		x: x,
		y: y,
		z: z,
	}
}

func (e *UpdatePlayerPosEvent) GetX() float64 {
	return e.x
}

func (e *UpdatePlayerPosEvent) GetY() float64 {
	return e.y
}

func (e *UpdatePlayerPosEvent) GetZ() float64 {
	return e.z
}

func (e *UpdatePlayerPosEvent) String() string {
	return fmt.Sprintf(
		"{ x: %f, y: %f, z: %f }",
		e.x, e.y, e.z,
	)
}

type ChanForConfirmKeepAliveEvent chan *ConfirmKeepAliveEvent

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

type Server struct {
	addr string // address

	max    int   // maximum number of players
	online int   // number of online players
	last   int32 // last number of entity

	favicon string // web image url
	desc    string // description of server

	rndDist int // render distance

	m0 map[ChunkPosStr]*Chunk
	m1 map[ChunkPosStr]map[uuid.UUID]*Player
}

func NewServer(
	addr string,
	max int,
	favicon string,
	desc string,
	rndDist int,
) (*Server, error) {
	// TODO: check addr is valid
	// TODO: check favicon is valid
	if rndDist < MinRndDist || MaxRndDist < rndDist {
		return nil, OutOfRndDistRangeError
	}

	return &Server{
		addr:    addr,
		max:     max,
		online:  0,
		last:    0,
		favicon: favicon,
		desc:    desc,
		rndDist: rndDist,
		m0:      make(map[ChunkPosStr]*Chunk),
		m1:      make(map[ChunkPosStr]map[uuid.UUID]*Player),
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

			chunk := s.GetChunk(j, i)

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
			chunk := s.GetChunk(j, i)

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

func (s *Server) handleConfirmKeepAliveEvent(
	chanForEvent ChanForConfirmKeepAliveEvent,
	cnt *Client,
	chanForError ChanForError,
	ctx context.Context,
) {
	lg := NewLogger(
		NewLgElement("handler", "ConfirmKeepAliveEvent"),
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

	flag := false
	var payload0 int64

	// TODO: update ping

	stop := false
	for {
		select {
		case <-time.After(CheckKeepAliveTime):
			if flag == true {
				break
			}
			payload0 = rand.Int63()
			if err := cnt.CheckKeepAlive(lg, payload0); err != nil {
				panic(err)
			}
			flag = true
		case event := <-chanForEvent:
			lg.Debug(
				"The event was received by the channel.",
				NewLgElement("event", event),
			)

			payload1 := event.GetPayload()

			if payload1 != payload0 {
				panic(DifferentKeepAlivePayloadError)
			}

			flag = false
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

func (s *Server) handleUpdatePlayerPosEvent(
	chanForEvent ChanForUpdatePlayerPosEvent,
	cnt *Client,
	player *Player,
	chanForError ChanForError,
	ctx context.Context,
) {
	lg := NewLogger(
		NewLgElement("handler", "UpdatePlayerPosEvent"),
		NewLgElement("client", cnt),
	)
	lg.Debug(
		"The handler for UpdatePlayerPosEvent was started.",
	)

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	dist := s.rndDist

	stop := false
	for {
		select {
		case event := <-chanForEvent:
			lg.Debug(
				"The event was received by the channel.",
				NewLgElement("event", event),
			)
			x := event.GetX()
			y := event.GetY()
			z := event.GetZ()
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

			//onGround := packet.GetOnGround()

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

	lg.Debug("The handler for UpdatePlayerPosEvent was ended")
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

	sx := float64(0)
	sy := float64(0)
	sz := float64(0)
	sYaw := float32(0)
	sPitch := float32(0)

	player := func() *Player {
		for {
			finish, uid, username, err := cnt.Loop2(lg, state)
			if err != nil {
				panic(err)
			}
			if finish == false {
				continue
			}

			eid := s.countEntity()
			player := NewPlayer(
				eid,
				uid,
				username,
				sx,
				sy,
				sz,
				sYaw,
				sPitch,
			)
			return player
		}
	}()

	eid := player.GetEid()
	uid := player.GetUid()
	username := player.GetUsername()

	lg.Info(
		"The player successfully logged in.",
		NewLgElement("eid", eid),
		NewLgElement("uid", uid),
		NewLgElement("username", username),
	)

	dist := s.rndDist

	if err := cnt.Init(lg, eid); err != nil {
		panic(err)
	}

	cx0, cz0 := toChunkPos(sx, sz)
	cx1, cz1, cx2, cz2 := findRect(
		cx0, cz0, dist,
	)
	if err := s.initChunks(
		lg,
		cx1, cz1, cx2, cz2,
		cnt,
	); err != nil {
		panic(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	chanForError := make(ChanForError, 1)

	chanForUpdatePlayerPosEvent := make(ChanForUpdatePlayerPosEvent, 1)
	chanForConfirmKeepAliveEvent := make(ChanForConfirmKeepAliveEvent, 1)

	go s.handleUpdatePlayerPosEvent(
		chanForUpdatePlayerPosEvent,
		cnt,
		player,
		chanForError,
		ctx,
	)

	go s.handleConfirmKeepAliveEvent(
		chanForConfirmKeepAliveEvent,
		cnt,
		chanForError,
		ctx,
	)

	stop := false
	for {
		select {
		case <-time.After(1):
			finish, err := cnt.Loop3(
				lg,
				chanForUpdatePlayerPosEvent,
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
	netType := NetType

	lg.Info(
		"It is started to render.",
		NewLgElement("addr", addr),
		NewLgElement("netType", netType),
	)

	ln, err := net.Listen(netType, addr)

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

func (s *Server) GetChunk(
	cx, cz int,
) *Chunk {
	key := toChunkPosStr(cx, cz)
	chunk, has := s.m0[key]
	if has == false {
		chunk = NewChunk()
	}
	return chunk
}

func (s *Server) SetChunk(
	cx, cz int,
	chunk *Chunk,
) {
	key := toChunkPosStr(cx, cz)
	s.m0[key] = chunk
}
