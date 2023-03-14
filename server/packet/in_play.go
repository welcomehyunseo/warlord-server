package packet

import (
	"fmt"
	"github.com/welcomehyunseo/warlord-server/server/data"
)

const InPacketIDToConfirmTeleport = 0x00
const InPacketIDToEnterChatText = 0x02
const InPacketIDToClickButton = 0x03
const InPacketIDToChangeSettings = 0x04
const InPacketIDToConfirmTransactionOfWindow = 0x05
const InPacketIDToClickWindow = 0x07
const InPacketIDToInteractWithEntity = 0x0A
const InPacketIDToConfirmKeepAlive = 0x0B
const InPacketIDToChangePosition = 0x0D
const InPacketIDToChangePositionAndLook = 0x0E
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
	arr []byte,
) error {
	data := data.NewDataWithBytes(arr)

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
	arr []byte,
) error {
	dt := data.NewDataWithBytes(arr)

	text, err := dt.ReadString()
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
	arr []byte,
) error {
	dt := data.NewDataWithBytes(arr)

	action, err := dt.ReadVarInt()
	if err != nil {
		return err
	}
	if action == 0 {
		p.respawn = true
		//p.stats = false
	} else {
		//p.respawn = false
		p.stats = true
	}

	return nil
}

func (p *InPacketToClickButton) IsRespawnAfterDeath() bool {
	return p.respawn
}

func (p *InPacketToClickButton) IsStatisticsMenuOpened() bool {
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
	local      string
	rndDist    int8
	chatMode   int32
	chatColors bool
	cape       bool
	jacket     bool
	lSleeve    bool
	rSleeve    bool
	lPants     bool
	rPants     bool
	hat        bool
	mh         int32
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
	arr []byte,
) error {
	dt := data.NewDataWithBytes(arr)

	local, err := dt.ReadString()
	if err != nil {
		return err
	}
	p.local = local

	rndDist, err := dt.ReadInt8()
	if err != nil {
		return err
	}
	p.rndDist = rndDist

	chatMode, err := dt.ReadVarInt()
	if err != nil {
		return err
	}
	p.chatMode = chatMode

	chatColors, err := dt.ReadBool()
	if err != nil {
		return err
	}
	p.chatColors = chatColors

	bitmask, err := dt.ReadUint8()
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
		p.lSleeve = true
	} else {
		p.lSleeve = false
	}
	if bitmask&uint8(8) == uint8(8) {
		p.rSleeve = true
	} else {
		p.rSleeve = false
	}
	if bitmask&uint8(16) == uint8(16) {
		p.lPants = true
	} else {
		p.lPants = false
	}
	if bitmask&uint8(32) == uint8(32) {
		p.rPants = true
	} else {
		p.rPants = false
	}
	if bitmask&uint8(64) == uint8(64) {
		p.hat = true
	} else {
		p.hat = false
	}

	mh, err := dt.ReadVarInt()
	if err != nil {
		return err
	}
	p.mh = mh

	return err
}

func (p *InPacketToChangeSettings) GetLocal() string {
	return p.local
}

func (p *InPacketToChangeSettings) GetRenderDistance() int8 {
	return p.rndDist
}

func (p *InPacketToChangeSettings) GetChatMode() int32 {
	return p.chatMode
}

func (p *InPacketToChangeSettings) IsChatColors() bool {
	return p.chatColors
}

func (p *InPacketToChangeSettings) GetSkins() (
	bool, bool, bool, bool, bool, bool, bool,
) {
	return p.cape, p.jacket, p.lSleeve, p.rSleeve, p.lPants, p.rPants, p.hat
}

func (p *InPacketToChangeSettings) IsCapeOn() bool {
	return p.cape
}

func (p *InPacketToChangeSettings) IsJacketOn() bool {
	return p.jacket
}

func (p *InPacketToChangeSettings) IsLeftSleeveOn() bool {
	return p.lSleeve
}

func (p *InPacketToChangeSettings) IsRightSleeveOn() bool {
	return p.rSleeve
}

func (p *InPacketToChangeSettings) IsLeftPantsOn() bool {
	return p.lPants
}

func (p *InPacketToChangeSettings) IsRightPantsOn() bool {
	return p.rPants
}

func (p *InPacketToChangeSettings) IsHatOn() bool {
	return p.hat
}

func (p *InPacketToChangeSettings) IsMainHandLeft() bool {
	return p.mh == 0
}

func (p *InPacketToChangeSettings) IsMainHandRight() bool {
	return p.mh == 1
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
			"lSleeve: %v, rSleeve: %v, "+
			"lPants: %v, rPants: %v, "+
			"hat: %v, "+
			"mh: %d "+
			"}",
		p.packet,
		p.local,
		p.rndDist,
		p.chatMode,
		p.chatColors,
		p.cape, p.jacket,
		p.lSleeve, p.rSleeve,
		p.lPants, p.rPants,
		p.hat,
		p.mh,
	)
}

type InPacketToConfirmTransactionOfWindow struct {
	*packet
	winID  int8
	actNum int16
	accept bool
}

func NewInPacketToConfirmTransactionOfWindow() *InPacketToConfirmTransactionOfWindow {
	return &InPacketToConfirmTransactionOfWindow{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToConfirmTransactionOfWindow,
		),
	}
}

func (p *InPacketToConfirmTransactionOfWindow) Unpack(
	arr []byte,
) error {
	dt := data.NewDataWithBytes(arr)

	winID, err := dt.ReadInt8()
	if err != nil {
		return err
	}
	p.winID = winID

	actNum, err := dt.ReadInt16()
	if err != nil {
		return err
	}
	p.actNum = actNum

	accept, err := dt.ReadBool()
	if err != nil {
		return err
	}
	p.accept = accept

	return nil
}

func (p *InPacketToConfirmTransactionOfWindow) GetWindowID() int8 {
	return p.winID
}

func (p *InPacketToConfirmTransactionOfWindow) GetActionNumber() int16 {
	return p.actNum
}

func (p *InPacketToConfirmTransactionOfWindow) IsAccepted() bool {
	return p.accept
}

func (p *InPacketToConfirmTransactionOfWindow) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"winId: %d, "+
			"act: %d "+
			"accept: %v "+
			"}",
		p.packet,
		p.winID,
		p.actNum,
		p.accept,
	)
}

type InPacketToClickWindow struct {
	*packet

	winID int8
	slot  int16
	btn   int8
	act   int16
	mode  int32
	//it    Item
}

func NewInPacketToClickWindow() *InPacketToClickWindow {
	return &InPacketToClickWindow{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToClickWindow,
		),
	}
}

func (p *InPacketToClickWindow) Unpack(
	arr []byte,
) error {
	dt := data.NewDataWithBytes(arr)

	winID, err := dt.ReadInt8()
	if err != nil {
		return err
	}
	p.winID = winID

	slot, err := dt.ReadInt16()
	if err != nil {
		return err
	}
	p.slot = slot

	btn, err := dt.ReadInt8()
	if err != nil {
		return err
	}
	p.btn = btn

	act, err := dt.ReadInt16()
	if err != nil {
		return err
	}
	p.act = act

	mode, err := dt.ReadVarInt()
	if err != nil {
		return err
	}
	p.mode = mode

	//it, err := ReadItem(
	//	dt,
	//)
	//if err != nil {
	//	return err
	//}
	//p.it = it

	return nil
}

func (p *InPacketToClickWindow) GetWindowID() int8 {
	return p.winID
}

func (p *InPacketToClickWindow) GetSlotNumber() int16 {
	return p.slot
}

func (p *InPacketToClickWindow) GetButtonEnum() int8 {
	return p.btn
}

func (p *InPacketToClickWindow) GetActionNumber() int16 {
	return p.act
}

func (p *InPacketToClickWindow) GetModeEnum() int32 {
	return p.mode
}

//func (p *InPacketToClickWindow) GetItem() Item {
//	return p.it
//}

func (p *InPacketToClickWindow) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"winId: %d, "+
			"slot: %d, "+
			"btn: %d, "+
			"act: %d "+
			"mode: %d "+
			//"it: %s "+
			"}",
		p.packet,
		p.winID,
		p.slot,
		p.btn,
		p.act,
		p.mode,
		//p.it,
	)
}

type InPacketToInteractWithEntity struct {
	*packet
	eid        int32
	num        int32
	tx, ty, tz float32
	hand       int32
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
	arr []byte,
) error {
	dt := data.NewDataWithBytes(arr)

	eid, err := dt.ReadVarInt()
	if err != nil {
		return err
	}
	p.eid = eid

	num, err := dt.ReadVarInt()
	if err != nil {
		return err
	}
	p.num = num

	if num == 2 {
		tx, err := dt.ReadFloat32()
		if err != nil {
			return err
		}
		p.tx = tx
		ty, err := dt.ReadFloat32()
		if err != nil {
			return err
		}
		p.ty = ty
		tz, err := dt.ReadFloat32()
		if err != nil {
			return err
		}
		p.tz = tz
	}
	if num == 0 || num == 2 {
		hand, err := dt.ReadVarInt()
		if err != nil {
			return err
		}
		p.hand = hand
	}
	return nil
}

// GetPosition
// GetTargetX

func (p *InPacketToInteractWithEntity) String() string {
	return fmt.Sprintf(
		"{ "+
			"packet: %+v, "+
			"eid: %d, "+
			"slot: %d, "+
			"tx: %f, ty: %f, tz: %f, "+
			"hand: %d "+
			"}",
		p.packet,
		p.eid,
		p.num,
		p.tx, p.ty, p.tz,
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
	arr []byte,
) error {
	dt := data.NewDataWithBytes(arr)

	var err error
	p.payload, err = dt.ReadInt64()
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

type InPacketToChangePosition struct {
	*packet
	x, y, z float64
	ground  bool
}

func NewInPacketToChangePosition() *InPacketToChangePosition {
	return &InPacketToChangePosition{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToChangePosition,
		),
	}
}

func (p *InPacketToChangePosition) Unpack(
	arr []byte,
) error {
	dt := data.NewDataWithBytes(arr)

	var err error
	p.x, err = dt.ReadFloat64()
	if err != nil {
		return err
	}
	p.y, err = dt.ReadFloat64()
	if err != nil {
		return err
	}
	p.z, err = dt.ReadFloat64()
	if err != nil {
		return err
	}
	p.ground, err = dt.ReadBool()
	if err != nil {
		return err
	}
	return err
}

func (p *InPacketToChangePosition) GetPosition() (
	float64, float64, float64,
) {
	return p.x, p.y, p.z
}

func (p *InPacketToChangePosition) GetX() float64 {
	return p.x
}

func (p *InPacketToChangePosition) GetY() float64 {
	return p.y
}

func (p *InPacketToChangePosition) GetZ() float64 {
	return p.z
}

func (p *InPacketToChangePosition) IsGround() bool {
	return p.ground
}

func (p *InPacketToChangePosition) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, x: %f, y: %f, z: %f, ground: %v }",
		p.packet, p.x, p.y, p.z, p.ground,
	)
}

type InPacketToChangePositionAndLook struct {
	*packet
	x, y, z    float64
	yaw, pitch float32
	ground     bool
}

func NewInPacketToChangePositionAndLook() *InPacketToChangePositionAndLook {
	return &InPacketToChangePositionAndLook{
		packet: newPacket(
			Inbound,
			PlayState,
			InPacketIDToChangePositionAndLook,
		),
	}
}

func (p *InPacketToChangePositionAndLook) Unpack(
	arr []byte,
) error {
	dt := data.NewDataWithBytes(arr)

	var err error
	p.x, err = dt.ReadFloat64()
	if err != nil {
		return err
	}
	p.y, err = dt.ReadFloat64()
	if err != nil {
		return err
	}
	p.z, err = dt.ReadFloat64()
	if err != nil {
		return err
	}
	p.yaw, err = dt.ReadFloat32()
	if err != nil {
		return err
	}
	p.pitch, err = dt.ReadFloat32()
	if err != nil {
		return err
	}
	p.ground, err = dt.ReadBool()
	if err != nil {
		return err
	}
	return err
}

func (p *InPacketToChangePositionAndLook) GetPosition() (
	float64, float64, float64,
) {
	return p.x, p.y, p.z
}

func (p *InPacketToChangePositionAndLook) GetX() float64 {
	return p.x
}

func (p *InPacketToChangePositionAndLook) GetY() float64 {
	return p.y
}

func (p *InPacketToChangePositionAndLook) GetZ() float64 {
	return p.z
}

func (p *InPacketToChangePositionAndLook) GetLook() (
	float32, float32,
) {
	return p.yaw, p.pitch
}

func (p *InPacketToChangePositionAndLook) GetYaw() float32 {
	return p.yaw
}

func (p *InPacketToChangePositionAndLook) GetPitch() float32 {
	return p.pitch
}

func (p *InPacketToChangePositionAndLook) IsGround() bool {
	return p.ground
}

func (p *InPacketToChangePositionAndLook) String() string {
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
	yaw, pitch float32
	ground     bool
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
	arr []byte,
) error {
	dt := data.NewDataWithBytes(arr)

	var err error
	p.yaw, err = dt.ReadFloat32()
	if err != nil {
		return err
	}
	p.pitch, err = dt.ReadFloat32()
	if err != nil {
		return err
	}
	p.ground, err = dt.ReadBool()
	if err != nil {
		return err
	}
	return err
}

func (p *InPacketToChangeLook) GetLook() (
	float32, float32,
) {
	return p.yaw, p.pitch
}

func (p *InPacketToChangeLook) GetYaw() float32 {
	return p.yaw
}

func (p *InPacketToChangeLook) GetPitch() float32 {
	return p.pitch
}

func (p *InPacketToChangeLook) IsGround() bool {
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
	arr []byte,
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
	arr []byte,
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
	arr []byte,
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
	arr []byte,
) error {

	return nil
}
