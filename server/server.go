package server

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net"
)

const (
	NetType = "tcp"    // network type of server
	McName  = "1.12.2" // minecraft version name
	ProtVer = 340      // protocol version

	CompThold = 16 // threshold for compression

	MinRndDist = 3  // minimum render distance
	MaxRndDist = 32 // maximum render distance
)

var OutOfRndDistRangeError = errors.New("it is out of maximum and minimum value of render distance")

type ChunkPosStr = string

func toChunkPosStr(
	cx int,
	cy int,
	cz int,
) ChunkPosStr {
	return fmt.Sprintf("(%d,%d,%d)", cx, cy, cz)
}

func playerPosToChunkPos(
	x float64,
	y float64,
	z float64,
) (
	int,
	int,
	int,
) {
	x0 := int(x) / 16
	y0 := int(y) / 16
	z0 := int(z) / 16
	if x0 < 0 {
		x0 = x0 - 16
	}
	//if y0 < 0 {
	//	y0 = y0 - 16
	//}
	if z0 < 0 {
		z0 = z0 - 16
	}
	return x0, y0, z0
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
	m1 map[ChunkPosStr][]*Player
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
		m1:      make(map[ChunkPosStr][]*Player),
	}, nil
}

func (s *Server) countLast() int32 {
	x := s.last
	s.last++
	return x
}

func (s *Server) initChunks(
	x float64,
	y float64,
	z float64,
	cnt *Client,
) error {
	rndDist := s.rndDist

	cx, cy, cz := playerPosToChunkPos(x, y, z)
	cx0, cy0, cz0 := cx+rndDist, cy+rndDist, cz+rndDist
	cx1, cy1, cz1 := cx-rndDist, cy-rndDist, cz-rndDist

	if cy0 > 15 {
		cy0 = 15
	}
	if cy1 < 0 {
		cy1 = 0
	}

	for i := cz0; i >= cz1; i-- {
		for j := cx0; j >= cx1; j-- {
			cc := NewChunkColumn()

			for k := cy0; k >= cy1; k-- {
				chunk := s.GetChunk(j, k, i)
				if chunk == nil {
					continue
				}
				cc.SetChunk(uint8(k), chunk)
			}

			err := cnt.LoadChunkColumn(true, true, int32(j), int32(i), cc)
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
		"player: %+V", player,
	)

	eid := player.GetEid()
	//uid := player.GetUid()
	//username := player.GetUsername()

	if err := cnt.Init(eid); err != nil {
		panic(err)
	}

	if err := s.initChunks(sx, sy, sz, cnt); err != nil {
		panic(err)
	}

	for {
		finish, err := cnt.Loop3(state)
		if err != nil {
			panic(err)
		}

		if finish == false {
			continue
		}
		break
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
	cx int,
	cy int,
	cz int,
) *Chunk {
	key := toChunkPosStr(cx, cy, cz)
	chunk := s.m0[key]

	return chunk
}

func (s *Server) SetChunk(
	cx int,
	cy int,
	cz int,
	chunk *Chunk,
) {
	key := toChunkPosStr(cx, cy, cz)
	s.m0[key] = chunk
}
