package server

import (
	"fmt"
	"github.com/google/uuid"
)

type entity struct {
	eid   int32
	uid   uuid.UUID
	x     float64
	y     float64
	z     float64
	yaw   float32
	pitch float32
	prevX float64
	prevY float64
	prevZ float64
}

func newEntity(
	eid int32,
	uid uuid.UUID,
	x float64,
	y float64,
	z float64,
	yaw float32,
	pitch float32,
) *entity {
	return &entity{
		eid:   eid,
		uid:   uid,
		x:     x,
		y:     y,
		z:     z,
		yaw:   yaw,
		pitch: pitch,
		prevX: x,
		prevY: y,
		prevZ: z,
	}
}

func (e *entity) UpdatePos(
	x float64,
	y float64,
	z float64,
) {
	e.prevX = e.x
	e.prevY = e.y
	e.prevZ = e.z
	e.x = x
	e.y = y
	e.z = z
}

func (e *entity) GetEid() int32 {
	return e.eid
}

func (e *entity) GetUid() uuid.UUID {
	return e.uid
}

func (e *entity) GetX() float64 {
	return e.x
}

func (e *entity) GetY() float64 {
	return e.y
}

func (e *entity) GetZ() float64 {
	return e.z
}

func (e *entity) GetYaw() float32 {
	return e.yaw
}

func (e *entity) GetPitch() float32 {
	return e.pitch
}

func (e *entity) GetPrevX() float64 {
	return e.prevX
}

func (e *entity) GetPrevY() float64 {
	return e.prevY
}

func (e *entity) GetPrevZ() float64 {
	return e.prevZ
}

func (e *entity) String() string {
	return fmt.Sprintf(
		"{ "+
			"eid: %d, "+
			"uid: %+v, "+
			"x: %f, "+
			"y: %f, "+
			"z: %f, "+
			"yaw: %f, "+
			"pitch: %f "+
			"}",
		e.eid, e.uid, e.x, e.y, e.z, e.yaw, e.pitch,
	)
}

type living struct {
	*entity
}

func newLiving(
	eid int32,
	uid uuid.UUID,
	x float64,
	y float64,
	z float64,
	yaw float32,
	pitch float32,
) *living {
	return &living{
		entity: newEntity(
			eid,
			uid,
			x,
			y,
			z,
			yaw,
			pitch,
		),
	}
}

func (l *living) String() string {
	return fmt.Sprintf(
		"{ entity: %+v }",
		l.entity,
	)
}

type Player struct {
	*living

	username string
}

func NewPlayer(
	eid int32,
	uid uuid.UUID,
	username string,
	x float64,
	y float64,
	z float64,
	yaw float32,
	pitch float32,

) *Player {
	return &Player{
		living: newLiving(
			eid,
			uid,
			x,
			y,
			z,
			yaw,
			pitch,
		),
		username: username,
	}
}

func (p *Player) GetUsername() string {
	return p.username
}

func (p *Player) String() string {
	return fmt.Sprintf(
		"{ living: %+v, username: %s }",
		p.living, p.username,
	)
}
