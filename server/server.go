package server

import (
	"fmt"
	"io"
	"net"
)

func NewE0875(err error) error {
	return fmt.Errorf("[E0875] err: %+v", err)
}

const (
	Address         = ":9999"
	Type            = "tcp"
	VersionName     = "1.12.2"
	VersionProtocol = 340
	Max             = 10
	Text            = "Hello, World!"
	Favicon         = ""
)

func readVarInt(r io.Reader) (int, int32, error) {
	v := int32(0)
	position := uint8(0)
	n0 := 0

	for {
		buf := make([]uint8, 1)
		n1, err := r.Read(buf)
		if err != nil {
			return 0, 0, err
		}
		n0 += n1
		b := buf[0]
		v |= int32(b&SegmentBits) << position

		if (b & ContinueBit) == 0 {
			break
		}

		position += 7
	}

	return n0, v, nil
}

func read(r io.Reader) (int32, *Data, error) {
	var err error

	_, l0, err := readVarInt(r)
	if err != nil {
		return 0, nil, err
	}

	l1, id, err := readVarInt(r)
	if err != nil {

		return 0, nil, err
	}

	l2 := int(l0) - l1
	buf := make([]uint8, l2)
	if l2 == 0 {
		return id, nil, nil
	}
	_, err = r.Read(buf)
	if err != nil {
		return 0, nil, err
	}

	data := NewData(buf...)

	return id, data, err
}

type Server struct {
	online int
}

func NewServer() *Server {
	return &Server{
		online: 0,
	}
}

func h3(
	lg *Logger,
	id int32,
	data *Data,
) error {
	switch id {
	case StartLoginPacketID:
		startLoginPacket := NewStartLoginPacket()
		startLoginPacket.Read(data)
		lg.Info("startLoginPacket: %+v", startLoginPacket)
		username := startLoginPacket.GetUsername()
		_, err := UsernameToPlayerID(username)
		if err != nil {
			return err
		}
		break
		//case EncryptionResponsePacketID:
		//	break
	}

	return nil
}

func h2(
	lg *Logger,
	id int32,
	data *Data,
	online int,
) *Data {
	switch id {
	case RequestPacketID:
		requestPacket := NewRequestPacket()
		requestPacket.Read(data)
		lg.Info("requestPacket: %+v", requestPacket)

		jsonResponse := &JsonResponse{
			Version: &Version{
				Name:     VersionName,
				Protocol: VersionProtocol,
			},
			Players: &Players{
				Max:    Max,
				Online: online,
				Sample: []*Sample{},
			},
			Description: &Description{
				Text: Text,
			},
			Favicon:            Favicon,
			PreviewsChat:       false,
			EnforcesSecureChat: false,
		}
		responsePacket := NewResponsePacket(jsonResponse)
		lg.Info("responsePacket: %+v", responsePacket)
		data := responsePacket.Write()
		lg.Info("data: %+v", data)
		return data
	case PingPacketID:
		pingPacket := NewPingPacket()
		pingPacket.Read(data)
		lg.Info("pingPacket: %+v", pingPacket)
		payload := pingPacket.GetPayload()

		pongPacket := NewPongPacket(payload)
		lg.Info("pongPacket: %+v", pongPacket)
		data := pongPacket.Write()
		lg.Info("data: %+v", data)
		return data
	}

	return nil
}

func h1(
	lg *Logger,
	state int32,
	id int32,
	data *Data,
) int32 {
	switch id {
	case HandshakePacketID:
		handshakePacket := NewHandshakePacket()
		handshakePacket.Read(data)
		lg.Info("handshakePacket: %+v", handshakePacket)
		nextState := handshakePacket.GetNextState()
		state = nextState
		break
	}
	return state
}

func h0(
	lg *Logger,
	c net.Conn,
	online int,
) error {
	state := HandshakingState
	for {
		lg.Info("state: %d", state)
		id, data, err := read(c)
		lg.Info("id: %d, data: %+v", id, data)
		if err != nil {
			return err
		}
		//fmt.Println("playerID:", playerID)
		//fmt.Printf("data: %+v\n", data)

		switch state {
		case HandshakingState:
			state = h1(lg, state, id, data)
			break
		case StatusState:
			data := h2(lg, id, data, online)
			if _, err = c.Write(data.GetBuf()); err != nil {
				return err
			}
			break
		case LoginState:
			if err := h3(lg, id, data); err != nil {
				return err
			}
			break
		}
	}
}

func (s *Server) Render() {
	ln, err := net.Listen(Type, Address)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = ln.Close()
	}()

	for {
		c, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go func() {
			lg := NewLogger("address: %s", c.RemoteAddr())
			lg.Info("start")
			defer func() {
				_ = c.Close()
				lg.Info("close")
			}()
			if err := h0(lg, c, s.online); err == io.EOF {
				return
			} else if err != nil {
				lg.Error(err)
				return
			}
		}()
	}

}

func (s *Server) GetOnline() int {
	return s.online
}
