package server

import (
	"fmt"
	"github.com/google/uuid"
)

type ChanForAddPlayerEvent chan *AddPlayerEvent
type ChanForRemovePlayerEvent chan *RemovePlayerEvent
type ChanForUpdateLatencyEvent chan *UpdateLatencyEvent

type ChanForDespawnEntityEvent chan *DespawnEntityEvent
type ChanForSpawnPlayerEvent chan *SpawnPlayerEvent
type ChanForUpdateChunkPosEvent chan *UpdateChunkPosEvent

type ChanForSetEntityLookEvent chan *SetEntityLookEvent
type ChanForUpdateLookEvent chan *UpdateLookEvent
type ChanForSetEntityRelativePosEvent chan *SetEntityRelativePosEvent
type ChanForUpdatePosEvent chan *UpdatePosEvent

type ChanForConfirmKeepAliveEvent chan *ConfirmKeepAliveEvent

type AddPlayerEvent struct {
	uid      uuid.UUID
	username string
	ctx      chan bool
}

func NewAddPlayerEvent(
	uid uuid.UUID,
	username string,
) *AddPlayerEvent {
	return &AddPlayerEvent{
		uid:      uid,
		username: username,
		ctx:      make(chan bool, 1),
	}
}

func (e *AddPlayerEvent) GetUUID() uuid.UUID {
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

type RemovePlayerEvent struct {
	uid uuid.UUID
}

func NewRemovePlayerEvent(
	uid uuid.UUID,
) *RemovePlayerEvent {
	return &RemovePlayerEvent{
		uid: uid,
	}
}

func (e *RemovePlayerEvent) GetUUID() uuid.UUID {
	return e.uid
}

func (e *RemovePlayerEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v } ",
		e.uid,
	)
}

type UpdateLatencyEvent struct {
	uid     uuid.UUID
	latency int32
}

func NewUpdateLatencyEvent(
	uid uuid.UUID,
	latency int32,
) *UpdateLatencyEvent {
	return &UpdateLatencyEvent{
		uid:     uid,
		latency: latency,
	}
}

func (e *UpdateLatencyEvent) GetUUID() uuid.UUID {
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

type SpawnPlayerEvent struct {
	eid   int32
	uid   uuid.UUID
	x     float64
	y     float64
	z     float64
	yaw   float32
	pitch float32
}

func NewSpawnPlayerEvent(
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
) *SpawnPlayerEvent {
	return &SpawnPlayerEvent{
		eid:   eid,
		uid:   uid,
		x:     x,
		y:     y,
		z:     z,
		yaw:   yaw,
		pitch: pitch,
	}
}

func (p *SpawnPlayerEvent) GetEID() int32 {
	return p.eid
}

func (p *SpawnPlayerEvent) GetUUID() uuid.UUID {
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

func (p *SpawnPlayerEvent) String() string {
	return fmt.Sprintf(
		"{ "+
			"eid: %d, "+
			"uid: %s, "+
			"x: %f, "+
			"y: %f, "+
			"z: %f, "+
			"yaw: %f, "+
			"pitch: %f "+
			"}",
		p.eid,
		p.uid,
		p.x,
		p.y,
		p.z,
		p.yaw,
		p.pitch,
	)
}

type UpdateChunkPosEvent struct {
	currCx int
	currCz int
	prevCx int
	prevCz int
}

func NewUpdateChunkPosEvent(
	currCx, currCz int,
	prevCx, prevCz int,
) *UpdateChunkPosEvent {
	return &UpdateChunkPosEvent{
		currCx: currCx,
		currCz: currCz,
		prevCx: prevCx,
		prevCz: prevCz,
	}
}

func (e *UpdateChunkPosEvent) GetCurrCx() int {
	return e.currCx
}

func (e *UpdateChunkPosEvent) GetCurrCz() int {
	return e.currCz
}

func (e *UpdateChunkPosEvent) GetPrevCx() int {
	return e.prevCx
}

func (e *UpdateChunkPosEvent) GetPrevCz() int {
	return e.prevCz
}

func (e *UpdateChunkPosEvent) String() string {
	return fmt.Sprintf(
		"{ currCx: %d, currCz: %d, prevCx: %d, prevCz: %d }",
		e.currCx, e.currCz, e.prevCx, e.prevCz,
	)
}

type SetEntityLookEvent struct {
	eid    int32
	yaw    float32
	pitch  float32
	ground bool
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
			"yaw: %f, "+
			"pitch: %f, "+
			"ground: %v "+
			"}",
		e.eid,
		e.yaw,
		e.pitch,
		e.ground,
	)
}

type UpdateLookEvent struct {
	yaw    float32
	pitch  float32
	ground bool
}

func NewUpdateLookEvent(
	yaw, pitch float32,
	ground bool,
) *UpdateLookEvent {
	return &UpdateLookEvent{
		yaw:    yaw,
		pitch:  pitch,
		ground: ground,
	}
}

func (e *UpdateLookEvent) GetYaw() float32 {
	return e.yaw
}

func (e *UpdateLookEvent) GetPitch() float32 {
	return e.pitch
}

func (e *UpdateLookEvent) GetGround() bool {
	return e.ground
}

func (e *UpdateLookEvent) String() string {
	return fmt.Sprintf(
		"{ "+
			"yaw: %f, "+
			"pitch: %f, "+
			"ground: %v "+
			"}",
		e.yaw,
		e.pitch,
		e.ground,
	)
}

type SetEntityRelativePosEvent struct {
	eid    int32
	deltaX int16
	deltaY int16
	deltaZ int16
	ground bool
}

func NewSetEntityRelativePosEvent(
	eid int32,
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

func (e *SetEntityRelativePosEvent) GetEID() int32 {
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

type UpdatePosEvent struct {
	x      float64
	y      float64
	z      float64
	ground bool
}

func NewUpdatePosEvent(
	x, y, z float64,
	ground bool,
) *UpdatePosEvent {
	return &UpdatePosEvent{
		x:      x,
		y:      y,
		z:      z,
		ground: ground,
	}
}

func (e *UpdatePosEvent) GetX() float64 {
	return e.x
}

func (e *UpdatePosEvent) GetY() float64 {
	return e.y
}

func (e *UpdatePosEvent) GetZ() float64 {
	return e.z
}

func (e *UpdatePosEvent) GetGround() bool {
	return e.ground
}

func (e *UpdatePosEvent) String() string {
	return fmt.Sprintf(
		"{ x: %f, y: %f, z: %f }",
		e.x, e.y, e.z,
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
