package server

type Metadata interface {
	Finish() error
	GetBytes() []byte
}

type metadata struct {
	data *Data
}

func newMetadata() *metadata {
	return &metadata{
		data: NewData(),
	}
}

func (md *metadata) writeUint8(
	index uint8, value uint8,
) error {
	data := md.data
	if err := data.WriteUint8(index); err != nil {
		return err
	}
	if err := data.WriteVarInt(0); err != nil {
		return err
	}
	if err := data.WriteUint8(value); err != nil {
		return err
	}

	return nil
}

func (md *metadata) writeVarInt(
	index uint8, value int32,
) error {
	data := md.data
	if err := data.WriteUint8(index); err != nil {
		return err
	}
	if err := data.WriteVarInt(1); err != nil {
		return err
	}
	if err := data.WriteVarInt(value); err != nil {
		return err
	}

	return nil
}

func (md *metadata) writeFloat32(
	index uint8, value float32,
) error {
	data := md.data
	if err := data.WriteUint8(index); err != nil {
		return err
	}
	if err := data.WriteVarInt(2); err != nil {
		return err
	}
	if err := data.WriteFloat32(value); err != nil {
		return err
	}

	return nil
}

func (md *metadata) writeString(
	index uint8, value string,
) error {
	data := md.data
	if err := data.WriteUint8(index); err != nil {
		return err
	}
	if err := data.WriteVarInt(3); err != nil {
		return err
	}
	if err := data.WriteString(value); err != nil {
		return err
	}
	return nil
}

//func (md *metadata) writeChat(
//	index uint8, value *Chat,
//) error {
//	data := md.data
//	if err := data.writeUint8(index); err != nil {
//		return err
//	}
//	if err := data.writeVarInt(4); err != nil {
//		return err
//	}
//	if err := data.WriteChat(value); err != nil {
//		return err
//	}
//	return nil
//}

func (md *metadata) writeBool(
	index uint8, value bool,
) error {
	data := md.data
	if err := data.WriteUint8(index); err != nil {
		return err
	}
	if err := data.WriteVarInt(6); err != nil {
		return err
	}
	if err := data.WriteBool(value); err != nil {
		return err
	}

	return nil
}

func (md *metadata) Finish() error {
	data := md.data
	if err := data.WriteUint8(0xff); err != nil {
		return err
	}
	return nil
}

func (md *metadata) GetBytes() []byte {
	return md.data.GetBytes()
}

type EntityMetadata struct {
	*metadata
}

func NewEntityMetadata() *EntityMetadata {
	return &EntityMetadata{
		metadata: newMetadata(),
	}
}

func (md *EntityMetadata) SetActions(
	sneaking bool,
	sprinting bool,
) error {
	bitmask := uint8(0x00)

	if sneaking == true {
		bitmask |= uint8(0x02)
	}
	if sprinting == true {
		bitmask |= uint8(0x08)
	}

	if err := md.writeUint8(0, bitmask); err != nil {
		return err
	}
	return nil
}

func (md *EntityMetadata) SetAirTick(tick int32) error {
	if err := md.writeVarInt(1, tick); err != nil {
		return err
	}
	return nil
}

func (md *EntityMetadata) SetCustomName(name string) error {
	if err := md.writeString(2, name); err != nil {
		return err
	}
	return nil
}

func (md *EntityMetadata) ShowCustomName() error {
	if err := md.writeBool(3, true); err != nil {
		return err
	}
	return nil
}

func (md *EntityMetadata) HideCustomName() error {
	if err := md.writeBool(3, false); err != nil {
		return err
	}
	return nil
}

func (md *EntityMetadata) StartSilent() error {
	if err := md.writeBool(4, true); err != nil {
		return err
	}
	return nil
}

func (md *EntityMetadata) StopSilent() error {
	if err := md.writeBool(4, false); err != nil {
		return err
	}
	return nil
}

func (md *EntityMetadata) EnableGravity() error {
	if err := md.writeBool(5, true); err != nil {
		return err
	}
	return nil
}

func (md *EntityMetadata) DisableGravity() error {
	if err := md.writeBool(5, false); err != nil {
		return err
	}
	return nil
}
