package server

import (
	"io"
	"net"
)

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

type UnknownPacket struct {
	id   int32
	data *Data
}

func newUnknownPacket(
	id int32,
	data *Data,
) *UnknownPacket {
	return &UnknownPacket{
		id:   id,
		data: data,
	}
}

type Server struct {
	online int
}

func NewServer() *Server {
	return &Server{
		online: 0,
	}
}

func h0(
	c net.Conn,
	online int,
) {
	state := HandshakingState
	for {
		id, data, err := read(c)
		if err == io.EOF {
			return
		} else if err != nil {
			panic(err)
		}
		//fmt.Println("id:", id)
		//fmt.Printf("data: %+v\n", data)

		switch state {
		case HandshakingState:
			switch id {
			case HandshakePacketID:
				handshakePacket := NewHandshakePacket()
				handshakePacket.Read(data)
				nextState := handshakePacket.GetNextState()
				state = nextState
				break
			}
			break
		case StatusState:
			switch id {
			case RequestPacketID:
				requestPacket := NewRequestPacket()
				requestPacket.Read(data)

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
				data := responsePacket.Write()
				_, _ = c.Write(data.GetBuf())

				break
			case PingPacketID:
				pingPacket := NewPingPacket()
				pingPacket.Read(data)
				payload := pingPacket.GetPayload()

				pongPacket := NewPongPacket(payload)
				data := pongPacket.Write()
				_, _ = c.Write(data.GetBuf())
				return
			}
			break
		case LoginState:
			break
		}
	}
}

func (s *Server) Render() {
	ln, _ := net.Listen(Type, Address)
	defer func() {
		_ = ln.Close()
	}()

	for {
		c, _ := ln.Accept()
		go h0(c, s.online)
	}

}

func (s *Server) GetOnline() int {
	return s.online
}
