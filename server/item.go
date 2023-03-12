package server

import (
	"errors"
	"fmt"
	"sync"
)

const (
	ItemIDOfWoodenSword            = int16(268)
	ItemMinimumAmountOfWoodenSword = uint8(1)
	ItemMaximumAmountOfWoodenSword = uint8(1)

	ItemIDOfStick            = int16(280)
	ItemMinimumAmountOfStick = uint8(1)
	ItemMaximumAmountOfStick = uint8(64)
)

func ReadItem(
	data *Data,
) (
	Item,
	error,
) {
	id, err := data.ReadInt16()
	if err != nil {
		return nil, err
	}
	if id == -1 {
		return nil, nil
	}

	amount, err := data.ReadUint8()
	if err != nil {
		return nil, err
	}
	if _, err := data.ReadInt16(); err != nil { // not implemented
		return nil, err
	}
	nbt := &ItemNbt{}
	if err := UnmarshalNbt(
		data,
		nbt,
	); err != nil {
		return nil, err
	}
	switch id {
	default:
		return nil, errors.New("there is unregistered id to read Item")
	case ItemIDOfWoodenSword:
		return NewWoodenSwordItem(amount, nbt), nil
	case ItemIDOfStick:
		return NewStickItem(amount, nbt), nil

	}
}

type Item interface {
	Write(
		data *Data,
	) error

	GetID() int16
	GetAmount() uint8
	GetNBT() *ItemNbt
}

type item struct {
	*sync.RWMutex

	id     int16
	amount uint8
	nbt    *ItemNbt
}

func newItem(
	id int16,
	amount uint8,
	nbt *ItemNbt,
) *item {

	return &item{
		new(sync.RWMutex),

		id,
		amount,
		nbt,
	}
}

func (i *item) Write(
	data *Data,
) error {
	if err := data.WriteInt16(
		i.id,
	); err != nil {
		return err
	}

	if err := data.WriteUint8(
		i.amount,
	); err != nil {
		return err
	}

	if err := data.WriteInt16(
		0, // not implemented
	); err != nil {
		return err
	}

	if err := MarshalNbt(
		data,
		i.nbt,
	); err != nil {
		return err
	}

	return nil
}

func (i *item) GetID() int16 {
	return i.id
}

func (i *item) GetAmount() uint8 {
	i.RLock()
	defer i.RUnlock()

	return i.amount
}

func (i *item) GetNBT() *ItemNbt {
	return i.nbt
}

func (i *item) String() string {
	return fmt.Sprintf(
		"{ id: %d, amount: %d, nbt: %s }",
		i.id, i.amount, i.nbt,
	)
}

type StickItem struct {
	*item
}

func NewStickItem(
	amount uint8,
	nbt *ItemNbt,
) *StickItem {
	return &StickItem{
		newItem(
			ItemIDOfStick,
			amount,
			nbt,
		),
	}
}

type WoodenSwordItem struct {
	*item
}

func NewWoodenSwordItem(
	amount uint8,
	nbt *ItemNbt,
) *WoodenSwordItem {
	return &WoodenSwordItem{
		newItem(
			ItemIDOfWoodenSword,
			amount,
			nbt,
		),
	}
}
