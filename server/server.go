package server

import (
	"github.com/google/uuid"
	"net"
)

const (
	Addr     = ":9999"
	Type     = "tcp"
	Mc       = "1.12.2"
	Protocol = 340
	Max      = 10
	Text     = "Hello, World!"
	Favicon  = ""
)

type Server struct {
	online int
	last   int32
	m0     map[uuid.UUID]net.Conn
	m1     map[uuid.UUID]*Player
}

func NewServer() *Server {
	return &Server{
		online: 0,
		last:   0,
		m0:     make(map[uuid.UUID]net.Conn),
		m1:     make(map[uuid.UUID]*Player),
	}
}

func (s *Server) countLast() int32 {
	x := s.last
	s.last++
	return x
}

func (s *Server) Render() {
	ln, err := net.Listen(Type, Addr)
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
		go func() {

			addr := conn.RemoteAddr()
			lg := NewLogger("addr: %s", addr)

			defer func() {
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

			var player *Player

			switch state {
			case StatusState:
				for {
					finish, err := cnt.Loop1(state, s.online)
					if err != nil {
						panic(err)
					}
					if finish == false {
						continue
					}
					break
				}
				break
			case LoginState:
				for {
					finish, uid, username, err := cnt.Loop2(state)
					if err != nil {
						panic(err)
					}
					if finish == false {
						continue
					}

					eid := s.countLast()
					player = NewPlayer(
						eid,
						uid,
						username,
						0,
						0,
						0,
						0,
						0,
					)
					break
				}
				break
			}

			lg.InfoWithVars(
				"Player was created.",
				"player: %+V", player,
			)

			eid := player.GetEid()
			if err := cnt.f0(eid); err != nil {
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

		}()
	}

}

func (s *Server) GetOnline() int {
	return s.online
}
