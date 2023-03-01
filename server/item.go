package server

import "math"

var (
	StickItem = NewItem(
		280, "Stick", "stick", 64, math.MaxInt16,
	)
	BowlItem = NewItem(
		281, "Bowl", "bowl", 64, math.MaxInt16,
	)
)

type Item struct {
	id                int16
	displayName, name string
	stackSize         uint8
	damage            int16
}

func NewItem(
	id int16,
	displayName, name string,
	stackSize uint8,
	damage int16,
) *Item {
	return &Item{
		id,
		displayName, name,
		stackSize,
		damage,
	}
}

func (i *Item) GetID() int16 {
	return i.id
}

func (i *Item) GetDisplayName() string {
	return i.displayName
}

func (i *Item) GetName() string {
	return i.name
}

func (i *Item) GetStackSize() uint8 {
	return i.stackSize
}

func (i *Item) GetDamage() int16 {
	return i.damage
}
