package server

import (
	"encoding/binary"
	"github.com/google/uuid"
	"math"
)

//var NegativeValueErr = errors.New("NegativeErr: negative value is not allowed")
//var NoMoreByteErr = errors.New("NoMoreByteErr: no more byte to read in Data")
//var NoMoreBytesErr = errors.New("NoMoreBytesErr: no more bytes to read in Data")

const (
	SegmentBits = uint8(0x7F)
	ContinueBit = uint8(0x80)

	BytesNumOfBool     = 1
	BytesNumOfInt8     = 1
	BytesNumOfUint8    = 1
	BytesNumOfInt16    = 2
	BytesNumOfUint16   = 2
	BytesNumOfInt32    = 4
	BytesNumOfInt64    = 8
	BytesNumOfFloat32  = 4
	BytesNumOfFloat64  = 8
	BytesNumOfVarInt   = 4
	BytesNumOfVarLong  = 8
	BytesNumOfPosition = 8
	BytesNumOfAngle    = 1
	BytesNumOfUUID     = 16
)

func compare(
	buf0 []uint8,
	buf1 []uint8,
) bool {
	l0 := len(buf0)
	l1 := len(buf1)
	if l0 != l1 {
		return false
	}
	for i := 0; i < l0; i++ {
		v0 := buf0[i]
		v1 := buf1[i]
		if v0 != v1 {
			return false
		}
	}
	return true
}

func split(
	buf []uint8,
	n int,
) ([]uint8, []uint8) {
	l := len(buf)
	return buf[0:n], buf[n:l]
}

func shift(
	buf []uint8,
) (uint8, []uint8) {
	buf0, buf1 := split(buf, 1)
	b := buf0[0]
	return b, buf1
}

func concat(
	buf0 []uint8,
	buf1 []uint8,
) []uint8 {
	l0 := len(buf0)
	l1 := len(buf1)
	buf2 := make([]uint8, l0+l1)
	for i := 0; i < l0; i++ {
		buf2[i] = buf0[i]
	}
	for i := 0; i < l1; i++ {
		buf2[i+l0] = buf1[i]
	}
	return buf2
}

func push(
	buf []uint8,
	v uint8,
) []uint8 {
	return concat(buf, []uint8{v})
}

type Data struct {
	buf []uint8
}

func NewData(bytes ...uint8) *Data {
	if bytes == nil {
		bytes = make([]uint8, 0)
	}

	return &Data{
		buf: bytes,
	}
}

func (d *Data) ReadBool() bool {
	v, buf := shift(d.buf)
	d.buf = buf
	return v == 0x1
}

func (d *Data) WriteBool(
	v bool,
) {
	if v == true {
		d.buf = push(d.buf, 0x01)
	} else {
		d.buf = push(d.buf, 0x00)
	}
}

func (d *Data) ReadInt8() int8 {
	v, buf := shift(d.buf)
	d.buf = buf

	return int8(v)
}

func (d *Data) WriteInt8(
	v int8,
) {
	d.buf = push(d.buf, uint8(v))
}

func (d *Data) ReadUint8() uint8 {
	v, buf := shift(d.buf)
	d.buf = buf

	return v
}

func (d *Data) WriteUint8(
	v uint8,
) {
	d.buf = push(d.buf, v)
}

func (d *Data) ReadInt16() int16 {
	buf0, buf1 := split(d.buf, BytesNumOfInt16)
	d.buf = buf1
	v := binary.BigEndian.Uint16(buf0)
	return int16(v)
}

func (d *Data) WriteInt16(
	v int16,
) {
	buf := make([]uint8, BytesNumOfInt16)
	binary.BigEndian.PutUint16(buf, uint16(v))
	d.buf = concat(d.buf, buf)
}

func (d *Data) ReadUint16() uint16 {
	buf0, buf1 := split(d.buf, BytesNumOfUint16)
	d.buf = buf1
	v := binary.BigEndian.Uint16(buf0)
	return v
}

func (d *Data) WriteUint16(
	v uint16,
) {
	buf := make([]uint8, BytesNumOfUint16)
	binary.BigEndian.PutUint16(buf, v)
	d.buf = concat(d.buf, buf)
}

func (d *Data) ReadInt32() int32 {
	buf0, buf1 := split(d.buf, BytesNumOfInt32)
	d.buf = buf1
	v := binary.BigEndian.Uint32(buf0)
	return int32(v)
}

func (d *Data) WriteInt32(
	v int32,
) {
	buf := make([]uint8, BytesNumOfInt32)
	binary.BigEndian.PutUint32(buf, uint32(v))
	d.buf = concat(d.buf, buf)
}

func (d *Data) ReadInt64() int64 {
	buf0, buf1 := split(d.buf, BytesNumOfInt64)
	d.buf = buf1
	v := binary.BigEndian.Uint64(buf0)
	return int64(v)
}

func (d *Data) WriteInt64(
	v int64,
) {
	buf := make([]uint8, BytesNumOfInt64)
	binary.BigEndian.PutUint64(buf, uint64(v))
	d.buf = concat(d.buf, buf)
}

func (d *Data) ReadFloat32() float32 {
	buf0, buf1 := split(d.buf, BytesNumOfFloat32)
	d.buf = buf1
	bits := binary.BigEndian.Uint32(buf0)
	v := math.Float32frombits(bits)
	return v
}

func (d *Data) WriteFloat32(
	v float32,
) {
	bits := math.Float32bits(v)
	buf := make([]uint8, BytesNumOfFloat32)
	binary.BigEndian.PutUint32(buf, bits)
	d.buf = concat(d.buf, buf)
}

func (d *Data) ReadFloat64() float64 {
	buf0, buf1 := split(d.buf, BytesNumOfFloat64)
	d.buf = buf1
	bits := binary.BigEndian.Uint64(buf0)
	v := math.Float64frombits(bits)
	return v
}

func (d *Data) WriteFloat64(
	v float64,
) {
	bits := math.Float64bits(v)
	buf := make([]uint8, BytesNumOfFloat64)
	binary.BigEndian.PutUint64(buf, bits)
	d.buf = concat(d.buf, buf)
}

func (d *Data) ReadString() string {
	l := d.ReadVarInt()
	buf0, buf1 := split(d.buf, int(l))
	d.buf = buf1
	s := string(buf0)
	return s
}

func (d *Data) WriteString(
	v string,
) {
	buf := []uint8(v)
	l := len(buf)
	d.WriteVarInt(int32(l))
	d.buf = concat(d.buf, buf)
}

func (d *Data) ReadVarInt() int32 {
	v := int32(0)
	position := uint8(0)

	for {
		b, buf := shift(d.buf)
		d.buf = buf
		v |= int32(b&SegmentBits) << position

		if (b & ContinueBit) == 0 {
			break
		}

		position += 7
	}

	return v
}

func (d *Data) WriteVarInt(
	v int32,
) {
	v0 := uint32(v)
	for {
		if (v0 & ^uint32(SegmentBits)) == 0 {
			b := uint8(v0)
			d.buf = push(d.buf, b)
			break
		}

		b := uint8(v0&uint32(SegmentBits)) | ContinueBit
		d.buf = push(d.buf, b)

		v0 >>= 7
	}
}

func (d *Data) ReadVarLong() int64 {
	v := int64(0)
	position := uint8(0)

	for {
		b, buf := shift(d.buf)
		d.buf = buf
		v |= int64(b&SegmentBits) << position

		if (b & ContinueBit) == 0 {
			break
		}

		position += 7
	}

	return v
}

func (d *Data) WriteVarLong(
	v int64,
) {
	v0 := uint64(v)
	for {
		if (v0 & ^uint64(SegmentBits)) == 0 {
			b := uint8(v0)
			d.buf = push(d.buf, b)
			break
		}

		b := uint8(v0&uint64(SegmentBits)) | ContinueBit
		d.buf = push(d.buf, b)

		v0 >>= 7
	}
}

func (d *Data) ReadPosition() (int, int, int) {
	v := d.ReadInt64()

	x := int(v >> 38)
	y := int(v & 0xFFF)
	z := int(v << 26 >> 38)

	if x >= 1<<25 {
		x -= 1 << 26
	}
	if y >= 1<<11 {
		y -= 1 << 12
	}
	if z >= 1<<25 {
		z -= 1 << 26
	}

	return x, y, z
}

// WritePosition encodes an integer position into the current Data.
// The position is consisted of a value x as a signed 26-bit integer, a value z as a signed 26-bit integer,
// and a value y as a signed 12-bit integer with two's complement and big-endian.
func (d *Data) WritePosition(
	x, y, z int,
) {
	buf := make([]uint8, BytesNumOfPosition)
	v := uint64(x&0x3FFFFFF)<<38 | uint64((z&0x3FFFFFF)<<12) | uint64(y&0xFFF)
	for i := 7; i >= 0; i-- {
		buf[i] = uint8(v)
		v >>= 8
	}
	d.buf = concat(d.buf, buf)
}

func frem(x, y float64) float64 {
	return x - y*math.Floor(x/y)
}

func (d *Data) ReadAngle() float64 {
	v, buf := shift(d.buf)
	d.buf = buf
	v0 := (360 * float64(v)) / math.MaxUint8
	v1 := frem(v0, 360)
	return v1
}

func (d *Data) WriteAngle(
	v float64,
) {
	v0 := frem(v, 360)
	b := uint8((math.MaxUint8 * v0) / 360)
	d.buf = push(d.buf, b)
}

func (d *Data) ReadUUID() uuid.UUID {
	buf0, buf1 := split(d.buf, BytesNumOfUUID)
	d.buf = buf1
	v, _ := uuid.FromBytes(buf0)

	return v
}

func (d *Data) WriteUUID(
	v uuid.UUID,
) {
	buf := make([]uint8, BytesNumOfUUID)
	for i := 0; i < BytesNumOfUUID; i++ {
		buf[i] = v[i]
	}
	d.buf = concat(d.buf, buf)
}

func (d *Data) ReadBuf(n int) []uint8 {
	b0, b1 := split(d.buf, n)
	d.buf = b1
	return b0
}

func (d *Data) WriteBuf(buf []uint8) {
	d.buf = concat(d.buf, buf)
}

func (d *Data) Write(v *Data) {
	d.buf = concat(d.buf, v.buf)
}

func (d *Data) GetLength() int {
	return len(d.buf)
}

func (d *Data) GetBuf() []uint8 {
	return d.buf
}
