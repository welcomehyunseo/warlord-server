package server

const (
	HandshakePacketID = 0x00

	RequestPacketID = 0x00
	PingPacketID    = 0x01

	StartLoginPacketID = 0x00
)

type InPacket interface {
	Read(*Data)

	GetBoundTo() int
	GetState() int
	GetID() int32
}

type HandshakePacket struct {
	*packet
	protocolVersion int32
	serverAddress   string
	serverPort      uint16
	nextState       int32
}

func NewHandshakePacket() *HandshakePacket {
	return &HandshakePacket{
		packet: newPacket(
			Inbound,
			HandshakingState,
			HandshakePacketID,
		),
	}
}

func (p *HandshakePacket) Read(
	data *Data,
) {
	p.protocolVersion = data.ReadVarInt()
	p.serverAddress = data.ReadString()
	p.serverPort = data.ReadUint16()
	p.nextState = data.ReadVarInt()
}

func (p *HandshakePacket) GetProtocolVersion() int32 {
	return p.protocolVersion
}

func (p *HandshakePacket) GetServerAddress() string {
	return p.serverAddress
}

func (p *HandshakePacket) GetServerPort() uint16 {
	return p.serverPort
}

func (p *HandshakePacket) GetNextState() int32 {
	return p.nextState
}

type RequestPacket struct {
	*packet
}

func NewRequestPacket() *RequestPacket {
	return &RequestPacket{
		packet: newPacket(
			Inbound,
			StatusState,
			RequestPacketID,
		),
	}
}

func (p *RequestPacket) Read(
	data *Data,
) {
}

type PingPacket struct {
	*packet
	payload int64
}

func NewPingPacket() *PingPacket {
	return &PingPacket{
		packet: newPacket(
			Inbound,
			StatusState,
			PingPacketID,
		),
	}
}

func (p *PingPacket) Read(
	data *Data,
) {
	p.payload = data.ReadInt64()
}

func (p *PingPacket) GetPayload() int64 {
	return p.payload
}

type StartLoginPacket struct {
	*packet
	username string
}

func NewStartLoginPacket() *StartLoginPacket {
	return &StartLoginPacket{
		packet: newPacket(
			Inbound,
			LoginState,
			StartLoginPacketID,
		),
	}
}

func (p *StartLoginPacket) Read(
	data *Data,
) {
	p.username = data.ReadString()
}

func (p *StartLoginPacket) GetUsername() string {
	return p.username
}
