package server

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
)

type Entity interface {
	GetEID() int32

	GetUID() uuid.UUID

	GetX() float64

	GetY() float64

	GetZ() float64

	GetXYZ() (
		float64, float64, float64,
	)

	GetYaw() float32

	GetPitch() float32

	GetYawPitch() (
		float32, float32,
	)

	GetGround() bool

	UpdatePos(
		x, y, z float64,
		ground bool,
	)

	UpdateLook(
		yaw, pitch float32,
		ground bool,
	)
}

type entity struct {
	*sync.RWMutex

	eid        int32
	uid        uuid.UUID
	x, y, z    float64
	yaw, pitch float32
	ground     bool
}

func newEntity(
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
) *entity {
	return &entity{
		new(sync.RWMutex),
		eid,
		uid,
		x, y, z,
		yaw, pitch,
		false,
	}
}

func (e *entity) GetEID() int32 {
	e.RLock()
	defer e.RUnlock()

	return e.eid
}

func (e *entity) GetUID() uuid.UUID {
	e.RLock()
	defer e.RUnlock()

	return e.uid
}

func (e *entity) GetX() float64 {
	e.RLock()
	defer e.RUnlock()

	return e.x
}

func (e *entity) GetY() float64 {
	e.RLock()
	defer e.RUnlock()

	return e.y
}

func (e *entity) GetZ() float64 {
	e.RLock()
	defer e.RUnlock()

	return e.z
}

func (e *entity) GetXYZ() (
	float64, float64, float64,
) {
	e.RLock()
	defer e.RUnlock()

	return e.x, e.y, e.z
}

func (e *entity) GetYaw() float32 {
	e.RLock()
	defer e.RUnlock()

	return e.yaw
}

func (e *entity) GetPitch() float32 {
	e.RLock()
	defer e.RUnlock()

	return e.pitch
}

func (e *entity) GetYawPitch() (
	float32, float32,
) {
	e.RLock()
	defer e.RUnlock()

	return e.yaw, e.pitch
}

func (e *entity) GetGround() bool {
	e.RLock()
	defer e.RUnlock()

	return e.ground
}

func (e *entity) UpdatePos(
	x, y, z float64,
	ground bool,
) {
	e.Lock()
	defer e.Unlock()

	e.x, e.y, e.z = x, y, z
	e.ground = ground
}

func (e *entity) UpdateLook(
	yaw, pitch float32,
	ground bool,
) {
	e.Lock()
	defer e.Unlock()

	e.yaw, e.pitch = yaw, pitch
	e.ground = ground
}

func (e *entity) String() string {
	e.RLock()
	defer e.RUnlock()

	return fmt.Sprintf(
		"{ "+
			"eid: %d, "+
			"uid: %s, "+
			"x: %f, y: %f, z: %f, "+
			"yaw: %f, pitch: %f, "+
			"}",
		e.eid,
		e.uid.String(),
		e.x, e.y, e.z,
		e.yaw, e.pitch,
	)
}

type ItemEntity struct {
	*entity

	*sync.RWMutex

	item Item
}

func NewItemEntity(
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
	item Item,
) *ItemEntity {
	return &ItemEntity{
		newEntity(
			eid,
			uid,
			x, y, z,
			yaw, pitch,
		),
		new(sync.RWMutex),

		item,
	}
}

func (e *ItemEntity) GetItem() Item {
	return e.item
}

type Living interface {
	Entity
}

type living struct {
	*entity

	*sync.RWMutex
}

func newLiving(
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
) *living {
	return &living{
		newEntity(
			eid,
			uid,
			x, y, z,
			yaw, pitch,
		),
		new(sync.RWMutex),
	}
}

func (c *living) String() string {
	c.RLock()
	defer c.RUnlock()

	return fmt.Sprintf(
		"{ entity: %s }",
		c.entity,
	)
}

type Player struct {
	*living

	*sync.RWMutex
}

func NewPlayer(
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
) *Player {
	return &Player{
		newLiving(
			eid,
			uid,
			x, y, z,
			yaw, pitch,
		),

		new(sync.RWMutex),
	}
}

func (e *Player) String() string {
	e.RLock()
	defer e.RUnlock()

	return fmt.Sprintf(
		"{ "+
			"living: %s, "+
			"}",
		e.living,
	)
}
