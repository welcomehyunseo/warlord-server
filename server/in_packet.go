package server

import "fmt"

const HandshakePacketID = 0x00

const RequestPacketID = 0x00
const PingPacketID = 0x01

const StartLoginPacketID = 0x00

const FinishTeleportPacketID = 0x00
const DemandPacketID = 0x03
const ChangeSettingsPacketID = 0x04
const ConfirmKeepAlivePacketID = 0x0B
const ChangePosPacketID = 0x0D
const ChangePosAndLookPacketID = 0x0E

type InPacket interface {
	*Packet

	Unpack(*Data)
}

type HandshakePacket struct {
	*packet
	version int32  // protocol version
	addr    string // server address
	port    uint16 // server port
	next    int32  // next state
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

func (p *HandshakePacket) Unpack(
	data *Data,
) {
	p.version = data.ReadVarInt()
	p.addr = data.ReadString()
	p.port = data.ReadUint16()
	p.next = data.ReadVarInt()
}

func (p *HandshakePacket) GetVersion() int32 {
	return p.version
}

func (p *HandshakePacket) GetAddr() string {
	return p.addr
}

func (p *HandshakePacket) GetPort() uint16 {
	return p.port
}

func (p *HandshakePacket) GetNext() int32 {
	return p.next
}

func (p *HandshakePacket) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"version: %d, "+
			"addr: %s, "+
			"port: %d, "+
			"next: %d "+
			"} ",
		p.packet, p.version, p.addr, p.port, p.next,
	)
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

func (p *RequestPacket) Unpack(
	data *Data,
) {
}

func (p *RequestPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v }",
		p.packet,
	)
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

func (p *PingPacket) Unpack(
	data *Data,
) {
	p.payload = data.ReadInt64()
}

func (p *PingPacket) GetPayload() int64 {
	return p.payload
}

func (p *PingPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, payload: %d }",
		p.packet, p.payload,
	)
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

func (p *StartLoginPacket) Unpack(
	data *Data,
) {
	p.username = data.ReadString()
}

func (p *StartLoginPacket) GetUsername() string {
	return p.username
}

func (p *StartLoginPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, username: %s }",
		p.packet, p.username,
	)
}

type FinishTeleportPacket struct {
	*packet
	payload int32
}

func NewFinishTeleportPacket() *FinishTeleportPacket {
	return &FinishTeleportPacket{
		packet: newPacket(
			Inbound,
			PlayState,
			FinishTeleportPacketID,
		),
	}
}

func (p *FinishTeleportPacket) Unpack(data *Data) {
	p.payload = data.ReadVarInt()
}

func (p *FinishTeleportPacket) GetPayload() int32 {
	return p.payload
}

func (p *FinishTeleportPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, payload: %d }",
		p.packet, p.payload,
	)
}

type DemandPacket struct {
	*packet
	respawn bool // when the client is ready to complete login and respawn after death
	stats   bool // when the client opens the statistics menu
}

func NewDemandPacket() *DemandPacket {
	return &DemandPacket{
		packet: newPacket(
			Inbound,
			PlayState,
			DemandPacketID,
		),
	}
}

func (p *DemandPacket) Unpack(data *Data) {
	action := data.ReadVarInt()
	if action == 0 {
		p.respawn = true
		p.stats = false
	} else {
		p.respawn = false
		p.stats = true
	}
}

func (p *DemandPacket) GetRespawn() bool {
	return p.respawn
}

func (p *DemandPacket) GetStats() bool {
	return p.stats
}

func (p *DemandPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, respawn: %v, stats: %v }",
		p.packet, p.respawn, p.stats,
	)
}

type ChangeSettingsPacket struct {
	*packet
	local       string
	rndDist     int8
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

func NewChangeSettingsPacket() *ChangeSettingsPacket {
	return &ChangeSettingsPacket{
		packet: newPacket(
			Inbound,
			PlayState,
			ChangeSettingsPacketID,
		),
	}
}

func (p *ChangeSettingsPacket) Unpack(
	data *Data,
) {
	p.local = data.ReadString()
	p.rndDist = data.ReadInt8()
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

func (p *ChangeSettingsPacket) GetLocal() string {
	return p.local
}

func (p *ChangeSettingsPacket) GetRndDist() int8 {
	return p.rndDist
}

func (p *ChangeSettingsPacket) GetChatMode() int32 {
	return p.chatMode
}

func (p *ChangeSettingsPacket) GetChatColors() bool {
	return p.chatColors
}

func (p *ChangeSettingsPacket) GetCape() bool {
	return p.cape
}

func (p *ChangeSettingsPacket) GetJacket() bool {
	return p.jacket
}

func (p *ChangeSettingsPacket) GetLeftSleeve() bool {
	return p.leftSleeve
}

func (p *ChangeSettingsPacket) GetRightSleeve() bool {
	return p.rightSleeve
}

func (p *ChangeSettingsPacket) GetLeftPants() bool {
	return p.leftPants
}

func (p *ChangeSettingsPacket) GetRightPants() bool {
	return p.rightPants
}

func (p *ChangeSettingsPacket) GetHat() bool {
	return p.hat
}

func (p *ChangeSettingsPacket) GetMainHand() int32 {
	return p.mainHand
}

func (p *ChangeSettingsPacket) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"local: %s, "+
			"rndDist: %d, "+
			"chatMode: %d, "+
			"chatColors: %v, "+
			"cape: %v, jacket: %v, "+
			"leftSleeve: %v, rightSleeve: %v, "+
			"leftPants: %v, rightPants: %v, "+
			"hat: %v, "+
			"mainHand: %d "+
			"}",
		p.packet,
		p.local,
		p.rndDist,
		p.chatMode,
		p.chatColors,
		p.cape, p.jacket,
		p.leftSleeve, p.rightSleeve,
		p.leftPants, p.rightPants,
		p.hat,
		p.mainHand,
	)
}

type ConfirmKeepAlivePacket struct {
	*packet
	payload int64
}

func NewConfirmKeepAlivePacket() *ConfirmKeepAlivePacket {
	return &ConfirmKeepAlivePacket{
		packet: newPacket(
			Inbound,
			PlayState,
			ConfirmKeepAlivePacketID,
		),
	}
}

func (p *ConfirmKeepAlivePacket) Unpack(data *Data) {
	p.payload = data.ReadInt64()
}

func (p *ConfirmKeepAlivePacket) GetPayload() int64 {
	return p.payload
}

func (p *ConfirmKeepAlivePacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, payload: %d }",
		p.packet, p.payload,
	)
}

type ChangePosPacket struct {
	*packet
	x      float64
	y      float64
	z      float64
	ground bool
}

func NewChangePlayerPosPacket() *ChangePosPacket {
	return &ChangePosPacket{
		packet: newPacket(
			Inbound,
			PlayState,
			ChangePosPacketID,
		),
	}
}

func (p *ChangePosPacket) Unpack(data *Data) {
	p.x = data.ReadFloat64()
	p.y = data.ReadFloat64()
	p.z = data.ReadFloat64()
	p.ground = data.ReadBool()
}

func (p *ChangePosPacket) GetX() float64 {
	return p.x
}

func (p *ChangePosPacket) GetY() float64 {
	return p.y
}

func (p *ChangePosPacket) GetZ() float64 {
	return p.z
}

func (p *ChangePosPacket) GetGround() bool {
	return p.ground
}

func (p *ChangePosPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, x: %f, y: %f, z: %f, ground: %v }",
		p.packet, p.x, p.y, p.z, p.ground,
	)
}

type ChangePosAndLookPacket struct {
	*packet
	x      float64
	y      float64
	z      float64
	yaw    float32
	pitch  float32
	ground bool
}

func NewChangePosAndLookPacket() *ChangePosAndLookPacket {
	return &ChangePosAndLookPacket{
		packet: newPacket(
			Inbound,
			PlayState,
			ChangePosAndLookPacketID,
		),
	}
}

func (p *ChangePosAndLookPacket) Unpack(data *Data) {
	p.x = data.ReadFloat64()
	p.y = data.ReadFloat64()
	p.z = data.ReadFloat64()
	p.yaw = data.ReadFloat32()
	p.pitch = data.ReadFloat32()
	p.ground = data.ReadBool()
}

func (p *ChangePosAndLookPacket) GetX() float64 {
	return p.x
}

func (p *ChangePosAndLookPacket) GetY() float64 {
	return p.y
}

func (p *ChangePosAndLookPacket) GetZ() float64 {
	return p.z
}

func (p *ChangePosAndLookPacket) GetYaw() float32 {
	return p.yaw
}

func (p *ChangePosAndLookPacket) GetPitch() float32 {
	return p.pitch
}

func (p *ChangePosAndLookPacket) GetGround() bool {
	return p.ground
}

func (p *ChangePosAndLookPacket) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"x: %f, y: %f, z: %f, "+
			"yaw: %f, pitch: %f, "+
			"ground: %v "+
			"}",
		p.packet,
		p.x, p.y, p.z,
		p.yaw, p.pitch,
		p.ground,
	)
}
