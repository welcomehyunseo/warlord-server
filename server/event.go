package server

import (
	"fmt"
)

type ChanForSpawnPlayerEvent chan *SpawnPlayerEvent
type ChanForDespawnEntityEvent chan *DespawnEntityEvent

type ChanForSetEntityRelativePosEvent chan *SetEntityRelativePosEvent
type ChanForSetEntityLookEvent chan *SetEntityLookEvent
type ChanForSetEntityMetadataEvent chan *SetEntityMetadataEvent

type ChanForAddPlayerEvent chan *AddPlayerEvent
type ChanForUpdateLatencyEvent chan *UpdateLatencyEvent
type ChanForRemovePlayerEvent chan *RemovePlayerEvent

type ChanForConfirmKeepAliveEvent chan *ConfirmKeepAliveEvent

type DespawnEntityEvent struct {
	eid EID
}

func NewDespawnEntityEvent(
	eid EID,
) *DespawnEntityEvent {
	return &DespawnEntityEvent{
		eid: eid,
	}
}

func (e *DespawnEntityEvent) GetEID() EID {
	return e.eid
}

func (e *DespawnEntityEvent) String() string {
	return fmt.Sprintf(
		"{ eid: %d }",
		e.eid,
	)
}

type SpawnPlayerEvent struct {
	eid                 EID
	uid                 UID
	x, y, z             float64
	yaw, pitch          float32
	sneaking, sprinting bool
}

func NewSpawnPlayerEvent(
	eid EID,
	uid UID,
	x, y, z float64,
	yaw, pitch float32,
	sneaking, sprinting bool,
) *SpawnPlayerEvent {
	return &SpawnPlayerEvent{
		eid,
		uid,
		x, y, z,
		yaw, pitch,
		sneaking, sprinting,
	}
}

func (p *SpawnPlayerEvent) GetEID() EID {
	return p.eid
}

func (p *SpawnPlayerEvent) GetUUID() UID {
	return p.uid
}

func (p *SpawnPlayerEvent) GetX() float64 {
	return p.x
}

func (p *SpawnPlayerEvent) GetY() float64 {
	return p.y
}

func (p *SpawnPlayerEvent) GetZ() float64 {
	return p.z
}

func (p *SpawnPlayerEvent) GetYaw() float32 {
	return p.yaw
}

func (p *SpawnPlayerEvent) GetPitch() float32 {
	return p.pitch
}

func (p *SpawnPlayerEvent) IsSneaking() bool {
	return p.sneaking
}

func (p *SpawnPlayerEvent) IsSprinting() bool {
	return p.sprinting
}

func (p *SpawnPlayerEvent) String() string {
	return fmt.Sprintf(
		"{ "+
			"eid: %d, "+
			"uid: %s, "+
			"x: %f, y: %f, z: %f, "+
			"yaw: %f, pitch: %f, "+
			"sneaking: %v, sprinting: %v "+
			"}",
		p.eid,
		p.uid,
		p.x, p.y, p.z,
		p.yaw, p.pitch,
		p.sneaking, p.sprinting,
	)
}

type SetEntityRelativePosEvent struct {
	eid    EID
	deltaX int16
	deltaY int16
	deltaZ int16
	ground bool
}

func NewSetEntityRelativePosEvent(
	eid EID,
	deltaX, deltaY, deltaZ int16,
	ground bool,
) *SetEntityRelativePosEvent {
	return &SetEntityRelativePosEvent{
		eid:    eid,
		deltaX: deltaX,
		deltaY: deltaY,
		deltaZ: deltaZ,
		ground: ground,
	}
}

func (e *SetEntityRelativePosEvent) GetEID() EID {
	return e.eid
}

func (e *SetEntityRelativePosEvent) GetDeltaX() int16 {
	return e.deltaX
}

func (e *SetEntityRelativePosEvent) GetDeltaY() int16 {
	return e.deltaY
}

func (e *SetEntityRelativePosEvent) GetDeltaZ() int16 {
	return e.deltaZ
}

func (e *SetEntityRelativePosEvent) GetGround() bool {
	return e.ground
}

func (e *SetEntityRelativePosEvent) String() string {
	return fmt.Sprintf(
		"{ "+
			"eid: %d, "+
			"deltaX: %d, "+
			"deltaY: %d, "+
			"deltaZ: %d, "+
			"ground: %v "+
			"}",
		e.eid,
		e.deltaX,
		e.deltaY,
		e.deltaZ,
		e.ground,
	)
}

type SetEntityLookEvent struct {
	eid        EID
	yaw, pitch float32
	ground     bool
}

func NewSetEntityLookEvent(
	eid EID,
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

func (e *SetEntityLookEvent) GetEID() EID {
	return e.eid
}

func (e *SetEntityLookEvent) GetYaw() float32 {
	return e.yaw
}

func (e *SetEntityLookEvent) GetPitch() float32 {
	return e.pitch
}

func (e *SetEntityLookEvent) GetGround() bool {
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

type SetEntityMetadataEvent struct {
	eid      EID
	metadata *EntityMetadata
}

func NewSetEntityMetadataEvent(
	eid EID,
	metadata *EntityMetadata,
) *SetEntityMetadataEvent {
	return &SetEntityMetadataEvent{
		eid,
		metadata,
	}
}

func (e *SetEntityMetadataEvent) GetEID() EID {
	return e.eid
}

func (e *SetEntityMetadataEvent) GetMetadata() *EntityMetadata {
	return e.metadata
}

func (e *SetEntityMetadataEvent) String() string {
	return fmt.Sprintf(
		"{ eid: %d, metadata: %s }",
		e.eid, e.metadata,
	)
}

type AddPlayerEvent struct {
	uid      UID
	username string
	ctx      chan bool
}

func NewAddPlayerEvent(
	uid UID,
	username string,
) *AddPlayerEvent {
	return &AddPlayerEvent{
		uid:      uid,
		username: username,
		ctx:      make(chan bool, 1),
	}
}

func (e *AddPlayerEvent) GetUUID() UID {
	return e.uid
}

func (e *AddPlayerEvent) GetUsername() string {
	return e.username
}

func (e *AddPlayerEvent) Done() {
	e.ctx <- true
}

func (e *AddPlayerEvent) Fail() {
	e.ctx <- false
}

func (e *AddPlayerEvent) Wait() {
	<-e.ctx
	close(e.ctx)
}

func (e *AddPlayerEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v, username: %s } ",
		e.uid, e.username,
	)
}

type UpdateLatencyEvent struct {
	uid     UID
	latency int32
}

func NewUpdateLatencyEvent(
	uid UID,
	latency int32,
) *UpdateLatencyEvent {
	return &UpdateLatencyEvent{
		uid:     uid,
		latency: latency,
	}
}

func (e *UpdateLatencyEvent) GetUUID() UID {
	return e.uid
}

func (e *UpdateLatencyEvent) GetLatency() int32 {
	return e.latency
}

func (e *UpdateLatencyEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v, latency: %d } ",
		e.uid, e.latency,
	)
}

type RemovePlayerEvent struct {
	uid UID
}

func NewRemovePlayerEvent(
	uid UID,
) *RemovePlayerEvent {
	return &RemovePlayerEvent{
		uid: uid,
	}
}

func (e *RemovePlayerEvent) GetUUID() UID {
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
