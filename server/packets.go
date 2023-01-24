package server

const (
	HandshakingState = int32(0)
	StatusState      = int32(1)
	LoginState       = int32(2)
	PlayState        = int32(3)

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
