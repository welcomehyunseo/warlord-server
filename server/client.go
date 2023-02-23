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
	"time"
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
		case EnterChatMessagePacketID:
			inPacket = NewEnterChatMessagePacket()
			break
		case ChangeSettingsPacketID:
			inPacket = NewChangeSettingsPacket()
			break
		case InteractWithEntityPacketID:
			inPacket = NewInteractWithEntityPacket()
			break
		case ConfirmKeepAlivePacketID:
			inPacket = NewConfirmKeepAlivePacket()
			break
		case ChangePosPacketID:
			inPacket = NewChangePosPacket()
			break
		case ChangeLookPacketID:
			inPacket = NewChangeLookPacket()
			break
		case ChangePosAndLookPacketID:
			inPacket = NewChangePosAndLookPacket()
			break
		case TakeActionPacketID:
			inPacket = NewTakeActionPacket()
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

	cid CID

	addr net.Addr

	conn net.Conn
}

func NewClient(
	cid CID,
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

		if err := cnt.write(outPacket); err != nil {
			return false, err
		}
		lg.Debug(
			"Client sent packet",
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
	lg.Debug("it is started to handle login state in Client")
	defer func() {
		lg.Debug("it is finished to handle login state in Client")
	}()

	state := LoginState
	inPacket, err := cnt.read(state)
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
	if err := cnt.write(enableCompPacket); err != nil {
		return NilUID, "", err
	}

	completeLoginPacket := NewCompleteLoginPacket(
		uid,
		username,
	)
	if err := cnt.writeWithComp(completeLoginPacket); err != nil {
		return NilUID, "", err
	}

	return uid, username, nil
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
	joinGamePacket := NewJoinGamePacket(
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

	// ChangeSettingsPacket
	if _, err := cnt.readWithComp(
		state,
	); err != nil {
		return err
	}

	// Plugin message
	if _, err := cnt.readWithComp(
		state,
	); err != nil {
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
	world Overworld,
	player Player,
	chanForConfirmKeepAliveEvent ChanForConfirmKeepAliveEvent,
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
		"Client read packet to loop for play state in Client",
		NewLgElement("InPacket", inPacket),
	)

	//eid := player.GetEid()

	var outPackets []OutPacket

	switch inPacket.(type) {
	case *InteractWithEntityPacket:
		interactWithEntityPacket :=
			inPacket.(*InteractWithEntityPacket)
		fmt.Println(interactWithEntityPacket)

		break
	case *ConfirmKeepAlivePacket: // 0x0B
		confirmKeepAlivePacket :=
			inPacket.(*ConfirmKeepAlivePacket)
		payload :=
			confirmKeepAlivePacket.GetPayload()
		confirmKeepAliveEvent :=
			NewConfirmKeepAliveEvent(
				payload,
			)
		chanForConfirmKeepAliveEvent <- confirmKeepAliveEvent

		break
	case *EnterChatMessagePacket:
		enterChatMessagePacket :=
			inPacket.(*EnterChatMessagePacket)
		text := enterChatMessagePacket.GetText()
		if err := player.EnterChatMessage(
			text,
		); err != nil {
			return err
		}

		break
	case *ChangePosPacket:
		changePosPacket := inPacket.(*ChangePosPacket)
		x, y, z :=
			changePosPacket.GetX(),
			changePosPacket.GetY(),
			changePosPacket.GetZ()
		ground := changePosPacket.GetGround()
		if err := player.UpdatePos(
			world,
			x, y, z,
			ground,
		); err != nil {
			return err
		}

		break
	case *ChangeLookPacket:
		changeLookPacket := inPacket.(*ChangeLookPacket)
		yaw, pitch :=
			changeLookPacket.GetYaw(),
			changeLookPacket.GetPitch()
		ground := changeLookPacket.GetGround()
		if err := player.UpdateLook(
			world,
			yaw, pitch,
			ground,
		); err != nil {
			return err
		}

		break
	case *ChangePosAndLookPacket:
		changePosAndLookPacket := inPacket.(*ChangePosAndLookPacket)
		x, y, z :=
			changePosAndLookPacket.GetX(),
			changePosAndLookPacket.GetY(),
			changePosAndLookPacket.GetZ()
		ground := changePosAndLookPacket.GetGround()
		if err := player.UpdatePos(
			world,
			x, y, z,
			ground,
		); err != nil {
			return err
		}
		yaw, pitch :=
			changePosAndLookPacket.GetYaw(),
			changePosAndLookPacket.GetPitch()
		if err := player.UpdateLook(
			world,
			yaw, pitch,
			ground,
		); err != nil {
			return err
		}

		break
	case *TakeActionPacket:
		takeActionPacket := inPacket.(*TakeActionPacket)
		startSneaking :=
			takeActionPacket.IsSneakingStarted()
		stopSneaking :=
			takeActionPacket.IsSneakingStopped()
		startSprinting :=
			takeActionPacket.IsSprintingStared()
		stopSprinting :=
			takeActionPacket.IsSprintingStopped()
		if startSneaking == true {
			if err := player.UpdateSneaking(
				world,
				true,
			); err != nil {
				return err
			}
		} else if stopSneaking == true {
			if err := player.UpdateSneaking(
				world,
				false,
			); err != nil {
				return err
			}
		} else if startSprinting == true {
			if err := player.UpdateSprinting(
				world,
				true,
			); err != nil {
				return err
			}
		} else if stopSprinting == true {
			if err := player.UpdateSprinting(
				world,
				false,
			); err != nil {
				return err
			}
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

func (cnt *Client) Init(
	lg *Logger,
) {
	lg.Debug("it is started to init Client")
	defer func() {
		lg.Debug("it is finished to init Client")
	}()
}

func (cnt *Client) HandleAddPlayerEvent(
	chanForEvent ChanForAddPlayerEvent,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"add-player-event-handler",
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle AddPlayerEvent")
	defer func() {
		lg.Debug("it is finished to handle AddPlayerEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				uid, username :=
					event.GetUUID(), event.GetUsername()
				if err := cnt.AddPlayer(
					uid, username,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				event.Fail()
				panic(err)
			}

			event.Done()
		}

		if stop == true {
			break
		}
	}
}

func (cnt *Client) HandleUpdateLatencyEvent(
	chanForEvent ChanForUpdateLatencyEvent,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"update-latency-event-handler",
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle UpdateLatencyEvent")
	defer func() {
		lg.Debug("it is finished to handle UpdateLatencyEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}

			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				uid, latency := event.GetUUID(), event.GetLatency()
				if err := cnt.UpdateLatency(
					uid, latency,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
		}

		if stop == true {
			break
		}
	}

}

func (cnt *Client) HandleRemovePlayerEvent(
	chanForEvent ChanForRemovePlayerEvent,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"remove-player-event-handler",
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle RemovePlayerEvent")
	defer func() {
		lg.Debug("it is finished to handle RemovePlayerEvent")
	}()

	defer func() {

		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				uid := event.GetUUID()
				if err := cnt.RemovePlayer(
					uid,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
		}

		if stop == true {
			break
		}
	}
}

func (cnt *Client) HandleConfirmKeepAliveEvent(
	world Overworld,
	chanForEvent ChanForConfirmKeepAliveEvent,
	uid UID,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"confirm-keep-alive-event-handler",
		NewLgElement("uid", uid),
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle ConfirmKeepAliveEvent")
	defer func() {
		lg.Debug("it is finished to handle ConfirmKeepAliveEvent")
	}()

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

		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				if err := func() error {
					payload1 := event.GetPayload()
					if payload1 != payload0 {
						return errors.New("payload for keep-alive must be same as given")
					}
					end := time.Now()
					latency := int32(end.Sub(start).Milliseconds())

					if err := world.UpdatePlayerLatency(
						uid,
						latency,
					); err != nil {
						return err
					}

					return nil
				}(); err != nil {
					return err
				}

				start = time.Time{}

				return nil
			}(); err != nil {
				panic(err)
			}
			break
		}

		if stop == true {
			break
		}
	}

}

func (cnt *Client) HandleSpawnPlayerEvent(
	chanForEvent ChanForSpawnPlayerEvent,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"spawn-player-event-handler",
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle SpawnPlayerEvent")
	defer func() {
		lg.Debug("it is finished to handle SpawnPlayerEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				eid, uid :=
					event.GetEID(), event.GetUUID()
				x, y, z :=
					event.GetX(), event.GetY(), event.GetZ()
				yaw, pitch :=
					event.GetYaw(), event.GetPitch()
				sneaking, sprinting :=
					event.IsSneaking(), event.IsSprinting()
				if err := cnt.SpawnPlayer(
					eid, uid,
					x, y, z,
					yaw, pitch,
					sneaking, sprinting,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
		}

		if stop == true {
			break
		}
	}
}

func (cnt *Client) HandleDespawnEntityEvent(
	chanForEvent ChanForDespawnEntityEvent,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"despawn-entity-event-handler",
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle DespawnEntityEvent")
	defer func() {
		lg.Debug("it is finished to handle DespawnEntityEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				eid := event.GetEID()
				if err := cnt.DespawnEntity(
					eid,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
		}

		if stop == true {
			break
		}
	}
}

func (cnt *Client) HandleSetEntityRelativePosEvent(
	chanForEvent ChanForSetEntityRelativePosEvent,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"set-entity-relative-pos-event-handler",
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle SetEntityRelativePosEvent")
	defer func() {
		lg.Debug("it is finished to handle SetEntityRelativePosEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

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
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
		}

		if stop == true {
			break
		}
	}
}

func (cnt *Client) HandleSetEntityLookEvent(
	chanForEvent ChanForSetEntityLookEvent,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"set-entity-look-event-handler",
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle SetEntityLookEvent")
	defer func() {
		lg.Debug("it is finished to handle SetEntityLookEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				eid := event.GetEID()
				yaw, pitch :=
					event.GetYaw(), event.GetPitch()
				ground := event.GetGround()
				if err := cnt.SetEntityLook(
					eid,
					yaw, pitch,
					ground,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
		}

		if stop == true {
			break
		}
	}
}

func (cnt *Client) HandleSetEntityMetadataEvent(
	chanForEvent ChanForSetEntityMetadataEvent,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"set-entity-look-event-handler",
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle SetEntityMetadataEvent")
	defer func() {
		lg.Debug("it is finished to handle SetEntityMetadataEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				eid := event.GetEID()
				metadata := event.GetMetadata()
				if err := cnt.SetEntityMetadata(
					eid,
					metadata,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
		}

		if stop == true {
			break
		}
	}
}

func (cnt *Client) HandleLoadChunkEvent(
	chanForEvent ChanForLoadChunkEvent,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"load-chunk-event-handler",
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle LoadChunkEvent")
	defer func() {
		lg.Debug("it is finished to handle LoadChunkEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				overworld, init :=
					event.GetOverworld(), event.GetInit()
				cx, cz :=
					event.GetCx(), event.GetCz()
				chunk :=
					event.GetChunk()
				if err := cnt.LoadChunk(
					overworld, init,
					cx, cz,
					chunk,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
		}

		if stop == true {
			break
		}
	}
}

func (cnt *Client) HandleUnloadChunkEvent(
	chanForEvent ChanForUnloadChunkEvent,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"unload-chunk-event-handler",
		NewLgElement("Client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle UnloadChunkEvent")
	defer func() {
		lg.Debug("it is finished to handle UnloadChunkEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				cx, cz :=
					event.GetCx(), event.GetCz()
				if err := cnt.UnloadChunk(
					cx, cz,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
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
	packet := NewAddPlayerPacket(
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

func (cnt *Client) Teleport(
	x, y, z float64,
	yaw, pitch float32,
) error {
	payload := rand.Int31()
	teleportPacket := NewTeleportPacket(
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
	//	inPacket.(*FinishTeleportPacket)
	//if ok == false {
	//	return errors.New("it is invalid packet to init play state")
	//}
	//payload1 := finishTeleportPacket.GetPayload()
	//if payload != payload1 {
	//	return errors.New("it is invalid payload of FinishTeleportPacket to init play state")
	//}

	return nil
}

func (cnt *Client) Respawn(
	dimension int32,
	difficulty uint8,
	gamemode uint8,
	level string,
) error {

	respawnPacket := NewRespawnPacket(
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

func (cnt *Client) UpdateLatency(
	uid UID,
	latency int32,
) error {

	packet := NewUpdateLatencyPacket(
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

	packet := NewRemovePlayerPacket(
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

	packet := NewCheckKeepAlivePacket(payload)
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
	packet := NewSpawnPlayerPacket(
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

	packet := NewDespawnEntityPacket(
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

	packet0 := NewSetEntityLookPacket(
		eid,
		yaw, pitch,
		ground,
	)
	if err := cnt.writeWithComp(packet0); err != nil {
		return err
	}

	packet1 := NewSetEntityHeadLookPacket(
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

	packet1 := NewSetEntityRelativePosPacket(
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

	packet1 := NewSetEntityMetadataPacket(
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
	packet := NewSendChunkDataPacket(
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

	packet := NewUnloadChunkPacket(
		cx, cz,
	)
	if err := cnt.writeWithComp(packet); err != nil {
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
