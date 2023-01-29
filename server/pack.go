package server

import "fmt"

type BoundType = int
type State = int32
type PacketID = int32

const (
	NilState         = State(-1)
	PlayState        = State(0)
	StatusState      = State(1)
	LoginState       = State(2)
	HandshakingState = State(3)

	Inbound  = BoundType(0)
	Outbound = BoundType(1)
)

type Packet interface {
	GetBoundTo() BoundType
	GetState() State
	GetID() PacketID
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

func (p *packet) GetBoundTo() BoundType {
	return p.bound
}

func (p *packet) GetState() State {
	return p.state
}

func (p *packet) GetID() PacketID {
	return p.id
}

func (p *packet) String() string {
	return fmt.Sprintf(
		"{ bound: %d, state: %d, id: %d }",
		p.bound, p.state, p.id,
	)
}
