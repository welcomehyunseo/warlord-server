package server

import (
	"errors"
	"github.com/google/uuid"
	"math/rand"
	"net"
)

type Client struct {
	cid uuid.UUID

	conn net.Conn

	lg *Logger
}

var UnknownPacketIDError = errors.New("current packet ID was unknown")

func NewClient(
	cid uuid.UUID,
	conn net.Conn,
	lg *Logger,
) *Client {

	lg.Info("Client was started with logging.")

	return &Client{
		cid:  cid,
		conn: conn,
		lg:   lg,
	}
}

func (cnt *Client) readVarInt() (
	int32,
	int,
	error,
) {
	conn := cnt.conn

	v := int32(0)
	position := uint8(0)
	length := 0

	for {
		buf := make([]uint8, 1)
		n, err := conn.Read(buf)
		if err != nil {
			return 0, 0, err
		}
		length += n
		b := buf[0]
		v |= int32(b&SegmentBits) << position

		if (b & ContinueBit) == 0 {
			break
		}

		position += 7
	}

	return v, length, nil
}

func (cnt *Client) read() (
	int32,
	*Data,
	error,
) {
	conn := cnt.conn
	lg := cnt.lg

	d0 := NewData()

	l0, _, err := cnt.readVarInt()
	if err != nil {
		return 0, nil, err
	}

	pid, l1, err := cnt.readVarInt()
	if err != nil {
		return 0, nil, err
	}

	l2 := int(l0) - l1
	buf := make([]uint8, l2)
	if l2 == 0 {
		lg.InfoWithVars(
			"Uninterpreted packet was read in the network.",
			"id: %d", pid,
		)
		return pid, d0, nil
	}
	_, err = conn.Read(buf)
	if err != nil {
		return 0, nil, err
	}

	d0.WriteBuf(buf)

	lg.InfoWithVars(
		"Uninterpreted packet was read in the network with Data.",
		"id: %d, data: %+v", pid, d0,
	)

	return pid, d0, nil
}

func (cnt *Client) write(
	data *Data,
) error {
	lg := cnt.lg
	conn := cnt.conn

	if _, err := conn.Write(data.GetBuf()); err != nil {
		return err
	}

	lg.InfoWithVars(
		"Data was wrote to the network.",
		"data: %+v", data,
	)

	return nil
}

func (cnt *Client) Loop0(
	state State,
) (
	State,
	error,
) {
	lg := cnt.lg

	defer func() {
		lg.Info("The Loop0 loop is ended.")
	}()

	lg.InfoWithVars(
		"The Loop0 loop is started with that state.",
		"state: %d", state,
	)
	pid, data, err := cnt.read()
	if err != nil {
		return NilState, err
	}

	switch pid {
	default:
		return NilState, UnknownPacketIDError
	case HandshakePacketID:
		handshakePacket := NewHandshakePacket()
		handshakePacket.Read(data)
		lg.InfoWithVars(
			"HandshakePacket was read.",
			"packet: %+V", handshakePacket,
		)
		state = handshakePacket.GetNextState()
		break
	}

	return state, nil
}

func (cnt *Client) Loop1(
	state State,
	online int,
) (
	bool,
	error,
) {
	lg := cnt.lg

	defer func() {
		lg.Info("The Loop1 loop is ended.")
	}()

	lg.InfoWithVars(
		"The Loop1 loop is started with that state.",
		"state: %d", state,
	)
	pid, d0, err := cnt.read()
	if err != nil {
		return true, err
	}

	finish := false
	var d1 *Data

	switch pid {
	default:
		return true, UnknownPacketIDError
	case RequestPacketID:
		requestPacket := NewRequestPacket()
		requestPacket.Read(d0)
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
		d1 = responsePacket.Write()
		break
	case PingPacketID:
		pingPacket := NewPingPacket()
		pingPacket.Read(d0)
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
		d1 = pongPacket.Write()
		finish = true
		break
	}

	if err := cnt.write(d1); err != nil {
		return true, err
	}

	return finish, nil
}

func (cnt *Client) Loop2(
	state State,
) (
	bool,
	uuid.UUID,
	string,
	error,
) {
	lg := cnt.lg

	defer func() {
		lg.Info("The Loop2 loop is ended.")
	}()

	lg.InfoWithVars(
		"The Loop2 loop is started with that state.",
		"state: %d", state,
	)
	pid, d0, err := cnt.read()
	if err != nil {
		return true, uuid.Nil, "", err
	}

	switch pid {
	default:
		return true, uuid.Nil, "", UnknownPacketIDError
	case StartLoginPacketID:
		startLoginPacket := NewStartLoginPacket()
		startLoginPacket.Read(d0)
		lg.InfoWithVars(
			"StartLoginPacket was read.",
			"packet: %+V", startLoginPacket,
		)

		username := startLoginPacket.GetUsername()
		lg.Info("Username starts converting to id.")
		uid, err := UsernameToUUID(username)
		if err != nil {
			return true, uuid.Nil, "", err
		}

		completeLoginPacket := NewCompleteLoginPacket(uid, username)
		lg.InfoWithVars(
			"CompleteLoginPacket was created.",
			"packet: %+V", completeLoginPacket,
		)
		data := completeLoginPacket.Write()
		if err := cnt.write(data); err != nil {
			return true, uuid.Nil, "", err
		}

		return true, uid, username, nil
	}
}

func (cnt *Client) Loop3(
	state State,
) (
	bool,
	error,
) {
	lg := cnt.lg

	defer func() {
		lg.Info("The Loop3 loop is ended.")
	}()

	lg.InfoWithVars(
		"The Loop3 loop is started with that state.",
		"state: %d", state,
	)

	finish := false

	pid, _, err := cnt.read()
	if err != nil {
		return true, err
	}

	switch pid {
	}

	return finish, nil
}

func (cnt *Client) f0(
	eid int32,
) error {
	lg := cnt.lg

	lg.Info("The normal sequence is started after login in the f0 func.")
	defer func() {
		lg.Info("The normal sequence was finished in the f0 func.")
	}()
	if err := func() error {
		packet := NewJoinGamePacket(
			eid,
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
		if err := cnt.write(data); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		id, data, err := cnt.read()
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
		_, _, err := cnt.read()
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
		if err := cnt.write(data); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	payload := rand.Int31()
	if err := func() error {
		packet := NewSetPlayerPosAndLookPacket(
			0,
			0,
			0,
			0,
			0,
			payload,
		)
		lg.InfoWithVars(
			"SetPlayerPosAndLookPacket was created.",
			"packet: %+V", packet,
		)
		data := packet.Write()
		if err := cnt.write(data); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		id, data, err := cnt.read()
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
		if err := cnt.write(d1); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) Close() {
	lg := cnt.lg

	_ = cnt.conn.Close()
	lg.Info("Client was ended with logging.")
}

func (cnt *Client) GetCID() uuid.UUID {
	return cnt.cid
}
