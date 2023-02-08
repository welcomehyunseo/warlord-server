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
) {
	data := md.data
	data.WriteUint8(index)
	data.WriteVarInt(0)
	data.WriteUint8(value)
}

func (md *Metadata) WriteVarInt(
	index uint8, value int32,
) {
	data := md.data
	data.WriteUint8(index)
	data.WriteVarInt(1)
	data.WriteVarInt(value)
}

func (md *Metadata) WriteFloat32(
	index uint8, value float32,
) {
	data := md.data
	data.WriteUint8(index)
	data.WriteVarInt(2)
	data.WriteFloat32(value)
}

func (md *Metadata) WriteString(
	index uint8, value string,
) {
	data := md.data
	data.WriteUint8(index)
	data.WriteVarInt(3)
	data.WriteString(value)
}

//
//func (md *Metadata) WriteChat(
//	index uint8, value ,
//) {
//	data := md.data
//	data.WriteUint8(index)
//	data.WriteVarInt(4)
//	data.Write
//}

//func (md *Metadata) WriteOptChat(
//	index uint8, value ,
//) {
//	data := md.data
//	data.WriteUint8(index)
//	data.WriteVarInt(5)
//	data.Write
//}

//func (md *Metadata) WriteSlot(
//	index uint8, value ,
//) {
//	data := md.data
//	data.WriteUint8(index)
//	data.WriteVarInt(6)
//	data.Write
//}

func (md *Metadata) WriteBoolean(
	index uint8, value bool,
) {
	data := md.data
	data.WriteUint8(index)
	data.WriteVarInt(7)
	data.WriteBool(value)
}

//func (md *Metadata) WriteRotation(
//	index uint8, x float32, y float32, z float32,
//) {
//	data := md.data
//	data.WriteUint8(index)
//	data.WriteVarInt(8)
//	data.Write
//}

//func (md *Metadata) WritePosition(
//	index uint8, ,
//) {
//	data := md.data
//	data.WriteUint8(index)
//	data.WriteVarInt(9)
//	data.Write
//}

//func (md *Metadata) WriteOptPosition(
//	index uint8, ,
//) {
//	data := md.data
//	data.WriteUint8(index)
//	data.WriteVarInt(9)
//	data.Write
//}

func (md *Metadata) Close() {
	data := md.data
	data.WriteUint8(0xff)
}

type EntityMetadata struct {
	*Metadata
}

//
//func NewEntityMetadata(
//	burning bool,
//	sneaking bool,
//	sprinting bool,
//	swimming bool,
//	invisible bool,
//	glowingEffect bool,
//	elytraFlying bool,
//) *EntityMetadata {
//
//}
