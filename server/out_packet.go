package server

import (
	"encoding/json"
	"github.com/google/uuid"
)

const ResponsePacketID = 0x00
const PongPacketID = 0x01

const DisconnectLoginPacketID = 0x00
const CompleteLoginPacketID = 0x02

const JoinGamePacketID = 0x23
const SetPlayerAbilitiesPacketID = 0x2C
const SetPlayerPosAndLookPacketID = 0x2F
const SetSpawnPosPacketID = 0x46

type OutPacket interface {
	Write() *Data

	GetBoundTo() int
	GetState() int
	GetID() int32
}

type Version struct {
	Name     string `json:"name"`
	Protocol int32  `json:"protocol"`
}

type Sample struct {
	Name string    `json:"name"`
	Id   uuid.UUID `json:"playerID"`
}

type Players struct {
	Max    int       `json:"max"`
	Online int       `json:"online"`
	Sample []*Sample `json:"sample"`
}

type Description struct {
	Text string `json:"text"`
}

type JsonResponse struct {
	Version            *Version     `json:"version"`
	Players            *Players     `json:"players"`
	Description        *Description `json:"description"`
	Favicon            string       `json:"favicon"`
	PreviewsChat       bool         `json:"previewsChat"`
	EnforcesSecureChat bool         `json:"enforcesSecureChat"`
}

type ResponsePacket struct {
	*packet
	jsonResponse *JsonResponse
}

func NewResponsePacket(
	jsonResponse *JsonResponse,
) *ResponsePacket {
	return &ResponsePacket{
		packet: newPacket(
			Outbound,
			StatusState,
			ResponsePacketID,
		),
		jsonResponse: jsonResponse,
	}
}

func (p *ResponsePacket) Write() *Data {
	d0 := NewData()
	d0.WriteVarInt(p.GetID())
	buf, _ := json.Marshal(p.jsonResponse)
	d0.WriteString(string(buf))

	length := d0.GetLength()
	d1 := NewData()
	d1.WriteVarInt(int32(length))
	d1.Write(d0)
	return d1
}

func (p *ResponsePacket) GetJsonResponse() *JsonResponse {
	return p.jsonResponse
}

type PongPacket struct {
	*packet
	payload int64
}

func NewPongPacket(
	payload int64,
) *PongPacket {
	return &PongPacket{
		packet: newPacket(
			Outbound,
			StatusState,
			PongPacketID,
		),
		payload: payload,
	}
}

func (p *PongPacket) Write() *Data {
	d0 := NewData()
	d0.WriteVarInt(p.GetID())
	d0.WriteInt64(p.payload)

	length := d0.GetLength()
	d1 := NewData()
	d1.WriteVarInt(int32(length))
	d1.Write(d0)
	return d1
}

func (p *PongPacket) GetPayload() int64 {
	return p.payload
}

type CompleteLoginPacket struct {
	*packet
	playerID uuid.UUID
	username string
}

func NewCompleteLoginPacket(
	playerID uuid.UUID,
	username string,
) *CompleteLoginPacket {
	return &CompleteLoginPacket{
		packet: newPacket(
			Outbound,
			LoginState,
			CompleteLoginPacketID,
		),
		playerID: playerID,
		username: username,
	}
}

func (p *CompleteLoginPacket) Write() *Data {
	d0 := NewData()
	d0.WriteVarInt(p.GetID())
	d0.WriteString(p.playerID.String())
	d0.WriteString(p.username)

	length := d0.GetLength()
	d1 := NewData()
	d1.WriteVarInt(int32(length))
	d1.Write(d0)
	return d1
}

func (p *CompleteLoginPacket) GetPlayerID() uuid.UUID {
	return p.playerID
}

func (p *CompleteLoginPacket) GetUsername() string {
	return p.username
}

type JoinGamePacket struct {
	*packet
	eid        int32
	gamemode   uint8
	dimension  int32
	difficulty uint8
	level      string
	debug      bool
}

func NewJoinGamePacket(
	eid int32,
	gamemode uint8,
	dimension int32,
	difficulty uint8,
	level string,
	debug bool,
) *JoinGamePacket {
	return &JoinGamePacket{
		packet: newPacket(
			Outbound,
			PlayState,
			JoinGamePacketID,
		),
		eid:        eid,
		gamemode:   gamemode,
		dimension:  dimension,
		difficulty: difficulty,
		level:      level,
		debug:      debug,
	}
}

func (p *JoinGamePacket) Write() *Data {
	d0 := NewData()
	d0.WriteVarInt(p.GetID())
	d0.WriteInt32(p.eid)
	d0.WriteUint8(p.gamemode)
	d0.WriteInt32(p.dimension)
	d0.WriteUint8(p.difficulty)
	d0.WriteUint8(0) // max is ignored
	d0.WriteString(p.level)
	d0.WriteBool(p.debug)

	length := d0.GetLength()
	d1 := NewData()
	d1.WriteVarInt(int32(length))
	d1.Write(d0)
	return d1
}

func (p *JoinGamePacket) GetEid() int32 {
	return p.eid
}

func (p *JoinGamePacket) GetGamemode() uint8 {
	return p.gamemode
}

func (p *JoinGamePacket) GetDimension() int32 {
	return p.dimension
}

func (p *JoinGamePacket) GetDifficulty() uint8 {
	return p.difficulty
}

func (p *JoinGamePacket) GetLevel() string {
	return p.level
}

func (p *JoinGamePacket) GetDebug() bool {
	return p.debug
}

type SetPlayerAbilitiesPacket struct {
	*packet
	invulnerable bool
	flying       bool
	allowFlying  bool
	instantBreak bool
	flyingSpeed  float32
	fovModifier  float32
}

func NewSetPlayerAbilitiesPacket(
	invulnerable bool,
	flying bool,
	allowFlying bool,
	instantBreak bool,
	flyingSpeed float32,
	fovModifier float32,
) *SetPlayerAbilitiesPacket {
	return &SetPlayerAbilitiesPacket{
		packet: newPacket(
			Outbound,
			PlayState,
			SetPlayerAbilitiesPacketID,
		),
		invulnerable: invulnerable,
		flying:       flying,
		allowFlying:  allowFlying,
		instantBreak: instantBreak,
		flyingSpeed:  flyingSpeed,
		fovModifier:  fovModifier,
	}
}

func (p *SetPlayerAbilitiesPacket) Write() *Data {
	d0 := NewData()
	d0.WriteVarInt(p.GetID())
	bitmask := uint8(0)
	if p.invulnerable == true {
		bitmask |= uint8(1)
	}
	if p.flying == true {
		bitmask |= uint8(2)
	}
	if p.allowFlying == true {
		bitmask |= uint8(4)
	}
	if p.instantBreak == true {
		bitmask |= uint8(8)
	}
	d0.WriteUint8(bitmask)
	d0.WriteFloat32(p.flyingSpeed)
	d0.WriteFloat32(p.fovModifier)

	length := d0.GetLength()
	d1 := NewData()
	d1.WriteVarInt(int32(length))
	d1.Write(d0)
	return d1
}

func (p *SetPlayerAbilitiesPacket) GetInvulnerable() bool {
	return p.invulnerable
}

func (p *SetPlayerAbilitiesPacket) GetFlying() bool {
	return p.flying
}

func (p *SetPlayerAbilitiesPacket) GetAllowFlying() bool {
	return p.allowFlying
}

func (p *SetPlayerAbilitiesPacket) GetInstantBreak() bool {
	return p.instantBreak
}

func (p *SetPlayerAbilitiesPacket) GetFlyingSpeed() float32 {
	return p.flyingSpeed
}

func (p *SetPlayerAbilitiesPacket) GetFovModifier() float32 {
	return p.fovModifier
}

type SetPlayerPosAndLookPacket struct {
	*packet
	x       float64
	y       float64
	z       float64
	yaw     float32
	pitch   float32
	payload int32
}

func NewSetPlayerPosAndLookPacket(
	x float64,
	y float64,
	z float64,
	yaw float32,
	pitch float32,
	payload int32,
) *SetPlayerPosAndLookPacket {
	return &SetPlayerPosAndLookPacket{
		packet: newPacket(
			Outbound,
			PlayState,
			SetPlayerPosAndLookPacketID,
		),
		x:       x,
		y:       y,
		z:       z,
		yaw:     yaw,
		pitch:   pitch,
		payload: payload,
	}
}

func (p *SetPlayerPosAndLookPacket) Write() *Data {
	d0 := NewData()
	d0.WriteVarInt(p.GetID())
	d0.WriteFloat64(p.x)
	d0.WriteFloat64(p.y)
	d0.WriteFloat64(p.z)
	d0.WriteFloat32(p.yaw)
	d0.WriteFloat32(p.pitch)
	d0.WriteInt8(0)
	d0.WriteVarInt(p.payload)

	length := d0.GetLength()
	d1 := NewData()
	d1.WriteVarInt(int32(length))
	d1.Write(d0)
	return d1
}

func (p *SetPlayerPosAndLookPacket) GetX() float64 {
	return p.x
}

func (p *SetPlayerPosAndLookPacket) GetY() float64 {
	return p.y
}

func (p *SetPlayerPosAndLookPacket) GetZ() float64 {
	return p.z
}

func (p *SetPlayerPosAndLookPacket) GetYaw() float32 {
	return p.yaw
}

func (p *SetPlayerPosAndLookPacket) GetPitch() float32 {
	return p.pitch
}

func (p *SetPlayerPosAndLookPacket) GetPayload() int32 {
	return p.payload
}

type SetSpawnPosPacket struct {
	*packet
	x int
	y int
	z int
}

func NewSetSpawnPosPacket(
	x int,
	y int,
	z int,
) *SetSpawnPosPacket {
	return &SetSpawnPosPacket{
		packet: newPacket(
			Outbound,
			PlayState,
			SetSpawnPosPacketID,
		),
		x: x,
		y: y,
		z: z,
	}
}

func (p *SetSpawnPosPacket) Write() *Data {
	d0 := NewData()
	d0.WriteVarInt(p.GetID())
	d0.WritePosition(p.x, p.y, p.z)

	length := d0.GetLength()
	d1 := NewData()
	d1.WriteVarInt(int32(length))
	d1.Write(d0)
	return d1
}

func (p *SetSpawnPosPacket) GetX() int {
	return p.x
}

func (p *SetSpawnPosPacket) GetY() int {
	return p.y
}

func (p *SetSpawnPosPacket) GetZ() int {
	return p.z
}
