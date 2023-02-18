package server

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"math/rand"
	"net"
	"sync"
)

type CID = uuid.UUID

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
) (
	int32,
	*Data,
	error,
) {
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

	if err := data.WriteBytes(buf); err != nil {
		return 0, nil, err
	}

	return pid, data, nil
}

func write(
	pid int32,
	data *Data,
) (
	*bytes.Buffer,
	error,
) {
	buf := bytes.NewBuffer(nil) // buffer of id and data of packet

	if _, err := writeVarInt(pid, buf); err != nil {
		return nil, err
	}

	if _, err := buf.Write(data.GetBytes()); err != nil {
		return nil, err
	}

	return buf, nil
}

func distribute(
	state State,
	pid int32,
	data *Data,
) (
	InPacket,
	error,
) {
	var inPacket InPacket

	switch state {
	case PlayState:
		switch pid {
		case FinishTeleportPacketID:
			inPacket = NewFinishTeleportPacket()
			break
		case ChangeSettingsPacketID:
			inPacket = NewChangeSettingsPacket()
			break
		case ConfirmKeepAlivePacketID:
			inPacket = NewConfirmKeepAlivePacket()
			break
		case ChangePosPacketID:
			inPacket = NewChangePosPacket()
			break
		case ChangePosAndLookPacketID:
			inPacket = NewChangePosAndLookPacket()
			break
		case ChangeLookPacketID:
			inPacket = NewChangeLookPacket()
			break
		}
		break
	case StatusState:
		switch pid {
		case RequestPacketID:
			inPacket = NewRequestPacket()
			break
		case PingPacketID:
			inPacket = NewPingPacket()
			break
		}
		break
	case LoginState:
		switch pid {
		case StartLoginPacketID:
			inPacket = NewStartLoginPacket()
			break
		}
		break
	case HandshakingState:
		switch pid {
		case HandshakePacketID:
			inPacket = NewHandshakePacket()
			break
		}
		break
	}
	if inPacket == nil {
		return nil, nil
	}

	if err := inPacket.Unpack(data); err != nil {
		return nil, err
	}
	return inPacket, nil
}

type Client struct {
	sync.Mutex

	cid uuid.UUID

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

func (cnt *Client) Read(
	state State,
) (
	InPacket,
	error,
) {
	cnt.Lock()
	defer cnt.Unlock()

	conn := cnt.conn

	l0, _, err := readVarInt(conn) // length of packet
	if err != nil {
		return nil, err
	}

	pid, data, err := read(int(l0), conn)
	if err != nil {
		return nil, err
	}

	return distribute(
		state,
		pid,
		data,
	)
}

func (cnt *Client) ReadWithComp(
	state State,
) (
	InPacket,
	error,
) {
	cnt.Lock()
	defer cnt.Unlock()

	conn := cnt.conn

	l0, _, err := readVarInt(conn) // length of packet
	if err != nil {
		return nil, err
	}

	l1, l2, err := readVarInt(conn) // uncompressed length of id and data of packet
	if err != nil {
		return nil, err
	}

	l3 := int(l0) - l2 // length of id and data of packet
	if l1 == 0 {
		pid, data, err := read(l3, conn)
		if err != nil {
			return nil, err
		}

		return distribute(
			state,
			pid,
			data,
		)
	} else if l1 < CompThold {
		return nil, errors.New("length of uncompressed id and data of packet is less than the threshold that set to read packet with compression in client")
	}

	arr := make([]uint8, l3)
	if _, err = conn.Read(arr); err != nil {
		return nil, err
	}

	buf, err := Uncompress(arr)
	if err != nil {
		return nil, err
	}

	pid, _, err := readVarInt(buf)
	if err != nil {
		return nil, err
	}

	data := NewData(buf.Bytes()...)

	return distribute(
		state,
		pid,
		data,
	)
}

func (cnt *Client) Write(
	packet OutPacket,
) error {
	cnt.Lock()
	defer cnt.Unlock()

	pid := packet.GetID()
	data, err := packet.Pack()
	if err != nil {
		return err
	}

	buf0, err := write(pid, data)
	if err != nil {
		return err
	}
	buf1 := buf0.Bytes()
	length := len(buf1)
	conn := cnt.conn
	if _, err := writeVarInt(int32(length), conn); err != nil {
		return err
	}
	if _, err := conn.Write(buf1); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) WriteWithComp(
	packet OutPacket,
) error {
	cnt.Lock()
	defer cnt.Unlock()

	conn := cnt.conn

	pid := packet.GetID()
	data, err := packet.Pack()
	if err != nil {
		return err
	}

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

	return nil
}

func (cnt *Client) HandleNonLoginState(
	lg *Logger,
	max, online int,
	text, favicon string,
) (
	bool, // stop
	error,
) {
	lg.Debug("it is started to handle non login state in client")
	defer func() {
		lg.Debug("it is finished to handle non login state in client")
	}()

	state := HandshakingState

	for {
		inPacket, err := cnt.Read(state)
		if err != nil {
			return false, err
		}

		lg.Debug(
			"client read packet",
			NewLgElement("InPacket", inPacket),
		)

		var outPacket OutPacket

		switch inPacket.(type) {
		case *HandshakePacket:
			handshakePacket := inPacket.(*HandshakePacket)
			state = handshakePacket.GetNext()
			break
		case *RequestPacket:
			responsePacket := NewResponsePacket(
				max, online, text, favicon,
			)
			outPacket = responsePacket
			break
		case *PingPacket:
			pingPacket := inPacket.(*PingPacket)
			payload := pingPacket.GetPayload()
			pongPacket := NewPongPacket(payload)
			outPacket = pongPacket
			break
		}

		if state == LoginState {
			return false, nil
		}

		if outPacket == nil {
			continue
		}

		if err := cnt.Write(outPacket); err != nil {
			return false, err
		}
		lg.Debug(
			"client sent packet",
			NewLgElement("OutPacket", outPacket),
		)

		if _, ok := outPacket.(*PongPacket); ok == true {
			return true, nil
		}
	}
}

func (cnt *Client) HandleLoginState(
	lg *Logger,
) (
	UID,
	string, // username
	error,
) {
	lg.Debug("it is started to handle login state in client")
	defer func() {
		lg.Debug("it is finished to handle login state in client")
	}()

	state := LoginState
	inPacket, err := cnt.Read(state)
	if err != nil {
		return NilUID, "", err
	}

	startLoginPacket, ok := inPacket.(*StartLoginPacket)
	if ok == false {
		return NilUID, "", errors.New("it is invalid inbound packet to handle login state")
	}
	username := startLoginPacket.GetUsername()
	uid, err := UsernameToUUID(username)
	if err != nil {
		return NilUID, "", err
	}

	enableCompPacket := NewEnableCompPacket(CompThold)
	if err := cnt.Write(enableCompPacket); err != nil {
		return NilUID, "", err
	}

	completeLoginPacket := NewCompleteLoginPacket(
		uid,
		username,
	)
	if err := cnt.WriteWithComp(completeLoginPacket); err != nil {
		return NilUID, "", err
	}

	return uid, username, nil
}

func (cnt *Client) JoinGame(
	lg *Logger,
	eid EID,
	spawnX, spawnY, spawnZ float64,
	spawnYaw, spawnPitch float32,
) error {
	lg.Debug("it is started to join game in client")
	defer func() {
		lg.Debug("it is finished to join game in client")
	}()

	state := PlayState
	joinGamePacket := NewJoinGamePacket(
		eid,
		0,
		0,
		2,
		"default",
		false,
	)
	if err := cnt.WriteWithComp(joinGamePacket); err != nil {
		return err
	}

	// ChangeSettingsPacket
	if _, err := cnt.ReadWithComp(state); err != nil {
		return err
	}

	// Plugin message
	if _, err := cnt.ReadWithComp(state); err != nil {
		return err
	}

	setAbilitiesPacket := NewSetAbilitiesPacket(
		false,
		false,
		false,
		false,
		0,
		0,
	)
	if err := cnt.WriteWithComp(setAbilitiesPacket); err != nil {
		return err
	}

	payload := rand.Int31()
	teleportPacket := NewTeleportPacket(
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
		payload,
	)
	if err := cnt.WriteWithComp(teleportPacket); err != nil {
		return err
	}

	inPacket2, err := cnt.ReadWithComp(state)
	if err != nil {
		return err
	}
	finishTeleportPacket, ok := inPacket2.(*FinishTeleportPacket)
	if ok == false {
		return errors.New("it is invalid packet to init play state")
	}
	payload1 := finishTeleportPacket.GetPayload()
	if payload != payload1 {
		return errors.New("it is invalid payload of FinishTeleportPacket to init play state")
	}

	chunk := NewChunk(0, 0)
	part := NewChunkPart()
	for z := 0; z < ChunkPartWidth; z++ {
		for x := 0; x < ChunkPartWidth; x++ {
			part.SetBlock(uint8(x), 0, uint8(z), StoneBlock)
		}
	}
	chunk.SetChunkPart(4, part)
	bitmask, data := chunk.GenerateData(true, true)
	sendChunkDataPacket := NewSendChunkDataPacket(
		0, 0,
		true,
		bitmask, data,
	)
	if err := cnt.WriteWithComp(sendChunkDataPacket); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) LoopForPlayState(
	lg *Logger,
	player *Player,
	chanForConfirmKeepAliveEvent ChanForConfirmKeepAliveEvent,
) error {
	lg.Debug("it is started to loop for play state in client")
	defer func() {
		lg.Debug("it is finished to loop for play state in client")
	}()

	state := PlayState
	inPacket, err := cnt.ReadWithComp(state)
	if err != nil {
		return err
	}

	lg.Debug(
		"client read packet to loop for play state in client",
		NewLgElement("InPacket", inPacket),
	)

	var outPackets []OutPacket

	switch inPacket.(type) {
	case *ConfirmKeepAlivePacket: // 0x0B
		confirmKeepAlivePacket := inPacket.(*ConfirmKeepAlivePacket)
		payload := confirmKeepAlivePacket.GetPayload()
		confirmKeepAliveEvent :=
			NewConfirmKeepAliveEvent(
				payload,
			)
		chanForConfirmKeepAliveEvent <- confirmKeepAliveEvent
		break
	}

	for _, outPacket := range outPackets {
		if err := cnt.WriteWithComp(outPacket); err != nil {
			return err
		}
		lg.Debug(
			"client sent packet to loop for play state in client",
			NewLgElement("OutPacket", outPacket),
		)
	}

	return nil
}

func (cnt *Client) Init(
	lg *Logger,
) {
	lg.Debug("it is started to init client")
	defer func() {
		lg.Debug("it is finished to init client")
	}()
}

func (cnt *Client) Close(
	lg *Logger,
) {
	lg.Debug("it is started to close client")
	defer func() {
		lg.Debug("it is finished to close client")
	}()
	_ = cnt.conn.Close()
}

func (cnt *Client) AddPlayer(
	lg *Logger,
	uid UID, username string,
) error {
	lg.Debug(
		"it is started to add player in client",
		NewLgElement("uid", uid),
		NewLgElement("username", username),
	)
	defer func() {
		lg.Debug("it is finished to add player in client")
	}()

	textureString, signature, err :=
		UUIDToTextureString(uid)
	if err != nil {
		return err
	}
	gamemode := int32(0)
	ping := int32(1000)
	displayName := &Chat{
		Text: username,
		Bold: true,
	}
	packet := NewAddPlayerPacket(
		uid,
		username,
		textureString,
		signature,
		gamemode,
		ping,
		displayName,
	)
	if err := cnt.WriteWithComp(packet); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) RemovePlayer(
	lg *Logger,
	uid UID,
) error {
	lg.Debug("it is started to remove player in client")
	defer func() {
		lg.Debug("it is finished to remove player in client")
	}()

	packet := NewRemovePlayerPacket(
		uid,
	)
	if err := cnt.WriteWithComp(packet); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) UpdateLatency(
	lg *Logger,
	uid UID,
	latency int32,
) error {
	lg.Debug("it is started to update latency in client")
	defer func() {
		lg.Debug("it is finished to update latency in client")
	}()

	packet := NewUpdateLatencyPacket(
		uid,
		latency,
	)
	if err := cnt.WriteWithComp(packet); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) CheckKeepAlive(
	lg *Logger,
	payload int64,
) error {
	lg.Debug(
		"it is started to check keep-alive in client",
		NewLgElement("payload", payload),
	)
	defer func() {
		lg.Debug("it is finished to check keep-alive in client")
	}()

	packet := NewCheckKeepAlivePacket(payload)
	if err := cnt.WriteWithComp(packet); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) GetCID() CID {
	return cnt.cid
}

func (cnt *Client) GetAddr() string {
	return cnt.conn.RemoteAddr().String()
}

func (cnt *Client) String() string {
	return fmt.Sprintf(
		"{ cid: %s, addr: %s }",
		cnt.cid, cnt.addr,
	)
}
