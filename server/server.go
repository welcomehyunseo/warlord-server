package server

import (
	"errors"
	"github.com/google/uuid"
	"io"
	"math/rand"
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

func read(
	lg *Logger,
	r io.Reader,
) (
	int32,
	*Data,
	error,
) {

	_, l0, err := readVarInt(r)
	if err != nil {
		return 0, nil, err
	}

	l1, pid, err := readVarInt(r)
	if err != nil {

		return 0, nil, err
	}

	l2 := int(l0) - l1
	buf := make([]uint8, l2)
	if l2 == 0 {
		return pid, nil, nil
	}
	_, err = r.Read(buf)
	if err != nil {
		return 0, nil, err
	}

	data := NewData(buf...)

	lg.InfoWithVars(
		"Uninterpreted packet was read.",
		"id: %d, data: %+v", pid, data,
	)

	return pid, data, nil
}

func send(
	lg *Logger,
	w io.Writer,
	data *Data,
) error {
	if _, err := w.Write(data.GetBuf()); err != nil {
		return err
	}
	lg.InfoWithVars("Data was sent.", "data: %+v", data)
	return nil
}

func h4(
	lg *Logger,
	c net.Conn,
) error {
	defer func() {
		lg.Info("The loop is ended in the handler h4.")
	}()

	for {
		_, _, err := read(lg, c)
		if err != nil {
			return err
		}

	}
}

func f1(
	lg *Logger,
	p *Player,
	c net.Conn,
) error {
	lg.Info("The connection is trying to start the normal sequence after login.")

	if err := func() error {
		packet := NewJoinGamePacket(
			p.GetEid(),
			1,
			0,
			2,
			"default",
			false,
		)
		lg.InfoWithVars(
			"JoinGamePacket was created.",
			"packet: %+V", packet,
		)
		data := packet.Write()
		if err := send(lg, c, data); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		id, data, err := read(lg, c)
		if err != nil {
			return err
		}
		if id != ChangeClientSettingsPacketID {
			return errors.New("packet must be ChangeClientSettingsPacket, but is not")
		}
		changeClientSettingsPacket := NewChangeClientSettingsPacket()
		changeClientSettingsPacket.Read(data)
		lg.InfoWithVars(
			"ChangeClientSettingsPacket was read.",
			"packet: %+V", changeClientSettingsPacket,
		)
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		_, _, err := read(lg, c)
		if err != nil {
			return err
		}
		// plugin message
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		packet := NewSetPlayerAbilitiesPacket(
			true,
			true,
			true,
			true,
			0.1,
			0.2,
		)
		lg.InfoWithVars(
			"SetPlayerAbilitiesPacket was created.",
			"packet: %+V", packet,
		)
		data := packet.Write()
		if err := send(lg, c, data); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	payload := rand.Int31()
	if err := func() error {
		packet := NewSetPlayerPosAndLookPacket(
			p.GetX(),
			p.GetY(),
			p.GetZ(),
			p.GetYaw(),
			p.GetPitch(),
			payload,
		)
		lg.InfoWithVars(
			"SetPlayerPosAndLookPacket was created.",
			"packet: %+V", packet,
		)
		data := packet.Write()
		if err := send(lg, c, data); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		id, data, err := read(lg, c)
		if err != nil {
			return err
		}
		if id != ConfirmTeleportPacketID {
			return errors.New("packet must be ConfirmTeleportPacket, but is not")
		}
		packet := NewConfirmTeleportPacket()
		packet.Read(data)
		lg.InfoWithVars(
			"ConfirmTeleportPacket was read.",
			"packet: %+V", packet,
		)
		payloadPrime := packet.GetPayload()
		if payload != payloadPrime {
			return errors.New(
				"the Payload value that read is not same the given",
			)
		}
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		cc := NewChunkColumn()
		cc.SetBiome(0, 0, VoidBiomeID)
		chunk := NewChunk()
		chunk.SetBlock(0, 0, 0, StoneBlock)
		cc.SetChunk(0, chunk)
		init := true
		overworld := true
		bitmask, d0 := cc.Write(init, overworld)

		packet := NewSendChunkDataPacket(
			0,
			0,
			init,
			bitmask,
			d0,
		)
		lg.InfoWithVars(
			"SendChunkDataPacket was created.",
			"packet: %+V", packet,
		)
		d1 := packet.Write()
		if err := send(lg, c, d1); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	lg.Info("The normal sequence was finished.")
	return nil
}

func f0(
	lg *Logger,
	c net.Conn,
) (
	uuid.UUID,
	string,
	error,
) {
	uid, username, err := func() (
		uuid.UUID,
		string,
		error,
	) {
		pid, data, err := read(lg, c)
		if pid != StartLoginPacketID {
			return uuid.Nil, "", errors.New(
				"packet that read must be StartLoginPacket, but is not",
			)
		}
		startLoginPacket := NewStartLoginPacket()
		startLoginPacket.Read(data)
		lg.InfoWithVars(
			"StartLoginPacket is read.",
			"packet: %+V", startLoginPacket,
		)
		username := startLoginPacket.GetUsername()
		uid, err := UsernameToUUID(username)
		if err != nil {
			return uuid.Nil, "", err
		}

		return uid, username, nil
	}()
	if err != nil {
		return uuid.Nil, "", err
	}

	if err := func() error {
		completeLoginPacket := NewCompleteLoginPacket(uid, username)
		lg.InfoWithVars(
			"CompleteLoginPacket was created.",
			"packet: %+V", completeLoginPacket,
		)
		data := completeLoginPacket.Write()
		if err := send(lg, c, data); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return uuid.Nil, "", err
	}
	return uid, username, nil

}

func h2(
	lg *Logger,
	pid int32,
	data *Data,
	online int,
) (*Data, bool) {
	switch pid {
	case RequestPacketID:
		requestPacket := NewRequestPacket()
		requestPacket.Read(data)
		lg.InfoWithVars(
			"RequestPacket was read.",
			"packet: %+V", requestPacket,
		)

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
		lg.InfoWithVars(
			"ResponsePacket was created.",
			"packet: %+V", responsePacket,
		)
		data := responsePacket.Write()
		return data, false
	case PingPacketID:
		pingPacket := NewPingPacket()
		pingPacket.Read(data)
		lg.InfoWithVars(
			"PingPacket was read.",
			"packet: %+V", pingPacket,
		)
		payload := pingPacket.GetPayload()

		pongPacket := NewPongPacket(payload)
		lg.InfoWithVars(
			"PongPacket was created.",
			"packet: %+V", pongPacket,
		)
		data := pongPacket.Write()
		return data, true
	}

	return nil, true
}

func h1(
	lg *Logger,
	state int32,
	pid int32,
	data *Data,
) int32 {
	switch pid {
	case HandshakePacketID:
		handshakePacket := NewHandshakePacket()
		handshakePacket.Read(data)
		lg.InfoWithVars(
			"HandshakePacket was read.",
			"packet: %+V", handshakePacket,
		)
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
		lg.Info("The loop is ended in the handler h0.")
	}()
	state := HandshakingState
	for {
		lg.InfoWithVars(
			"The loop is started with that state in the handler h0.",
			"state: %d, ", state,
		)
		pid, data, err := read(lg, c)
		if err != nil {
			return false, err
		}

		switch state {
		case HandshakingState:
			lg.Info("The HandshakingState handler h1 is started.")
			state = h1(lg, state, pid, data)
			if state == LoginState {
				return true, nil
			}
			break
		case StatusState:
			lg.Info("The StatusState handler h2 is started.")
			data, finish := h2(lg, pid, data, online)
			if err := send(lg, c, data); err != nil {
				return false, err
			}
			if finish == true {
				return false, nil
			}
			break
		}
	}
}

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

			login, err := h0(lg, c, s.online)
			if err == io.EOF {
				return
			} else if err != nil {
				lg.Error(err)
				return
			} else if login == false {
				return
			}

			lg.Info("The login sequence is started.")
			uid, username, err := f0(lg, c)
			if err != nil {
				lg.Error(err)
				lg.Info("The login sequence is failure.")
				return
			}

			s.online++
			defer func() {
				s.online--
			}()
			eid := s.last
			s.last++
			p := NewPlayer(eid, uid, username, 0, 0, 0, 0, 0)
			s.m0[uid] = c
			s.m1[uid] = p

			if err := f1(lg, p, c); err == io.EOF {
				return
			} else if err != nil {
				lg.Error(err)
				return
			}

			if err := h4(lg, c); err == io.EOF {
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
