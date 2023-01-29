package server

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
)

const ResponsePacketID = 0x00
const PongPacketID = 0x01

const DisconnectLoginPacketID = 0x00
const CompleteLoginPacketID = 0x02
const EnableCompressionPacketID = 0x03

const UnloadChunkPacketID = 0x1D
const SendChunkDataPacketID = 0x20
const JoinGamePacketID = 0x23
const SetPlayerAbilitiesPacketID = 0x2C
const SetPlayerPosAndLookPacketID = 0x2F
const SetSpawnPosPacketID = 0x46

type OutPacket interface {
	Packet

	Write() *Data
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
	data := NewData()
	buf, _ := json.Marshal(p.jsonResponse)
	data.WriteString(string(buf))

	return data
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
	data := NewData()
	data.WriteInt64(p.payload)

	return data
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
	data := NewData()
	data.WriteString(p.playerID.String())
	data.WriteString(p.username)

	return data
}

func (p *CompleteLoginPacket) GetPlayerID() uuid.UUID {
	return p.playerID
}

func (p *CompleteLoginPacket) GetUsername() string {
	return p.username
}

type EnableCompressionPacket struct {
	*packet
	threshold int32
}

func NewEnableCompressionPacket(
	threshold int32,
) *EnableCompressionPacket {
	return &EnableCompressionPacket{
		packet: newPacket(
			Outbound,
			LoginState,
			EnableCompressionPacketID,
		),
		threshold: threshold,
	}
}

func (p *EnableCompressionPacket) Write() *Data {
	data := NewData()
	data.WriteVarInt(p.threshold)

	return data
}

func (p *EnableCompressionPacket) GetThreshold() int32 {
	return p.threshold
}

type UnloadChunkPacket struct {
	*packet
	cx int32
	cz int32
}

func NewUnloadChunkPacket(
	cx, cz int32,
) *UnloadChunkPacket {
	return &UnloadChunkPacket{
		packet: newPacket(
			Outbound,
			PlayState,
			UnloadChunkPacketID,
		),
		cx: cx,
		cz: cz,
	}
}

func (p *UnloadChunkPacket) Write() *Data {
	data := NewData()
	data.WriteInt32(p.cx)
	data.WriteInt32(p.cz)

	return data
}

func (p *UnloadChunkPacket) GetCx() int32 {
	return p.cx
}

func (p *UnloadChunkPacket) GetCz() int32 {
	return p.cz
}

type SendChunkDataPacket struct {
	*packet
	cx      int32
	cz      int32
	init    bool
	bitmask uint16
	data    []uint8
}

func NewSendChunkDataPacket(
	cx, cz int32,
	init bool,
	bitmask uint16,
	data []uint8,
) *SendChunkDataPacket {
	return &SendChunkDataPacket{
		packet: newPacket(
			Outbound,
			PlayState,
			SendChunkDataPacketID,
		),
		cx:      cx,
		cz:      cz,
		init:    init,
		bitmask: bitmask,
		data:    data,
	}
}

func (p *SendChunkDataPacket) Write() *Data {
	data := NewData()
	data.WriteInt32(p.cx)
	data.WriteInt32(p.cz)
	data.WriteBool(p.init)
	data.WriteVarInt(int32(p.bitmask))
	l0 := len(p.data)
	data.WriteVarInt(int32(l0))
	data.WriteBytes(p.data)

	l1 := 0
	data.WriteVarInt(int32(l1)) // block entities

	return data
}

func (p *SendChunkDataPacket) GetCx() int32 {
	return p.cx
}

func (p *SendChunkDataPacket) GetCz() int32 {
	return p.cz
}

func (p *SendChunkDataPacket) GetInit() bool {
	return p.init
}

func (p *SendChunkDataPacket) GetBitmask() uint16 {
	return p.bitmask
}

func (p *SendChunkDataPacket) GetData() []uint8 {
	return p.data
}

func (p *SendChunkDataPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, cx: %d, cz: %d, init: %v, bitmask: %d, data: [...] }",
		p.packet, p.cx, p.cz, p.init, p.bitmask,
	)
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
	data := NewData()
	data.WriteInt32(p.eid)
	data.WriteUint8(p.gamemode)
	data.WriteInt32(p.dimension)
	data.WriteUint8(p.difficulty)
	data.WriteUint8(0) // max is ignored
	data.WriteString(p.level)
	data.WriteBool(p.debug)

	return data
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
	invulnerable,
	flying,
	allowFlying,
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
	data := NewData()
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
	data.WriteUint8(bitmask)
	data.WriteFloat32(p.flyingSpeed)
	data.WriteFloat32(p.fovModifier)

	return data
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
	data := NewData()
	data.WriteFloat64(p.x)
	data.WriteFloat64(p.y)
	data.WriteFloat64(p.z)
	data.WriteFloat32(p.yaw)
	data.WriteFloat32(p.pitch)
	data.WriteInt8(0)
	data.WriteVarInt(p.payload)

	return data
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
	data := NewData()
	data.WritePosition(p.x, p.y, p.z)

	return data
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
