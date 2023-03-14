package packet

import "fmt"

const (
	Inbound  = int(0)
	Outbound = int(1)

	NilState         = int32(-1)
	PlayState        = int32(0)
	StatusState      = int32(1)
	LoginState       = int32(2)
	HandshakingState = int32(3)
)

type Packet interface {
	GetBoundTo() int
	GetState() int32
	GetID() int32
}

type packet struct {
	bound int
	state int32
	id    int32
}

func newPacket(
	bound int,
	state int32,
	id int32,
) *packet {
	return &packet{
		bound: bound,
		state: state,
		id:    id,
	}
}

func (p *packet) GetBoundTo() int {
	return p.bound
}

func (p *packet) GetState() int32 {
	return p.state
}

func (p *packet) GetID() int32 {
	return p.id
}

func (p *packet) String() string {
	return fmt.Sprintf(
		"{ bound: %d, state: %d, id: %d }",
		p.bound, p.state, p.id,
	)
}

type InPacket interface {
	Packet

	Unpack(
		[]byte,
	) error
}

type OutPacket interface {
	Packet

	Pack() (
		[]byte,
		error,
	)
}
