package server

import (
	"fmt"
	"github.com/google/uuid"
)

type EID int32
type UID uuid.UUID

var NilUID = UID(uuid.Nil)

type Entity interface {
	GetEid() EID
	GetUid() UID

	GetX() float64
	GetY() float64
	GetZ() float64

	GetPrevX() float64
	GetPrevY() float64
	GetPrevZ() float64
	SetPos(
		x, y, z float64,
		ground bool,
	)
	UpdatePos(
		world Overworld,
		x, y, z float64,
		ground bool,
	) error

	GetYaw() float32
	GetPitch() float32
	SetLook(
		yaw, pitch float32,
		ground bool,
	)
	UpdateLook(
		world Overworld,
		yaw, pitch float32,
		ground bool,
	) error

	IsGround() bool

	IsSneaking() bool
	UpdateSneaking(
		world Overworld,
		sneaking bool,
	) error

	IsSprinting() bool
	UpdateSprinting(
		world Overworld,
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

func (e *entity) GetPrevX() float64 {
	return e.prevX
}

func (e *entity) GetPrevY() float64 {
	return e.prevY
}

func (e *entity) GetPrevZ() float64 {
	return e.prevZ
}

func (e *entity) SetPos(
	x, y, z float64,
	ground bool,
) {
	e.prevX = e.x
	e.prevY = e.y
	e.prevZ = e.z
	e.x = x
	e.y = y
	e.z = z
	e.ground = ground
}

func (e *entity) UpdatePos(
	world Overworld,
	x, y, z float64,
	ground bool,
) error {
	e.SetPos(
		x, y, z,
		ground,
	)

	return nil
}

func (e *entity) GetYaw() float32 {
	return e.yaw
}

func (e *entity) GetPitch() float32 {
	return e.pitch
}

func (e *entity) SetLook(
	yaw, pitch float32,
	ground bool,
) {
	e.yaw = yaw
	e.pitch = pitch
	e.ground = ground
}

func (e *entity) UpdateLook(
	world Overworld,
	yaw, pitch float32,
	ground bool,
) error {
	e.SetLook(
		yaw, pitch,
		ground,
	)

	return nil
}

func (e *entity) IsGround() bool {
	return e.ground
}

func (e *entity) IsSneaking() bool {
	return e.sneaking
}

func (e *entity) UpdateSneaking(
	world Overworld,
	sneaking bool,
) error {
	e.sneaking = sneaking

	return nil
}

func (e *entity) IsSprinting() bool {
	return e.sprinting
}

func (e *entity) UpdateSprinting(
	world Overworld,
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

	EnterChatMessage(
		text string,
	) error
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

func (p *player) EnterChatMessage(
	text string,
) error {

	return nil
}

func (p *player) UpdatePos(
	world Overworld,
	x, y, z float64,
	ground bool,
) error {
	if err := p.living.UpdatePos(
		world,
		x, y, z,
		ground,
	); err != nil {
		return err
	}

	eid := p.eid
	prevX, prevY, prevZ :=
		p.prevX, p.prevY, p.prevZ
	deltaX, deltaY, deltaZ :=
		int16(((x*32)-(prevX*32))*128),
		int16(((y*32)-(prevY*32))*128),
		int16(((z*32)-(prevZ*32))*128)
	if err := world.UpdatePlayerPos(
		eid,
		deltaX, deltaY, deltaZ,
		ground,
	); err != nil {
		return err
	}

	uid := p.uid
	yaw, pitch :=
		p.yaw, p.pitch
	sneaking, sprinting :=
		p.sneaking, p.sprinting
	if err := world.UpdatePlayerChunk(
		eid, uid,
		x, y, z,
		prevX, prevY, prevZ,
		yaw, pitch,
		sneaking, sprinting,
	); err != nil {
		return err
	}

	return nil
}

func (p *player) UpdateLook(
	world Overworld,
	yaw, pitch float32,
	ground bool,
) error {
	if err := p.living.UpdateLook(
		world,
		yaw, pitch,
		ground,
	); err != nil {
		return err
	}

	eid := p.eid
	if err := world.UpdatePlayerLook(
		eid,
		yaw, pitch,
		ground,
	); err != nil {
		return err
	}

	return nil
}

func (p *player) UpdateSneaking(
	world Overworld,
	sneaking bool,
) error {
	if err := p.living.UpdateSneaking(
		world,
		sneaking,
	); err != nil {
		return err
	}

	eid := p.eid
	if err := world.UpdatePlayerSneaking(
		eid,
		sneaking,
	); err != nil {
		return err
	}

	return nil
}

func (p *player) UpdateSprinting(
	world Overworld,
	sprinting bool,
) error {
	if err := p.living.UpdateSprinting(
		world,
		sprinting,
	); err != nil {
		return err
	}

	eid := p.eid
	if err := world.UpdatePlayerSprinting(
		eid,
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
