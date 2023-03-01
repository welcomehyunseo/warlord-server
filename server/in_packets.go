package server

import "fmt"

const InPacketIDToHandshake = 0x00

const InPacketIDToRequest = 0x00
const InPacketIDToPing = 0x01

const InPacketIDToStartLogin = 0x00

const InPacketIDToConfirmTeleport = 0x00
const InPacketIDToEnterChatText = 0x02
const InPacketIDToClickButton = 0x03
const InPacketIDToChangeSettings = 0x04
const InPacketIDToInteractWithEntity = 0x0A
const InPacketIDToConfirmKeepAlive = 0x0B
const InPacketIDToChangePos = 0x0D
const InPacketIDToChangePosAndLook = 0x0E
const InPacketIDToChangeLook = 0x0F
const InPacketIDToDoActions = 0x15
const InPacketIDToStartSneaking = 0x15
const InPacketIDToStopSneaking = 0x15
const InPacketIDToLeaveBed = 0x15
const InPacketIDToStartSprinting = 0x15
const InPacketIDToStopSprinting = 0x15
const InPacketIDToStartJumpWithHorse = 0x15
const InPacketIDToStopJumpWithHorse = 0x15
const InPacketIDToOpenHorseInventory = 0x15
const InPacketIDToStartFlyingWithElytra = 0x15

type InPacketToHandshake struct {
	*packet
	version int32
	addr    string
	port    uint16
	next    int32
}

func NewInPacketToHandshake() *InPacketToHandshake {
	return &InPacketToHandshake{
		packet: newPacket(
			Inbound,
			HandshakingState,
			InPacketIDToHandshake,
		),
	}
}

func (p *InPacketToHandshake) Unpack(
	data *Data,
) error {
	var err error
	p.version, err = data.ReadVarInt()
	if err != nil {
		return err
	}
	p.addr, err = data.ReadString()
	if err != nil {
		return err
	}
	p.port, err = data.ReadUint16()
	if err != nil {
		return err
	}
	p.next, err = data.ReadVarInt()
	if err != nil {
		return err
	}
	return err
}

func (p *InPacketToHandshake) GetVersion() int32 {
	return p.version
}

func (p *InPacketToHandshake) GetAddr() string {
	return p.addr
}

func (p *InPacketToHandshake) GetPort() uint16 {
	return p.port
}

func (p *InPacketToHandshake) GetNext() int32 {
	return p.next
}

func (p *InPacketToHandshake) String() string {
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

type InPacketToRequest struct {
	*packet
}

func NewInPacketToRequest() *InPacketToRequest {
	return &InPacketToRequest{
		packet: newPacket(
			Inbound,
			StatusState,
			InPacketIDToRequest,
		),
	}
}

func (p *InPacketToRequest) Unpack(
	data *Data,
) error {
	var err error
	return err
}

func (p *InPacketToRequest) String() string {
	return fmt.Sprintf(
		"{ packet: %+v }",
		p.packet,
	)
}

type InPacketToPing struct {
	*packet
	payload int64
}

func NewInPacketToPing() *InPacketToPing {
	return &InPacketToPing{
		packet: newPacket(
			Inbound,
			StatusState,
			InPacketIDToPing,
		),
	}
}

func (p *InPacketToPing) Unpack(
	data *Data,
) error {
	var err error
	p.payload, err = data.ReadInt64()
	if err != nil {
		return err
	}
	return err
}

func (p *InPacketToPing) GetPayload() int64 {
	return p.payload
}

func (p *InPacketToPing) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, payload: %d }",
		p.packet, p.payload,
	)
}

type InPacketToStartLogin struct {
	*packet
	username string
}

func NewInPacketToStartLogin() *InPacketToStartLogin {
	return &InPacketToStartLogin{
		packet: newPacket(
			Inbound,
			LoginState,
			InPacketIDToStartLogin,
		),
	}
}

func (p *InPacketToStartLogin) Unpack(
	data *Data,
) error {
	var err error
	p.username, err = data.ReadString()
	if err != nil {
		return nil
	}
	return err
}

func (p *InPacketToStartLogin) GetUsername() string {
	return p.username
}

func (p *InPacketToStartLogin) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, username: %s }",
		p.packet, p.username,
	)
}

type InPacketToEnterChatText struct {
	*packet
	text string
}

func NewInPacketToEnterChatText() *InPacketToEnterChatText {
	return &InPacketToEnterChatText{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToEnterChatText,
		),
	}
}

func (p *InPacketToEnterChatText) Unpack(
	data *Data,
) error {
	text, err := data.ReadString()
	if err != nil {
		return err
	}
	p.text = text

	return nil
}

func (p *InPacketToEnterChatText) GetText() string {
	return p.text
}

func (p *InPacketToEnterChatText) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, text: %s }",
		p.packet, p.text,
	)
}

type InPacketToConfirmTeleport struct {
	*packet
	payload int32
}

func NewInPacketToConfirmTeleport() *InPacketToConfirmTeleport {
	return &InPacketToConfirmTeleport{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToConfirmTeleport,
		),
	}
}

func (p *InPacketToConfirmTeleport) Unpack(
	data *Data,
) error {
	var err error
	p.payload, err = data.ReadVarInt()
	if err != nil {
		return err
	}
	return err
}

func (p *InPacketToConfirmTeleport) GetPayload() int32 {
	return p.payload
}

func (p *InPacketToConfirmTeleport) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, payload: %d }",
		p.packet, p.payload,
	)
}

type InPacketToClickButton struct {
	*packet

	respawn bool // when the Client is ready to complete login and respawn after death
	stats   bool // when the Client opens the statistics menu
}

func NewInPacketToClickButton() *InPacketToClickButton {
	return &InPacketToClickButton{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToClickButton,
		),
	}
}

func (p *InPacketToClickButton) Unpack(
	data *Data,
) error {
	var err error
	action, err := data.ReadVarInt()
	if err != nil {
		return err
	}
	if action == 0 {
		p.respawn = true
		p.stats = false
	} else {
		p.respawn = false
		p.stats = true
	}
	return err
}

func (p *InPacketToClickButton) IsRespawnStarted() bool {
	return p.respawn
}

func (p *InPacketToClickButton) IsStatsMenuOpened() bool {
	return p.stats
}

func (p *InPacketToClickButton) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, respawn: %v, stats: %v }",
		p.packet, p.respawn, p.stats,
	)
}

type InPacketToChangeSettings struct {
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

func NewInPacketToChangeSettings() *InPacketToChangeSettings {
	return &InPacketToChangeSettings{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToChangeSettings,
		),
	}
}

func (p *InPacketToChangeSettings) Unpack(
	data *Data,
) error {
	var err error
	p.local, err = data.ReadString()
	if err != nil {
		return err
	}
	p.rndDist, err = data.ReadInt8()
	if err != nil {
		return err
	}
	p.chatMode, err = data.ReadVarInt()
	if err != nil {
		return err
	}
	p.chatColors, err = data.ReadBool()
	if err != nil {
		return err
	}
	bitmask, err := data.ReadUint8()
	if err != nil {
		return err
	}
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
	p.mainHand, err = data.ReadVarInt()
	if err != nil {
		return err
	}
	return err
}

func (p *InPacketToChangeSettings) GetLocal() string {
	return p.local
}

func (p *InPacketToChangeSettings) GetRndDist() int8 {
	return p.rndDist
}

func (p *InPacketToChangeSettings) GetChatMode() int32 {
	return p.chatMode
}

func (p *InPacketToChangeSettings) GetChatColors() bool {
	return p.chatColors
}

func (p *InPacketToChangeSettings) GetCape() bool {
	return p.cape
}

func (p *InPacketToChangeSettings) GetJacket() bool {
	return p.jacket
}

func (p *InPacketToChangeSettings) GetLeftSleeve() bool {
	return p.leftSleeve
}

func (p *InPacketToChangeSettings) GetRightSleeve() bool {
	return p.rightSleeve
}

func (p *InPacketToChangeSettings) GetLeftPants() bool {
	return p.leftPants
}

func (p *InPacketToChangeSettings) GetRightPants() bool {
	return p.rightPants
}

func (p *InPacketToChangeSettings) GetHat() bool {
	return p.hat
}

func (p *InPacketToChangeSettings) GetMainHand() int32 {
	return p.mainHand
}

func (p *InPacketToChangeSettings) String() string {
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

type InPacketToInteractWithEntity struct {
	*packet
	target                    EID
	num                       int32
	targetX, targetY, targetZ float32
	hand                      int32
}

func NewInPacketToInteractWithEntity() *InPacketToInteractWithEntity {
	return &InPacketToInteractWithEntity{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToInteractWithEntity,
		),
	}
}

func (p *InPacketToInteractWithEntity) Unpack(
	data *Data,
) error {
	target, err := data.ReadVarInt()
	if err != nil {
		return err
	}
	p.target = EID(target)
	num, err := data.ReadVarInt()
	if err != nil {
		return err
	}
	p.num = num
	if num == 2 {
		targetX, err := data.ReadFloat32()
		if err != nil {
			return err
		}
		p.targetX = targetX
		targetY, err := data.ReadFloat32()
		if err != nil {
			return err
		}
		p.targetY = targetY
		targetZ, err := data.ReadFloat32()
		if err != nil {
			return err
		}
		p.targetZ = targetZ
	}
	if num == 0 || num == 2 {
		hand, err := data.ReadVarInt()
		if err != nil {
			return err
		}
		p.hand = hand
	}
	return nil
}

func (p *InPacketToInteractWithEntity) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, "+
			"target: %d, "+
			"num: %d, "+
			"targetX: %f, targetY: %f, targetZ: %f, "+
			"hand: %d "+
			"}",
		p.packet,
		p.target,
		p.num,
		p.targetX, p.targetY, p.targetZ,
		p.hand,
	)
}

type InPacketToConfirmKeepAlive struct {
	*packet
	payload int64
}

func NewInPacketToConfirmKeepAlive() *InPacketToConfirmKeepAlive {
	return &InPacketToConfirmKeepAlive{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToConfirmKeepAlive,
		),
	}
}

func (p *InPacketToConfirmKeepAlive) Unpack(
	data *Data,
) error {
	var err error
	p.payload, err = data.ReadInt64()
	if err != nil {
		return err
	}
	return err
}

func (p *InPacketToConfirmKeepAlive) GetPayload() int64 {
	return p.payload
}

func (p *InPacketToConfirmKeepAlive) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, payload: %d }",
		p.packet, p.payload,
	)
}

type InPacketToChangePos struct {
	*packet
	x      float64
	y      float64
	z      float64
	ground bool
}

func NewInPacketToChangePos() *InPacketToChangePos {
	return &InPacketToChangePos{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToChangePos,
		),
	}
}

func (p *InPacketToChangePos) Unpack(
	data *Data,
) error {
	var err error
	p.x, err = data.ReadFloat64()
	if err != nil {
		return err
	}
	p.y, err = data.ReadFloat64()
	if err != nil {
		return err
	}
	p.z, err = data.ReadFloat64()
	if err != nil {
		return err
	}
	p.ground, err = data.ReadBool()
	if err != nil {
		return err
	}
	return err
}

func (p *InPacketToChangePos) GetX() float64 {
	return p.x
}

func (p *InPacketToChangePos) GetY() float64 {
	return p.y
}

func (p *InPacketToChangePos) GetZ() float64 {
	return p.z
}

func (p *InPacketToChangePos) GetGround() bool {
	return p.ground
}

func (p *InPacketToChangePos) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, x: %f, y: %f, z: %f, ground: %v }",
		p.packet, p.x, p.y, p.z, p.ground,
	)
}

type InPacketToChangePosAndLook struct {
	*packet
	x      float64
	y      float64
	z      float64
	yaw    float32
	pitch  float32
	ground bool
}

func NewInPacketToChangePosAndLook() *InPacketToChangePosAndLook {
	return &InPacketToChangePosAndLook{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToChangePosAndLook,
		),
	}
}

func (p *InPacketToChangePosAndLook) Unpack(
	data *Data,
) error {
	var err error
	p.x, err = data.ReadFloat64()
	if err != nil {
		return err
	}
	p.y, err = data.ReadFloat64()
	if err != nil {
		return err
	}
	p.z, err = data.ReadFloat64()
	if err != nil {
		return err
	}
	p.yaw, err = data.ReadFloat32()
	if err != nil {
		return err
	}
	p.pitch, err = data.ReadFloat32()
	if err != nil {
		return err
	}
	p.ground, err = data.ReadBool()
	if err != nil {
		return err
	}
	return err
}

func (p *InPacketToChangePosAndLook) GetX() float64 {
	return p.x
}

func (p *InPacketToChangePosAndLook) GetY() float64 {
	return p.y
}

func (p *InPacketToChangePosAndLook) GetZ() float64 {
	return p.z
}

func (p *InPacketToChangePosAndLook) GetYaw() float32 {
	return p.yaw
}

func (p *InPacketToChangePosAndLook) GetPitch() float32 {
	return p.pitch
}

func (p *InPacketToChangePosAndLook) GetGround() bool {
	return p.ground
}

func (p *InPacketToChangePosAndLook) String() string {
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

type InPacketToChangeLook struct {
	*packet
	yaw    float32
	pitch  float32
	ground bool
}

func NewInPacketToChangeLook() *InPacketToChangeLook {
	return &InPacketToChangeLook{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToChangeLook,
		),
	}
}

func (p *InPacketToChangeLook) Unpack(
	data *Data,
) error {
	var err error
	p.yaw, err = data.ReadFloat32()
	if err != nil {
		return err
	}
	p.pitch, err = data.ReadFloat32()
	if err != nil {
		return err
	}
	p.ground, err = data.ReadBool()
	if err != nil {
		return err
	}
	return err
}

func (p *InPacketToChangeLook) GetYaw() float32 {
	return p.yaw
}

func (p *InPacketToChangeLook) GetPitch() float32 {
	return p.pitch
}

func (p *InPacketToChangeLook) GetGround() bool {
	return p.ground
}

func (p *InPacketToChangeLook) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"yaw: %f, pitch: %f, "+
			"ground: %v "+
			"}",
		p.packet,
		p.yaw, p.pitch,
		p.ground,
	)
}

type InPacketToStartSneaking struct {
	*packet
}

func NewInPacketToStartSneaking() *InPacketToStartSneaking {
	return &InPacketToStartSneaking{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToStartSneaking,
		),
	}
}

func (p *InPacketToStartSneaking) Unpack(
	data *Data,
) error {
	return nil
}

type InPacketToStopSneaking struct {
	*packet
}

func NewInPacketToStopSneaking() *InPacketToStopSneaking {
	return &InPacketToStopSneaking{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToStopSneaking,
		),
	}
}

func (p *InPacketToStopSneaking) Unpack(
	data *Data,
) error {
	return nil
}

type InPacketToStartSprinting struct {
	*packet
}

func NewInPacketToStartSprinting() *InPacketToStartSprinting {
	return &InPacketToStartSprinting{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToStartSprinting,
		),
	}
}

func (p *InPacketToStartSprinting) Unpack(
	data *Data,
) error {
	return nil
}

type InPacketToStopSprinting struct {
	*packet
}

func NewInPacketToStopSprinting() *InPacketToStopSprinting {
	return &InPacketToStopSprinting{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToStopSprinting,
		),
	}
}

func (p *InPacketToStopSprinting) Unpack(
	data *Data,
) error {
	return nil
}
