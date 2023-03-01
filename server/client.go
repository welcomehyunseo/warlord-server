package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"
)

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
		case InPacketIDToConfirmTeleport:
			inPacket = NewInPacketToConfirmTeleport()
			break
		case InPacketIDToEnterChatText:
			inPacket = NewInPacketToEnterChatText()
			break
		case InPacketIDToChangeSettings:
			inPacket = NewInPacketToChangeSettings()
			break
		case InPacketIDToInteractWithEntity:
			inPacket = NewInPacketToInteractWithEntity()
			break
		case InPacketIDToConfirmKeepAlive:
			inPacket = NewInPacketToConfirmKeepAlive()
			break
		case InPacketIDToChangePos:
			inPacket = NewInPacketToChangePos()
			break
		case InPacketIDToChangeLook:
			inPacket = NewInPacketToChangeLook()
			break
		case InPacketIDToChangePosAndLook:
			inPacket = NewInPacketToChangePosAndLook()
			break
		case InPacketIDToDoActions:
			_, err := data.ReadVarInt() // TODO: what is exactly?
			if err != nil {
				return nil, err
			}
			actionID, err := data.ReadVarInt()
			if err != nil {
				return nil, err
			}
			switch actionID {
			case 0:
				inPacket = NewInPacketToStartSneaking()
				break
			case 1:
				inPacket = NewInPacketToStopSneaking()
				break
			case 3:
				inPacket = NewInPacketToStartSprinting()
				break
			case 4:
				inPacket = NewInPacketToStopSprinting()
				break
			}
			break
		}
		break
	case StatusState:
		switch pid {
		case InPacketIDToRequest:
			inPacket = NewInPacketToRequest()
			break
		case InPacketIDToPing:
			inPacket = NewInPacketToPing()
			break
		}
		break
	case LoginState:
		switch pid {
		case InPacketIDToStartLogin:
			inPacket = NewInPacketToStartLogin()
			break
		}
		break
	case HandshakingState:
		switch pid {
		case InPacketIDToHandshake:
			inPacket = NewInPacketToHandshake()
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
	*sync.Mutex

	addr net.Addr

	conn net.Conn
}

func NewClient(
	conn net.Conn,
) *Client {
	addr := conn.RemoteAddr()

	return &Client{
		new(sync.Mutex),
		addr,
		conn,
	}
}

func (cnt *Client) read(
	state State,
) (
	InPacket,
	error,
) {

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

func (cnt *Client) readWithComp(
	state State,
) (
	InPacket,
	error,
) {
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
		return nil, errors.New("length of uncompressed id and data of packet is less than the threshold that set to read packet with compression in Client")
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

func (cnt *Client) write(
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

func (cnt *Client) writeWithComp(
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

	if l0 < CompThold {
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
	lg.Debug("it is started to handle non login state in Client")
	defer func() {
		lg.Debug("it is finished to handle non login state in Client")
	}()

	state := HandshakingState

	for {
		inPacket, err := cnt.read(state)
		if err != nil {
			return false, err
		}

		lg.Debug(
			"Client read packet",
			NewLgElement("InPacket", inPacket),
		)

		var outPacket OutPacket

		switch inPacket.(type) {
		case *InPacketToHandshake:
			handshakePacket := inPacket.(*InPacketToHandshake)
			state = handshakePacket.GetNext()
			break
		case *InPacketToRequest:
			responsePacket := NewOutPacketToResponse(
				max, online, text, favicon,
			)
			outPacket = responsePacket
			break
		case *InPacketToPing:
			pingPacket := inPacket.(*InPacketToPing)
			payload := pingPacket.GetPayload()
			pongPacket := NewOutPacketToPong(payload)
			outPacket = pongPacket
			break
		}

		if state == LoginState {
			return false, nil
		}

		if outPacket == nil {
			continue
		}

		if err := cnt.write(outPacket); err != nil {
			return false, err
		}
		lg.Debug(
			"Client sent packet",
			NewLgElement("OutPacket", outPacket),
		)

		if _, ok := outPacket.(*OutPacketToPong); ok == true {
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
	lg.Debug("it is started to handle login state in Client")
	defer func() {
		lg.Debug("it is finished to handle login state in Client")
	}()

	state := LoginState
	inPacket, err := cnt.read(state)
	if err != nil {
		return NilUID, "", err
	}

	startLoginPacket, ok := inPacket.(*InPacketToStartLogin)
	if ok == false {
		return NilUID, "", errors.New("it is invalid inbound packet to handle login state")
	}
	username := startLoginPacket.GetUsername()
	uid, err := UsernameToUUID(username)
	if err != nil {
		return NilUID, "", err
	}

	enableCompPacket := NewOutPacketToEnableComp(CompThold)
	if err := cnt.write(enableCompPacket); err != nil {
		return NilUID, "", err
	}

	completeLoginPacket := NewOutPacketToCompleteLogin(
		uid,
		username,
	)
	if err := cnt.writeWithComp(completeLoginPacket); err != nil {
		return NilUID, "", err
	}

	return uid, username, nil
}

func (cnt *Client) Init(
	lg *Logger,
) {
	lg.Debug("it is started to init Client")
	defer func() {
		lg.Debug("it is finished to init Client")
	}()
}

func (cnt *Client) JoinGame(
	lg *Logger,
	eid EID,
) error {
	lg.Debug(
		"it is started to join game in Client",
		NewLgElement("eid", eid),
	)
	defer func() {
		lg.Debug("it is finished to join game in Client")
	}()

	state := PlayState
	joinGamePacket := NewOutPacketToJoinGame(
		eid,
		0,
		0,
		2,
		"default",
		false,
	)
	if err := cnt.writeWithComp(
		joinGamePacket,
	); err != nil {
		return err
	}

	// InPacketToChangeSettings
	if _, err := cnt.readWithComp(
		state,
	); err != nil {
		return err
	}

	// Plugin msg
	if _, err := cnt.readWithComp(
		state,
	); err != nil {
		return err
	}

	setAbilitiesPacket := NewOutPacketToSetAbilities(
		false,
		false,
		false,
		false,
		0,
		0,
	)
	if err := cnt.writeWithComp(
		setAbilitiesPacket,
	); err != nil {
		return err
	}

	if err := cnt.Teleport(
		0, 0, 0,
		0, 0,
	); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) LoopForPlayState(
	lg *Logger,
	headCmdMgr *HeadCmdMgr,
	worldCmdMgr *WorldCmdMgr,
	dim *Dimension,
	chanForCKAEvent ChanForConfirmKeepAliveEvent,
) error {
	lg.Debug("it is started to loop for play state in Client")
	defer func() {
		lg.Debug("it is finished to loop for play state in Client")
	}()

	state := PlayState
	inPacket, err := cnt.readWithComp(state)
	if err != nil {
		return err
	}

	lg.Debug(
		"client read packet to loop for play state in Client",
		NewLgElement("InPacket", inPacket),
	)

	//eid := player.GetEID()

	var outPackets []OutPacket

	switch inPacket.(type) {
	case *InPacketToInteractWithEntity:
		IWEPacket := inPacket.(*InPacketToInteractWithEntity)
		fmt.Println(IWEPacket)

		break
	case *InPacketToConfirmKeepAlive: // 0x0B
		CKAPacket :=
			inPacket.(*InPacketToConfirmKeepAlive)
		payload := CKAPacket.GetPayload()
		CKAEvent := NewConfirmKeepAliveEvent(
			payload,
		)
		chanForCKAEvent <- CKAEvent
		break
	case *InPacketToEnterChatText:
		enterChatMessagePacket :=
			inPacket.(*InPacketToEnterChatText)
		text := enterChatMessagePacket.GetText()
		if err := dim.EnterChatText(
			headCmdMgr,
			worldCmdMgr,
			text,
			cnt,
		); err != nil {
			msg := &Chat{
				Text: fmt.Sprintf(
					"[error] %s", err,
				),
				Color: "dark_red",
			}
			if err := cnt.SendChatMessage(
				msg,
			); err != nil {
				return err
			}
		}
		break
	case *InPacketToChangePos:
		CPPacket := inPacket.(*InPacketToChangePos)
		x, y, z :=
			CPPacket.GetX(),
			CPPacket.GetY(),
			CPPacket.GetZ()
		ground := CPPacket.GetGround()
		if err := dim.UpdatePlayerPos(
			x, y, z,
			ground,
		); err != nil {
			return err
		}
		break
	case *InPacketToChangeLook:
		CLPacket := inPacket.(*InPacketToChangeLook)
		yaw, pitch :=
			CLPacket.GetYaw(),
			CLPacket.GetPitch()
		ground := CLPacket.GetGround()
		if err := dim.UpdatePlayerLook(
			yaw, pitch,
			ground,
		); err != nil {
			return err
		}
		break
	case *InPacketToChangePosAndLook:
		CPALPacket := inPacket.(*InPacketToChangePosAndLook)
		x, y, z :=
			CPALPacket.GetX(),
			CPALPacket.GetY(),
			CPALPacket.GetZ()
		ground := CPALPacket.GetGround()
		if err := dim.UpdatePlayerPos(
			x, y, z,
			ground,
		); err != nil {
			return err
		}
		yaw, pitch :=
			CPALPacket.GetYaw(),
			CPALPacket.GetPitch()
		if err := dim.UpdatePlayerLook(
			yaw, pitch,
			ground,
		); err != nil {
			return err
		}
		break
	case *InPacketToStartSneaking:
		//packet := inPacket.(*InPacketToStartSneaking)
		if err := dim.UpdatePlayerSneaking(
			true,
		); err != nil {
			return err
		}
		break
	case *InPacketToStopSneaking:
		//packet := inPacket.(*InPacketToStopSneaking)
		if err := dim.UpdatePlayerSneaking(
			false,
		); err != nil {
			return err
		}
		break
	case *InPacketToStartSprinting:
		//packet := inPacket.(*InPacketToStartSprinting)
		if err := dim.UpdatePlayerSprinting(
			true,
		); err != nil {
			return err
		}
		break
	case *InPacketToStopSprinting:
		//packet := inPacket.(*InPacketToStopSprinting)
		if err := dim.UpdatePlayerSprinting(
			false,
		); err != nil {
			return err
		}
		break
	}

	for _, outPacket := range outPackets {
		if err := cnt.writeWithComp(outPacket); err != nil {
			return err
		}
		lg.Debug(
			"Client sent packet to loop for play state in Client",
			NewLgElement("OutPacket", outPacket),
		)
	}

	return nil
}

func (cnt *Client) HandleCommonEvents(
	chanForAPEvent ChanForAddPlayerEvent,
	chanForULEvent ChanForUpdateLatencyEvent,
	chanForRPEvent ChanForRemovePlayerEvent,
	chanForSPEvent ChanForSpawnPlayerEvent,
	chanForDEEvent ChanForDespawnEntityEvent,
	chanForSERPEvent ChanForSetEntityRelativePosEvent,
	chanForSELEvent ChanForSetEntityLookEvent,
	chanForSEMEvent ChanForSetEntityMetadataEvent,
	chanForLCEvent ChanForLoadChunkEvent,
	chanForUnCEvent ChanForUnloadChunkEvent,
	chanForError ChanForError,
	ctx context.Context,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer wg.Done()

	lg := NewLogger(
		"common-events-handler",
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle events")
	defer lg.Debug("it is finished to handle events")

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event := <-chanForAPEvent:
			uid, username :=
				event.GetUUID(),
				event.GetUsername()
			if err := cnt.AddPlayer(
				uid, username,
			); err != nil {
				event.Fail()
				panic(err)
			}
			event.Done()
			break
		case event := <-chanForULEvent:
			uid, latency :=
				event.GetUUID(),
				event.GetLatency()
			if err := cnt.UpdateLatency(
				uid, latency,
			); err != nil {
				panic(err)
			}
			break
		case event := <-chanForRPEvent:
			uid := event.GetUUID()
			if err := cnt.RemovePlayer(
				uid,
			); err != nil {
				event.Fail()
				panic(err)
			}
			event.Done()
			break
		case event := <-chanForSPEvent:
			eid, uid :=
				event.GetEID(),
				event.GetUUID()
			x, y, z :=
				event.GetX(),
				event.GetY(),
				event.GetZ()
			yaw, pitch :=
				event.GetYaw(),
				event.GetPitch()
			sneaking, sprinting :=
				event.IsSneaking(),
				event.IsSprinting()
			if err := cnt.SpawnPlayer(
				eid, uid,
				x, y, z,
				yaw, pitch,
				sneaking, sprinting,
			); err != nil {
				panic(err)
			}
			break
		case event := <-chanForDEEvent:
			eid := event.GetEID()
			if err := cnt.DespawnEntity(
				eid,
			); err != nil {
				panic(err)
			}
			break
		case event := <-chanForSERPEvent:
			eid := event.GetEID()
			deltaX, deltaY, deltaZ :=
				event.GetDeltaX(),
				event.GetDeltaY(),
				event.GetDeltaZ()
			ground := event.GetGround()
			if err := cnt.SetEntityRelativePos(
				eid,
				deltaX, deltaY, deltaZ,
				ground,
			); err != nil {
				panic(err)
			}
			break
		case event := <-chanForSELEvent:
			eid := event.GetEID()
			yaw, pitch :=
				event.GetYaw(),
				event.GetPitch()
			ground := event.GetGround()
			if err := cnt.SetEntityLook(
				eid,
				yaw, pitch,
				ground,
			); err != nil {
				panic(err)
			}
			break
		case event := <-chanForSEMEvent:
			eid := event.GetEID()
			metadata := event.GetMetadata()
			if err := cnt.SetEntityMetadata(
				eid,
				metadata,
			); err != nil {
				panic(err)
			}
			break
		case event := <-chanForLCEvent:
			overworld, init :=
				event.GetOverworld(),
				event.GetInit()
			cx, cz :=
				event.GetCx(),
				event.GetCz()
			chunk :=
				event.GetChunk()
			if err := cnt.LoadChunk(
				overworld, init,
				cx, cz,
				chunk,
			); err != nil {
				panic(err)
			}
			break
		case event := <-chanForUnCEvent:
			cx, cz :=
				event.GetCx(),
				event.GetCz()
			if err := cnt.UnloadChunk(
				cx, cz,
			); err != nil {
				panic(err)
			}
			break
		case <-ctx.Done():
			stop = true
			break
		}

		if stop == true {
			break
		}
	}
}

func (cnt *Client) HandleUpdateChunkEvent(
	chanForEvent ChanForUpdateChunkEvent,
	dim *Dimension,
	chanForError ChanForError,
	ctx context.Context,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer wg.Done()

	lg := NewLogger(
		"update-chunk-event-handler",
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle events")
	defer lg.Debug("it is finished to handle events")

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event := <-chanForEvent:
			currCx, currCz :=
				event.GetCurrCx(),
				event.GetCurrCz()
			prevCx, prevCz :=
				event.GetPrevCx(),
				event.GetPrevCz()
			if err := dim.UpdatePlayerChunk(
				prevCx, prevCz,
				currCx, currCz,
			); err != nil {
				panic(err)
			}
			break
		case <-ctx.Done():
			stop = true
			break
		}

		if stop == true {
			break
		}
	}
}

func (cnt *Client) HandleConfirmKeepAliveEvent(
	chanForCKAEvent ChanForConfirmKeepAliveEvent,
	dim *Dimension,
	chanForError ChanForError,
	ctx context.Context,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer wg.Done()

	lg := NewLogger(
		"confirm-keep-alive-event-handler",
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle events")
	defer lg.Debug("it is finished to handle events")

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	start := time.Time{}
	var payload0 int64

	stop := false
	for {
		select {
		case <-time.After(DelayForCheckKeepAlive):
			if start.IsZero() == false {
				break
			}

			payload0 = rand.Int63()
			if err := cnt.CheckKeepAlive(
				payload0,
			); err != nil {
				panic(err)
			}

			start = time.Now()

			break
		case event := <-chanForCKAEvent:
			payload1 := event.GetPayload()
			if payload1 != payload0 {
				err := errors.New("payload for keep-alive must be same as given")
				panic(err)
			}

			end := time.Now()
			latency := int32(end.Sub(start).Milliseconds())
			if err := dim.UpdateLatency(
				latency,
			); err != nil {
				panic(err)
			}
			start = time.Time{}

			break
		case <-ctx.Done():
			stop = true
			break
		}

		if stop == true {
			break
		}
	}
}

func (cnt *Client) Close(
	lg *Logger,
) {
	lg.Debug("it is started to close Client")
	defer func() {
		lg.Debug("it is finished to close Client")
	}()
	_ = cnt.conn.Close()
}

func (cnt *Client) SendChatMessage(
	msg *Chat,
) error {
	SCMPacket := NewOutPacketToSendChatMessage(
		msg,
	)
	if err := cnt.writeWithComp(
		SCMPacket,
	); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) Teleport(
	x, y, z float64,
	yaw, pitch float32,
) error {
	payload := rand.Int31()
	teleportPacket := NewOutPacketToTeleport(
		x, y, z,
		yaw, pitch,
		payload,
	)
	if err := cnt.writeWithComp(
		teleportPacket,
	); err != nil {
		return err
	}

	// TODO: check payload
	//state := PlayState
	//inPacket, err := cnt.readWithComp(state)
	//if err != nil {
	//	return err
	//}
	//finishTeleportPacket, ok :=
	//	inPacket.(*InPacketToConfirmTeleport)
	//if ok == false {
	//	return errors.New("it is invalid packet to init play state")
	//}
	//payload1 := finishTeleportPacket.GetPayload()
	//if payload != payload1 {
	//	return errors.New("it is invalid payload of InPacketToConfirmTeleport to init play state")
	//}

	return nil
}

func (cnt *Client) Respawn(
	dimension int32,
	difficulty uint8,
	gamemode uint8,
	level string,
) error {

	respawnPacket := NewOutPacketToRespawn(
		dimension,
		difficulty,
		gamemode,
		level,
	)
	if err := cnt.writeWithComp(
		respawnPacket,
	); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) AddPlayer(
	uid UID, username string,
) error {
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
	packet := NewOutPacketToAddPlayer(
		uid,
		username,
		textureString,
		signature,
		gamemode,
		ping,
		displayName,
	)
	if err := cnt.writeWithComp(packet); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) UpdateLatency(
	uid UID,
	latency int32,
) error {

	packet := NewOutPacketToUpdateLatency(
		uid,
		latency,
	)
	if err := cnt.writeWithComp(packet); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) RemovePlayer(
	uid UID,
) error {
	packet := NewOutPacketToRemovePlayer(
		uid,
	)
	if err := cnt.writeWithComp(packet); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) CheckKeepAlive(
	payload int64,
) error {

	packet := NewOutPacketToCheckKeepAlive(payload)
	if err := cnt.writeWithComp(packet); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) SpawnPlayer(
	eid EID, uid UID,
	x, y, z float64,
	yaw, pitch float32,
	sneaking, sprinting bool,
) error {
	metadata := NewEntityMetadata()
	if err := metadata.SetActions(
		sneaking, sprinting,
	); err != nil {
		return err
	}
	packet := NewOutPacketToSpawnPlayer(
		eid, uid,
		x, y, z,
		yaw, pitch,
		metadata,
	)
	if err := cnt.writeWithComp(packet); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) DespawnEntity(
	eid EID,
) error {

	packet := NewOutPacketToDespawnEntity(
		eid,
	)
	if err := cnt.writeWithComp(packet); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) SetEntityLook(
	eid EID,
	yaw, pitch float32,
	ground bool,
) error {

	packet0 := NewOutPacketToSetEntityLook(
		eid,
		yaw, pitch,
		ground,
	)
	if err := cnt.writeWithComp(packet0); err != nil {
		return err
	}

	packet1 := NewOutPacketToSetEntityHeadLook(
		eid,
		yaw,
	)
	if err := cnt.writeWithComp(packet1); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) SetEntityRelativePos(
	eid EID,
	deltaX, deltaY, deltaZ int16,
	ground bool,
) error {

	packet1 := NewOutPacketToSetEntityRltvPos(
		eid,
		deltaX, deltaY, deltaZ,
		ground,
	)
	if err := cnt.writeWithComp(packet1); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) SetEntityMetadata(
	eid EID,
	metadata *EntityMetadata,
) error {

	packet1 := NewOutPacketToSetEntityMd(
		eid,
		metadata,
	)
	if err := cnt.writeWithComp(packet1); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) LoadChunk(
	overworld, init bool,
	cx, cz int32,
	chunk *Chunk,
) error {

	bitmask, data := chunk.GenerateData(init, overworld)
	packet := NewOutPacketToSendChunkData(
		cx, cz,
		init,
		bitmask, data,
	)
	if err := cnt.writeWithComp(packet); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) UnloadChunk(
	cx, cz int32,
) error {

	packet := NewOutPacketToUnloadChunk(
		cx, cz,
	)
	if err := cnt.writeWithComp(packet); err != nil {
		return err
	}

	return nil
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
