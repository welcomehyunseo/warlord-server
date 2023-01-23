package server

import (
	"encoding/binary"
	"math"
)

//var NegativeValueErr = errors.New("NegativeErr: negative value is not allowed")
//var NoMoreByteErr = errors.New("NoMoreByteErr: no more byte to read in Data")
//var NoMoreBytesErr = errors.New("NoMoreBytesErr: no more bytes to read in Data")

const SegmentBits = uint8(0x7F)
const ContinueBit = uint8(0x80)

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
	buf0, buf1 := split(d.buf, 2)
	d.buf = buf1
	v := binary.BigEndian.Uint16(buf0)
	return int16(v)
}

func (d *Data) WriteInt16(
	v int16,
) {
	buf := make([]uint8, 2)
	binary.BigEndian.PutUint16(buf, uint16(v))
	d.buf = concat(d.buf, buf)
}

func (d *Data) ReadUint16() uint16 {
	buf0, buf1 := split(d.buf, 2)
	d.buf = buf1
	v := binary.BigEndian.Uint16(buf0)
	return v
}

func (d *Data) WriteUint16(
	v uint16,
) {
	buf := make([]uint8, 2)
	binary.BigEndian.PutUint16(buf, v)
	d.buf = concat(d.buf, buf)
}

func (d *Data) ReadInt32() int32 {
	buf0, buf1 := split(d.buf, 4)
	d.buf = buf1
	v := binary.BigEndian.Uint32(buf0)
	return int32(v)
}

func (d *Data) WriteInt32(
	v int32,
) {
	buf := make([]uint8, 4)
	binary.BigEndian.PutUint32(buf, uint32(v))
	d.buf = concat(d.buf, buf)
}

func (d *Data) ReadInt64() int64 {
	buf0, buf1 := split(d.buf, 8)
	d.buf = buf1
	v := binary.BigEndian.Uint64(buf0)
	return int64(v)
}

func (d *Data) WriteInt64(
	v int64,
) {
	buf := make([]uint8, 8)
	binary.BigEndian.PutUint64(buf, uint64(v))
	d.buf = concat(d.buf, buf)
}

func (d *Data) ReadFloat32() float32 {
	buf0, buf1 := split(d.buf, 4)
	d.buf = buf1
	bits := binary.BigEndian.Uint32(buf0)
	v := math.Float32frombits(bits)
	return v
}

func (d *Data) WriteFloat32(
	v float32,
) {
	bits := math.Float32bits(v)
	buf := make([]uint8, 4)
	binary.BigEndian.PutUint32(buf, bits)
	d.buf = concat(d.buf, buf)
}

func (d *Data) ReadFloat64() float64 {
	buf0, buf1 := split(d.buf, 8)
	d.buf = buf1
	bits := binary.BigEndian.Uint64(buf0)
	v := math.Float64frombits(bits)
	return v
}

func (d *Data) WriteFloat64(
	v float64,
) {
	bits := math.Float64bits(v)
	buf := make([]uint8, 8)
	binary.BigEndian.PutUint64(buf, bits)
	d.buf = concat(d.buf, buf)
}

//func (d *Data) ReadString() string {
//
//}
//
//func (d *Data) WriteString(
//	v string,
//) {
//
//}

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
