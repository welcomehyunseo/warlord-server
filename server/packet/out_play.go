package packet

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/welcomehyunseo/warlord-server/server/data"
	"github.com/welcomehyunseo/warlord-server/server/item"
	"github.com/welcomehyunseo/warlord-server/server/metadata"
)

const OutPacketIDToSPawnObject = 0x00

// const OutPacketIDToSpawnMobileEntity = 0x03
const OutPacketIDToSpawnArmorStand = 0x03
const OutPacketIDToSpawnPlayer = 0x05
const OutPacketIDToSendChatMessage = 0x0F
const OutPacketIDToAcceptTransactionOfWindow = 0x11
const OutPacketIDToRejectTransactionOfWindow = 0x11
const OutPacketIDToCloseWindow = 0x12
const OutPacketIDToSetSlotsInWindow = 0x14
const OutPacketIDToSetSlotInWindow = 0x16
const OutPacketIDToUnloadChunk = 0x1D
const OutPacketIDToCheckKeepAlive = 0x1F
const OutPacketIDToSendChunkData = 0x20
const OutPacketIDToJoinGame = 0x23
const OutPacketIDToSetEntityRelativeMove = 0x26
const OutPacketIDToSetEntityLook = 0x28
const OutPacketIDToSetAbilities = 0x2C
const OutPacketIDToAddPlayer = 0x2E
const OutPacketIDToUpdateLatency = 0x2E
const OutPacketIDToRemovePlayer = 0x2E
const OutPacketIDToTeleport = 0x2F
const OutPacketIDToDespawnEntity = 0x32
const OutPacketIDToRespawn = 0x35
const OutPacketIDToSetEntityHeadLook = 0x36
const OutPacketIDToSetItemEntityMetadata = 0x3C
const OutPacketIDToSetPlayerMetadata = 0x3C
const OutPacketIDToSetArmorStandMetadata = 0x3C
const OutPacketIDToSetEntityVelocity = 0x3E
const OutPacketIDToSetEntityEquipment = 0x3F
const OutPacketIDToSetSpawnPosition = 0x46

type OutPacketToSpawnObject struct {
	*packet
	eid        int32
	uid        uuid.UUID
	n          int8
	x, y, z    float64
	pitch, yaw float32
	data       int32
	vx, vy, vz int16
}

func NewOutPacketToSpawnObject(
	eid int32,
	uid uuid.UUID,
	n int8,
	x, y, z float64,
	pitch, yaw float32,
	data int32,
	vx, vy, vz int16,
) *OutPacketToSpawnObject {
	return &OutPacketToSpawnObject{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSPawnObject,
		),
		eid,
		uid,
		n,
		x, y, z,
		pitch, yaw,
		data,
		vx, vy, vz,
	}
}

func (p *OutPacketToSpawnObject) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteVarInt(
		p.eid,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteUUID(
		p.uid,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteInt8(
		p.n,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteFloat64(
		p.x,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteFloat64(
		p.y,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteFloat64(
		p.z,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteAngle(
		p.pitch,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteAngle(
		p.yaw,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteInt32(
		p.data,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteInt16(
		p.vx,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteInt16(
		p.vy,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteInt16(
		p.vz,
	); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

type OutPacketToSpawnArmorStand struct {
	*packet
	eid        int32
	uid        uuid.UUID
	x, y, z    float64
	yaw, pitch float32
	md         *metadata.ArmorStandMetadata
}

func NewOutPacketToSpawnArmorStand(
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
	md *metadata.ArmorStandMetadata,
) *OutPacketToSpawnArmorStand {
	return &OutPacketToSpawnArmorStand{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSpawnArmorStand,
		),
		eid,
		uid,
		x, y, z,
		yaw, pitch,
		md,
	}
}

func (p *OutPacketToSpawnArmorStand) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteVarInt(
		p.eid,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteUUID(
		p.uid,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteVarInt(
		30,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteFloat64(
		p.x,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteFloat64(
		p.y,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteFloat64(
		p.z,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteAngle(
		p.yaw,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteAngle(
		p.pitch,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteAngle(
		0,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteInt16(
		0,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteInt16(
		0,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteInt16(
		0,
	); err != nil {
		return nil, err
	}

	if err := p.md.Write(
		dt,
	); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

type OutPacketToSpawnPlayer struct {
	*packet
	eid        int32
	uid        uuid.UUID
	x, y, z    float64
	yaw, pitch float32
	md         *metadata.PlayerMetadata
}

func NewOutPacketToSpawnPlayer(
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
	md *metadata.PlayerMetadata,
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
		md,
	}
}

func (p *OutPacketToSpawnPlayer) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteVarInt(
		p.eid,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteUUID(
		p.uid,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteFloat64(
		p.x,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteFloat64(
		p.y,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteFloat64(
		p.z,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteAngle(
		p.yaw,
	); err != nil {
		return nil, err
	}

	if err := dt.WriteAngle(
		p.pitch,
	); err != nil {
		return nil, err
	}

	if err := p.md.Write(
		dt,
	); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToSpawnPlayer) GetEID() int32 {
	return p.eid
}

func (p *OutPacketToSpawnPlayer) GetUID() uuid.UUID {
	return p.uid
}

func (p *OutPacketToSpawnPlayer) GetPosition() (
	float64, float64, float64,
) {
	return p.x, p.y, p.z
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

func (p *OutPacketToSpawnPlayer) GetLook() (
	float32, float32,
) {
	return p.yaw, p.pitch
}

func (p *OutPacketToSpawnPlayer) GetYaw() float32 {
	return p.yaw
}

func (p *OutPacketToSpawnPlayer) GetPitch() float32 {
	return p.pitch
}

func (p *OutPacketToSpawnPlayer) GetMetadata() *metadata.PlayerMetadata {
	return p.md
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
	msg *data.Chat
}

func NewOutPacketToSendChatMessage(
	msg *data.Chat,
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
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteChat(p.msg); err != nil {
		return nil, err
	}
	if err := dt.WriteUint8(0); err != nil { // 0: chat box
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToSendChatMessage) GetMessage() *data.Chat {
	return p.msg
}

func (p *OutPacketToSendChatMessage) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, msg: %s }",
		p.packet, p.msg,
	)
}

type OutPacketToAcceptTransactionOfWindow struct {
	*packet

	windowID int8
	actNum   int16
}

func NewOutPacketToAcceptTransactionOfWindow(
	windowID int8,
	actNum int16,
) *OutPacketToAcceptTransactionOfWindow {
	return &OutPacketToAcceptTransactionOfWindow{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToAcceptTransactionOfWindow,
		),
		windowID,
		actNum,
	}
}

func (p *OutPacketToAcceptTransactionOfWindow) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteInt8(
		p.windowID,
	); err != nil {
		return nil, err
	}
	if err := dt.WriteInt16(
		p.actNum,
	); err != nil {
		return nil, err
	}
	if err := dt.WriteBool(
		true,
	); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToAcceptTransactionOfWindow) GetWindowID() int8 {
	return p.windowID
}

func (p *OutPacketToAcceptTransactionOfWindow) GetActionNumber() int16 {
	return p.actNum
}

type OutPacketToRejectTransactionOfWindow struct {
	*packet

	winID  int8
	actNum int16
}

func NewOutPacketToRejectTransactionOfWindow(
	winID int8,
	actNum int16,
) *OutPacketToRejectTransactionOfWindow {
	return &OutPacketToRejectTransactionOfWindow{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToRejectTransactionOfWindow,
		),
		winID,
		actNum,
	}
}

func (p *OutPacketToRejectTransactionOfWindow) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteInt8(
		p.winID,
	); err != nil {
		return nil, err
	}
	if err := dt.WriteInt16(
		p.actNum,
	); err != nil {
		return nil, err
	}
	if err := dt.WriteBool(
		false,
	); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToRejectTransactionOfWindow) GetWindowID() int8 {
	return p.winID
}

func (p *OutPacketToRejectTransactionOfWindow) GetActionNumber() int16 {
	return p.actNum
}

type OutPacketToCloseWindow struct {
	*packet
	winID int8
}

func NewOutPacketToCloseWindow(
	winID int8,
) *OutPacketToCloseWindow {
	return &OutPacketToCloseWindow{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToCloseWindow,
		),
		winID,
	}
}

func (p *OutPacketToCloseWindow) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteInt8(
		p.winID,
	); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToCloseWindow) GetWindowID() int8 {
	return p.winID
}

type OutPacketToSetSlotInWindow struct {
	*packet

	winID int8
	slot  int16
	it    item.Item
}

func NewOutPacketToSetSlotInWindow(
	winID int8,
	slot int16,
	it item.Item,
) *OutPacketToSetSlotInWindow {
	return &OutPacketToSetSlotInWindow{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetSlotInWindow,
		),
		winID,
		slot,
		it,
	}
}

func (p *OutPacketToSetSlotInWindow) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()

	if err := dt.WriteInt8(
		p.winID,
	); err != nil {
		return nil, err
	}
	if err := dt.WriteInt16(
		p.slot,
	); err != nil {
		return nil, err
	}
	if err := p.it.Write(
		dt,
	); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToSetSlotInWindow) GetWindowID() int8 {
	return p.winID
}

func (p *OutPacketToSetSlotInWindow) GetSlotNumber() int16 {
	return p.slot
}

func (p *OutPacketToSetSlotInWindow) GetItem() item.Item {
	return p.it
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
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteInt32(
		p.cx,
	); err != nil {
		return nil, err
	}
	if err := dt.WriteInt32(
		p.cz,
	); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToUnloadChunk) GetChunkPosition() (
	int32, int32,
) {
	return p.cx, p.cz
}

func (p *OutPacketToUnloadChunk) GetChunkX() int32 {
	return p.cx
}

func (p *OutPacketToUnloadChunk) GetChunkZ() int32 {
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
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteInt64(p.payload); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
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
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteInt32(p.cx); err != nil {
		return nil, err
	}
	if err := dt.WriteInt32(p.cz); err != nil {
		return nil, err
	}
	if err := dt.WriteBool(p.init); err != nil {
		return nil, err
	}
	if err := dt.WriteVarInt(int32(p.bitmask)); err != nil {
		return nil, err
	}
	l0 := len(p.data)
	if err := dt.WriteVarInt(int32(l0)); err != nil {
		return nil, err
	}
	if err := dt.WriteBytes(p.data); err != nil {
		return nil, err
	}
	l1 := 0
	if err := dt.WriteVarInt(int32(l1)); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToSendChunkData) GetChunkPosition() (
	int32, int32,
) {
	return p.cx, p.cz
}

func (p *OutPacketToSendChunkData) GetChunkX() int32 {
	return p.cx
}

func (p *OutPacketToSendChunkData) GetChunkZ() int32 {
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
	eid        int32
	gamemode   uint8
	dimension  int32
	difficulty uint8
	level      string
	debug      bool
}

func NewOutPacketToJoinGame(
	eid int32,
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
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteInt32(p.eid); err != nil {
		return nil, err
	}
	if err := dt.WriteUint8(p.gamemode); err != nil {
		return nil, err
	}
	if err := dt.WriteInt32(p.dimension); err != nil {
		return nil, err
	}
	if err := dt.WriteUint8(p.difficulty); err != nil {
		return nil, err
	}
	if err := dt.WriteUint8(0); err != nil { // max is ignored;
		return nil, err
	}
	if err := dt.WriteString(p.level); err != nil {
		return nil, err
	}
	if err := dt.WriteBool(p.debug); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToJoinGame) GetEID() int32 {
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

type OutPacketToSetEntityRelativeMove struct {
	*packet
	eid        int32
	dx, dy, dz int16
	ground     bool
}

func NewOutPacketToSetEntityRelativeMove(
	eid int32,
	dx, dy, dz int16,
	ground bool,
) *OutPacketToSetEntityRelativeMove {
	return &OutPacketToSetEntityRelativeMove{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetEntityRelativeMove,
		),
		eid,
		dx, dy, dz,
		ground,
	}
}

func (p *OutPacketToSetEntityRelativeMove) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteVarInt(p.eid); err != nil {
		return nil, err
	}
	if err := dt.WriteInt16(p.dx); err != nil {
		return nil, err
	}
	if err := dt.WriteInt16(p.dy); err != nil {
		return nil, err
	}
	if err := dt.WriteInt16(p.dz); err != nil {
		return nil, err
	}
	if err := dt.WriteBool(p.ground); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToSetEntityRelativeMove) GetEID() int32 {
	return p.eid
}

func (p *OutPacketToSetEntityRelativeMove) GetDifferences() (
	int16, int16, int16,
) {
	return p.dx, p.dy, p.dz
}

func (p *OutPacketToSetEntityRelativeMove) GetDeltaX() int16 {
	return p.dx
}

func (p *OutPacketToSetEntityRelativeMove) GetDeltaY() int16 {
	return p.dy
}

func (p *OutPacketToSetEntityRelativeMove) GetDeltaZ() int16 {
	return p.dz
}

func (p *OutPacketToSetEntityRelativeMove) IsGround() bool {
	return p.ground
}

func (p *OutPacketToSetEntityRelativeMove) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"eid: %d, "+
			"dx: %d, dy: %d, dz: %d, "+
			"ground: %v "+
			"}",
		p.packet,
		p.eid,
		p.dx, p.dy, p.dz,
		p.ground,
	)
}

type OutPacketToSetEntityLook struct {
	*packet
	eid        int32
	yaw, pitch float32
	ground     bool
}

func NewOutPacketToSetEntityLook(
	eid int32,
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
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteVarInt(p.eid); err != nil {
		return nil, err
	}
	if err := dt.WriteAngle(p.yaw); err != nil {
		return nil, err
	}
	if err := dt.WriteAngle(p.pitch); err != nil {
		return nil, err
	}
	if err := dt.WriteBool(p.ground); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToSetEntityLook) GetEID() int32 {
	return p.eid
}

func (p *OutPacketToSetEntityLook) GetYaw() float32 {
	return p.yaw
}

func (p *OutPacketToSetEntityLook) GetPitch() float32 {
	return p.pitch
}

func (p *OutPacketToSetEntityLook) IsGround() bool {
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
	[]byte,
	error,
) {
	dt := data.NewData()
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
	if err := dt.WriteUint8(bitmask); err != nil {
		return nil, err
	}
	if err := dt.WriteFloat32(p.flyingSpeed); err != nil {
		return nil, err
	}
	if err := dt.WriteFloat32(p.fovModifier); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToSetAbilities) IsInvulnerable() bool {
	return p.invulnerable
}

func (p *OutPacketToSetAbilities) IsFlying() bool {
	return p.flying
}

func (p *OutPacketToSetAbilities) IsCanFly() bool {
	return p.canFly
}

func (p *OutPacketToSetAbilities) IsInstantBreak() bool {
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
	uid                uuid.UUID
	username           string
	texture, signature string
	gamemode           int32
	latency            int32
	displayName        *data.Chat
}

func NewOutPacketToAddPlayer(
	uid uuid.UUID, username string,
	texture, signature string,
	gamemode int32,
	latency int32,
	displayName *data.Chat,
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
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteVarInt(0); err != nil {
		return nil, err
	}
	if err := dt.WriteVarInt(1); err != nil {
		return nil, err
	}

	if err := dt.WriteUUID(p.uid); err != nil {
		return nil, err
	}
	if err := dt.WriteString(p.username); err != nil {
		return nil, err
	}
	if err := dt.WriteVarInt(1); err != nil {
		return nil, err
	}
	if err := dt.WriteString("texture"); err != nil {
		return nil, err
	}
	if err := dt.WriteString(p.texture); err != nil {
		return nil, err
	}
	if err := dt.WriteBool(true); err != nil {
		return nil, err
	}
	if err := dt.WriteString(p.signature); err != nil {
		return nil, err
	}
	if err := dt.WriteVarInt(p.gamemode); err != nil {
		return nil, err
	}
	if err := dt.WriteVarInt(p.latency); err != nil {
		return nil, err
	}
	if err := dt.WriteBool(true); err != nil {
		return nil, err
	}
	if err := dt.WriteChat(p.displayName); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToAddPlayer) GetUID() uuid.UUID {
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

func (p *OutPacketToAddPlayer) GetDisplayName() *data.Chat {
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
	uid uuid.UUID
	ms  int32
}

func NewOutPacketToUpdateLatency(
	uid uuid.UUID,
	ms int32,
) *OutPacketToUpdateLatency {
	return &OutPacketToUpdateLatency{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToUpdateLatency,
		),
		uid,
		ms,
	}
}

func (p *OutPacketToUpdateLatency) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteVarInt(2); err != nil {
		return nil, err
	}
	if err := dt.WriteVarInt(1); err != nil {
		return nil, err
	}
	if err := dt.WriteUUID(p.uid); err != nil {
		return nil, err
	}
	if err := dt.WriteVarInt(p.ms); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToUpdateLatency) GetUID() uuid.UUID {
	return p.uid
}

func (p *OutPacketToUpdateLatency) GetValue() int32 {
	return p.ms
}

func (p *OutPacketToUpdateLatency) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"uid: %s, "+
			"ms: %d "+
			"}",
		p.packet,
		p.uid,
		p.ms,
	)
}

type OutPacketToRemovePlayer struct {
	*packet
	uid uuid.UUID
}

func NewOutPacketToRemovePlayer(
	uid uuid.UUID,
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
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteVarInt(4); err != nil {
		return nil, err
	}
	if err := dt.WriteVarInt(1); err != nil {
		return nil, err
	}
	if err := dt.WriteUUID(p.uid); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToRemovePlayer) GetUID() uuid.UUID {
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
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteFloat64(p.x); err != nil {
		return nil, err
	}
	if err := dt.WriteFloat64(p.y); err != nil {
		return nil, err
	}
	if err := dt.WriteFloat64(p.z); err != nil {
		return nil, err
	}
	if err := dt.WriteFloat32(p.yaw); err != nil {
		return nil, err
	}
	if err := dt.WriteFloat32(p.pitch); err != nil {
		return nil, err
	}
	if err := dt.WriteInt8(0); err != nil {
		return nil, err
	}
	if err := dt.WriteVarInt(p.payload); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToTeleport) GetPosition() (
	float64, float64, float64,
) {
	return p.x, p.y, p.z
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

func (p *OutPacketToTeleport) GetLook() (
	float32, float32,
) {
	return p.yaw, p.pitch
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
	eid int32
}

func NewOutPacketToDespawnEntity(
	eid int32,
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
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteVarInt(1); err != nil {
		return nil, err
	}
	if err := dt.WriteVarInt(p.eid); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToDespawnEntity) GetEID() int32 {
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
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteInt32(p.dimension); err != nil {
		return nil, err
	}
	if err := dt.WriteUint8(p.difficulty); err != nil {
		return nil, err
	}
	if err := dt.WriteUint8(p.gamemode); err != nil {
		return nil, err
	}
	if err := dt.WriteString(p.level); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
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
	eid int32
	yaw float32
}

func NewOutPacketToSetEntityHeadLook(
	eid int32,
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
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteVarInt(p.eid); err != nil {
		return nil, err
	}
	if err := dt.WriteAngle(p.yaw); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToSetEntityHeadLook) GetEID() int32 {
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

type OutPacketToSetItemEntityMetadata struct {
	*packet
	eid int32
	md  *metadata.ItemEntityMetadata
}

func NewOutPacketToSetItemEntityMetadata(
	eid int32,
	md *metadata.ItemEntityMetadata,
) *OutPacketToSetItemEntityMetadata {
	return &OutPacketToSetItemEntityMetadata{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetItemEntityMetadata,
		),
		eid,
		md,
	}
}

func (p *OutPacketToSetItemEntityMetadata) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()

	if err := dt.WriteVarInt(
		p.eid,
	); err != nil {
		return nil, err
	}

	if err := p.md.Write(
		dt,
	); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

type OutPacketToSetPlayerMetadata struct {
	*packet
	eid int32
	md  *metadata.PlayerMetadata
}

func NewOutPacketToSetPlayerMetadata(
	eid int32,
	md *metadata.PlayerMetadata,
) *OutPacketToSetPlayerMetadata {
	return &OutPacketToSetPlayerMetadata{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetPlayerMetadata,
		),
		eid,
		md,
	}
}

func (p *OutPacketToSetPlayerMetadata) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()

	if err := dt.WriteVarInt(
		p.eid,
	); err != nil {
		return nil, err
	}

	if err := p.md.Write(
		dt,
	); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToSetPlayerMetadata) GetEID() int32 {
	return p.eid
}

func (p *OutPacketToSetPlayerMetadata) GetMetadata() *metadata.PlayerMetadata {
	return p.md
}

func (p *OutPacketToSetPlayerMetadata) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, eid: %d, md: {...} }",
		p.packet, p.eid,
	)
}

type OutPacketToSetArmorStandMetadata struct {
	*packet
	eid int32
	md  *metadata.ArmorStandMetadata
}

func NewOutPacketToSetArmorStandMetadata(
	eid int32,
	md *metadata.ArmorStandMetadata,
) *OutPacketToSetArmorStandMetadata {
	return &OutPacketToSetArmorStandMetadata{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetArmorStandMetadata,
		),
		eid,
		md,
	}
}

func (p *OutPacketToSetArmorStandMetadata) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()

	if err := dt.WriteVarInt(
		p.eid,
	); err != nil {
		return nil, err
	}

	if err := p.md.Write(
		dt,
	); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

type OutPacketToSetEntityVelocity struct {
	*packet
	eid     int32
	x, y, z int16
}

func NewOutPacketToSetEntityVelocity(
	eid int32,
	x, y, z int16,
) *OutPacketToSetEntityVelocity {
	return &OutPacketToSetEntityVelocity{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetEntityVelocity,
		),
		eid,
		x, y, z,
	}
}

func (p *OutPacketToSetEntityVelocity) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteVarInt(p.eid); err != nil {
		return nil, err
	}
	if err := dt.WriteInt16(p.x); err != nil {
		return nil, err
	}
	if err := dt.WriteInt16(p.y); err != nil {
		return nil, err
	}
	if err := dt.WriteInt16(p.z); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToSetEntityVelocity) GetEID() int32 {
	return p.eid
}

func (p *OutPacketToSetEntityVelocity) GetVector() (
	int16, int16, int16,
) {
	return p.x, p.y, p.z
}

func (p *OutPacketToSetEntityVelocity) GetX() int16 {
	return p.x
}

func (p *OutPacketToSetEntityVelocity) GetY() int16 {
	return p.y
}

func (p *OutPacketToSetEntityVelocity) GetZ() int16 {
	return p.z
}

type OutPacketToSetEntityEquipment struct {
	*packet
	eid  int32
	n    int32
	item item.Item
}

func NewOutPacketToSetEntityEquipment(
	eid int32,
	n int32,
	item item.Item,
) *OutPacketToSetEntityEquipment {
	return &OutPacketToSetEntityEquipment{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetEntityEquipment,
		),
		eid,
		n,
		item,
	}
}

func (p *OutPacketToSetEntityEquipment) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteVarInt(
		p.eid,
	); err != nil {
		return nil, err
	}
	if err := dt.WriteVarInt(
		p.n,
	); err != nil {
		return nil, err
	}
	if err := p.item.Write(
		dt,
	); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

type OutPacketToSetSpawnPosition struct {
	*packet
	x, y, z int
}

func NewOutPacketToSetSpawnPosition(
	x, y, z int,
) *OutPacketToSetSpawnPosition {
	return &OutPacketToSetSpawnPosition{
		newPacket(
			Outbound,
			PlayState,
			OutPacketIDToSetSpawnPosition,
		),
		x, y, z,
	}
}

func (p *OutPacketToSetSpawnPosition) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WritePosition(p.x, p.y, p.z); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToSetSpawnPosition) GetPosition() (
	int, int, int,
) {
	return p.x, p.y, p.z
}

func (p *OutPacketToSetSpawnPosition) GetX() int {
	return p.x
}

func (p *OutPacketToSetSpawnPosition) GetY() int {
	return p.y
}

func (p *OutPacketToSetSpawnPosition) GetZ() int {
	return p.z
}

func (p *OutPacketToSetSpawnPosition) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"x: %d, y: %d, z: %d "+
			"}",
		p.packet,
		p.x, p.y, p.z,
	)
}
