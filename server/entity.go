package server

import (
	"fmt"
	"github.com/google/uuid"
)

type EID int32
type UID uuid.UUID

var NilUID = UID(uuid.Nil)

type entity struct {
	eid                 EID
	uid                 UID
	x, y, z             float64
	yaw, pitch          float32
	prevX, prevY, prevZ float64
	sneaking, sprinting bool
}

func newEntity(
	eid EID,
	uid UID,
	x, y, z float64,
	yaw, pitch float32,
) *entity {
	return &entity{
		eid: eid,
		uid: uid,
		x:   x, y: y, z: z,
		yaw: yaw, pitch: pitch,
		prevX: x, prevY: y, prevZ: z,
	}
}

func (e *entity) GetEid() EID {
	return e.eid
}

func (e *entity) GetUid() UID {
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

func (e *entity) UpdateLook(
	yaw, pitch float32,
) {
	e.yaw = yaw
	e.pitch = pitch
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

func (e *entity) UpdatePos(
	x, y, z float64,
) {
	e.prevX = e.x
	e.prevY = e.y
	e.prevZ = e.z
	e.x = x
	e.y = y
	e.z = z
}

func (e *entity) IsSneaking() bool {
	return e.sneaking
}

func (e *entity) StartSneaking() {
	e.sneaking = true
}

func (e *entity) StopSneaking() {
	e.sneaking = false
}

func (e *entity) IsSprinting() bool {
	return e.sprinting
}

func (e *entity) StartSprinting() {
	e.sprinting = true
}

func (e *entity) StopSprinting() {
	e.sprinting = false
}

func (e *entity) String() string {
	return fmt.Sprintf(
		"{ "+
			"eid: %d, "+
			"uid: %+v, "+
			"x: %f, y: %f, z: %f, "+
			"yaw: %f, pitch: %f, "+
			"prevX: %f, prevY: %f, prevZ: %f, "+
			"sneaking: %v, sprinting: %v "+
			"}",
		e.eid,
		e.uid,
		e.x, e.y, e.z,
		e.yaw, e.pitch,
		e.prevX, e.prevY, e.prevZ,
		e.sneaking, e.sprinting,
	)
}

type living struct {
	*entity
}

func newLiving(
	eid EID,
	uid UID,
	x, y, z float64,
	yaw, pitch float32,
) *living {
	return &living{
		entity: newEntity(
			eid,
			uid,
			x, y, z,
			yaw, pitch,
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
	eid EID,
	uid UID,
	username string,
	x, y, z float64,
	yaw, pitch float32,

) *Player {
	return &Player{
		living: newLiving(
			eid,
			uid,
			x, y, z,
			yaw, pitch,
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
