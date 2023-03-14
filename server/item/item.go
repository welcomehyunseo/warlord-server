package item

import (
	"fmt"
	"github.com/welcomehyunseo/warlord-server/server/data"
	"github.com/welcomehyunseo/warlord-server/server/nbt"
	"sync"
)

const (
	WoodenSwordItemID            = int16(268)
	WoodenSwordItemMinimumAmount = uint8(1)
	WoodenSwordItemMaximumAmount = uint8(1)

	StickItemID            = int16(280)
	StickItemMinimumAmount = uint8(1)
	StickItemMaximumAmount = uint8(64)
)

//func ReadItem(
//	dt *data.Data,
//) (
//	Item,
//	error,
//) {
//	id, err := dt.ReadInt16()
//	if err != nil {
//		return nil, err
//	}
//	if id == -1 {
//		return nil, nil
//	}
//
//	amt, err := dt.ReadUint8()
//	if err != nil {
//		return nil, err
//	}
//	if _, err := dt.ReadInt16(); err != nil { // not implemented
//		return nil, err
//	}
//
//	tag := &nbt.ItemNbt{}
//	if err := nbt.UnmarshalNbt(
//		dt,
//		tag,
//	); err != nil {
//		return nil, err
//	}
//
//	switch id {
//	default:
//		return nil, errors.New("there is unregistered id to read Item")
//	case WoodenSwordItemID:
//		return NewWoodenSwordItem(amt, tag), nil
//	case StickItemID:
//		return NewStickItem(amt, tag), nil
//
//	}
//}

type Item interface {
	Write(
		*data.Data,
	) error

	GetID() int16
	GetAmount() uint8
	GetTag() *nbt.ItemNbt
}

type item struct {
	*sync.RWMutex

	id  int16
	amt uint8
	tag *nbt.ItemNbt
}

func newItem(
	id int16,
	amt uint8,
	tag *nbt.ItemNbt,
) *item {

	return &item{
		new(sync.RWMutex),

		id,
		amt,
		tag,
	}
}

func (i *item) Write(
	dt *data.Data,
) error {
	if err := dt.WriteInt16(
		i.id,
	); err != nil {
		return err
	}

	if err := dt.WriteUint8(
		i.amt,
	); err != nil {
		return err
	}

	if err := dt.WriteInt16(
		0, // not implemented
	); err != nil {
		return err
	}

	if err := nbt.MarshalNbt(
		dt,
		i.tag,
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

	return i.amt
}

func (i *item) GetTag() *nbt.ItemNbt {
	return i.tag
}

func (i *item) String() string {
	return fmt.Sprintf(
		"{ id: %d, amt: %d, tag: %s }",
		i.id, i.amt, i.tag,
	)
}

type StickItem struct {
	*item
}

func NewStickItem(
	amt uint8,
	tag *nbt.ItemNbt,
) *StickItem {
	return &StickItem{
		newItem(
			StickItemID,
			amt,
			tag,
		),
	}
}

type WoodenSwordItem struct {
	*item
}

func NewWoodenSwordItem(
	amt uint8,
	tag *nbt.ItemNbt,
) *WoodenSwordItem {
	return &WoodenSwordItem{
		newItem(
			WoodenSwordItemID,
			amt,
			tag,
		),
	}
}
