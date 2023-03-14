package metadata

import (
	"errors"
	"github.com/welcomehyunseo/warlord-server/server/data"
)

const (
	typeIDOfByte    = int32(0)
	typeIDOfVarInt  = int32(1)
	typeIDOfFloat32 = int32(2)
	typeIDOfString  = int32(3)
	typeIDOfChat    = int32(4)
	typeIDOfItem    = int32(5)
	typeIDOfBool    = int32(6)
)

func writeUint8(
	dt *data.Data,
	i uint8, v uint8,
) error {
	if err := dt.WriteUint8(
		i,
	); err != nil {
		return err
	}

	if err := dt.WriteVarInt(
		typeIDOfByte,
	); err != nil {
		return err
	}

	if err := dt.WriteUint8(
		v,
	); err != nil {
		return err
	}

	return nil
}

func writeVarInt(
	dt *data.Data,
	i uint8, v int32,
) error {
	if err := dt.WriteUint8(
		i,
	); err != nil {
		return err
	}

	if err := dt.WriteVarInt(
		typeIDOfVarInt,
	); err != nil {
		return err
	}

	if err := dt.WriteVarInt(
		v,
	); err != nil {
		return err
	}

	return nil
}

func writeFloat32(
	dt *data.Data,
	i uint8, v float32,
) error {
	if err := dt.WriteUint8(
		i,
	); err != nil {
		return err
	}

	if err := dt.WriteVarInt(
		typeIDOfFloat32,
	); err != nil {
		return err
	}

	if err := dt.WriteFloat32(
		v,
	); err != nil {
		return err
	}

	return nil
}

func writeString(
	dt *data.Data,
	i uint8, v string,
) error {
	if err := dt.WriteUint8(
		i,
	); err != nil {
		return err
	}

	if err := dt.WriteVarInt(
		typeIDOfString,
	); err != nil {
		return err
	}

	if err := dt.WriteString(
		v,
	); err != nil {
		return err
	}
	return nil
}

func writeBool(
	dt *data.Data,
	i uint8, v bool,
) error {
	if err := dt.WriteUint8(
		i,
	); err != nil {
		return err
	}

	if err := dt.WriteVarInt(
		typeIDOfBool,
	); err != nil {
		return err
	}

	if err := dt.WriteBool(
		v,
	); err != nil {
		return err
	}

	return nil
}

type metadata struct {
	m0 map[uint8]uint8
	m1 map[uint8]int32
	m2 map[uint8]float32
	m3 map[uint8]string
	m6 map[uint8]bool
}

func newMetadata() *metadata {
	return &metadata{
		make(map[uint8]uint8),
		make(map[uint8]int32),
		make(map[uint8]float32),
		make(map[uint8]string),
		make(map[uint8]bool),
	}
}

func (md *metadata) Write(
	dt *data.Data,
) error {

	for i, v := range md.m0 {
		if err := writeUint8(
			dt,
			i, v,
		); err != nil {
			return err
		}
	}

	for i, v := range md.m1 {
		if err := writeVarInt(
			dt,
			i, v,
		); err != nil {
			return err
		}
	}

	for i, v := range md.m2 {
		if err := writeFloat32(
			dt,
			i, v,
		); err != nil {
			return err
		}
	}

	for i, v := range md.m3 {
		if err := writeString(
			dt,
			i, v,
		); err != nil {
			return err
		}
	}

	for i, v := range md.m6 {
		if err := writeBool(
			dt,
			i, v,
		); err != nil {
			return err
		}
	}

	if err := dt.WriteUint8(
		0xff,
	); err != nil {
		return err
	}

	return nil
}

type entityMetadata struct {
	*metadata
}

func newEntityMetadata() *entityMetadata {
	return &entityMetadata{
		metadata: newMetadata(),
	}
}

func (md *entityMetadata) SetActions(
	sneaking bool,
	sprinting bool,
) error {
	i := uint8(0)
	m := md.m0
	if _, has := m[i]; has == true {
		return errors.New("it is already existed field to set actions of entity metadata")
	}

	v := uint8(0x00)
	if sneaking == true {
		v |= uint8(0x02)
	}
	if sprinting == true {
		v |= uint8(0x08)
	}

	m[i] = v
	return nil
}

func (md *entityMetadata) SetAirTicks(
	v int32,
) error {
	i := uint8(1)
	m := md.m1
	if _, has := m[i]; has == true {
		return errors.New("it is already existed field to set air ticks of entity metadata")
	}

	m[i] = v
	return nil
}

func (md *entityMetadata) SetCustomName(
	v string,
) error {
	i := uint8(2)
	m := md.m3
	if _, has := m[i]; has == true {
		return errors.New("it is already existed field to set custom name of entity metadata")
	}

	m[i] = v
	return nil
}

func (md *entityMetadata) ShowCustomName(
	v bool,
) error {
	i := uint8(3)
	m := md.m6
	if _, has := m[i]; has == true {
		return errors.New("it is already existed field to show custom name of entity metadata")
	}

	m[i] = v
	return nil
}

func (md *entityMetadata) SetSilent(
	v bool,
) error {
	i := uint8(4)
	m := md.m6
	if _, has := m[i]; has == true {
		return errors.New("it is already existed field to set silent of entity metadata")
	}

	m[i] = v
	return nil
}

func (md *entityMetadata) SetGravity(
	v bool,
) error {
	i := uint8(5)
	m := md.m6
	if _, has := m[i]; has == true {
		return errors.New("it is already existed field to set gravity of entity metadata")
	}

	m[i] = v
	return nil
}

type livingMetadata struct {
	*entityMetadata
}

func newLivingMetadata() *livingMetadata {
	return &livingMetadata{
		newEntityMetadata(),
	}
}

type PlayerMetadata struct {
	*livingMetadata
}

func NewPlayerMetadata() *PlayerMetadata {
	return &PlayerMetadata{
		newLivingMetadata(),
	}
}
