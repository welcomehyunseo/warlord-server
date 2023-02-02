package server

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"math/rand"
	"net"
)

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

type Client struct {
	cid  uuid.UUID
	addr net.Addr

	conn net.Conn
}

func NewClient(
	cid uuid.UUID,
	conn net.Conn,
) *Client {
	addr := conn.RemoteAddr()

	return &Client{
		cid:  cid,
		addr: addr,
		conn: conn,
	}
}

func (cnt *Client) read(
	lg *Logger,
) (
	int32,
	*Data,
	error,
) {
	lg.Debug("It is started to read a packet.")

	conn := cnt.conn

	l0, _, err := readVarInt(conn) // length of packet
	if err != nil {
		return 0, nil, err
	}

	pid, data, err := read(int(l0), conn)
	if err != nil {
		return 0, nil, err
	}

	lg.Debug("It is finished to read a packet.")
	return pid, data, nil
}

func (cnt *Client) readWithComp(
	lg *Logger,
) (
	int32,
	*Data,
	error,
) {
	lg.Debug(
		"It is started to read a packet with compression.",
	)

	conn := cnt.conn

	l0, _, err := readVarInt(conn) // length of packet
	if err != nil {
		return 0, nil, err
	}

	l1, l2, err := readVarInt(conn) // uncompressed length of id and data of packet
	if err != nil {
		return 0, nil, err
	}

	l3 := int(l0) - l2 // length of id and data of packet
	if l1 == 0 {
		pid, data, err := read(l3, conn)
		if err != nil {
			return 0, nil, err
		}

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

	lg.Debug(
		"It is finished to read a packet with compression.",
	)
	return pid, data, nil
}

func (cnt *Client) write(
	lg *Logger,
	packet OutPacket,
) error {
	lg.Debug(
		"It is started to generateData the packet.",
		NewLgElement("packet", packet),
	)

	conn := cnt.conn

	pid := packet.GetID()
	data := packet.Pack()

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

	lg.Debug(
		"It is finished to generateData the packet.",
	)
	return nil
}

func (cnt *Client) writeWithComp(
	lg *Logger,
	packet OutPacket,
) error {
	lg.Debug(
		"It is started to generateData the packet with compression.",
		NewLgElement("packet", packet),
	)

	conn := cnt.conn

	pid := packet.GetID()
	data := packet.Pack()

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

	lg.Debug(
		"It is finished to generateData the packet with compression.",
	)
	return nil
}

func (cnt *Client) Loop0(
	lg *Logger,
	state State,
) (
	State,
	error,
) {
	lg.Debug(
		"The sequence of Loop0 is started.",
		NewLgElement("state", state),
	)

	pid, data, err := cnt.read(lg)
	if err != nil {
		return NilState, err
	}

	switch pid {
	default:
		return NilState, UnknownPacketIDError
	case HandshakePacketID:
		packet := NewHandshakePacket()
		packet.Unpack(data)
		lg.Debug(
			"HandshakePacket was created.",
			NewLgElement("packet", packet),
		)
		state = packet.GetNext()
		break
	}

	lg.Debug(
		"The sequence of Loop0 is finished.",
	)
	return state, nil
}

func (cnt *Client) Loop1(
	lg *Logger,
	state State,
	max int,
	online int,
	text string,
	favicon string,
) (
	bool,
	error,
) {
	lg.Debug(
		"The sequence of Loop1 is started.",
		NewLgElement("state", state),
	)

	pid, data, err := cnt.read(lg)
	if err != nil {
		return true, err
	}

	stop := false

	switch pid {
	default:
		return true, UnknownPacketIDError
	case RequestPacketID:
		packet0 := NewRequestPacket()
		packet0.Unpack(data)
		lg.Debug(
			"RequestPacket was created.",
			NewLgElement("packet", packet0),
		)
		packet1 := NewResponsePacket(max, online, text, favicon)
		lg.Debug(
			"ResponsePacket was created.",
			NewLgElement("packet", packet1),
		)
		if err := cnt.write(lg, packet1); err != nil {
			return true, err
		}
		break
	case PingPacketID:
		packet0 := NewPingPacket()
		packet0.Unpack(data)
		lg.Debug(
			"PingPacket was created.",
			NewLgElement("packet", packet0),
		)
		payload := packet0.GetPayload()

		packet1 := NewPongPacket(payload)
		lg.Debug(
			"PingPacket was created.",
			NewLgElement("packet", packet1),
		)
		if err := cnt.write(lg, packet1); err != nil {
			return true, err
		}
		stop = true
		break
	}

	lg.Debug(
		"The sequence of Loop1 is finished.",
		NewLgElement("stop", stop),
	)
	return stop, nil
}

func (cnt *Client) Loop2(
	lg *Logger,
	state State,
) (
	bool,
	uuid.UUID,
	string,
	error,
) {
	lg.Debug(
		"The sequence of Loop2 is started.",
		NewLgElement("state", state),
	)

	pid, d0, err := cnt.read(lg)
	if err != nil {
		return true, uuid.Nil, "", err
	}

	switch pid {
	default:
		return true, uuid.Nil, "", UnknownPacketIDError
	case StartLoginPacketID:
		packet0 := NewStartLoginPacket()
		packet0.Unpack(d0)
		lg.Debug(
			"StartLoginPacket was created.",
			NewLgElement("packet", packet0),
		)

		username := packet0.GetUsername()
		uid, err := UsernameToUUID(username)
		if err != nil {
			return true, uuid.Nil, "", err
		}
		lg.Debug(
			"It is finished to convert username to UUID.",
			NewLgElement("uid", uid),
		)

		packet1 := NewEnableCompPacket(CompThold)
		lg.Debug(
			"EnableCompPacket was created.",
			NewLgElement("packet", packet1),
		)
		if err := cnt.write(lg, packet1); err != nil {
			return true, uuid.Nil, "", err
		}

		packet2 := NewCompleteLoginPacket(uid, username)
		lg.Debug(
			"CompleteLoginPacket was created.",
			NewLgElement("packet", packet2),
		)
		if err := cnt.writeWithComp(lg, packet2); err != nil {
			return true, uuid.Nil, "", err
		}

		lg.Debug(
			"The sequence of Loop2 is finished.",
		)
		return true, uid, username, nil
	}
}

func (cnt *Client) Loop3(
	lg *Logger,
	chanForUpdatePosEvent ChanForUpdatePosEvent,
	chanForConfirmKeepAliveEvent ChanForConfirmKeepAliveEvent,
	state State,
) (
	bool, // finish
	error,
) {
	lg.Debug(
		"The sequence of Loop3 is started.",
		NewLgElement("state", state),
	)

	stop := false

	pid, data, err := cnt.readWithComp(lg)
	if err != nil {
		return true, err
	}

	switch pid {
	case ConfirmKeepAlivePacketID:
		packet := NewConfirmKeepAlivePacket()
		packet.Unpack(data)
		lg.Debug(
			"ConfirmKeepAlivePacket was created.",
			NewLgElement("packet", packet),
		)
		payload := packet.GetPayload()
		chanForConfirmKeepAliveEvent <- NewConfirmKeepAliveEvent(payload)
		break
	case ChangePosPacketID:
		packet := NewChangePlayerPosPacket()
		packet.Unpack(data)
		lg.Debug(
			"ChangePosPacket was created.",
			NewLgElement("packet", packet),
		)
		x, y, z :=
			packet.GetX(), packet.GetY(), packet.GetZ()
		chanForUpdatePosEvent <- NewUpdatePosEvent(
			x, y, z,
		)
		break
	case ChangePosAndLookPacketID:
		packet := NewChangePosAndLookPacket()
		packet.Unpack(data)
		lg.Debug(
			"ChangePosAndLookPacket was created.",
			NewLgElement("packet", packet),
		)
		x, y, z :=
			packet.GetX(), packet.GetY(), packet.GetZ()
		chanForUpdatePosEvent <- NewUpdatePosEvent(
			x, y, z,
		)
		break
	}

	lg.Debug(
		"The sequence of Loop3 is finished.",
		NewLgElement("stop", stop),
	)
	return stop, nil
}

func (cnt *Client) Init(
	lg *Logger,
	eid int32,
	spawnX, spawnY, spawnZ float64,
	spawnYaw, spawnPitch float32,
) error {
	lg.Debug(
		"It is started to init.",
		NewLgElement("eid", eid),
	)
	if err := func() error {
		packet := NewJoinGamePacket(
			eid,
			1,
			0,
			2,
			"default",
			false,
		)
		lg.Debug(
			"JoinGamePacket was created.",
			NewLgElement("packet", packet),
		)
		if err := cnt.writeWithComp(lg, packet); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		id, data, err := cnt.readWithComp(lg)
		if err != nil {
			return err
		}
		if id != ChangeSettingsPacketID {
			return InvalidPacketIDError
		}
		packet := NewChangeSettingsPacket()
		packet.Unpack(data)
		lg.Debug(
			"ChangeSettingsPacket was created.",
			NewLgElement("packet", packet),
		)
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		_, _, err := cnt.readWithComp(lg)
		if err != nil {
			return err
		}
		// plugin message
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		packet := NewSetAbilitiesPacket(
			false,
			false,
			false,
			false,
			0,
			0,
		)
		lg.Debug(
			"SetAbilitiesPacket was created.",
			NewLgElement("packet", packet),
		)
		if err := cnt.writeWithComp(lg, packet); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	payload := rand.Int31()
	if err := func() error {
		packet := NewTeleportPacket(
			spawnX, spawnY, spawnZ,
			spawnYaw, spawnPitch,
			payload,
		)
		lg.Debug(
			"TeleportPacket was created.",
			NewLgElement("packet", packet),
		)
		if err := cnt.writeWithComp(lg, packet); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		id, data, err := cnt.readWithComp(lg)
		if err != nil {
			return err
		}
		if id != FinishTeleportPacketID {
			return InvalidPacketIDError
		}
		packet := NewFinishTeleportPacket()
		packet.Unpack(data)
		lg.Debug(
			"FinishTeleportPacket was created.",
			NewLgElement("packet", packet),
		)
		payloadPrime := packet.GetPayload()
		if payload != payloadPrime {
			return InvalidPayloadError
		}
		return nil
	}(); err != nil {
		return err
	}

	lg.Debug("It is finished to init.")
	return nil
}

func (cnt *Client) CheckKeepAlive(
	lg *Logger,
	payload int64,
) error {
	lg.Debug(
		"It is started to check keep-alive of player.",
		NewLgElement("payload", payload),
	)

	packet := NewCheckKeepAlivePacket(payload)
	if err := cnt.writeWithComp(lg, packet); err != nil {
		return err
	}

	lg.Debug("It is finished to check keep-alive of player.")
	return nil
}

func (cnt *Client) LoadChunk(
	lg *Logger,
	overworld, init bool,
	cx, cz int32,
	chunk *Chunk,
) error {
	lg.Debug(
		"It is started to load chunk.",
		NewLgElement("overworld", overworld),
		NewLgElement("init", init),
		NewLgElement("cx", cx),
		NewLgElement("cz", cz),
		NewLgElement("chunk", chunk),
	)
	bitmask, data := chunk.GenerateData(init, overworld)
	lg.Debug(
		"It was finished to generateData data.",
		NewLgElement("bitmask", bitmask),
		NewLgElement("data", "[...]"),
	)
	packet := NewSendChunkDataPacket(
		cx, cz,
		init,
		bitmask,
		data,
	)
	if err := cnt.writeWithComp(lg, packet); err != nil {
		return err
	}

	lg.Debug("It is finished to load chunk.")
	return nil
}

func (cnt *Client) UnloadChunk(
	lg *Logger,
	cx, cz int32,
) error {
	lg.Debug(
		"It is started to unload chunk.",
	)

	packet := NewUnloadChunkPacket(
		cx, cz,
	)
	if err := cnt.writeWithComp(lg, packet); err != nil {
		return err
	}
	lg.Debug(
		"It is finished to unload chunk.",
	)
	return nil
}

func (cnt *Client) AddPlayer(
	lg *Logger,
	uid uuid.UUID,
	username string,
) error {
	lg.Debug(
		"It is started to add player.",
	)

	textureString, signature, err := UUIDToTextureString(uid)
	if err != nil {
		return err
	}
	gamemode := int32(0)
	ping := int32(1000)
	displayName := username
	packet := NewAddPlayerPacket(
		uid,
		username,
		textureString,
		signature,
		gamemode,
		ping,
		displayName,
	)
	if err := cnt.writeWithComp(lg, packet); err != nil {
		return err
	}

	lg.Debug(
		"It is finished to add player.",
	)

	return nil
}

func (cnt *Client) RemovePlayer(
	lg *Logger,
	uid uuid.UUID,
) error {
	lg.Debug(
		"It is started to remove player",
	)

	packet := NewRemovePlayerPacket(
		uid,
	)
	if err := cnt.writeWithComp(lg, packet); err != nil {
		return err
	}

	lg.Debug(
		"It is finished to remove player",
	)

	return nil
}

func (cnt *Client) UpdateLatency(
	lg *Logger,
	uid uuid.UUID,
	latency int32,
) error {
	lg.Debug(
		"It is started to update latency",
	)

	packet := NewUpdateLatencyPacket(
		uid,
		latency,
	)
	if err := cnt.writeWithComp(lg, packet); err != nil {
		return err
	}

	lg.Debug(
		"It is finished to update latency",
	)

	return nil
}

func (cnt *Client) SpawnPlayer(
	lg *Logger,
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
) error {
	lg.Debug(
		"It is started to spawn player",
	)

	packet := NewSpawnPlayerPacket(
		eid, uid,
		x, y, z,
		yaw, pitch,
	)
	if err := cnt.writeWithComp(lg, packet); err != nil {
		return err
	}

	lg.Debug(
		"It is finished to spawn player",
	)

	return nil
}

func (cnt *Client) RelativeMove(
	lg *Logger,
	eid int32,
	deltaX, deltaY, deltaZ int16,
	ground bool,
) error {
	lg.Debug(
		"It is started to move relatively.",
	)

	packet := NewRelativeMovePacket(
		eid,
		deltaX, deltaY, deltaZ,
		ground,
	)
	if err := cnt.writeWithComp(lg, packet); err != nil {
		return err
	}

	lg.Debug(
		"It is finished to move relatively.",
	)

	return nil
}

func (cnt *Client) DespawnEntity(
	lg *Logger,
	eid int32,
) error {
	lg.Debug(
		"It is started to despawn player",
	)

	packet := NewDespawnEntityPacket(eid)
	if err := cnt.writeWithComp(lg, packet); err != nil {
		return err
	}

	lg.Debug(
		"It is finished to despawn player",
	)

	return nil
}

func (cnt *Client) Close(
	lg *Logger,
) {
	lg.Info("Client is closed.")

	_ = cnt.conn.Close()
}

func (cnt *Client) GetCID() uuid.UUID {
	return cnt.cid
}

func (cnt *Client) String() string {
	return fmt.Sprintf(
		"{ cid: %s, addr: %s }",
		cnt.cid, cnt.addr,
	)
}
