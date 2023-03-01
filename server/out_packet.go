package server

import (
	"fmt"
	"github.com/google/uuid"
)

const OutPacketIDToResponse = 0x00
const OutPacketIDToPong = 0x01

const OutPacketIDToRejectLogin = 0x00
const OutPacketIDToCompleteLogin = 0x02
const OutPacketIDToEnableComp = 0x03

const OutPacketIDToSpawnPlayer = 0x05
const OutPacketIDToSendChatMessage = 0x0F
const OutPacketIDToUnloadChunk = 0x1D
const OutPacketIDToCheckKeepAlive = 0x1F
const OutPacketIDToSendChunkData = 0x20
const OutPacketIDToJoinGame = 0x23
const OutPacketIDToSetEntityRltvPos = 0x26
const OutPacketIDToSetEntityLook = 0x28
const OutPacketIDToSetAbilities = 0x2C
const OutPacketIDToAddPlayer = 0x2E
const OutPacketIDToUpdateLatency = 0x2E
const OutPacketIDToRemovePlayer = 0x2E
const OutPacketIDToTeleport = 0x2F
const OutPacketIDToDespawnEntity = 0x32
const OutPacketIDToRespawn = 0x35
const OutPacketIDToSetEntityHeadLook = 0x36
const OutPacketIDToSetEntityMd = 0x3C
const OutPacketIDToSetSpawnPos = 0x46

type OutPacket interface {
	Packet

	Pack() (
		*Data,
		error,
	)
}

type OutPacketToResponse struct {
	*packet
	max     int    // maximum number of players
	online  int    // current number of players
	text    string // string for description
	favicon string // a png image string that is base64 encoded
}

func NewOutPacketToResponse(
	max, online int,
	text, favicon string,
) *OutPacketToResponse {
	return &OutPacketToResponse{
		newPacket(
			Outbound,
			StatusState,
			OutPacketIDToResponse,
		),
		max, online,
		text, favicon,
	}
}

func (p *OutPacketToResponse) Pack() (
	*Data,
	error,
) {
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
	if err := data.WriteString(jsonString); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToResponse) GetMax() int {
	return p.max
}

func (p *OutPacketToResponse) GetOnline() int {
	return p.online
}

func (p *OutPacketToResponse) GetText() string {
	return p.text
}

func (p *OutPacketToResponse) GetFavicon() string {
	return p.favicon
}

func (p *OutPacketToResponse) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, max: %d, online: %d, Text: %s, favicon: %s }",
		p.packet, p.max, p.online, p.text, p.favicon,
	)
}

type OutPacketToPong struct {
	*packet
	payload int64
}

func NewOutPacketToPong(
	payload int64,
) *OutPacketToPong {
	return &OutPacketToPong{
		newPacket(
			Outbound,
			StatusState,
			OutPacketIDToPong,
		),
		payload,
	}
}

func (p *OutPacketToPong) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteInt64(p.payload); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToPong) GetPayload() int64 {
	return p.payload
}

func (p *OutPacketToPong) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, payload: %d }",
		p.packet, p.payload,
	)
}

type OutPacketToCompleteLogin struct {
	*packet
	uid      UID
	username string
}

func NewOutPacketToCompleteLogin(
	uid UID,
	username string,
) *OutPacketToCompleteLogin {
	return &OutPacketToCompleteLogin{
		packet: newPacket(
			Outbound,
			LoginState,
			OutPacketIDToCompleteLogin,
		),
		uid:      uid,
		username: username,
	}
}

func (p *OutPacketToCompleteLogin) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteString(uuid.UUID(p.uid).String()); err != nil {
		return nil, err
	}
	if err := data.WriteString(p.username); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToCompleteLogin) GetUUID() UID {
	return p.uid
}

func (p *OutPacketToCompleteLogin) GetUsername() string {
	return p.username
}

func (p *OutPacketToCompleteLogin) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, uid: %s, username: %s }",
		p.packet, p.uid, p.username,
	)
}

type OutPacketToEnableComp struct {
	*packet
	threshold int32
}

func NewOutPacketToEnableComp(
	threshold int32,
) *OutPacketToEnableComp {
	return &OutPacketToEnableComp{
		packet: newPacket(
			Outbound,
			LoginState,
			OutPacketIDToEnableComp,
		),
		threshold: threshold,
	}
}

func (p *OutPacketToEnableComp) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteVarInt(p.threshold); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToEnableComp) GetThreshold() int32 {
	return p.threshold
}

func (p *OutPacketToEnableComp) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, threshold: %d }",
		p.packet, p.threshold,
	)
}

type OutPacketToSpawnPlayer struct {
	*packet
	eid        EID
	uid        UID
	x, y, z    float64
	yaw, pitch float32
	metadata   Metadata
}

func NewOutPacketToSpawnPlayer(
	eid EID,
	uid UID,
	x, y, z float64,
	yaw, pitch float32,
	metadata Metadata,
) *OutPacketToSpawnPlayer {
	return &OutPacketToSpawnPlayer{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSpawnPlayer,
		),
		eid,
		uid,
		x, y, z,
		yaw, pitch,
		metadata,
	}
}

func (p *OutPacketToSpawnPlayer) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteVarInt(int32(p.eid)); err != nil {
		return nil, err
	}
	if err := data.WriteUUID(uuid.UUID(p.uid)); err != nil {
		return nil, err
	}
	if err := data.WriteFloat64(p.x); err != nil {
		return nil, err
	}
	if err := data.WriteFloat64(p.y); err != nil {
		return nil, err
	}
	if err := data.WriteFloat64(p.z); err != nil {
		return nil, err
	}
	if err := data.WriteAngle(p.yaw); err != nil {
		return nil, err
	}
	if err := data.WriteAngle(p.pitch); err != nil {
		return nil, err
	}
	if err := data.WriteMetadata(p.metadata); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToSpawnPlayer) GetEID() EID {
	return p.eid
}

func (p *OutPacketToSpawnPlayer) GetUUID() UID {
	return p.uid
}

func (p *OutPacketToSpawnPlayer) GetX() float64 {
	return p.x
}

func (p *OutPacketToSpawnPlayer) GetY() float64 {
	return p.y
}

func (p *OutPacketToSpawnPlayer) GetZ() float64 {
	return p.z
}

func (p *OutPacketToSpawnPlayer) GetYaw() float32 {
	return p.yaw
}

func (p *OutPacketToSpawnPlayer) GetPitch() float32 {
	return p.pitch
}

func (p *OutPacketToSpawnPlayer) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"eid: %d, "+
			"uid: %s, "+
			"x: %f, y: %f, z: %f, "+
			"yaw: %f, pitch: %f "+
			"}",
		p.packet,
		p.eid,
		p.uid,
		p.x, p.y, p.z,
		p.yaw, p.pitch,
	)
}

type OutPacketToSendChatMessage struct {
	*packet
	msg *Chat
}

func NewOutPacketToSendChatMessage(
	msg *Chat,
) *OutPacketToSendChatMessage {
	return &OutPacketToSendChatMessage{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSendChatMessage,
		),
		msg,
	}
}

func (p *OutPacketToSendChatMessage) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteChat(p.msg); err != nil {
		return nil, err
	}
	if err := data.WriteUint8(0); err != nil { // 0: chat box
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToSendChatMessage) GetMessage() *Chat {
	return p.msg
}

func (p *OutPacketToSendChatMessage) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, msg: %s }",
		p.packet, p.msg,
	)
}

type OutPacketToUnloadChunk struct {
	*packet
	cx, cz int32
}

func NewOutPacketToUnloadChunk(
	cx, cz int32,
) *OutPacketToUnloadChunk {
	return &OutPacketToUnloadChunk{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToUnloadChunk,
		),
		cx, cz,
	}
}

func (p *OutPacketToUnloadChunk) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteInt32(p.cx); err != nil {
		return nil, err
	}
	if err := data.WriteInt32(p.cz); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToUnloadChunk) GetCx() int32 {
	return p.cx
}

func (p *OutPacketToUnloadChunk) GetCz() int32 {
	return p.cz
}

func (p *OutPacketToUnloadChunk) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, cx: %d, cz: %d }",
		p.packet, p.cx, p.cz,
	)
}

type OutPacketToCheckKeepAlive struct {
	*packet
	payload int64
}

func NewOutPacketToCheckKeepAlive(
	payload int64,
) *OutPacketToCheckKeepAlive {
	return &OutPacketToCheckKeepAlive{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToCheckKeepAlive,
		),
		payload,
	}
}

func (p *OutPacketToCheckKeepAlive) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteInt64(p.payload); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToCheckKeepAlive) GetPayload() int64 {
	return p.payload
}

func (p *OutPacketToCheckKeepAlive) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, payload: %d }",
		p.packet, p.payload,
	)
}

type OutPacketToSendChunkData struct {
	*packet
	cx, cz  int32
	init    bool
	bitmask uint16
	data    []uint8
}

func NewOutPacketToSendChunkData(
	cx, cz int32,
	init bool,
	bitmask uint16,
	data []uint8,
) *OutPacketToSendChunkData {
	return &OutPacketToSendChunkData{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSendChunkData,
		),
		cx, cz,
		init,
		bitmask,
		data,
	}
}

func (p *OutPacketToSendChunkData) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteInt32(p.cx); err != nil {
		return nil, err
	}
	if err := data.WriteInt32(p.cz); err != nil {
		return nil, err
	}
	if err := data.WriteBool(p.init); err != nil {
		return nil, err
	}
	if err := data.WriteVarInt(int32(p.bitmask)); err != nil {
		return nil, err
	}
	l0 := len(p.data)
	if err := data.WriteVarInt(int32(l0)); err != nil {
		return nil, err
	}
	if err := data.WriteBytes(p.data); err != nil {
		return nil, err
	}
	l1 := 0
	if err := data.WriteVarInt(int32(l1)); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToSendChunkData) GetCx() int32 {
	return p.cx
}

func (p *OutPacketToSendChunkData) GetCz() int32 {
	return p.cz
}

func (p *OutPacketToSendChunkData) GetInit() bool {
	return p.init
}

func (p *OutPacketToSendChunkData) GetBitmask() uint16 {
	return p.bitmask
}

func (p *OutPacketToSendChunkData) GetData() []uint8 {
	return p.data
}

func (p *OutPacketToSendChunkData) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, cx: %d, cz: %d, init: %v, bitmask: %d, data: [...] }",
		p.packet, p.cx, p.cz, p.init, p.bitmask,
	)
}

type OutPacketToJoinGame struct {
	*packet
	eid        EID
	gamemode   uint8
	dimension  int32
	difficulty uint8
	level      string
	debug      bool
}

func NewOutPacketToJoinGame(
	eid EID,
	gamemode uint8,
	dimension int32,
	difficulty uint8,
	level string,
	debug bool,
) *OutPacketToJoinGame {
	return &OutPacketToJoinGame{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToJoinGame,
		),
		eid,
		gamemode,
		dimension,
		difficulty,
		level,
		debug,
	}
}

func (p *OutPacketToJoinGame) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteInt32(int32(p.eid)); err != nil {
		return nil, err
	}
	if err := data.WriteUint8(p.gamemode); err != nil {
		return nil, err
	}
	if err := data.WriteInt32(p.dimension); err != nil {
		return nil, err
	}
	if err := data.WriteUint8(p.difficulty); err != nil {
		return nil, err
	}
	if err := data.WriteUint8(0); err != nil { // max is ignored;
		return nil, err
	}
	if err := data.WriteString(p.level); err != nil {
		return nil, err
	}
	if err := data.WriteBool(p.debug); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToJoinGame) GetEid() EID {
	return p.eid
}

func (p *OutPacketToJoinGame) GetGamemode() uint8 {
	return p.gamemode
}

func (p *OutPacketToJoinGame) GetDimension() int32 {
	return p.dimension
}

func (p *OutPacketToJoinGame) GetDifficulty() uint8 {
	return p.difficulty
}

func (p *OutPacketToJoinGame) GetLevel() string {
	return p.level
}

func (p *OutPacketToJoinGame) GetDebug() bool {
	return p.debug
}

func (p *OutPacketToJoinGame) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"eid: %d, "+
			"gamemode: %d, "+
			"dimension: %d, "+
			"difficulty: %d, "+
			"level: %s, "+
			"debug: %v "+
			"}",
		p.packet, p.eid, p.gamemode, p.dimension, p.difficulty, p.level, p.debug,
	)
}

type OutPacketToSetEntityRltvPos struct {
	*packet
	eid                    EID
	deltaX, deltaY, deltaZ int16
	ground                 bool
}

func NewOutPacketToSetEntityRltvPos(
	eid EID,
	deltaX, deltaY, deltaZ int16,
	ground bool,
) *OutPacketToSetEntityRltvPos {
	return &OutPacketToSetEntityRltvPos{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetEntityRltvPos,
		),
		eid,
		deltaX, deltaY, deltaZ,
		ground,
	}
}

func (p *OutPacketToSetEntityRltvPos) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteVarInt(int32(p.eid)); err != nil {
		return nil, err
	}
	if err := data.WriteInt16(p.deltaX); err != nil {
		return nil, err
	}
	if err := data.WriteInt16(p.deltaY); err != nil {
		return nil, err
	}
	if err := data.WriteInt16(p.deltaZ); err != nil {
		return nil, err
	}
	if err := data.WriteBool(p.ground); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToSetEntityRltvPos) GetEID() EID {
	return p.eid
}

func (p *OutPacketToSetEntityRltvPos) GetDeltaX() int16 {
	return p.deltaX
}

func (p *OutPacketToSetEntityRltvPos) GetDeltaY() int16 {
	return p.deltaY
}

func (p *OutPacketToSetEntityRltvPos) GetDeltaZ() int16 {
	return p.deltaZ
}

func (p *OutPacketToSetEntityRltvPos) GetGround() bool {
	return p.ground
}

func (p *OutPacketToSetEntityRltvPos) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"eid: %d, "+
			"deltaX: %d, deltaY: %d, deltaZ: %d, "+
			"ground: %v "+
			"}",
		p.packet,
		p.eid,
		p.deltaX, p.deltaY, p.deltaZ,
		p.ground,
	)
}

type OutPacketToSetEntityLook struct {
	*packet
	eid        EID
	yaw, pitch float32
	ground     bool
}

func NewOutPacketToSetEntityLook(
	eid EID,
	yaw, pitch float32,
	ground bool,
) *OutPacketToSetEntityLook {
	return &OutPacketToSetEntityLook{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetEntityLook,
		),
		eid,
		yaw, pitch,
		ground,
	}
}

func (p *OutPacketToSetEntityLook) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteVarInt(int32(p.eid)); err != nil {
		return nil, err
	}
	if err := data.WriteAngle(p.yaw); err != nil {
		return nil, err
	}
	if err := data.WriteAngle(p.pitch); err != nil {
		return nil, err
	}
	if err := data.WriteBool(p.ground); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToSetEntityLook) GetEID() EID {
	return p.eid
}

func (p *OutPacketToSetEntityLook) GetYaw() float32 {
	return p.yaw
}

func (p *OutPacketToSetEntityLook) GetPitch() float32 {
	return p.pitch
}

func (p *OutPacketToSetEntityLook) GetGround() bool {
	return p.ground
}

func (p *OutPacketToSetEntityLook) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"eid: %d, "+
			"yaw: %f, pitch: %f, "+
			"ground: %v "+
			"}",
		p.packet,
		p.eid,
		p.yaw, p.pitch,
		p.ground,
	)
}

type OutPacketToSetAbilities struct {
	*packet
	invulnerable, flying, canFly, instantBreak bool
	flyingSpeed                                float32
	fovModifier                                float32
}

func NewOutPacketToSetAbilities(
	invulnerable, flying, canFly, instantBreak bool,
	flyingSpeed float32,
	fovModifier float32,
) *OutPacketToSetAbilities {
	return &OutPacketToSetAbilities{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetAbilities,
		),
		invulnerable, flying, canFly, instantBreak,
		flyingSpeed,
		fovModifier,
	}
}

func (p *OutPacketToSetAbilities) Pack() (
	*Data,
	error,
) {
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
	if err := data.WriteUint8(bitmask); err != nil {
		return nil, err
	}
	if err := data.WriteFloat32(p.flyingSpeed); err != nil {
		return nil, err
	}
	if err := data.WriteFloat32(p.fovModifier); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToSetAbilities) GetInvulnerable() bool {
	return p.invulnerable
}

func (p *OutPacketToSetAbilities) GetFlying() bool {
	return p.flying
}

func (p *OutPacketToSetAbilities) GetCanFly() bool {
	return p.canFly
}

func (p *OutPacketToSetAbilities) GetInstantBreak() bool {
	return p.instantBreak
}

func (p *OutPacketToSetAbilities) GetFlyingSpeed() float32 {
	return p.flyingSpeed
}

func (p *OutPacketToSetAbilities) GetFovModifier() float32 {
	return p.fovModifier
}

func (p *OutPacketToSetAbilities) String() string {
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

type OutPacketToAddPlayer struct {
	*packet
	uid                UID
	username           string
	texture, signature string
	gamemode           int32
	latency            int32
	displayName        *Chat
}

func NewOutPacketToAddPlayer(
	uid UID, username string,
	texture, signature string,
	gamemode int32,
	latency int32,
	displayName *Chat,
) *OutPacketToAddPlayer {
	return &OutPacketToAddPlayer{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToAddPlayer,
		),
		uid, username,
		texture, signature,
		gamemode,
		latency,
		displayName,
	}
}

func (p *OutPacketToAddPlayer) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteVarInt(0); err != nil {
		return nil, err
	}
	if err := data.WriteVarInt(1); err != nil {
		return nil, err
	}

	if err := data.WriteUUID(uuid.UUID(p.uid)); err != nil {
		return nil, err
	}
	if err := data.WriteString(p.username); err != nil {
		return nil, err
	}
	if err := data.WriteVarInt(1); err != nil {
		return nil, err
	}
	if err := data.WriteString("texture"); err != nil {
		return nil, err
	}
	if err := data.WriteString(p.texture); err != nil {
		return nil, err
	}
	if err := data.WriteBool(true); err != nil {
		return nil, err
	}
	if err := data.WriteString(p.signature); err != nil {
		return nil, err
	}
	if err := data.WriteVarInt(p.gamemode); err != nil {
		return nil, err
	}
	if err := data.WriteVarInt(p.latency); err != nil {
		return nil, err
	}
	if err := data.WriteBool(true); err != nil {
		return nil, err
	}
	if err := data.WriteChat(p.displayName); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToAddPlayer) GetUid() UID {
	return p.uid
}

func (p *OutPacketToAddPlayer) GetUsername() string {
	return p.username
}

func (p *OutPacketToAddPlayer) GetTexture() string {
	return p.texture
}

func (p *OutPacketToAddPlayer) GetSignature() string {
	return p.signature
}

func (p *OutPacketToAddPlayer) GetGamemode() int32 {
	return p.gamemode
}

func (p *OutPacketToAddPlayer) GetLatency() int32 {
	return p.latency
}

func (p *OutPacketToAddPlayer) GetDisplayName() *Chat {
	return p.displayName
}

func (p *OutPacketToAddPlayer) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"uid: %s, username: %s, "+
			"texture: %s, "+
			"signature: %s, "+
			"gamemode: %d, "+
			"latency: %d, "+
			"displayName: %+v "+
			"}",
		p.packet,
		p.uid, p.username,
		p.texture,
		p.signature,
		p.gamemode,
		p.latency,
		p.displayName,
	)
}

type OutPacketToUpdateLatency struct {
	*packet
	uid     UID
	latency int32
}

func NewOutPacketToUpdateLatency(
	uid UID,
	latency int32,
) *OutPacketToUpdateLatency {
	return &OutPacketToUpdateLatency{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToUpdateLatency,
		),
		uid,
		latency,
	}
}

func (p *OutPacketToUpdateLatency) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteVarInt(2); err != nil {
		return nil, err
	}
	if err := data.WriteVarInt(1); err != nil {
		return nil, err
	}
	if err := data.WriteUUID(uuid.UUID(p.uid)); err != nil {
		return nil, err
	}
	if err := data.WriteVarInt(p.latency); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToUpdateLatency) GetUUID() UID {
	return p.uid
}

func (p *OutPacketToUpdateLatency) GetLatency() int32 {
	return p.latency
}

func (p *OutPacketToUpdateLatency) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"uid: %s, "+
			"latency: %d "+
			"}",
		p.packet,
		p.uid,
		p.latency,
	)
}

type OutPacketToRemovePlayer struct {
	*packet
	uid UID
}

func NewOutPacketToRemovePlayer(
	uid UID,
) *OutPacketToRemovePlayer {
	return &OutPacketToRemovePlayer{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToRemovePlayer,
		),
		uid,
	}
}

func (p *OutPacketToRemovePlayer) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteVarInt(4); err != nil {
		return nil, err
	}
	if err := data.WriteVarInt(1); err != nil {
		return nil, err
	}
	if err := data.WriteUUID(uuid.UUID(p.uid)); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToRemovePlayer) GetUUID() UID {
	return p.uid
}

func (p *OutPacketToRemovePlayer) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, uid: %s }",
		p.packet, p.uid,
	)
}

type OutPacketToTeleport struct {
	*packet
	x, y, z    float64
	yaw, pitch float32
	payload    int32
}

func NewOutPacketToTeleport(
	x, y, z float64,
	yaw, pitch float32,
	payload int32,
) *OutPacketToTeleport {
	return &OutPacketToTeleport{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToTeleport,
		),
		x, y, z,
		yaw, pitch,
		payload,
	}
}

func (p *OutPacketToTeleport) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteFloat64(p.x); err != nil {
		return nil, err
	}
	if err := data.WriteFloat64(p.y); err != nil {
		return nil, err
	}
	if err := data.WriteFloat64(p.z); err != nil {
		return nil, err
	}
	if err := data.WriteFloat32(p.yaw); err != nil {
		return nil, err
	}
	if err := data.WriteFloat32(p.pitch); err != nil {
		return nil, err
	}
	if err := data.WriteInt8(0); err != nil {
		return nil, err
	}
	if err := data.WriteVarInt(p.payload); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToTeleport) GetX() float64 {
	return p.x
}

func (p *OutPacketToTeleport) GetY() float64 {
	return p.y
}

func (p *OutPacketToTeleport) GetZ() float64 {
	return p.z
}

func (p *OutPacketToTeleport) GetYaw() float32 {
	return p.yaw
}

func (p *OutPacketToTeleport) GetPitch() float32 {
	return p.pitch
}

func (p *OutPacketToTeleport) GetPayload() int32 {
	return p.payload
}

func (p *OutPacketToTeleport) String() string {
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

type OutPacketToDespawnEntity struct {
	*packet
	eid EID
}

func NewOutPacketToDespawnEntity(
	eid EID,
) *OutPacketToDespawnEntity {
	return &OutPacketToDespawnEntity{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToDespawnEntity,
		),
		eid,
	}
}

func (p *OutPacketToDespawnEntity) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteVarInt(1); err != nil {
		return nil, err
	}
	if err := data.WriteVarInt(int32(p.eid)); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToDespawnEntity) GetEID() EID {
	return p.eid
}

func (p *OutPacketToDespawnEntity) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"eid: %d "+
			"}",
		p.packet,
		p.eid,
	)
}

type OutPacketToRespawn struct {
	*packet
	dimension  int32
	difficulty uint8
	gamemode   uint8
	level      string
}

func NewOutPacketToRespawn(
	dimension int32,
	difficulty uint8,
	gamemode uint8,
	level string,
) *OutPacketToRespawn {
	return &OutPacketToRespawn{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToRespawn,
		),
		dimension,
		difficulty,
		gamemode,
		level,
	}
}

func (p *OutPacketToRespawn) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteInt32(p.dimension); err != nil {
		return nil, err
	}
	if err := data.WriteUint8(p.difficulty); err != nil {
		return nil, err
	}
	if err := data.WriteUint8(p.gamemode); err != nil {
		return nil, err
	}
	if err := data.WriteString(p.level); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToRespawn) GetDimension() int32 {
	return p.dimension
}

func (p *OutPacketToRespawn) GetDifficulty() uint8 {
	return p.difficulty
}

func (p *OutPacketToRespawn) GetGamemode() uint8 {
	return p.gamemode
}

func (p *OutPacketToRespawn) GetLevel() string {
	return p.level
}

func (p *OutPacketToRespawn) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"dimension: %d, "+
			"difficulty: %d, "+
			"gamemode: %d, "+
			"level: %s "+
			"}",
		p.packet,
		p.dimension,
		p.difficulty,
		p.gamemode,
		p.level,
	)
}

type OutPacketToSetEntityHeadLook struct {
	*packet
	eid EID
	yaw float32
}

func NewOutPacketToSetEntityHeadLook(
	eid EID,
	yaw float32,
) *OutPacketToSetEntityHeadLook {
	return &OutPacketToSetEntityHeadLook{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetEntityHeadLook,
		),
		eid, yaw,
	}
}

func (p *OutPacketToSetEntityHeadLook) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteVarInt(int32(p.eid)); err != nil {
		return nil, err
	}
	if err := data.WriteAngle(p.yaw); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToSetEntityHeadLook) GetEID() EID {
	return p.eid
}

func (p *OutPacketToSetEntityHeadLook) GetYaw() float32 {
	return p.yaw
}

func (p *OutPacketToSetEntityHeadLook) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"eid: %d, "+
			"yaw: %f "+
			"}",
		p.packet,
		p.eid,
		p.yaw,
	)
}

type OutPacketToSetEntityMd struct {
	*packet
	eid EID
	md  *EntityMetadata
}

func NewOutPacketToSetEntityMd(
	eid EID,
	md *EntityMetadata,
) *OutPacketToSetEntityMd {
	return &OutPacketToSetEntityMd{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetEntityMd,
		),
		eid,
		md,
	}
}

func (p *OutPacketToSetEntityMd) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WriteVarInt(int32(p.eid)); err != nil {
		return nil, err
	}
	if err := data.WriteMetadata(p.md); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToSetEntityMd) GetEID() EID {
	return p.eid
}

func (p *OutPacketToSetEntityMd) GetMetadata() *EntityMetadata {
	return p.md
}

func (p *OutPacketToSetEntityMd) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, eid: %d, md: {...} }",
		p.packet, p.eid,
	)
}

type OutPacketToSetSpawnPos struct {
	*packet
	x, y, z int
}

func NewOutPacketToSetSpawnPos(
	x, y, z int,
) *OutPacketToSetSpawnPos {
	return &OutPacketToSetSpawnPos{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetSpawnPos,
		),
		x, y, z,
	}
}

func (p *OutPacketToSetSpawnPos) Pack() (
	*Data,
	error,
) {
	data := NewData()
	if err := data.WritePosition(p.x, p.y, p.z); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *OutPacketToSetSpawnPos) GetX() int {
	return p.x
}

func (p *OutPacketToSetSpawnPos) GetY() int {
	return p.y
}

func (p *OutPacketToSetSpawnPos) GetZ() int {
	return p.z
}

func (p *OutPacketToSetSpawnPos) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"x: %d, y: %d, z: %d "+
			"}",
		p.packet,
		p.x, p.y, p.z,
	)
}
