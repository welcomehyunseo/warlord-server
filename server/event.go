package server

import (
	"fmt"
	"github.com/google/uuid"
)

type ChanForUpdatePlayerPosEvent chan *UpdatePlayerPosEvent
type ChanForConfirmKeepAliveEvent chan *ConfirmKeepAliveEvent
type ChanForAddToPlayerListEvent chan *AddToPlayerListEvent
type ChanForRemoveToPlayerListEvent chan *RemoveToPlayerListEvent

type UpdatePlayerPosEvent struct {
	x float64
	y float64
	z float64
}

func NewUpdatePlayerPosEvent(
	x, y, z float64,
) *UpdatePlayerPosEvent {
	return &UpdatePlayerPosEvent{
		x: x,
		y: y,
		z: z,
	}
}

func (e *UpdatePlayerPosEvent) GetX() float64 {
	return e.x
}

func (e *UpdatePlayerPosEvent) GetY() float64 {
	return e.y
}

func (e *UpdatePlayerPosEvent) GetZ() float64 {
	return e.z
}

func (e *UpdatePlayerPosEvent) String() string {
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

type AddToPlayerListEvent struct {
	uid      uuid.UUID
	username string
}

func NewAddToPlayerListEvent(
	uid uuid.UUID,
	username string,
) *AddToPlayerListEvent {
	return &AddToPlayerListEvent{
		uid:      uid,
		username: username,
	}
}

func (p *AddToPlayerListEvent) GetUUID() uuid.UUID {
	return p.uid
}

func (p *AddToPlayerListEvent) GetUsername() string {
	return p.username
}

func (p *AddToPlayerListEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v, username: %s } ",
		p.uid, p.username,
	)
}

type RemoveToPlayerListEvent struct {
	uid uuid.UUID
}

func NewRemoveToPlayerListEvent(
	uid uuid.UUID,
) *RemoveToPlayerListEvent {
	return &RemoveToPlayerListEvent{
		uid: uid,
	}
}

func (p *RemoveToPlayerListEvent) GetUUID() uuid.UUID {
	return p.uid
}

func (p *RemoveToPlayerListEvent) String() string {
	return fmt.Sprintf(
		"{ uid: %+v } ",
		p.uid,
	)
}

type PlayerListItem struct {
	uid      uuid.UUID
	username string
}

func NewPlayerListItem(
	uid uuid.UUID,
	username string,
) *PlayerListItem {
	return &PlayerListItem{
		uid:      uid,
		username: username,
	}
}

func (i *PlayerListItem) GetUUID() uuid.UUID {
	return i.uid
}

func (i *PlayerListItem) GetUsername() string {
	return i.username
}

func (i *PlayerListItem) String() string {
	return fmt.Sprintf(
		"{ uid: %+v, username: %s } ",
		i.uid, i.username,
	)
}
