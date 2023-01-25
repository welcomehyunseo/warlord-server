package server

import (
	"errors"
	"io"
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
) (*Data, bool, error) {
	switch id {
	case StartLoginPacketID:
		startLoginPacket := NewStartLoginPacket()
		startLoginPacket.Read(data)
		lg.InfoWithVars("StartLoginPacket is read.", "packet: %+V", startLoginPacket)
		username := startLoginPacket.GetUsername()
		playerID, err := UsernameToPlayerID(username)
		if err != nil {
			return nil, false, err
		}
		completeLoginPacket := NewCompleteLoginPacket(playerID, username)
		lg.InfoWithVars("CompleteLoginPacket was created.", "packet: %+V", completeLoginPacket)
		data := completeLoginPacket.Write()
		lg.InfoWithVars("Data was wrote.", "data: %+V", data)
		return data, true, nil
		//case EncryptionResponsePacketID:
		//	break
	}

	return nil, false, errors.New("no Data")
}

func h2(
	lg *Logger,
	id int32,
	data *Data,
	online int,
) (*Data, bool) {
	switch id {
	case RequestPacketID:
		requestPacket := NewRequestPacket()
		requestPacket.Read(data)
		lg.InfoWithVars("RequestPacket was read.", "packet: %+V", requestPacket)

		jsonResponse := &JsonResponse{
			Version: &Version{
				Name:     Mc,
				Protocol: Protocol,
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
		lg.InfoWithVars("ResponsePacket was created.", "packet: %+V", responsePacket)
		data := responsePacket.Write()
		lg.InfoWithVars("Data was wrote.", "data: %+V", data)
		return data, false
	case PingPacketID:
		pingPacket := NewPingPacket()
		pingPacket.Read(data)
		lg.InfoWithVars("PingPacket was read.", "packet: %+V", pingPacket)
		payload := pingPacket.GetPayload()

		pongPacket := NewPongPacket(payload)
		lg.InfoWithVars("PongPacket was created.", "packet: %+V", pongPacket)
		data := pongPacket.Write()
		lg.InfoWithVars("Data was wrote.", "data: %+V", data)
		return data, true
	}

	return nil, true
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
		lg.InfoWithVars("HandshakePacket was read.", "packet: %+V", handshakePacket)
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
) (bool, error) {
	defer func() {
		lg.Info("The loop is ended.")
	}()
	state := HandshakingState
	for {
		lg.InfoWithVars("The loop is started with that state.", "state: %d, ", state)
		id, data, err := read(c)
		if err != nil {
			return false, err
		}
		lg.InfoWithVars("Unknown packet was read.", "id: %d, data: %+V", id, data)

		switch state {
		case HandshakingState:
			lg.Info("The HandshakingState handler is started.")
			state = h1(lg, state, id, data)
			break
		case StatusState:
			lg.Info("The StatusState handler is started.")
			data, finish := h2(lg, id, data, online)

			if _, err = c.Write(data.GetBuf()); err != nil {
				return false, err
			}
			if finish == true {
				return false, nil
			}
			break
		case LoginState:
			lg.Info("The LoginState handler is started.")
			data, success, err := h3(lg, id, data)
			if err != nil {
				return false, err
			}
			if _, err = c.Write(data.GetBuf()); err != nil {
				return false, err
			}
			if success == true {
				return true, nil
			}
			break
		}
	}
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
		c, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go func() {
			lg := NewLogger("addr: %s", c.RemoteAddr())
			lg.Info("The connection is started with logging.")

			defer func() {
				_ = c.Close()
				lg.Info("The connection is closed with logging.")
			}()

			success, err := h0(lg, c, s.online)
			if err == io.EOF {
				return
			} else if err != nil {
				lg.Error(err)
				return
			}
			if success == false {
				lg.Info("The login is failure.")
				return
			}

		}()
	}

}

func (s *Server) GetOnline() int {
	return s.online
}
