package server

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"math"
)

var InsufLenOfBytesError = errors.New("the array of bytes has insufficient length")
var LessThanMinBytesError = errors.New("there are less than minimum number of bytes")
var OverThanMaxBytesError = errors.New("there are over than maximum number of bytes")
var VarIntIsTooBigError = errors.New("VarInt is too big")
var VarLongIsTooBigError = errors.New("VarLong is too big")
var Not26BitIntValueRangeError = errors.New("it is a not 26-bit integer value range")
var Not12BitIntValueRangeError = errors.New("it is a not 12-bit integer value range")

const (
	SegmentBits = uint8(0x7F)
	ContinueBit = uint8(0x80)

	MinBytesNumOfBool     = 1
	MinBytesNumOfInt8     = 1
	MinBytesNumOfUint8    = 1
	MinBytesNumOfInt16    = 2
	MinBytesNumOfUint16   = 2
	MinBytesNumOfInt32    = 4
	MinBytesNumOfInt64    = 8
	MinBytesNumOfFloat32  = 4
	MinBytesNumOfFloat64  = 8
	MinBytesNumOfString   = 0
	MinBytesNumOfChat     = 0
	MinBytesNumOfVarInt   = 4
	MinBytesNumOfVarLong  = 8
	MinBytesNumOfPosition = 8
	MinBytesNumOfAngle    = 1
	MinBytesNumOfUUID     = 16

	MaxBytesNumOfBool     = 1
	MaxBytesNumOfInt8     = 1
	MaxBytesNumOfUint8    = 1
	MaxBytesNumOfInt16    = 2
	MaxBytesNumOfUint16   = 2
	MaxBytesNumOfInt32    = 4
	MaxBytesNumOfInt64    = 8
	MaxBytesNumOfFloat32  = 4
	MaxBytesNumOfFloat64  = 8
	MaxBytesNumOfString   = 32767 * 4
	MaxBytesNumOfChat     = 262144 * 4
	MaxBytesNumOfVarInt   = 4
	MaxBytesNumOfVarLong  = 8
	MaxBytesNumOfPosition = 8
	MaxBytesNumOfAngle    = 1
	MaxBytesNumOfUUID     = 16

	MinNumOf26BitInt = -33554432
	MaxNumOf26BitInt = 33554431
	MinNumOf12BitInt = -2048
	MaxNumOf12BitInt = 2047

	MinValueOfUint16 = 0
	MaxValueOfUint16 = 65535
)

type Chat struct {
	Text          string  `json:"text,omitempty"`
	Bold          bool    `json:"bold,omitempty"`
	Italic        bool    `json:"italic,omitempty"`
	Underlined    bool    `json:"underlined,omitempty"`
	Strikethrough bool    `json:"strikethrough,omitempty"`
	Obfuscated    bool    `json:"obfuscated,omitempty"`
	Font          string  `json:"font,omitempty"`
	Color         string  `json:"color,omitempty"`
	Insertion     string  `json:"insertion,omitempty"`
	Extra         []*Chat `json:"extra,omitempty"`
}

func (c *Chat) String() string {
	return fmt.Sprintf(
		"{ "+
			"text: %s, "+
			"bold: %v, "+
			"italic: %v, "+
			"underlined: %v, "+
			"strikethrough: %v, "+
			"obfuscated: %v, "+
			"font: %s, "+
			"color: %s, "+
			"insertion: %s, "+
			"extra: %v "+
			"}",
		c.Text,
		c.Bold,
		c.Italic,
		c.Underlined,
		c.Strikethrough,
		c.Obfuscated,
		c.Font,
		c.Color,
		c.Insertion,
		c.Extra,
	)
}

func compare(
	arr0 []uint8,
	arr1 []uint8,
) bool {
	l0 := len(arr0)
	l1 := len(arr1)
	if l0 != l1 {
		return false
	}
	for i := 0; i < l0; i++ {
		v0 := arr0[i]
		v1 := arr1[i]
		if v0 != v1 {
			return false
		}
	}
	return true
}

func split(
	arr []byte,
	n int,
) (
	[]byte,
	[]byte,
	error,
) {
	l := len(arr)
	if l < n {
		return nil, nil, InsufLenOfBytesError
	}

	return arr[0:n], arr[n:l], nil
}

func shift(
	arr []byte,
) (
	byte,
	[]byte,
	error,
) {
	arr0, arr1, err := split(arr, 1)
	if err != nil {
		return 0x00, nil, err
	}

	b := arr0[0]
	return b, arr1, nil
}

func concat(
	arr0 []byte,
	arr1 []byte,
) []byte {
	l0 := len(arr0)
	l1 := len(arr1)
	arr2 := make([]byte, l0+l1)

	copy(arr2[0:], arr0)
	copy(arr2[l0:], arr1)

	return arr2
}

func push(
	arr0 []byte,
	b byte,
) []byte {
	l0 := len(arr0)
	arr1 := make([]byte, l0+1)

	copy(arr1, arr0)
	arr1[l0] = b

	return arr1
}

type Data struct {
	arr []byte
}

func NewData() *Data {
	return &Data{
		make([]byte, 0),
	}
}

func NewDataWithBytes(
	arr []byte,
) *Data {
	return &Data{
		arr,
	}
}

func (d *Data) ReadBool() (
	bool,
	error,
) {
	v, arr, err := shift(d.arr)
	if err != nil {
		return false, err
	}
	d.arr = arr
	return v == 0x1, nil
}

func (d *Data) WriteBool(
	v bool,
) error {
	if v == true {
		d.arr = push(d.arr, 0x01)
	} else {
		d.arr = push(d.arr, 0x00)
	}
	return nil
}

func (d *Data) ReadInt8() (
	int8,
	error,
) {
	v, arr, err := shift(d.arr)
	if err != nil {
		return 0, err
	}
	d.arr = arr

	return int8(v), nil
}

func (d *Data) WriteInt8(
	v int8,
) error {
	d.arr = push(d.arr, byte(v))

	return nil
}

func (d *Data) ReadBytes(n int) (
	[]byte,
	error,
) {
	b0, b1, err := split(d.arr, n)
	if err != nil {
		return nil, err
	}
	d.arr = b1
	return b0, nil
}

func (d *Data) WriteBytes(b []byte) error {
	d.arr = concat(d.arr, b)
	return nil
}

func (d *Data) ReadUint8() (
	byte,
	error,
) {
	v, arr, err := shift(d.arr)
	if err != nil {
		return 0, err
	}
	d.arr = arr

	return v, nil
}

func (d *Data) WriteUint8(
	v byte,
) error {
	d.arr = push(d.arr, v)
	return nil
}

func (d *Data) ReadInt16() (
	int16,
	error,
) {
	arr0, arr1, err := split(d.arr, MaxBytesNumOfInt16)
	if err != nil {
		return 0, err
	}
	d.arr = arr1
	v := binary.BigEndian.Uint16(arr0)
	return int16(v), nil
}

func (d *Data) WriteInt16(
	v int16,
) error {
	arr := make([]uint8, MaxBytesNumOfInt16)
	binary.BigEndian.PutUint16(arr, uint16(v))
	d.arr = concat(d.arr, arr)

	return nil
}

func (d *Data) ReadUint16() (
	uint16,
	error,
) {
	arr0, arr1, err := split(d.arr, MaxBytesNumOfUint16)
	if err != nil {
		return 0, err
	}
	d.arr = arr1
	v := binary.BigEndian.Uint16(arr0)
	return v, nil
}

func (d *Data) WriteUint16(
	v uint16,
) error {
	arr := make([]uint8, MaxBytesNumOfUint16)
	binary.BigEndian.PutUint16(arr, v)
	d.arr = concat(d.arr, arr)

	return nil
}

func (d *Data) ReadInt32() (
	int32,
	error,
) {
	arr0, arr1, err := split(d.arr, MaxBytesNumOfInt32)
	if err != nil {
		return 0, err
	}
	d.arr = arr1
	v := binary.BigEndian.Uint32(arr0)
	return int32(v), nil
}

func (d *Data) WriteInt32(
	v int32,
) error {
	arr := make([]uint8, MaxBytesNumOfInt32)
	binary.BigEndian.PutUint32(arr, uint32(v))
	d.arr = concat(d.arr, arr)

	return nil
}

func (d *Data) ReadInt64() (
	int64,
	error,
) {
	arr0, arr1, err := split(d.arr, MaxBytesNumOfInt64)
	if err != nil {
		return 0, err
	}
	d.arr = arr1
	v := binary.BigEndian.Uint64(arr0)
	return int64(v), nil
}

func (d *Data) WriteInt64(
	v int64,
) error {
	arr := make([]uint8, MaxBytesNumOfInt64)
	binary.BigEndian.PutUint64(arr, uint64(v))
	d.arr = concat(d.arr, arr)

	return nil
}

func (d *Data) ReadFloat32() (
	float32,
	error,
) {
	arr0, arr1, err := split(d.arr, MaxBytesNumOfFloat32)
	if err != nil {
		return 0, err
	}
	d.arr = arr1
	bits := binary.BigEndian.Uint32(arr0)
	v := math.Float32frombits(bits)
	return v, err
}

func (d *Data) WriteFloat32(
	v float32,
) error {
	bits := math.Float32bits(v)
	arr := make([]uint8, MaxBytesNumOfFloat32)
	binary.BigEndian.PutUint32(arr, bits)
	d.arr = concat(d.arr, arr)

	return nil
}

func (d *Data) ReadFloat64() (
	float64,
	error,
) {
	arr0, arr1, err := split(d.arr, MaxBytesNumOfFloat64)
	if err != nil {
		return 0, err
	}
	d.arr = arr1
	bits := binary.BigEndian.Uint64(arr0)
	v := math.Float64frombits(bits)
	return v, nil
}

func (d *Data) WriteFloat64(
	v float64,
) error {
	bits := math.Float64bits(v)
	arr := make([]uint8, MaxBytesNumOfFloat64)
	binary.BigEndian.PutUint64(arr, bits)
	d.arr = concat(d.arr, arr)
	return nil
}

func (d *Data) ReadString() (
	string,
	error,
) {
	l, err := d.ReadVarInt()
	if err != nil {
		return "", err
	}
	if l < MinBytesNumOfString {
		return "", LessThanMinBytesError
	}
	if MaxBytesNumOfString < l {
		return "", OverThanMaxBytesError
	}

	arr0, arr1, err := split(d.arr, int(l))
	if err != nil {
		return "", err
	}
	d.arr = arr1
	s := string(arr0)
	return s, nil
}

func (d *Data) WriteString(
	v string,
) error {
	arr := []byte(v)

	length := len(arr)
	if length < MinBytesNumOfString {
		return LessThanMinBytesError
	}
	if MaxBytesNumOfString < length {
		return OverThanMaxBytesError
	}

	if err := d.WriteVarInt(int32(length)); err != nil {
		return err
	}
	d.arr = concat(d.arr, arr)

	return nil
}

func (d *Data) WriteChat(
	v *Chat,
) error {
	arr, err := json.Marshal(v)
	if err != nil {
		return err
	}

	length := len(arr)
	if length < MinBytesNumOfChat {
		return LessThanMinBytesError
	}
	if MaxBytesNumOfChat < length {
		return OverThanMaxBytesError
	}

	if err := d.WriteVarInt(int32(length)); err != nil {
		return err
	}
	d.arr = concat(d.arr, arr)
	return nil
}

func (d *Data) ReadVarInt() (
	int32,
	error,
) {
	v := int32(0)
	position := uint8(0)

	for {
		b, arr, err := shift(d.arr)
		if err != nil {
			return 0, err
		}
		d.arr = arr
		v |= int32(b&SegmentBits) << position

		if (b & ContinueBit) == 0 {
			break
		}

		position += 7

		if position >= 32 {
			return 0, VarIntIsTooBigError
		}
	}

	return v, nil
}

func (d *Data) WriteVarInt(
	v int32,
) error {
	v0 := uint32(v)
	for {
		if (v0 & ^uint32(SegmentBits)) == 0 {
			b := uint8(v0)
			d.arr = push(d.arr, b)
			break
		}

		b := uint8(v0&uint32(SegmentBits)) | ContinueBit
		d.arr = push(d.arr, b)

		v0 >>= 7
	}

	return nil
}

func (d *Data) ReadVarLong() (
	int64,
	error,
) {
	v := int64(0)
	position := uint8(0)

	for {
		b, arr, err := shift(d.arr)
		if err != nil {
			return 0, err
		}
		d.arr = arr
		v |= int64(b&SegmentBits) << position

		if (b & ContinueBit) == 0 {
			break
		}

		position += 7
		if position >= 64 {
			return 0, VarLongIsTooBigError
		}
	}

	return v, nil
}

func (d *Data) WriteVarLong(
	v int64,
) error {
	v0 := uint64(v)
	for {
		if (v0 & ^uint64(SegmentBits)) == 0 {
			b := uint8(v0)
			d.arr = push(d.arr, b)
			break
		}

		b := uint8(v0&uint64(SegmentBits)) | ContinueBit
		d.arr = push(d.arr, b)

		v0 >>= 7
	}

	return nil
}

func (d *Data) WriteMetadata(
	v Metadata,
) error {
	if err := v.Finish(); err != nil {
		return err
	}

	arr := v.GetBytes()
	d.arr = concat(d.arr, arr)
	return nil
}

func (d *Data) ReadPosition() (
	int, int, int,
	error,
) {
	v, err := d.ReadInt64()
	if err != nil {
		return 0, 0, 0, err
	}

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

	return x, y, z, nil
}

// WritePosition encodes an integer position into the current Data.
// The position is consisted of a value x as a signed 26-bit integer, a value z as a signed 26-bit integer,
// and a value y as a signed 12-bit integer with two's complement and big-endian.
func (d *Data) WritePosition(
	x, y, z int,
) error {
	if x < MinNumOf26BitInt || MaxNumOf26BitInt < x ||
		z < MinNumOf26BitInt || MaxNumOf26BitInt < z {
		return Not26BitIntValueRangeError
	}
	if y < MinNumOf12BitInt || MaxNumOf12BitInt < y {
		return Not12BitIntValueRangeError
	}

	arr := make([]uint8, MaxBytesNumOfPosition)
	v := uint64(x&0x3FFFFFF)<<38 | uint64((z&0x3FFFFFF)<<12) | uint64(y&0xFFF)
	for i := 7; i >= 0; i-- {
		arr[i] = uint8(v)
		v >>= 8
	}
	d.arr = concat(d.arr, arr)
	return nil
}

func frem(x, y float32) float32 {
	return x - y*float32(math.Floor(float64(x/y)))
}

func (d *Data) ReadAngle() (
	float32,
	error,
) {
	v, arr, err := shift(d.arr)
	if err != nil {
		return 0, err
	}
	d.arr = arr
	v0 := (360 * float32(v)) / math.MaxUint8
	v1 := frem(v0, 360)
	return v1, nil
}

func (d *Data) WriteAngle(
	v float32,
) error {
	v0 := frem(v, 360)
	b := uint8((math.MaxUint8 * v0) / 360)
	d.arr = push(d.arr, b)

	return nil
}

func (d *Data) ReadUUID() (
	uuid.UUID,
	error,
) {
	arr0, arr1, err := split(d.arr, MaxBytesNumOfUUID)
	if err != nil {
		return uuid.Nil, err
	}
	d.arr = arr1
	v, err := uuid.FromBytes(arr0)
	if err != nil {
		return uuid.Nil, err
	}

	return v, nil
}

func (d *Data) WriteUUID(
	v uuid.UUID,
) error {
	arr := make([]byte, MaxBytesNumOfUUID)
	for i := 0; i < MaxBytesNumOfUUID; i++ {
		arr[i] = v[i]
	}
	d.arr = concat(d.arr, arr)

	return nil
}

func (d *Data) GetBytes() []byte {
	return d.arr
}

func (d *Data) GetLength() int {
	return len(d.arr)
}

func (d *Data) Concat(
	v *Data,
) error {
	d.arr = concat(d.arr, v.arr)

	return nil
}
