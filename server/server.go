package server

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net"
	"sort"
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

type PosStr = string

func toPosStr(
	x int,
	y int,
	z int,
) PosStr {
	return fmt.Sprintf("(%d,%d,%d)", x, y, z)
}

func toChunkCellPos(
	x float64,
	y float64,
	z float64,
) (
	int,
	int,
	int,
) {
	cx := int(x) / 16
	cy := int(y) / 16
	cz := int(z) / 16
	if cx < 0 {
		cx = cx - 16
	}
	//if cy < 0 {
	//	cy = cy - 16
	//}
	if cz < 0 {
		cz = cz - 16
	}
	return cx, cy, cz
}

// findCube returns endpoints of cube at the distance d from the center (x, y, z).
// The order is from maximum point to minimum point.
func findCube(
	cx int,
	cy int,
	cz int,
	d int,
) (
	int, int, int,
	int, int, int,
) {
	return cx + d, cy + d, cz + d, cx - d, cy - d, cz - d
}

func isCubesOverlap(
	cx0 int, // current
	cy0 int,
	cz0 int,
	cx1 int, // prev
	cy1 int,
	cz1 int,
	d int,
) bool {
	cx2, cy2, cz2, cx3, cy3, cz3 := findCube(cx0, cy0, cz0, 2*d)
	if cx1 < cx3 || cy1 < cy3 || cz1 < cz3 ||
		cx2 < cx1 || cy2 < cy1 || cz2 < cz1 {
		return false
	}

	return true
}

// subCubes returns endpoints of overlapping cube between cubes c0 and c1 overlapped to each other.
// The parameters (x0, y0, z0) and (x1, y1, z1) are the center points of the cubes.
// The order is from maximum point to minimum point.
func subCubes(
	cx0 int,
	cy0 int,
	cz0 int,
	cx1 int,
	cy1 int,
	cz1 int,
	d int,
) (
	int, int, int,
	int, int, int,
) {
	cx2, cy2, cz2, cx3, cy3, cz3 := findCube(cx0, cy0, cz0, d)
	cx4, cy4, cz4, cx5, cy5, cz5 := findCube(cx1, cy1, cz1, d)
	l0 := []int{cx2, cx3, cx4, cx5}
	l1 := []int{cy2, cy3, cy4, cy5}
	l2 := []int{cz2, cz3, cz4, cz5}
	sort.Ints(l0)
	sort.Ints(l1)
	sort.Ints(l2)
	return l0[2], l1[2], l2[2], l0[1], l1[1], l2[1]
}

type Server struct {
	addr string // address

	max    int   // maximum number of players
	online int   // number of online players
	last   int32 // last number of entity

	favicon string // web image url
	desc    string // description of server

	rndDist int // render distance

	m0 map[PosStr]*ChunkCell
	m1 map[PosStr][]*Player
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
		m0:      make(map[PosStr]*ChunkCell),
		m1:      make(map[PosStr][]*Player),
	}, nil
}

func (s *Server) countLast() int32 {
	x := s.last
	s.last++
	return x
}

func (s *Server) initChunks(
	cx int,
	cy int,
	cz int,
	cnt *Client,
) error {
	rndDist := s.rndDist

	cx0, cy0, cz0, cx1, cy1, cz1 := findCube(
		cx, cy, cz, rndDist,
	)

	if cy0 > 15 {
		cy0 = 15
	}
	if cy1 < 0 {
		cy1 = 0
	}

	for i := cz0; i >= cz1; i-- {
		for j := cx0; j >= cx1; j-- {
			cc := NewChunkCol()

			for k := cy0; k >= cy1; k-- {
				chunk := s.GetChunkCell(j, k, i)
				if chunk == nil {
					continue
				}
				cc.SetChunkCell(uint8(k), chunk)
			}

			err := cnt.LoadChunk(
				true,
				true,
				int32(j),
				int32(i),
				cc,
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
		"player: %+V", player,
	)

	rndDist := s.rndDist

	eid := player.GetEid()
	//uid := player.GetUid()
	//username := player.GetUsername()

	if err := cnt.Init(eid); err != nil {
		panic(err)
	}

	cx, cy, cz := toChunkCellPos(sx, sy, sz)
	if err := s.initChunks(cx, cy, cz, cnt); err != nil {
		panic(err)
	}

	for {
		move, finish, err := cnt.Loop3(
			player,
			state,
		)
		if err != nil {
			panic(err)
		}
		if move == true {
			x0, y0, z0 :=
				player.GetX(), player.GetY(), player.GetZ()
			x1, y1, z1 :=
				player.GetPrevX(), player.GetPrevY(), player.GetPrevZ()
			cx0, cy0, cz0 := toChunkCellPos(x0, y0, z0)
			cx1, cy1, cz1 := toChunkCellPos(x1, y1, z1)

			if isCubesOverlap(cx0, cy0, cz0, cx1, cy1, cz1, rndDist) == false {
				if err := s.initChunks(cx0, cy0, cz0, cnt); err != nil {
					panic(err)
				}

				// unload chunks
				continue
			}

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

func (s *Server) GetChunkCell(
	cx int,
	cy int,
	cz int,
) *ChunkCell {
	key := toPosStr(cx, cy, cz)
	chunk := s.m0[key]

	return chunk
}

func (s *Server) SetChunkCell(
	cx int,
	cy int,
	cz int,
	cell *ChunkCell,
) {
	key := toPosStr(cx, cy, cz)
	s.m0[key] = cell
}
