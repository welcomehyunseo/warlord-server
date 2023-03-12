package server

import (
	"fmt"
	"github.com/google/uuid"
)

type ChanForConfirmKeepAliveEvent chan *ConfirmKeepAliveEvent

type ChanForAddPlayerEvent chan *AddPlayerEvent
type ChanForUpdateLatencyEvent chan *UpdateLatencyEvent
type ChanForRemovePlayerEvent chan *RemovePlayerEvent

type ChanForSpawnPlayerEvent chan *SpawnPlayerEvent
type ChanForSetEntityRelativeMoveEvent chan *SetEntityRelativeMoveEvent
type ChanForSetEntityLookEvent chan *SetEntityLookEvent
type ChanForSetEntityActionsEvent chan *SetEntityActionsEvent
type ChanForSetEntityVelocityEvent chan *SetEntityVelocityEvent
type ChanForDespawnEntityEvent chan *DespawnEntityEvent
type ChanForLoadChunkEvent chan *LoadChunkEvent
type ChanForUnloadChunkEvent chan *UnloadChunkEvent
type ChanForUpdateChunkEvent chan *UpdateChunkEvent

type ChanForClickWindowEvent chan *ClickWindowEvent

type waitable struct {
	ctx chan bool
}

func newWaitable() *waitable {
	return &waitable{
		make(chan bool, 1),
	}
}

func (w *waitable) Done() {
	w.ctx <- true
}

func (w *waitable) Fail() {
	w.ctx <- false
}

func (w *waitable) Wait() {
	<-w.ctx
	close(w.ctx)
}

type AddPlayerEvent struct {
	*waitable
	uid      uuid.UUID
	username string
}

func NewAddPlayerEvent(
	uid uuid.UUID,
	username string,
) *AddPlayerEvent {
	return &AddPlayerEvent{
		newWaitable(),
		uid,
		username,
	}
}

func (e *AddPlayerEvent) GetUID() uuid.UUID {
	return e.uid
}

func (e *AddPlayerEvent) GetUsername() string {
	return e.username
}

func (e *AddPlayerEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v, username: %s } ",
		e.uid, e.username,
	)
}

type UpdateLatencyEvent struct {
	uid uuid.UUID
	ms  int32
}

func NewUpdateLatencyEvent(
	uid uuid.UUID,
	ms int32,
) *UpdateLatencyEvent {
	return &UpdateLatencyEvent{
		uid,
		ms,
	}
}

func (e *UpdateLatencyEvent) GetUID() uuid.UUID {
	return e.uid
}

func (e *UpdateLatencyEvent) GetMilliseconds() int32 {
	return e.ms
}

func (e *UpdateLatencyEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v, ms: %d } ",
		e.uid, e.ms,
	)
}

type RemovePlayerEvent struct {
	*waitable
	uid uuid.UUID
}

func NewRemovePlayerEvent(
	uid uuid.UUID,
) *RemovePlayerEvent {
	return &RemovePlayerEvent{
		newWaitable(),
		uid,
	}
}

func (e *RemovePlayerEvent) GetUID() uuid.UUID {
	return e.uid
}

func (e *RemovePlayerEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v } ",
		e.uid,
	)
}

type ConfirmKeepAliveEvent struct {
	payload int64
}

func NewConfirmKeepAliveEvent(
	payload int64,
) *ConfirmKeepAliveEvent {
	return &ConfirmKeepAliveEvent{
		payload: payload,
	}
}

func (e *ConfirmKeepAliveEvent) GetPayload() int64 {
	return e.payload
}

func (e *ConfirmKeepAliveEvent) String() string {
	return fmt.Sprintf(
		"{ payload: %d }", e.payload,
	)
}

type UpdateChunkEvent struct {
	prevCx, prevCz int32
	currCx, currCz int32
}

func NewUpdateChunkEvent(
	prevCx, prevCz int32,
	currCx, currCz int32,
) *UpdateChunkEvent {
	return &UpdateChunkEvent{
		prevCx, prevCz,
		currCx, currCz,
	}
}

func (e *UpdateChunkEvent) GetPrevChunkPosition() (
	int32, int32,
) {
	return e.prevCx, e.prevCz
}

func (e *UpdateChunkEvent) GetPrevChunkX() int32 {
	return e.prevCx
}

func (e *UpdateChunkEvent) GetPrevChunkZ() int32 {
	return e.prevCz
}

func (e *UpdateChunkEvent) GetCurrChunkPosition() (
	int32, int32,
) {
	return e.currCx, e.currCz
}

func (e *UpdateChunkEvent) GetCurrChunkX() int32 {
	return e.currCx
}

func (e *UpdateChunkEvent) GetCurrChunkZ() int32 {
	return e.currCz
}

type SpawnPlayerEvent struct {
	eid        int32
	uid        uuid.UUID
	x, y, z    float64
	yaw, pitch float32
}

func NewSpawnPlayerEvent(
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
) *SpawnPlayerEvent {
	return &SpawnPlayerEvent{
		eid,
		uid,
		x, y, z,
		yaw, pitch,
	}
}

func (e *SpawnPlayerEvent) GetEID() int32 {
	return e.eid
}

func (e *SpawnPlayerEvent) GetUID() uuid.UUID {
	return e.uid
}

func (e *SpawnPlayerEvent) GetPosition() (
	float64, float64, float64,
) {
	return e.x, e.y, e.z
}

func (e *SpawnPlayerEvent) GetX() float64 {
	return e.x
}

func (e *SpawnPlayerEvent) GetY() float64 {
	return e.y
}

func (e *SpawnPlayerEvent) GetZ() float64 {
	return e.z
}

func (e *SpawnPlayerEvent) GetLook() (
	float32, float32,
) {
	return e.yaw, e.pitch
}

func (e *SpawnPlayerEvent) GetYaw() float32 {
	return e.yaw
}

func (e *SpawnPlayerEvent) GetPitch() float32 {
	return e.pitch
}

func (e *SpawnPlayerEvent) String() string {
	return fmt.Sprintf(
		"{ "+
			"eid: %d, "+
			"uid: %s, "+
			"x: %f, y: %f, z: %f, "+
			"yaw: %f, pitch: %f, "+
			"}",
		e.eid,
		e.uid,
		e.x, e.y, e.z,
		e.yaw, e.pitch,
	)
}

type DespawnEntityEvent struct {
	eid int32
}

func NewDespawnEntityEvent(
	eid int32,
) *DespawnEntityEvent {
	return &DespawnEntityEvent{
		eid: eid,
	}
}

func (e *DespawnEntityEvent) GetEID() int32 {
	return e.eid
}

func (e *DespawnEntityEvent) String() string {
	return fmt.Sprintf(
		"{ eid: %d }",
		e.eid,
	)
}

type SetEntityRelativeMoveEvent struct {
	eid        int32
	dx, dy, dz int16
	ground     bool
}

func NewSetEntityRelativeMoveEvent(
	eid int32,
	dx, dy, dz int16,
	ground bool,
) *SetEntityRelativeMoveEvent {
	return &SetEntityRelativeMoveEvent{
		eid,
		dx, dy, dz,
		ground,
	}
}

func (e *SetEntityRelativeMoveEvent) GetEID() int32 {
	return e.eid
}

func (e *SetEntityRelativeMoveEvent) GetDifferences() (
	int16, int16, int16,
) {
	return e.dx, e.dy, e.dz
}

func (e *SetEntityRelativeMoveEvent) GetDeltaX() int16 {
	return e.dx
}

func (e *SetEntityRelativeMoveEvent) GetDeltaY() int16 {
	return e.dy
}

func (e *SetEntityRelativeMoveEvent) GetDeltaZ() int16 {
	return e.dz
}

func (e *SetEntityRelativeMoveEvent) IsGround() bool {
	return e.ground
}

func (e *SetEntityRelativeMoveEvent) String() string {
	return fmt.Sprintf(
		"{ "+
			"eid: %d, "+
			"dx: %d, dy: %d, dz: %d, "+
			"ground: %v "+
			"}",
		e.eid,
		e.dx, e.dy, e.dz,
		e.ground,
	)
}

type SetEntityLookEvent struct {
	eid        int32
	yaw, pitch float32
	ground     bool
}

func NewSetEntityLookEvent(
	eid int32,
	yaw, pitch float32,
	ground bool,
) *SetEntityLookEvent {
	return &SetEntityLookEvent{
		eid:    eid,
		yaw:    yaw,
		pitch:  pitch,
		ground: ground,
	}
}

func (e *SetEntityLookEvent) GetEID() int32 {
	return e.eid
}

func (e *SetEntityLookEvent) GetLook() (
	float32, float32,
) {
	return e.yaw, e.pitch
}

func (e *SetEntityLookEvent) GetYaw() float32 {
	return e.yaw
}

func (e *SetEntityLookEvent) GetPitch() float32 {
	return e.pitch
}

func (e *SetEntityLookEvent) IsGround() bool {
	return e.ground
}

func (e *SetEntityLookEvent) String() string {
	return fmt.Sprintf(
		"{ "+
			"eid: %d, "+
			"yaw: %f, pitch: %f, "+
			"ground: %v "+
			"}",
		e.eid,
		e.yaw, e.pitch,
		e.ground,
	)
}

type SetEntityActionsEvent struct {
	eid                 int32
	sneaking, sprinting bool
}

func NewSetEntityActionsEvent(
	eid int32,
	sneaking, sprinting bool,
) *SetEntityActionsEvent {
	return &SetEntityActionsEvent{
		eid,
		sneaking, sprinting,
	}
}

func (e *SetEntityActionsEvent) GetEID() int32 {
	return e.eid
}

func (e *SetEntityActionsEvent) IsSneaking() bool {
	return e.sneaking
}

func (e *SetEntityActionsEvent) IsSprinting() bool {
	return e.sprinting
}

func (e *SetEntityActionsEvent) String() string {
	return fmt.Sprintf(
		"{ eid: %d, sneaking: %v, sprinting: %v }",
		e.eid, e.sneaking, e.sprinting,
	)
}

type SetEntityVelocityEvent struct {
	eid     int32
	x, y, z int16
}

func NewSetEntityVelocityEvent(
	eid int32,
	x, y, z int16,
) *SetEntityVelocityEvent {
	return &SetEntityVelocityEvent{
		eid,
		x, y, z,
	}
}

func (e *SetEntityVelocityEvent) GetEID() int32 {
	return e.eid
}

func (e *SetEntityVelocityEvent) GetVector() (
	int16, int16, int16,
) {
	return e.x, e.y, e.z
}

func (e *SetEntityVelocityEvent) GetX() int16 {
	return e.x
}

func (e *SetEntityVelocityEvent) GetY() int16 {
	return e.y
}

func (e *SetEntityVelocityEvent) GetZ() int16 {
	return e.z
}

type LoadChunkEvent struct {
	ow, init bool
	cx, cz   int32
	chunk    *Chunk
}

func NewLoadChunkEvent(
	ow, init bool,
	cx, cz int32,
	chunk *Chunk,
) *LoadChunkEvent {
	return &LoadChunkEvent{
		ow, init,
		cx, cz,
		chunk,
	}
}

func (e *LoadChunkEvent) IsOverworld() bool {
	return e.ow
}

func (e *LoadChunkEvent) IsInit() bool {
	return e.init
}

func (e *LoadChunkEvent) GetChunkPosition() (
	int32, int32,
) {
	return e.cx, e.cz
}

func (e *LoadChunkEvent) GetChunkX() int32 {
	return e.cx
}

func (e *LoadChunkEvent) GetChunkZ() int32 {
	return e.cz
}

func (e *LoadChunkEvent) GetChunk() *Chunk {
	return e.chunk
}

func (e *LoadChunkEvent) String() string {
	return fmt.Sprintf(
		"{ "+
			"ow: %v, "+
			"init: %v, "+
			"cx: %d, cz: %d, "+
			"chunk: %s "+
			"}",
		e.ow,
		e.init,
		e.cx, e.cz,
		e.chunk,
	)
}

type UnloadChunkEvent struct {
	cx, cz int32
}

func NewUnloadChunkEvent(
	cx, cz int32,
) *UnloadChunkEvent {
	return &UnloadChunkEvent{
		cx, cz,
	}
}

func (e *UnloadChunkEvent) GetChunkPosition() (
	int32, int32,
) {
	return e.cx, e.cz
}

func (e *UnloadChunkEvent) GetChunkX() int32 {
	return e.cx
}

func (e *UnloadChunkEvent) GetChunkZ() int32 {
	return e.cz
}

func (e *UnloadChunkEvent) String() string {
	return fmt.Sprintf(
		"{ "+
			"cx: %d, "+
			"cz: %d, "+
			"}",
		e.cx,
		e.cz,
	)
}

type ClickWindowEvent struct {
	winID int8
	slot  int16
	btn   int8
	act   int16
	mode  int32
}

func NewClickWindowEvent(
	winID int8,
	slot int16,
	btn int8,
	act int16,
	mode int32,
) *ClickWindowEvent {
	return &ClickWindowEvent{
		winID,
		slot,
		btn,
		act,
		mode,
	}
}
func (e *ClickWindowEvent) GetWindowID() int8 {
	return e.winID
}

func (e *ClickWindowEvent) GetSlotEnum() int16 {
	return e.slot
}

func (e *ClickWindowEvent) GetButtonEnum() int8 {
	return e.btn
}

func (e *ClickWindowEvent) GetActionNumber() int16 {
	return e.act
}

func (e *ClickWindowEvent) GetModeEnum() int32 {
	return e.mode
}

func (e *ClickWindowEvent) String() string {
	return fmt.Sprintf(
		"{ "+
			"winId: %d, "+
			"slot: %d, "+
			"btn: %d, "+
			"act: %d "+
			"mode: %d "+
			//"item: %s "+
			"}",
		e.winID,
		e.slot,
		e.btn,
		e.act,
		e.mode,
		//p.item,
	)
}
