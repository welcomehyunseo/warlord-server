package server

const HandshakePacketID = 0x00

const RequestPacketID = 0x00
const PingPacketID = 0x01

const StartLoginPacketID = 0x00

const TakeActionPacketID = 0x03
const ChangeClientSettingsPacketID = 0x04

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

type TakeActionPacket struct {
	*packet
	respawn bool
	stats   bool
}

func NewTakeActionPacket() *TakeActionPacket {
	return &TakeActionPacket{
		packet: newPacket(
			Inbound,
			LoginState,
			TakeActionPacketID,
		),
	}
}

func (p *TakeActionPacket) Read(data *Data) {
	action := data.ReadVarInt()
	if action == 0 {
		p.respawn = true
		p.stats = false
	} else {
		p.respawn = false
		p.stats = true
	}
}

func (p *TakeActionPacket) GetRespawn() bool {
	return p.respawn
}

func (p *TakeActionPacket) GetStats() bool {
	return p.stats
}

type ChangeClientSettingsPacket struct {
	*packet
	local       string
	viewDist    int8
	chatMode    int32
	chatColors  bool
	cape        bool
	jacket      bool
	leftSleeve  bool
	rightSleeve bool
	leftPants   bool
	rightPants  bool
	hat         bool
	mainHand    int32
}

func NewChangeClientSettingsPacket() *ChangeClientSettingsPacket {
	return &ChangeClientSettingsPacket{
		packet: newPacket(
			Inbound,
			StatusState,
			ChangeClientSettingsPacketID,
		),
	}
}

func (p *ChangeClientSettingsPacket) Read(
	data *Data,
) {
	p.local = data.ReadString()
	p.viewDist = data.ReadInt8()
	p.chatMode = data.ReadVarInt()
	p.chatColors = data.ReadBool()
	bitmask := data.ReadUint8()
	if bitmask&uint8(1) == uint8(1) {
		p.cape = true
	} else {
		p.cape = false
	}
	if bitmask&uint8(2) == uint8(2) {
		p.jacket = true
	} else {
		p.jacket = false
	}
	if bitmask&uint8(4) == uint8(4) {
		p.leftSleeve = true
	} else {
		p.leftSleeve = false
	}
	if bitmask&uint8(8) == uint8(8) {
		p.rightSleeve = true
	} else {
		p.rightSleeve = false
	}
	if bitmask&uint8(16) == uint8(16) {
		p.leftPants = true
	} else {
		p.leftPants = false
	}
	if bitmask&uint8(32) == uint8(32) {
		p.rightPants = true
	} else {
		p.rightPants = false
	}
	if bitmask&uint8(64) == uint8(64) {
		p.hat = true
	} else {
		p.hat = false
	}
	p.mainHand = data.ReadVarInt()
}

func (p *ChangeClientSettingsPacket) getLocal() string {
	return p.local
}

func (p *ChangeClientSettingsPacket) getViewDist() int8 {
	return p.viewDist
}

func (p *ChangeClientSettingsPacket) getChatMode() int32 {
	return p.chatMode
}

func (p *ChangeClientSettingsPacket) getChatColors() bool {
	return p.chatColors
}

func (p *ChangeClientSettingsPacket) getCape() bool {
	return p.cape
}

func (p *ChangeClientSettingsPacket) getJacket() bool {
	return p.jacket
}

func (p *ChangeClientSettingsPacket) getLeftSleeve() bool {
	return p.leftSleeve
}

func (p *ChangeClientSettingsPacket) getRightSleeve() bool {
	return p.rightSleeve
}

func (p *ChangeClientSettingsPacket) getLeftPants() bool {
	return p.leftPants
}

func (p *ChangeClientSettingsPacket) getRightPants() bool {
	return p.rightPants
}

func (p *ChangeClientSettingsPacket) getHat() bool {
	return p.hat
}

func (p *ChangeClientSettingsPacket) getMainHand() int32 {
	return p.mainHand
}
