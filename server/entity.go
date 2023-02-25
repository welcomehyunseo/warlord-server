package server

import (
	"fmt"
	"github.com/google/uuid"
)

type EID int32
type UID uuid.UUID

var NilUID = UID(uuid.Nil)

type Entity interface {
	GetEID() EID
	GetUID() UID

	GetX() float64
	GetY() float64
	GetZ() float64

	GetPrevX() float64
	GetPrevY() float64
	GetPrevZ() float64
	UpdatePos(
		x, y, z float64,
		ground bool,
	) error

	GetYaw() float32
	GetPitch() float32
	UpdateLook(
		yaw, pitch float32,
		ground bool,
	) error

	IsGround() bool

	IsSneaking() bool
	UpdateSneaking(
		sneaking bool,
	) error

	IsSprinting() bool
	UpdateSprinting(
		sprinting bool,
	) error
}

type entity struct {
	eid                 EID
	uid                 UID
	x, y, z             float64
	prevX, prevY, prevZ float64
	yaw, pitch          float32
	ground              bool
	sneaking, sprinting bool
}

func newEntity(
	eid EID,
	uid UID,
) *entity {
	return &entity{
		eid: eid,
		uid: uid,
	}
}

func (e *entity) GetEID() EID {
	return e.eid
}

func (e *entity) GetUID() UID {
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
	ground bool,
) error {
	e.prevX = e.x
	e.prevY = e.y
	e.prevZ = e.z
	e.x = x
	e.y = y
	e.z = z
	e.ground = ground

	return nil
}

func (e *entity) GetYaw() float32 {
	return e.yaw
}

func (e *entity) GetPitch() float32 {
	return e.pitch
}

func (e *entity) UpdateLook(
	yaw, pitch float32,
	ground bool,
) error {
	e.yaw = yaw
	e.pitch = pitch
	e.ground = ground

	return nil
}

func (e *entity) IsGround() bool {
	return e.ground
}

func (e *entity) IsSneaking() bool {
	return e.sneaking
}

func (e *entity) UpdateSneaking(
	sneaking bool,
) error {
	e.sneaking = sneaking

	return nil
}

func (e *entity) IsSprinting() bool {
	return e.sprinting
}

func (e *entity) UpdateSprinting(
	sprinting bool,
) error {
	e.sprinting = sprinting

	return nil
}

func (e *entity) String() string {
	return fmt.Sprintf(
		"{ "+
			"eid: %d, "+
			"uid: %s, "+
			"x: %f, y: %f, z: %f, "+
			"yaw: %f, pitch: %f, "+
			"prevX: %f, prevY: %f, prevZ: %f, "+
			"sneaking: %v, sprinting: %v "+
			"}",
		e.eid,
		uuid.UUID(e.uid).String(),
		e.x, e.y, e.z,
		e.yaw, e.pitch,
		e.prevX, e.prevY, e.prevZ,
		e.sneaking, e.sprinting,
	)
}

type Living interface {
	Entity
}

type living struct {
	*entity
}

func newLiving(
	eid EID,
	uid UID,
) *living {
	return &living{
		entity: newEntity(
			eid,
			uid,
		),
	}
}

func (l *living) String() string {
	return fmt.Sprintf(
		"{ entity: %+v }",
		l.entity,
	)
}

type Player interface {
	Living

	GetUsername() string
}

type player struct {
	*living

	username string
}

func newPlayer(
	eid EID,
	uid UID,
	username string,
) *player {
	return &player{
		living: newLiving(
			eid,
			uid,
		),
		username: username,
	}
}

func (p *player) UpdatePos(
	x, y, z float64,
	ground bool,
) error {
	if err := p.living.UpdatePos(
		x, y, z,
		ground,
	); err != nil {
		return err
	}

	return nil
}

func (p *player) UpdateLook(
	yaw, pitch float32,
	ground bool,
) error {
	if err := p.living.UpdateLook(
		yaw, pitch,
		ground,
	); err != nil {
		return err
	}

	return nil
}

func (p *player) UpdateSneaking(
	sneaking bool,
) error {
	if err := p.living.UpdateSneaking(
		sneaking,
	); err != nil {
		return err
	}

	return nil
}

func (p *player) UpdateSprinting(
	sprinting bool,
) error {
	if err := p.living.UpdateSprinting(
		sprinting,
	); err != nil {
		return err
	}

	return nil
}

func (p *player) GetUsername() string {
	return p.username
}

func (p *player) String() string {
	return fmt.Sprintf(
		"{ living: %+v, username: %s }",
		p.living, p.username,
	)
}

type Guest struct {
	*player
}

func NewGuest(
	eid EID,
	uid UID,
	username string,
) *Guest {
	return &Guest{
		player: newPlayer(
			eid,
			uid,
			username,
		),
	}
}

func (p *Guest) String() string {
	return fmt.Sprintf(
		"{ player: %+v }",
		p.player,
	)
}

type Warlord struct {
	*player
}

func NewWarlord(
	eid EID,
	uid UID,
	username string,
) *Warlord {
	return &Warlord{
		player: newPlayer(
			eid,
			uid,
			username,
		),
	}
}

func (p *Warlord) String() string {
	return fmt.Sprintf(
		"{ player: %+v }",
		p.player,
	)
}
