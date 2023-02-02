package server

import (
	"fmt"
	"github.com/google/uuid"
)

const ResponsePacketID = 0x00
const PongPacketID = 0x01

const RejectLoginPacketID = 0x00
const CompleteLoginPacketID = 0x02
const EnableCompPacketID = 0x03

const SpawnPlayerPacketID = 0x05
const UnloadChunkPacketID = 0x1D
const CheckKeepAlivePacketID = 0x1F
const SendChunkDataPacketID = 0x20
const JoinGamePacketID = 0x23
const RelativeMovePacketID = 0x26
const SetAbilitiesPacketID = 0x2C
const AddPlayerPacketID = 0x2E
const RemovePlayerPacketID = 0x2E
const UpdateLatencyPacketID = 0x2E
const TeleportPacketID = 0x2F
const DespawnEntityPacketID = 0x32
const SetSpawnPosPacketID = 0x46

type OutPacket interface {
	Packet

	Pack() *Data
}

type ResponsePacket struct {
	*packet
	max     int    // maximum number of players
	online  int    // current number of players
	text    string // string for description
	favicon string // a png image string that is base64 encoded
}

func NewResponsePacket(
	max int,
	online int,
	text string,
	favicon string,
) *ResponsePacket {
	return &ResponsePacket{
		packet: newPacket(
			Outbound,
			StatusState,
			ResponsePacketID,
		),
		max:     max,
		online:  online,
		text:    text,
		favicon: favicon,
	}
}

func (p *ResponsePacket) Pack() *Data {
	data := NewData()
	jsonString := fmt.Sprintf(
		"{"+
			"\"version\":{\"name\":\"%s\",\"protocol\":%d},"+
			"\"players\":{\"max\":%d,\"online\":%d,\"sample\":[]},"+
			"\"description\":{\"text\":\"%s\"},"+
			"\"favicon\":\"%s\","+
			"\"previewsChat\":%v,"+
			"\"enforcesSecureChat\":%v"+
			"}",
		"1.12.2", 340,
		p.max, p.online,
		p.text, p.favicon,
		true, true,
	)
	data.WriteString(jsonString)

	return data
}

func (p *ResponsePacket) GetMax() int {
	return p.max
}

func (p *ResponsePacket) GetOnline() int {
	return p.online
}

func (p *ResponsePacket) GetText() string {
	return p.text
}

func (p *ResponsePacket) GetFavicon() string {
	return p.favicon
}

func (p *ResponsePacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, max: %d, online: %d, text: %s, favicon: %s }",
		p.packet, p.max, p.online, p.text, p.favicon,
	)
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

func (p *PongPacket) Pack() *Data {
	data := NewData()
	data.WriteInt64(p.payload)

	return data
}

func (p *PongPacket) GetPayload() int64 {
	return p.payload
}

func (p *PongPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, payload: %d }",
		p.packet, p.payload,
	)
}

type CompleteLoginPacket struct {
	*packet
	uid      uuid.UUID
	username string
}

func NewCompleteLoginPacket(
	uid uuid.UUID,
	username string,
) *CompleteLoginPacket {
	return &CompleteLoginPacket{
		packet: newPacket(
			Outbound,
			LoginState,
			CompleteLoginPacketID,
		),
		uid:      uid,
		username: username,
	}
}

func (p *CompleteLoginPacket) Pack() *Data {
	data := NewData()
	data.WriteString(p.uid.String())
	data.WriteString(p.username)

	return data
}

func (p *CompleteLoginPacket) GetUUID() uuid.UUID {
	return p.uid
}

func (p *CompleteLoginPacket) GetUsername() string {
	return p.username
}

func (p *CompleteLoginPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, uid: %s, username: %s }",
		p.packet, p.uid, p.username,
	)
}

type EnableCompPacket struct {
	*packet
	threshold int32
}

func NewEnableCompPacket(
	threshold int32,
) *EnableCompPacket {
	return &EnableCompPacket{
		packet: newPacket(
			Outbound,
			LoginState,
			EnableCompPacketID,
		),
		threshold: threshold,
	}
}

func (p *EnableCompPacket) Pack() *Data {
	data := NewData()
	data.WriteVarInt(p.threshold)

	return data
}

func (p *EnableCompPacket) GetThreshold() int32 {
	return p.threshold
}

func (p *EnableCompPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, threshold: %d }",
		p.packet, p.threshold,
	)
}

type SpawnPlayerPacket struct {
	*packet
	eid   int32
	uid   uuid.UUID
	x     float64
	y     float64
	z     float64
	yaw   float32
	pitch float32
	//metadata
}

func NewSpawnPlayerPacket(
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
) *SpawnPlayerPacket {
	return &SpawnPlayerPacket{
		packet: newPacket(
			Outbound,
			PlayState,
			SpawnPlayerPacketID,
		),
		eid:   eid,
		uid:   uid,
		x:     x,
		y:     y,
		z:     z,
		yaw:   yaw,
		pitch: pitch,
	}
}

func (p *SpawnPlayerPacket) Pack() *Data {
	data := NewData()
	data.WriteVarInt(p.eid)
	data.WriteUUID(p.uid)
	data.WriteFloat64(p.x)
	data.WriteFloat64(p.y)
	data.WriteFloat64(p.z)
	data.WriteFloat32(p.yaw)
	data.WriteFloat32(p.pitch)
	data.WriteUint8(0xff)

	return data
}

func (p *SpawnPlayerPacket) GetEID() int32 {
	return p.eid
}

func (p *SpawnPlayerPacket) GetUUID() uuid.UUID {
	return p.uid
}

func (p *SpawnPlayerPacket) GetX() float64 {
	return p.x
}

func (p *SpawnPlayerPacket) GetY() float64 {
	return p.y
}

func (p *SpawnPlayerPacket) GetZ() float64 {
	return p.z
}

func (p *SpawnPlayerPacket) GetYaw() float32 {
	return p.yaw
}

func (p *SpawnPlayerPacket) GetPitch() float32 {
	return p.pitch
}

func (p *SpawnPlayerPacket) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"eid: %d, "+
			"uid: %s, "+
			"x: %f, "+
			"y: %f, "+
			"z: %f, "+
			"yaw: %f, "+
			"pitch: %f "+
			"}",
		p.packet,
		p.eid,
		p.uid,
		p.x,
		p.y,
		p.z,
		p.yaw,
		p.pitch,
	)
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

func (p *UnloadChunkPacket) Pack() *Data {
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

func (p *UnloadChunkPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, cx: %d, cz: %d }",
		p.packet, p.cx, p.cz,
	)
}

type CheckKeepAlivePacket struct {
	*packet
	payload int64
}

func NewCheckKeepAlivePacket(
	payload int64,
) *CheckKeepAlivePacket {
	return &CheckKeepAlivePacket{
		packet: newPacket(
			Outbound,
			PlayState,
			CheckKeepAlivePacketID,
		),
		payload: payload,
	}
}

func (p *CheckKeepAlivePacket) Pack() *Data {
	data := NewData()
	data.WriteInt64(p.payload)

	return data
}

func (p *CheckKeepAlivePacket) GetPayload() int64 {
	return p.payload
}

func (p *CheckKeepAlivePacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, payload: %d }",
		p.packet, p.payload,
	)
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

func (p *SendChunkDataPacket) Pack() *Data {
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

func (p *JoinGamePacket) Pack() *Data {
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

func (p *JoinGamePacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, eid: %d, gamemode: %d, dimension: %d, difficulty: %d, level: %s, debug: %v }",
		p.packet, p.eid, p.gamemode, p.dimension, p.difficulty, p.level, p.debug,
	)
}

type RelativeMovePacket struct {
	*packet
	eid    int32
	deltaX int16
	deltaY int16
	deltaZ int16
	ground bool
}

func NewRelativeMovePacket(
	eid int32,
	deltaX, deltaY, deltaZ int16,
	ground bool,
) *RelativeMovePacket {
	return &RelativeMovePacket{
		packet: newPacket(
			Outbound,
			PlayState,
			RelativeMovePacketID,
		),
		eid:    eid,
		deltaX: deltaX,
		deltaY: deltaY,
		deltaZ: deltaZ,
		ground: ground,
	}
}

func (p *RelativeMovePacket) Pack() *Data {
	data := NewData()
	data.WriteVarInt(p.eid)
	data.WriteInt16(p.deltaX)
	data.WriteInt16(p.deltaY)
	data.WriteInt16(p.deltaZ)
	data.WriteBool(p.ground)

	return data
}

func (p *RelativeMovePacket) GetEID() int32 {
	return p.eid
}

func (p *RelativeMovePacket) GetDeltaX() int16 {
	return p.deltaX
}

func (p *RelativeMovePacket) GetDeltaY() int16 {
	return p.deltaY
}

func (p *RelativeMovePacket) GetDeltaZ() int16 {
	return p.deltaZ
}

func (p *RelativeMovePacket) GetGround() bool {
	return p.ground
}

func (p *RelativeMovePacket) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"eid: %d, "+
			"deltaX: %d, "+
			"deltaY: %d, "+
			"deltaZ: %d, "+
			"ground: %v "+
			"}",
		p.packet,
		p.eid,
		p.deltaX,
		p.deltaY,
		p.deltaZ,
		p.ground,
	)
}

type SetAbilitiesPacket struct {
	*packet
	invulnerable bool
	flying       bool
	canFly       bool
	instantBreak bool
	flyingSpeed  float32
	fovModifier  float32
}

func NewSetAbilitiesPacket(
	invulnerable, flying, canFly, instantBreak bool,
	flyingSpeed float32,
	fovModifier float32,
) *SetAbilitiesPacket {
	return &SetAbilitiesPacket{
		packet: newPacket(
			Outbound,
			PlayState,
			SetAbilitiesPacketID,
		),
		invulnerable: invulnerable,
		flying:       flying,
		canFly:       canFly,
		instantBreak: instantBreak,
		flyingSpeed:  flyingSpeed,
		fovModifier:  fovModifier,
	}
}

func (p *SetAbilitiesPacket) Pack() *Data {
	data := NewData()
	bitmask := uint8(0)
	if p.invulnerable == true {
		bitmask |= uint8(1)
	}
	if p.flying == true {
		bitmask |= uint8(2)
	}
	if p.canFly == true {
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

func (p *SetAbilitiesPacket) GetInvulnerable() bool {
	return p.invulnerable
}

func (p *SetAbilitiesPacket) GetFlying() bool {
	return p.flying
}

func (p *SetAbilitiesPacket) GetCanFly() bool {
	return p.canFly
}

func (p *SetAbilitiesPacket) GetInstantBreak() bool {
	return p.instantBreak
}

func (p *SetAbilitiesPacket) GetFlyingSpeed() float32 {
	return p.flyingSpeed
}

func (p *SetAbilitiesPacket) GetFovModifier() float32 {
	return p.fovModifier
}

func (p *SetAbilitiesPacket) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"invulnerable: %v, "+
			"flying: %v, "+
			"canFly: %v, "+
			"instantBreak: %v, "+
			"flyingSpeed: %f, "+
			"fovModifier: %f "+
			"}",
		p.packet,
		p.invulnerable,
		p.flying,
		p.canFly,
		p.instantBreak,
		p.flyingSpeed,
		p.fovModifier,
	)
}

type AddPlayerPacket struct {
	*packet
	uid         uuid.UUID
	username    string
	texture     string
	signature   string
	gamemode    int32
	latency     int32
	displayName string
}

func NewAddPlayerPacket(
	uid uuid.UUID,
	username string,
	texture string,
	signature string,
	gamemode int32,
	latency int32,
	displayName string,
) *AddPlayerPacket {
	return &AddPlayerPacket{
		packet: newPacket(
			Outbound,
			PlayState,
			AddPlayerPacketID,
		),
		uid:         uid,
		username:    username,
		texture:     texture,
		signature:   signature,
		gamemode:    gamemode,
		latency:     latency,
		displayName: displayName,
	}
}

func (p *AddPlayerPacket) Pack() *Data {
	data := NewData()
	data.WriteVarInt(0)
	data.WriteVarInt(1)

	data.WriteUUID(p.uid)
	data.WriteString(p.username)
	data.WriteVarInt(1)
	data.WriteString("texture")
	data.WriteString(p.texture)
	data.WriteBool(true)
	data.WriteString(p.signature)
	data.WriteVarInt(p.gamemode)
	data.WriteVarInt(p.latency)
	data.WriteBool(false)
	//data.WriteString(p.displayName)  // TODO

	return data
}

func (p *AddPlayerPacket) GetUid() uuid.UUID {
	return p.uid
}

func (p *AddPlayerPacket) GetUsername() string {
	return p.username
}

func (p *AddPlayerPacket) GetTexture() string {
	return p.texture
}

func (p *AddPlayerPacket) GetSignature() string {
	return p.signature
}

func (p *AddPlayerPacket) GetGamemode() int32 {
	return p.gamemode
}

func (p *AddPlayerPacket) GetLatency() int32 {
	return p.latency
}

func (p *AddPlayerPacket) GetDisplayName() string {
	return p.displayName
}

func (p *AddPlayerPacket) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"uid: %s, "+
			"username: %s, "+
			"texture: %s, "+
			"signature: %s, "+
			"gamemode: %d, "+
			"latency: %d, "+
			"displayName: %s "+
			"}",
		p.packet,
		p.uid,
		p.username,
		p.texture,
		p.signature,
		p.gamemode,
		p.latency,
		p.displayName,
	)
}

type RemovePlayerPacket struct {
	*packet
	uid uuid.UUID
}

func NewRemovePlayerPacket(
	uid uuid.UUID,
) *RemovePlayerPacket {
	return &RemovePlayerPacket{
		packet: newPacket(
			Outbound,
			PlayState,
			RemovePlayerPacketID,
		),
		uid: uid,
	}
}

func (p *RemovePlayerPacket) Pack() *Data {
	data := NewData()
	data.WriteVarInt(4)
	data.WriteVarInt(1)

	data.WriteUUID(p.uid)

	return data
}

func (p *RemovePlayerPacket) GetUUID() uuid.UUID {
	return p.uid
}

func (p *RemovePlayerPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, uid: %s }",
		p.packet, p.uid,
	)
}

type UpdateLatencyPacket struct {
	*packet
	uid     uuid.UUID
	latency int32
}

func NewUpdateLatencyPacket(
	uid uuid.UUID,
	latency int32,
) *UpdateLatencyPacket {
	return &UpdateLatencyPacket{
		packet: newPacket(
			Outbound,
			PlayState,
			UpdateLatencyPacketID,
		),
		uid:     uid,
		latency: latency,
	}
}

func (p *UpdateLatencyPacket) Pack() *Data {
	data := NewData()
	data.WriteVarInt(2)
	data.WriteVarInt(1)

	data.WriteUUID(p.uid)
	data.WriteVarInt(p.latency)

	return data
}

func (p *UpdateLatencyPacket) GetUUID() uuid.UUID {
	return p.uid
}

func (p *UpdateLatencyPacket) GetLatency() int32 {
	return p.latency
}

func (p *UpdateLatencyPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, uid: %s, latency: %d }",
		p.packet, p.uid, p.latency,
	)
}

type TeleportPacket struct {
	*packet
	x       float64
	y       float64
	z       float64
	yaw     float32
	pitch   float32
	payload int32
}

func NewTeleportPacket(
	x, y, z float64,
	yaw, pitch float32,
	payload int32,
) *TeleportPacket {
	return &TeleportPacket{
		packet: newPacket(
			Outbound,
			PlayState,
			TeleportPacketID,
		),
		x:       x,
		y:       y,
		z:       z,
		yaw:     yaw,
		pitch:   pitch,
		payload: payload,
	}
}

func (p *TeleportPacket) Pack() *Data {
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

func (p *TeleportPacket) GetX() float64 {
	return p.x
}

func (p *TeleportPacket) GetY() float64 {
	return p.y
}

func (p *TeleportPacket) GetZ() float64 {
	return p.z
}

func (p *TeleportPacket) GetYaw() float32 {
	return p.yaw
}

func (p *TeleportPacket) GetPitch() float32 {
	return p.pitch
}

func (p *TeleportPacket) GetPayload() int32 {
	return p.payload
}

func (p *TeleportPacket) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"x: %f, y: %f, z: %f, "+
			"yaw: %f, pitch: %f, "+
			"payload: %d "+
			"}",
		p.packet,
		p.x, p.y, p.z,
		p.yaw, p.pitch,
		p.payload,
	)
}

type DespawnEntityPacket struct {
	*packet
	eid int32
}

func NewDespawnEntityPacket(
	eid int32,
) *DespawnEntityPacket {
	return &DespawnEntityPacket{
		packet: newPacket(
			Outbound,
			PlayState,
			DespawnEntityPacketID,
		),
		eid: eid,
	}
}

func (p *DespawnEntityPacket) Pack() *Data {
	data := NewData()
	data.WriteVarInt(1)
	data.WriteVarInt(p.eid)

	return data
}

func (p *DespawnEntityPacket) GetEID() int32 {
	return p.eid
}

func (p *DespawnEntityPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, eid: %d }",
		p.packet, p.eid,
	)
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

func (p *SetSpawnPosPacket) Pack() *Data {
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

func (p *SetSpawnPosPacket) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, x: %d, y: %d, z: %d }",
		p.packet, p.x, p.y, p.z,
	)
}
