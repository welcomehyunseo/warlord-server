package packet

import (
	"fmt"
	"github.com/welcomehyunseo/warlord-server/server/data"
)

const InPacketIDToHandshake = 0x00

const InPacketIDToRequest = 0x00
const InPacketIDToPing = 0x01

const InPacketIDToStartLogin = 0x00

type InPacketToHandshake struct {
	*packet
	ver  int32
	addr string
	port uint16
	next int32
}

func NewInPacketToHandshake() *InPacketToHandshake {
	return &InPacketToHandshake{
		packet: newPacket(
			Inbound,
			HandshakingState,
			InPacketIDToHandshake,
		),
	}
}

func (p *InPacketToHandshake) Unpack(
	arr []byte,
) error {
	dt := data.NewDataWithBytes(arr)

	ver, err := dt.ReadVarInt()
	if err != nil {
		return err
	}
	p.ver = ver

	addr, err := dt.ReadString()
	if err != nil {
		return err
	}
	p.addr = addr

	port, err := dt.ReadUint16()
	if err != nil {
		return err
	}
	p.port = port

	next, err := dt.ReadVarInt()
	if err != nil {
		return err
	}
	p.next = next

	return nil
}

func (p *InPacketToHandshake) GetVersion() int32 {
	return p.ver
}

func (p *InPacketToHandshake) GetAddress() string {
	return p.addr
}

func (p *InPacketToHandshake) GetPort() uint16 {
	return p.port
}

func (p *InPacketToHandshake) GetNestState() int32 {
	return p.next
}

func (p *InPacketToHandshake) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"ver: %d, "+
			"addr: %s, "+
			"port: %d, "+
			"next: %d "+
			"} ",
		p.packet,
		p.ver,
		p.addr,
		p.port,
		p.next,
	)
}

type InPacketToRequest struct {
	*packet
}

func NewInPacketToRequest() *InPacketToRequest {
	return &InPacketToRequest{
		packet: newPacket(
			Inbound,
			StatusState,
			InPacketIDToRequest,
		),
	}
}

func (p *InPacketToRequest) Unpack(
	arr []byte,
) error {

	return nil
}

func (p *InPacketToRequest) String() string {
	return fmt.Sprintf(
		"{ packet: %+v }",
		p.packet,
	)
}

type InPacketToPing struct {
	*packet
	payload int64
}

func NewInPacketToPing() *InPacketToPing {
	return &InPacketToPing{
		packet: newPacket(
			Inbound,
			StatusState,
			InPacketIDToPing,
		),
	}
}

func (p *InPacketToPing) Unpack(
	arr []byte,
) error {
	dt := data.NewDataWithBytes(arr)

	payload, err := dt.ReadInt64()
	if err != nil {
		return err
	}
	p.payload = payload

	return nil
}

func (p *InPacketToPing) GetPayload() int64 {
	return p.payload
}

func (p *InPacketToPing) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, payload: %d }",
		p.packet, p.payload,
	)
}

type InPacketToStartLogin struct {
	*packet
	username string
}

func NewInPacketToStartLogin() *InPacketToStartLogin {
	return &InPacketToStartLogin{
		packet: newPacket(
			Inbound,
			LoginState,
			InPacketIDToStartLogin,
		),
	}
}

func (p *InPacketToStartLogin) Unpack(
	arr []byte,
) error {
	dt := data.NewDataWithBytes(arr)

	username, err := dt.ReadString()
	if err != nil {
		return nil
	}
	p.username = username

	return nil
}

func (p *InPacketToStartLogin) GetUsername() string {
	return p.username
}

func (p *InPacketToStartLogin) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, username: %s }",
		p.packet, p.username,
	)
}
