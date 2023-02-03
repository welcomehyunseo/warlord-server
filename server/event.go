package server

import (
	"fmt"
	"github.com/google/uuid"
)

type ChanForUpdatePosEvent chan *UpdatePosEvent
type ChanForConfirmKeepAliveEvent chan *ConfirmKeepAliveEvent
type ChanForAddPlayerEvent chan *AddPlayerEvent
type ChanForRemovePlayerEvent chan *RemovePlayerEvent
type ChanForUpdateLatencyEvent chan *UpdateLatencyEvent
type ChanForSpawnPlayerEvent chan *SpawnPlayerEvent
type ChanForRelativeMoveEvent chan *RelativeMoveEvent
type ChanForDespawnEntityEvent chan *DespawnEntityEvent

type UpdatePosEvent struct {
	x float64
	y float64
	z float64
}

func NewUpdatePosEvent(
	x, y, z float64,
) *UpdatePosEvent {
	return &UpdatePosEvent{
		x: x,
		y: y,
		z: z,
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

func (p *AddPlayerEvent) GetUUID() uuid.UUID {
	return p.uid
}

func (p *AddPlayerEvent) GetUsername() string {
	return p.username
}

func (p *AddPlayerEvent) Done() {
	p.ctx <- true
}

func (p *AddPlayerEvent) Fail() {
	p.ctx <- false
}

func (p *AddPlayerEvent) Wait() {
	<-p.ctx
	close(p.ctx)
}

func (p *AddPlayerEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v, username: %s } ",
		p.uid, p.username,
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

func (p *RemovePlayerEvent) GetUUID() uuid.UUID {
	return p.uid
}

func (p *RemovePlayerEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v } ",
		p.uid,
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

func (p *UpdateLatencyEvent) GetUUID() uuid.UUID {
	return p.uid
}

func (p *UpdateLatencyEvent) GetLatency() int32 {
	return p.latency
}

func (p *UpdateLatencyEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v, latency: %d } ",
		p.uid, p.latency,
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

type RelativeMoveEvent struct {
	eid    int32
	deltaX int16
	deltaY int16
	deltaZ int16
	ground bool
}

func NewRelativeMoveEvent(
	eid int32,
	deltaX, deltaY, deltaZ int16,
	ground bool,
) *RelativeMoveEvent {
	return &RelativeMoveEvent{
		eid:    eid,
		deltaX: deltaX,
		deltaY: deltaY,
		deltaZ: deltaZ,
		ground: ground,
	}
}

func (p *RelativeMoveEvent) Pack() *Data {
	data := NewData()
	data.WriteVarInt(p.eid)
	data.WriteInt16(p.deltaX)
	data.WriteInt16(p.deltaY)
	data.WriteInt16(p.deltaZ)
	data.WriteBool(p.ground)

	return data
}

func (p *RelativeMoveEvent) GetEID() int32 {
	return p.eid
}

func (p *RelativeMoveEvent) GetDeltaX() int16 {
	return p.deltaX
}

func (p *RelativeMoveEvent) GetDeltaY() int16 {
	return p.deltaY
}

func (p *RelativeMoveEvent) GetDeltaZ() int16 {
	return p.deltaZ
}

func (p *RelativeMoveEvent) GetGround() bool {
	return p.ground
}

func (p *RelativeMoveEvent) String() string {
	return fmt.Sprintf(
		"{ "+
			"eid: %d, "+
			"deltaX: %d, "+
			"deltaY: %d, "+
			"deltaZ: %d, "+
			"ground: %v "+
			"}",
		p.eid,
		p.deltaX,
		p.deltaY,
		p.deltaZ,
		p.ground,
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

func (p *DespawnEntityEvent) GetEID() int32 {
	return p.eid
}

func (p *DespawnEntityEvent) String() string {
	return fmt.Sprintf(
		"{ eid: %d }",
		p.eid,
	)
}
