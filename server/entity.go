package server

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/welcomehyunseo/warlord-server/server/item"
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

func (e *entity) GetPosition() (
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

func (e *entity) GetLook() (
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

	it item.Item
}

func NewItemEntity(
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
	it item.Item,
) *ItemEntity {
	return &ItemEntity{
		newEntity(
			eid,
			uid,
			x, y, z,
			yaw, pitch,
		),
		new(sync.RWMutex),

		it,
	}
}

func (e *ItemEntity) GetItem() item.Item {
	return e.it
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

	username string

	slots [45]item.Item
	hldIt item.Item
}

func NewPlayer(
	eid int32,
	uid uuid.UUID, username string,
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

		username,

		[45]item.Item{},
		nil,
	}
}

func (e *Player) clickLeftInInventoryWindow(
	slot int16,
) error {
	hldIt := e.hldIt
	if hldIt == nil {
		it := e.slots[slot]

		if it != nil {
			e.slots[slot] = nil
			e.hldIt = it
		}
	} else {
		it := e.slots[slot]

		if it == nil {
			e.slots[slot] = hldIt
			e.hldIt = nil
		} else {
			e.slots[slot] = hldIt
			e.hldIt = it
		}
	}

	return nil
}

func (e *Player) clickRightInInventoryWindow(
	slot int16,
) error {
	return nil
}

func (e *Player) pressShiftAndClickLeftInInventoryWindow(
	slot int16,
) error {
	return nil
}

func (e *Player) pressShiftAndClickRightInInventoryWindow(
	slot int16,
) error {
	return nil
}

func (e *Player) ClickInventoryWindow(
	slot int16,
	btn int8,
	mode int32,
) error {
	e.Lock()
	defer e.Unlock()

	switch mode {
	case 0:
		switch btn {
		case 0:
			if err := e.clickLeftInInventoryWindow(
				slot,
			); err != nil {
				return err
			}
			break
		case 1:
			if err := e.clickRightInInventoryWindow(
				slot,
			); err != nil {
				return err
			}
			break
		}
		break
	case 1:
		switch btn {
		case 0:
			if err := e.pressShiftAndClickLeftInInventoryWindow(
				slot,
			); err != nil {
				return err
			}
			break
		case 1:
			if err := e.pressShiftAndClickRightInInventoryWindow(
				slot,
			); err != nil {
				return err
			}
			break
		}
		break
	}

	fmt.Println("slots:", e.slots)
	fmt.Println("hldIt:", e.hldIt)
	return nil
}

func (e *Player) GetUsername() string {
	return e.username
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

type ArmorStand struct {
	*living

	*sync.RWMutex
}

func NewArmorStand(
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
) *ArmorStand {
	return &ArmorStand{
		newLiving(
			eid,
			uid,
			x, y, z,
			yaw, pitch,
		),

		new(sync.RWMutex),
	}
}

type ItemStand struct {
	*ArmorStand

	*sync.RWMutex

	it item.Item
}

func NewItemStand(
	eid int32,
	uid uuid.UUID,
	x, y, z float64,
	yaw, pitch float32,
	it item.Item,
) *ItemStand {
	return &ItemStand{
		NewArmorStand(
			eid,
			uid,
			x, y, z,
			yaw, pitch,
		),

		new(sync.RWMutex),

		it,
	}
}

func (e *ItemStand) GetItem() item.Item {
	return e.it
}
