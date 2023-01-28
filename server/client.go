package server

import (
	"bytes"
	"errors"
	"github.com/google/uuid"
	"io"
	"math/rand"
	"net"
)

type Client struct {
	cid uuid.UUID

	conn net.Conn

	lg *Logger
}

var InvalidPacketIDError = errors.New("current packet ID was invalid")
var InvalidPayloadError = errors.New("payload does not match to given")
var UnknownPacketIDError = errors.New("current packet ID was unknown")
var LessThanThresholdError = errors.New("length of uncompressed id and data of packet is less than the threshold that set")

func readVarInt(
	r io.Reader,
) (
	int32,
	int,
	error,
) {
	v := int32(0)
	position := uint8(0)
	length := 0

	for {
		buf := make([]uint8, 1)
		if _, err := r.Read(buf); err != nil {
			return 0, 0, err
		}
		length += 1
		b := buf[0]
		//fmt.Println("b:", b)
		v |= int32(b&SegmentBits) << position

		if (b & ContinueBit) == 0 {
			break
		}

		position += 7
	}

	return v, length, nil
}

func writeVarInt(
	v int32,
	w io.Writer,
) (
	int,
	error,
) {
	v0 := uint32(v)
	arr := make([]uint8, 0)

	for {
		if (v0 & ^uint32(SegmentBits)) == 0 {
			b := uint8(v0)
			arr = append(arr, b)
			break
		}

		b := uint8(v0&uint32(SegmentBits)) | ContinueBit
		arr = append(arr, b)

		v0 >>= 7
	}

	if _, err := w.Write(arr); err != nil {
		return 0, err
	}
	return len(arr), nil
}

func read(
	length int, // length of packet
	r io.Reader,
) (int32, *Data, error) {
	pid, l0, err := readVarInt(r)
	if err != nil {
		return 0, nil, err
	}

	data := NewData()

	l1 := length - l0
	if l1 == 0 {
		return pid, data, nil
	}

	buf := make([]uint8, l1)
	if _, err = r.Read(buf); err != nil {
		return 0, nil, err
	}

	data.WriteBytes(buf)

	return pid, data, nil
}

func write(
	pid int32,
	data *Data,
) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil) // buffer of id and data of packet

	if _, err := writeVarInt(pid, buf); err != nil {
		return nil, err
	}

	if _, err := buf.Write(data.GetBytes()); err != nil {
		return nil, err
	}

	return buf, nil
}

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

func (cnt *Client) read() (
	int32,
	*Data,
	error,
) {
	conn := cnt.conn
	lg := cnt.lg

	l0, _, err := readVarInt(conn) // length of packet
	if err != nil {
		return 0, nil, err
	}

	pid, data, err := read(int(l0), conn)
	if err != nil {
		return 0, nil, err
	}

	lg.InfoWithVars(
		"Uninterpreted packet was read in the connection.",
		"id: %d", pid,
	)

	return pid, data, nil
}

func (cnt *Client) readWithComp() (
	int32,
	*Data,
	error,
) {
	conn := cnt.conn
	lg := cnt.lg

	l0, _, err := readVarInt(conn) // length of packet
	if err != nil {
		return 0, nil, err
	}
	//fmt.Println("l0:", l0)

	l1, l2, err := readVarInt(conn) // uncompressed length of id and data of packet
	if err != nil {
		return 0, nil, err
	}

	//fmt.Println("l1:", l1)
	//fmt.Println("l2:", l2)
	l3 := int(l0) - l2 // length of id and data of packet
	//fmt.Println("l3:", l3)
	if l1 == 0 {
		pid, data, err := read(l3, conn)
		if err != nil {
			return 0, nil, err
		}

		lg.InfoWithVars(
			"Uninterpreted packet was read "+
				"in the connection with non-compression.",
			"id: %d", pid,
		)

		return pid, data, nil
	} else if l1 < CompThold {
		return 0, nil, LessThanThresholdError
	}

	arr := make([]uint8, l3)
	if _, err = conn.Read(arr); err != nil {
		return 0, nil, err
	}

	buf, err := Uncompress(arr)
	if err != nil {
		return 0, nil, err
	}

	pid, _, err := readVarInt(buf)
	if err != nil {
		return 0, nil, err
	}

	data := NewData(buf.Bytes()...)

	lg.InfoWithVars(
		"Uninterpreted packet was read "+
			"in the connection with compression.",
		"id: %d", pid,
	)

	return pid, data, nil
}

func (cnt *Client) write(
	packet OutPacket,
) error {
	lg := cnt.lg
	conn := cnt.conn

	pid := packet.GetID()
	data := packet.Write()

	buf, err := write(pid, data)
	if err != nil {
		return err
	}
	arr := buf.Bytes()
	length := len(arr)
	if _, err := writeVarInt(int32(length), conn); err != nil {
		return err
	}
	if _, err := conn.Write(arr); err != nil {
		return err
	}

	lg.Info("Packet was wrote in the connection.")

	return nil
}

func (cnt *Client) writeWithComp(
	packet OutPacket,
) error {
	lg := cnt.lg
	conn := cnt.conn

	pid := packet.GetID()
	data := packet.Write()

	buf0, err := write(pid, data)
	if err != nil {
		return err
	}
	arr0 := buf0.Bytes()
	l0 := len(arr0) // length of packet before compression

	if l0 <= CompThold {
		buf1 := bytes.NewBuffer(nil)
		l1, err := writeVarInt(int32(0), buf1)
		if err != nil {
			return err
		}

		arr1 := buf1.Bytes()
		l2 := l0 + l1

		if _, err := writeVarInt(int32(l2), conn); err != nil {
			return err
		}
		if _, err := conn.Write(arr1); err != nil {
			return err
		}
		if _, err := conn.Write(arr0); err != nil {
			return err
		}

		lg.InfoWithVars(
			"Packet was wrote "+
				"in the connection with non-compression.",
			"packet: %+v", packet,
		)
		return nil
	}

	buf1, err := Compress(arr0)
	if err != nil {
		return err
	}
	arr1 := buf1.Bytes()
	l1 := len(arr1) // length of packet after compression

	buf2 := bytes.NewBuffer(nil)
	l2, err := writeVarInt(int32(l0), buf2)
	if err != nil {
		return err
	}
	arr2 := buf2.Bytes()

	l3 := l2 + l1
	if _, err := writeVarInt(int32(l3), conn); err != nil {
		return err
	}
	if _, err := conn.Write(arr2); err != nil {
		return err
	}
	if _, err := conn.Write(arr1); err != nil {
		return err
	}

	lg.InfoWithVars(
		"Packet was wrote "+
			"in the connection with compression.",
		"packet: %+v", packet,
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
			"packet: %+v", handshakePacket,
		)
		state = handshakePacket.GetNextState()
		break
	}

	return state, nil
}

func (cnt *Client) Loop1(
	state State,
	max int,
	online int,
	desc string,
	favicon string,
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
	pid, data, err := cnt.read()
	if err != nil {
		return true, err
	}

	finish := false

	switch pid {
	default:
		return true, UnknownPacketIDError
	case RequestPacketID:
		requestPacket := NewRequestPacket()
		requestPacket.Read(data)
		lg.InfoWithVars(
			"RequestPacket was read.",
			"packet: %+v", requestPacket,
		)

		jsonResponse := &JsonResponse{
			Version: &Version{
				Name:     McName,
				Protocol: ProtVer,
			},
			Players: &Players{
				Max:    max,
				Online: online,
				Sample: []*Sample{},
			},
			Description: &Description{
				Text: desc,
			},
			Favicon:            favicon,
			PreviewsChat:       false,
			EnforcesSecureChat: false,
		}
		responsePacket := NewResponsePacket(jsonResponse)
		lg.InfoWithVars(
			"ResponsePacket was created.",
			"packet: %+v", responsePacket,
		)
		if err := cnt.write(responsePacket); err != nil {
			return true, err
		}
		break
	case PingPacketID:
		pingPacket := NewPingPacket()
		pingPacket.Read(data)
		lg.InfoWithVars(
			"PingPacket was read.",
			"packet: %+v", pingPacket,
		)
		payload := pingPacket.GetPayload()

		pongPacket := NewPongPacket(payload)
		lg.InfoWithVars(
			"PongPacket was created.",
			"packet: %+v", pongPacket,
		)
		if err := cnt.write(pongPacket); err != nil {
			return true, err
		}
		finish = true
		break
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
			"packet: %+v", startLoginPacket,
		)

		username := startLoginPacket.GetUsername()
		lg.Info("Username starts converting to id.")
		uid, err := UsernameToUUID(username)
		if err != nil {
			return true, uuid.Nil, "", err
		}

		enableCompressionPacket := NewEnableCompressionPacket(CompThold)
		lg.InfoWithVars(
			"EnableCompressionPacket was created.",
			"packet: %+v", enableCompressionPacket,
		)
		if err := cnt.write(enableCompressionPacket); err != nil {
			return true, uuid.Nil, "", err
		}

		completeLoginPacket := NewCompleteLoginPacket(uid, username)
		lg.InfoWithVars(
			"CompleteLoginPacket was created.",
			"packet: %+v", completeLoginPacket,
		)
		if err := cnt.writeWithComp(completeLoginPacket); err != nil {
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

	pid, _, err := cnt.readWithComp()
	if err != nil {
		return true, err
	}

	switch pid {
	case ChangePlayerPosPacketID:
		break
	case ChangePlayerPosAndLookPacketID:
		break
	}

	return finish, nil
}

func (cnt *Client) Init(
	eid int32,
) error {
	lg := cnt.lg

	lg.Info("The normal sequence is started after login in the Init func.")
	defer func() {
		lg.Info("The normal sequence was finished in the Init func.")
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
			"packet: %+v", packet,
		)
		if err := cnt.writeWithComp(packet); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		id, data, err := cnt.readWithComp()
		if err != nil {
			return err
		}
		if id != ChangeClientSettingsPacketID {
			return InvalidPacketIDError
		}
		changeClientSettingsPacket := NewChangeClientSettingsPacket()
		changeClientSettingsPacket.Read(data)
		lg.InfoWithVars(
			"ChangeClientSettingsPacket was read.",
			"packet: %+v", changeClientSettingsPacket,
		)
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		_, _, err := cnt.readWithComp()
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
			"packet: %+v", packet,
		)
		if err := cnt.writeWithComp(packet); err != nil {
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
			"packet: %+v", packet,
		)
		if err := cnt.writeWithComp(packet); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		id, data, err := cnt.readWithComp()
		if err != nil {
			return err
		}
		if id != ConfirmTeleportPacketID {
			return InvalidPacketIDError
		}
		packet := NewConfirmTeleportPacket()
		packet.Read(data)
		lg.InfoWithVars(
			"ConfirmTeleportPacket was read.",
			"packet: %+v", packet,
		)
		payloadPrime := packet.GetPayload()
		if payload != payloadPrime {
			return InvalidPayloadError
		}
		return nil
	}(); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) LoadChunk(
	overworld bool,
	init bool,
	cx, cz int32,
	cc *Chunk,
) error {
	lg := cnt.lg

	lg.InfoWithVars(
		"The chunk column is started to loaded.",
		"cx: %d, cz: %d",
		cx, cz,
	)

	bitmask, d0 := cc.Write(init, overworld)
	packet := NewSendChunkDataPacket(
		cx,
		cz,
		init,
		bitmask,
		d0,
	)
	lg.InfoWithVars(
		"SendChunkDataPacket was created.",
		"packet: %+v", packet,
	)
	if err := cnt.writeWithComp(packet); err != nil {
		return err
	}

	lg.Info("The chunk column is finished to load.")
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
