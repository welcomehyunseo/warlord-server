package server

import (
	"context"
	"errors"
	"github.com/google/uuid"
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
)

func findRect(
	cx, cz int, // player pos
	d int, // positive
) (int, int, int, int) {

	return cx + d, cz + d, cx - d, cz - d
}

var OutOfRndDistRangeError = errors.New("it is out of maximum and minimum value of render distance")

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

func (s *Server) countLast() int32 {
	x := s.last
	s.last++
	return x
}

func (s *Server) updateChunks(
	cx0, cz0, cx1, cz1, // current chunk range
	cx2, cz2, cx3, cz3 int, // previous chunk range
	cnt *Client,
) error {
	l0 := []int{cx0, cx1, cx2, cx3}
	l1 := []int{cz0, cz1, cz2, cz3}
	sort.Ints(l0)
	sort.Ints(l1)

	cx4, cz4, cx5, cz5 := l0[2], l1[2], l0[1], l1[1]

	for i := cz0; i >= cz1; i-- {
		for j := cx0; j >= cx1; j-- {
			if cx5 <= j && j <= cx4 && cz5 <= i && i <= cz4 {
				continue
			}

			chunk := s.GetChunk(j, i)

			err := cnt.LoadChunk(
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
				int32(j),
				int32(i),
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Server) unloadChunks(
	cx0, cz0, // max
	cx1, cz1 int, // min
	cnt *Client,
) error {
	for i := cz0; i >= cz1; i-- {
		for j := cx0; j >= cx1; j-- {
			err := cnt.UnloadChunk(
				int32(j),
				int32(i),
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Server) initChunks(
	cx0, cz0, // max
	cx1, cz1 int, // min
	cnt *Client,
) error {
	for i := cz0; i >= cz1; i-- {
		for j := cx0; j >= cx1; j-- {
			chunk := s.GetChunk(j, i)

			err := cnt.LoadChunk(
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

	return nil
}

func (s *Server) handleConnection(
	conn net.Conn,
) {
	addr := conn.RemoteAddr()
	lg := NewLogger("addr: %s", addr)

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
	cnt := NewClient(cid, conn, lg)

	defer func() {
		cnt.Close()
	}()

	state := HandshakingState

	for {
		next, err := cnt.Loop0(state)
		if err != nil {
			panic(err)
		}
		state = next
		break
	}

	if state == StatusState {
		for {
			finish, err := cnt.Loop1(
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
			finish, uid, username, err := cnt.Loop2(state)
			if err != nil {
				panic(err)
			}
			if finish == false {
				continue
			}

			eid := s.countLast()
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

	lg.InfoWithVars(
		"Player was created.",
		"player: %+v", player,
	)

	dist := s.rndDist

	eid := player.GetEid()
	//uid := player.GetUid()
	//username := player.GetUsername()

	if err := cnt.Init(eid); err != nil {
		panic(err)
	}

	cx0, cz0 := toChunkPos(sx, sz)
	cx1, cz1, cx2, cz2 := findRect(
		cx0, cz0, dist,
	)
	if err := s.initChunks(
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

	chanForErrors := make(chan any, 1)

	chanForUpdatePlayerPosEvent := make(chan *UpdatePlayerPosEvent, 1)

	go func() {
		lg.Info("The handling thread is started for ChangePlayerPosPacket.")

		defer func() {
			lg.Info("The handling thread is ended for ChangePlayerPosPacket.")

			if err := recover(); err != nil {
				chanForErrors <- err
			}
		}()

		for {
			select {
			case event := <-chanForUpdatePlayerPosEvent:
				lg.Info(
					"The channel was received " +
						"ChangePlayerPosPacket in the thread.",
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
						cx2, cz2, cx3, cz3,
						cx4, cz4, cx5, cz5,
						cnt,
					); err != nil {
						panic(err)
					}
				}

				//onGround := packet.GetOnGround()
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-time.After(1):
			finish, err := cnt.Loop3(
				chanForUpdatePlayerPosEvent,
				state,
			)
			if err != nil {
				panic(err)
			}
			if finish == false {
				continue
			}
		case err := <-chanForErrors:
			panic(err)
		}
	}
}

func (s *Server) Render() {
	addr := s.addr
	ln, err := net.Listen(NetType, addr)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
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
	chunk := s.m0[key]

	return chunk
}

func (s *Server) SetChunk(
	cx, cz int,
	chunk *Chunk,
) {
	key := toChunkPosStr(cx, cz)
	s.m0[key] = chunk
}
