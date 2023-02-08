package server

type Metadata struct {
	data *Data
}

func NewMetadata() *Metadata {
	return &Metadata{
		data: NewData(),
	}
}

func (md *Metadata) WriteUint8(
	index uint8, value uint8,
) error {
	data := md.data
	data.WriteUint8(index)
	data.WriteVarInt(0)
	data.WriteUint8(value)

	return nil
}

func (md *Metadata) WriteVarInt(
	index uint8, value int32,
) error {
	data := md.data
	data.WriteUint8(index)
	data.WriteVarInt(1)
	data.WriteVarInt(value)

	return nil
}

func (md *Metadata) WriteFloat32(
	index uint8, value float32,
) error {
	data := md.data
	data.WriteUint8(index)
	data.WriteVarInt(2)
	data.WriteFloat32(value)

	return nil
}

func (md *Metadata) WriteString(
	index uint8, value string,
) error {
	data := md.data
	data.WriteUint8(index)
	data.WriteVarInt(3)
	if err := data.WriteString(value); err != nil {
		return err
	}
	return nil
}

func (md *Metadata) WriteChat(
	index uint8, value *Chat,
) error {
	data := md.data
	data.WriteUint8(index)
	data.WriteVarInt(4)
	if err := data.WriteChat(value); err != nil {
		return err
	}
	return nil
}

func (md *Metadata) WriteOptChat(
	index uint8, value *Chat,
) error {
	data := md.data
	data.WriteUint8(index)
	data.WriteVarInt(5)
	data.WriteBool(true)
	if err := data.WriteChat(value); err != nil {
		return err
	}

	return nil
}

func (md *Metadata) WriteBoolean(
	index uint8, value bool,
) error {
	data := md.data
	data.WriteUint8(index)
	data.WriteVarInt(7)
	data.WriteBool(value)

	return nil
}

func (md *Metadata) Close() {
	data := md.data
	data.WriteUint8(0xff)
}

type EntityMetadata struct {
	*Metadata
}

func NewEntityMetadata() *EntityMetadata {
	return &EntityMetadata{
		Metadata: NewMetadata(),
	}
}

func (md *EntityMetadata) Burning() {

}

func (md *EntityMetadata) Sneaking() {

}

func (md *EntityMetadata) Sprinting() {

}

func (md *EntityMetadata) Swimming() {

}

func (md *EntityMetadata) Invisible() {

}

func (md *EntityMetadata) GlowingEffect() {

}

func (md *EntityMetadata) ElytraFlying() {

}
