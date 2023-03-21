package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/welcomehyunseo/warlord-server/server/data"
	"github.com/welcomehyunseo/warlord-server/server/item"
	"github.com/welcomehyunseo/warlord-server/server/metadata"
	"github.com/welcomehyunseo/warlord-server/server/packet"
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
	l := 0

	for {
		buf := make([]byte, 1)
		if _, err := r.Read(buf); err != nil {
			return 0, 0, err
		}
		l += 1
		b := buf[0]

		v |= int32(b&data.SegmentBits) << position

		if (b & data.ContinueBit) == 0 {
			break
		}

		position += 7

		if position >= 32 {
			return 0, 0, errors.New("it is too big to read VarInt")
		}
	}

	return v, l, nil
}

func writeVarInt(
	v int32,
) (
	[]byte,
	error,
) {
	v0 := uint32(v)

	arr := make([]byte, 0)

	for {
		if (v0 & ^uint32(data.SegmentBits)) == 0 {
			v1 := uint8(v0)
			arr = append(arr, v1)
			break
		}

		v1 := uint8(v0&uint32(data.SegmentBits)) | data.ContinueBit
		arr = append(arr, v1)

		v0 >>= 7
	}

	return arr, nil
}

func read(
	length int, // length of packet
	r io.Reader,
) (
	int32,
	[]byte,
	error,
) {
	pid, l0, err := readVarInt(r)
	if err != nil {
		return 0, nil, err
	}

	l1 := length - l0
	arr := make([]byte, l1)

	if l1 == 0 {
		return pid, arr, nil
	}

	if _, err = r.Read(arr); err != nil {
		return 0, nil, err
	}

	return pid, arr, nil
}

func write(
	pid int32,
	arr []byte,
) (
	[]byte,
	error,
) {
	arr1, err := writeVarInt(
		pid,
	)
	if err != nil {
		return nil, err
	}

	arr2 := data.Concat(arr1, arr)

	return arr2, nil
}

func distribute(
	state int32,
	pid int32,
	arr []byte,
) (
	packet.InPacket,
	error,
) {
	//fmt.Println("pid:", pid)

	var inPacket packet.InPacket

	switch state {
	case packet.PlayState:
		switch pid {
		case packet.InPacketIDToConfirmTeleport:
			inPacket = packet.NewInPacketToConfirmTeleport()
			break
		case packet.InPacketIDToEnterChatText:
			inPacket = packet.NewInPacketToEnterChatText()
			break
		case packet.InPacketIDToChangeSettings:
			inPacket = packet.NewInPacketToChangeSettings()
			break
		case packet.InPacketIDToConfirmTransactionOfWindow:
			inPacket = packet.NewInPacketToConfirmTransactionOfWindow()
			break
		case packet.InPacketIDToClickWindow:
			inPacket = packet.NewInPacketToClickWindow()
			break
		case packet.InPacketIDToInteractWithEntity:
			inPacket = packet.NewInPacketToInteractWithEntity()
			break
		case packet.InPacketIDToConfirmKeepAlive:
			inPacket = packet.NewInPacketToConfirmKeepAlive()
			break
		case packet.InPacketIDToChangePosition:
			inPacket = packet.NewInPacketToChangePosition()
			break
		case packet.InPacketIDToChangeLook:
			inPacket = packet.NewInPacketToChangeLook()
			break
		case packet.InPacketIDToChangePositionAndLook:
			inPacket = packet.NewInPacketToChangePositionAndLook()
			break
		case packet.InPacketIDToDoActions:
			data := data.NewDataWithBytes(arr)
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
				inPacket = packet.NewInPacketToStartSneaking()
				break
			case 1:
				inPacket = packet.NewInPacketToStopSneaking()
				break
			case 3:
				inPacket = packet.NewInPacketToStartSprinting()
				break
			case 4:
				inPacket = packet.NewInPacketToStopSprinting()
				break
			}

			return inPacket, nil
		}
		break
	case packet.StatusState:
		switch pid {
		case packet.InPacketIDToRequest:
			inPacket = packet.NewInPacketToRequest()
			break
		case packet.InPacketIDToPing:
			inPacket = packet.NewInPacketToPing()
			break
		}
		break
	case packet.LoginState:
		switch pid {
		case packet.InPacketIDToStartLogin:
			inPacket = packet.NewInPacketToStartLogin()
			break
		}
		break
	case packet.HandshakingState:
		switch pid {
		case packet.InPacketIDToHandshake:
			inPacket = packet.NewInPacketToHandshake()
			break
		}
		break
	}
	if inPacket == nil {
		return nil, nil
	}

	if err := inPacket.Unpack(
		arr,
	); err != nil {
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
	state int32,
) (
	packet.InPacket,
	error,
) {

	conn := cnt.conn

	l0, _, err := readVarInt(
		conn,
	) // length of packet
	if err != nil {
		return nil, err
	}

	pid, arr, err := read(
		int(l0), conn,
	)
	if err != nil {
		return nil, err
	}

	return distribute(
		state,
		pid,
		arr,
	)
}

func (cnt *Client) readWithComp(
	state int32,
) (
	packet.InPacket,
	error,
) {
	conn := cnt.conn

	l0, _, err := readVarInt(
		conn,
	) // length of packet
	if err != nil {
		return nil, err
	}

	l1, l2, err := readVarInt(
		conn,
	) // uncompressed length of id and data of packet
	if err != nil {
		return nil, err
	}

	l3 := int(l0) - l2 // length of winId and data of packet
	if l1 == 0 {
		pid, arr0, err := read(
			l3, conn,
		)
		if err != nil {
			return nil, err
		}

		return distribute(
			state,
			pid,
			arr0,
		)
	} else if l1 < CompThold {
		return nil, errors.New("length of uncompressed packet ID and bytes of packet is less than the threshold that set to read packet with compression in Client")
	}

	arr0 := make([]byte, l3)
	if _, err = conn.Read(
		arr0,
	); err != nil {
		return nil, err
	}

	buf, err := Uncompress(arr0)
	if err != nil {
		return nil, err
	}

	pid, _, err := readVarInt(buf)
	if err != nil {
		return nil, err
	}

	arr1 := buf.Bytes()

	return distribute(
		state,
		pid,
		arr1,
	)
}

func (cnt *Client) write(
	packet packet.OutPacket,
) error {
	cnt.Lock()
	defer cnt.Unlock()

	conn := cnt.conn

	pid := packet.GetID()
	arr0, err := packet.Pack()
	if err != nil {
		return err
	}

	arr1, err := write(
		pid, arr0,
	)
	if err != nil {
		return err
	}

	length := len(arr1)

	arr2, err := writeVarInt(
		int32(length),
	)
	if err != nil {
		return err
	}

	arr3 := data.Concat(arr2, arr1)
	if _, err := conn.Write(
		arr3,
	); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) writeWithComp(
	packet packet.OutPacket,
) error {
	cnt.Lock()
	defer cnt.Unlock()

	conn := cnt.conn

	pid := packet.GetID()
	arr0, err := packet.Pack()
	if err != nil {
		return err
	}

	arr1, err := write(pid, arr0)
	if err != nil {
		return err
	}
	l1 := len(arr1) // length of packet before compression

	if l1 < CompThold {
		arr2, err := writeVarInt(
			int32(0),
		)
		if err != nil {
			return err
		}

		l2 := len(arr2)
		l3 := l1 + l2
		arr3, err := writeVarInt(
			int32(l3),
		)
		if err != nil {
			return err
		}

		arr4 := data.Concat(arr3, arr2)
		arr5 := data.Concat(arr4, arr1)

		if _, err := conn.Write(
			arr5,
		); err != nil {
			return err
		}

		return nil
	}

	buf1, err := Compress(arr1)
	if err != nil {
		return err
	}
	arr2 := buf1.Bytes()
	l2 := len(arr2) // length of packet after compression

	arr4, err := writeVarInt(
		int32(l1),
	)
	if err != nil {
		return err
	}
	l4 := len(arr4)

	l5 := l4 + l2

	arr6, err := writeVarInt(
		int32(l5),
	)
	if err != nil {
		return err
	}

	arr7 := data.Concat(arr6, arr4)
	arr8 := data.Concat(arr7, arr2)

	if _, err := conn.Write(
		arr8,
	); err != nil {
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

	state := packet.HandshakingState

	for {
		inPacket, err := cnt.read(state)
		if err != nil {
			return false, err
		}

		lg.Debug(
			"Client read packet",
			NewLgElement("InPacket", inPacket),
		)

		var outPacket packet.OutPacket

		switch inPacket.(type) {
		case *packet.InPacketToHandshake:
			handshakePacket := inPacket.(*packet.InPacketToHandshake)
			state = handshakePacket.GetNestState()
			break
		case *packet.InPacketToRequest:
			responsePacket := packet.NewOutPacketToResponse(
				max, online, text, favicon,
			)
			outPacket = responsePacket
			break
		case *packet.InPacketToPing:
			pingPacket := inPacket.(*packet.InPacketToPing)
			pongPacket := packet.NewOutPacketToPong(
				pingPacket.GetPayload(),
			)
			outPacket = pongPacket
			break
		}

		if state == packet.LoginState {
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

		if _, ok := outPacket.(*packet.OutPacketToPong); ok == true {
			return true, nil
		}
	}
}

func (cnt *Client) HandleLoginState(
	lg *Logger,
) (
	uuid.UUID,
	string, // username
	error,
) {
	lg.Debug("it is started to handle login state in Client")
	defer func() {
		lg.Debug("it is finished to handle login state in Client")
	}()

	state := packet.LoginState
	inPacket, err := cnt.read(state)
	if err != nil {
		return uuid.Nil, "", err
	}

	startLoginPacket, ok := inPacket.(*packet.InPacketToStartLogin)
	if ok == false {
		return uuid.Nil, "", errors.New("it is invalid inbound packet to handle login state")
	}
	username := startLoginPacket.GetUsername()
	uid, err := UsernameToUID(username)
	if err != nil {
		return uuid.Nil, "", err
	}

	enableCompPacket := packet.NewOutPacketToEnableComp(CompThold)
	if err := cnt.write(enableCompPacket); err != nil {
		return uuid.Nil, "", err
	}

	completeLoginPacket := packet.NewOutPacketToCompleteLogin(
		uid,
		username,
	)
	if err := cnt.writeWithComp(completeLoginPacket); err != nil {
		return uuid.Nil, "", err
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
	eid int32,
) error {
	lg.Debug(
		"it is started to join game in Client",
		NewLgElement("eid", eid),
	)
	defer func() {
		lg.Debug("it is finished to join game in Client")
	}()

	state := packet.PlayState
	JGPacket := packet.NewOutPacketToJoinGame(
		eid,
		0,
		0,
		2,
		"default",
		false,
	)
	if err := cnt.writeWithComp(
		JGPacket,
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

	SAPacket := packet.NewOutPacketToSetAbilities(
		false,
		false,
		false,
		false,
		0,
		0,
	)
	if err := cnt.writeWithComp(
		SAPacket,
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

func (cnt *Client) LoopToPlaying(
	lg *Logger,
	dim *Dimension,
	CHForCKAEvent ChanForConfirmKeepAliveEvent,
) error {
	lg.Debug("it is started to loop for play state in Client")
	defer func() {
		lg.Debug("it is finished to loop for play state in Client")
	}()

	state := packet.PlayState
	inPacket, err := cnt.readWithComp(state)
	if err != nil {
		return err
	}

	lg.Debug(
		"client read packet to loop for play state in Client",
		NewLgElement("InPacket", inPacket),
	)

	//eid := playerConnection.GetEID()

	var outPackets []packet.OutPacket

	switch inPacket.(type) {
	case *packet.InPacketToClickWindow:
		pk := inPacket.(*packet.InPacketToClickWindow)
		if err := dim.ClickWindow(
			pk.GetWindowID(),
			pk.GetSlotNumber(),
			pk.GetButtonEnum(),
			pk.GetActionNumber(),
			pk.GetModeEnum(),
		); err != nil {
			return err
		}
		break
	case *packet.InPacketToConfirmKeepAlive: // 0x0B
		pk := inPacket.(*packet.InPacketToConfirmKeepAlive)
		CHForCKAEvent <- NewConfirmKeepAliveEvent(
			pk.GetPayload(),
		)
		break
	case *packet.InPacketToEnterChatText:
		pk := inPacket.(*packet.InPacketToEnterChatText)
		if err := dim.EnterChatText(
			pk.GetText(),
			cnt,
		); err != nil {
			if err := cnt.SendErrorMessage(
				err,
			); err != nil {
				return err
			}
		}
		break
	case *packet.InPacketToChangePosition:
		pk := inPacket.(*packet.InPacketToChangePosition)
		x, y, z := pk.GetPosition()
		if err := dim.UpdatePos(
			x, y, z,
			pk.IsGround(),
		); err != nil {
			return err
		}
		//fmt.Println(pk)
		break
	case *packet.InPacketToChangeLook:
		pk := inPacket.(*packet.InPacketToChangeLook)
		yaw, pitch := pk.GetLook()
		if err := dim.UpdateLook(
			yaw, pitch,
			pk.IsGround(),
		); err != nil {
			return err
		}
		break
	case *packet.InPacketToChangePositionAndLook:
		pk := inPacket.(*packet.InPacketToChangePositionAndLook)
		x, y, z := pk.GetPosition()
		ground := pk.IsGround()
		if err := dim.UpdatePos(
			x, y, z,
			ground,
		); err != nil {
			return err
		}
		yaw, pitch := pk.GetLook()
		if err := dim.UpdateLook(
			yaw, pitch,
			ground,
		); err != nil {
			return err
		}
		break
		//case *InPacketToStartSneaking:
		//	if err := dim.UpdatePlayerSneaking(
		//		true,
		//	); err != nil {
		//		return err
		//	}
		//	break
		//case *InPacketToStopSneaking:
		//	if err := dim.UpdatePlayerSneaking(
		//		false,
		//	); err != nil {
		//		return err
		//	}
		//	break
		//case *InPacketToStartSprinting:
		//	if err := dim.UpdatePlayerSprinting(
		//		true,
		//	); err != nil {
		//		return err
		//	}
		//	break
		//case *InPacketToStopSprinting:
		//	if err := dim.UpdatePlayerSprinting(
		//		false,
		//	); err != nil {
		//		return err
		//	}
		//	break
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
	CHForAPEvent ChanForAddPlayerEvent,
	CHForULEvent ChanForUpdateLatencyEvent,
	CHForRPEvent ChanForRemovePlayerEvent,
	CHForSPEvent ChanForSpawnPlayerEvent,
	CHForSISEvent ChanForSpawnItemStandEvent,
	CHForSERMEvent ChanForSetEntityRelativeMoveEvent,
	CHForSELEvent ChanForSetEntityLookEvent,
	CHForSEAEvent ChanForSetEntityActionsEvent,
	CHForDEEvent ChanForDespawnEntityEvent,
	CHForLCEvent ChanForLoadChunkEvent,
	CHForUCEvent ChanForUnloadChunkEvent,
	CHForError ChanForError,
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
			CHForError <- err
		}
	}()

	stop := false
	for {
		select {
		case e := <-CHForAPEvent:
			if err := cnt.AddPlayer(
				e.GetUID(),
				e.GetUsername(),
			); err != nil {
				e.Fail()
				panic(err)
			}
			e.Done()
		case e := <-CHForULEvent:
			if err := cnt.UpdateLatency(
				e.GetUID(),
				e.GetMilliseconds(),
			); err != nil {
				panic(err)
			}
		case e := <-CHForRPEvent:
			if err := cnt.RemovePlayer(
				e.GetUID(),
			); err != nil {
				e.Fail()
				panic(err)
			}
			e.Done()
		case e := <-CHForSPEvent:
			eid := e.GetEID()
			uid := e.GetUID()
			x, y, z := e.GetPosition()
			yaw, pitch := e.GetLook()
			if err := cnt.SpawnPlayer(
				eid, uid,
				x, y, z,
				yaw, pitch,
			); err != nil {
				panic(err)
			}
		case e := <-CHForSISEvent:
			if err := cnt.SpawnItemStand(
				e.GetEID(),
				e.GetUID(),
				e.GetX(), e.GetY(), e.GetZ(),
				e.GetYaw(), e.GetPitch(),
				e.GetItem(),
			); err != nil {
				panic(err)
			}
		case e := <-CHForSERMEvent:
			dx, dy, dz := e.GetDifferences()
			if err := cnt.SetEntityRelativeMove(
				e.GetEID(),
				dx, dy, dz,
				e.IsGround(),
			); err != nil {
				panic(err)
			}
		case e := <-CHForSELEvent:
			eid := e.GetEID()
			yaw, pitch := e.GetLook()
			ground := e.IsGround()
			if err := cnt.SetEntityLook(
				eid,
				yaw, pitch,
				ground,
			); err != nil {
				panic(err)
			}
		case e := <-CHForDEEvent:
			eid := e.GetEID()
			if err := cnt.DespawnEntity(
				eid,
			); err != nil {
				panic(err)
			}
		case e := <-CHForLCEvent:
			ow, init := e.IsOverworld(), e.IsInit()
			cx, cz := e.GetChunkPosition()
			chunk := e.GetChunk()
			if err := cnt.LoadChunk(
				ow, init,
				cx, cz,
				chunk,
			); err != nil {
				panic(err)
			}
		case e := <-CHForUCEvent:
			cx, cz := e.GetChunkPosition()
			if err := cnt.UnloadChunk(
				cx, cz,
			); err != nil {
				panic(err)
			}
		case <-ctx.Done():
			stop = true
		}

		if stop {
			break
		}
	}
}

func (cnt *Client) HandleConfirmKeepAliveEvent(
	CHForCKAEvent ChanForConfirmKeepAliveEvent,
	dim *Dimension,
	CHForError ChanForError,
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
			CHForError <- err
		}
	}()

	start := time.Time{}
	var payload int64

	stop := false
	for {
		select {
		case <-time.After(DelayForCheckKeepAlive):
			if start.IsZero() == false {
				break
			}

			payload = rand.Int63()
			if err := cnt.CheckKeepAlive(
				payload,
			); err != nil {
				panic(err)
			}

			start = time.Now()

			break
		case e := <-CHForCKAEvent:
			if e.GetPayload() != payload {
				err := errors.New("payload for keep-alive must be same as given")
				panic(err)
			}

			end := time.Now()
			ms := int32(end.Sub(start).Milliseconds())
			if err := dim.UpdateLatency(
				ms,
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

func (cnt *Client) CloseWindow(
	winID int8,
) error {
	CWPacket := packet.NewOutPacketToCloseWindow(
		winID,
	)
	if err := cnt.writeWithComp(
		CWPacket,
	); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) SendErrorMessage(
	err any,
) error {
	msg := &data.Chat{
		Text: fmt.Sprintf(
			"[error] %s", err,
		),
		Color: "dark_red",
	}
	SCMPacket := packet.NewOutPacketToSendChatMessage(
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
	p := packet.NewOutPacketToTeleport(
		x, y, z,
		yaw, pitch,
		payload,
	)
	if err := cnt.writeWithComp(
		p,
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
	x, y, z float64,
	yaw, pitch float32,
) error {
	RPacket0 := packet.NewOutPacketToRespawn(
		-1,
		2,
		0,
		"default",
	)
	if err := cnt.writeWithComp(
		RPacket0,
	); err != nil {
		return err
	}

	payload0 := rand.Int31()
	TPacket0 := packet.NewOutPacketToTeleport(
		0, 0, 0,
		0, 0,
		payload0,
	)
	if err := cnt.writeWithComp(
		TPacket0,
	); err != nil {
		return err
	}

	// TODO: check payload

	RPacket1 := packet.NewOutPacketToRespawn(
		0,
		2,
		0,
		"default",
	)
	if err := cnt.writeWithComp(
		RPacket1,
	); err != nil {
		return err
	}

	payload1 := rand.Int31()
	TPacket1 := packet.NewOutPacketToTeleport(
		x, y, z,
		yaw, pitch,
		payload1,
	)
	if err := cnt.writeWithComp(
		TPacket1,
	); err != nil {
		return err
	}

	// TODO: check payload

	return nil
}

func (cnt *Client) AddPlayer(
	uid uuid.UUID, username string,
) error {
	textureStr, signature, err :=
		UIDToTextureString(
			uid,
		)
	if err != nil {
		return err
	}

	gamemode := int32(0)
	ping := int32(1000)
	displayName := &data.Chat{
		Text: username,
		Bold: true,
	}
	APPacket := packet.NewOutPacketToAddPlayer(
		uid,
		username,
		textureStr,
		signature,
		gamemode,
		ping,
		displayName,
	)
	if err := cnt.writeWithComp(APPacket); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) UpdateLatency(
	uid uuid.UUID,
	ms int32,
) error {
	PLPacket := packet.NewOutPacketToUpdateLatency(
		uid,
		ms,
	)
	if err := cnt.writeWithComp(PLPacket); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) RemovePlayer(
	uid uuid.UUID,
) error {
	packet := packet.NewOutPacketToRemovePlayer(
		uid,
	)
	if err := cnt.writeWithComp(packet); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) SpawnPlayer(
	eid int32, uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
) error {
	md := metadata.NewPlayerMetadata()
	pk := packet.NewOutPacketToSpawnPlayer(
		eid, uid,
		x, y, z,
		yaw, pitch,
		md,
	)
	if err := cnt.writeWithComp(
		pk,
	); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) SpawnItemStand(
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
	it item.Item,
) error {
	md := metadata.NewArmorStandMetadata()
	md.SetActions(
		false,
		false,
		true,
	)
	md.SetRightArmRotation(0, 0, 270)
	pk0 := packet.NewOutPacketToSpawnArmorStand(
		eid,
		uid,
		x-0.3, y-1.27, z-0.3,
		yaw, pitch,
		md,
	)
	if err := cnt.writeWithComp(
		pk0,
	); err != nil {
		return err
	}

	pk1 := packet.NewOutPacketToSetEntityEquipment(
		eid,
		0,
		it,
	)
	if err := cnt.writeWithComp(
		pk1,
	); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) SetEntityLook(
	eid int32,
	yaw, pitch float32,
	ground bool,
) error {

	packet0 := packet.NewOutPacketToSetEntityLook(
		eid,
		yaw, pitch,
		ground,
	)
	if err := cnt.writeWithComp(packet0); err != nil {
		return err
	}

	packet1 := packet.NewOutPacketToSetEntityHeadLook(
		eid,
		yaw,
	)
	if err := cnt.writeWithComp(packet1); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) SetEntityRelativeMove(
	eid int32,
	dx, dy, dz int16,
	ground bool,
) error {

	pk := packet.NewOutPacketToSetEntityRelativeMove(
		eid,
		dx, dy, dz,
		ground,
	)
	if err := cnt.writeWithComp(
		pk,
	); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) SetPlayerMetadata(
	eid int32,
	md *metadata.PlayerMetadata,
) error {

	pk := packet.NewOutPacketToSetPlayerMetadata(
		eid,
		md,
	)
	if err := cnt.writeWithComp(
		pk,
	); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) DespawnEntity(
	eid int32,
) error {

	packet := packet.NewOutPacketToDespawnEntity(
		eid,
	)
	if err := cnt.writeWithComp(packet); err != nil {
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
	packet := packet.NewOutPacketToSendChunkData(
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

	packet := packet.NewOutPacketToUnloadChunk(
		cx, cz,
	)
	if err := cnt.writeWithComp(packet); err != nil {
		return err
	}

	return nil
}

func (cnt *Client) CheckKeepAlive(
	payload int64,
) error {

	packet := packet.NewOutPacketToCheckKeepAlive(payload)
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
		"{ addr: %s }",
		cnt.addr,
	)
}
