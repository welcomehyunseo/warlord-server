package server

type State = int32

const (
	NilState         = State(-1)
	PlayState        = State(0)
	StatusState      = State(1)
	LoginState       = State(2)
	HandshakingState = State(3)

	Inbound  = 0
	Outbound = 1
)

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
