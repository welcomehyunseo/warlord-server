package server

//
//import (
//	"github.com/google/uuid"
//	"math"
//	"testing"
//)
//
//func TestReadBool(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x01},
//		{0x00},
//	}
//	yValues := []bool{
//		true,
//		false,
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadBool()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteBool(
//	t *testing.T,
//) {
//	xValues := []bool{
//		true,
//		false,
//	}
//	yValues := [][]uint8{
//		{0x01},
//		{0x00},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteBool(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestReadInt8(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x7F},
//		{0x00},
//		{0xff},
//		{0x80},
//	}
//	yValues := []int8{
//		math.MaxInt8,
//		0,
//		-1,
//		math.MinInt8,
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadInt8()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteInt8(
//	t *testing.T,
//) {
//	xValues := []int8{
//		math.MaxInt8,
//		0,
//		-1,
//		math.MinInt8,
//	}
//	yValues := [][]uint8{
//		{0x7F},
//		{0x00},
//		{0xff},
//		{0x80},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteInt8(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestReadUint8(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x00},
//		{0xff},
//	}
//	yValues := []uint8{
//		0,
//		math.MaxUint8,
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadUint8()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteUint8(
//	t *testing.T,
//) {
//	xValues := []uint8{
//		0,
//		math.MaxUint8,
//	}
//	yValues := [][]uint8{
//		{0x00},
//		{0xff},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteUint8(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestReadInt16(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x7F, 0xFF},
//		{0x00, 0x00},
//		{0xff, 0xff},
//		{0x80, 0x00},
//	}
//	yValues := []int16{
//		math.MaxInt16,
//		0,
//		-1,
//		math.MinInt16,
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadInt16()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteInt16(
//	t *testing.T,
//) {
//	xValues := []int16{
//		math.MaxInt16,
//		0,
//		-1,
//		math.MinInt16,
//	}
//	yValues := [][]uint8{
//		{0x7F, 0xFF},
//		{0x00, 0x00},
//		{0xff, 0xff},
//		{0x80, 0x00},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteInt16(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestReadUint16(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x00, 0x00},
//		{0xff, 0xff},
//	}
//	yValues := []uint16{
//		0,
//		math.MaxUint16,
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadUint16()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteUit16(
//	t *testing.T,
//) {
//	xValues := []uint16{
//		0,
//		math.MaxUint16,
//	}
//	yValues := [][]uint8{
//		{0x00, 0x00},
//		{0xff, 0xff},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteUint16(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestReadInt32(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x7F, 0xFF, 0xFF, 0xFF},
//		{0x00, 0x00, 0x00, 0x00},
//		{0xff, 0xff, 0xff, 0xff},
//		{0x80, 0x00, 0x00, 0x00},
//	}
//	yValues := []int32{
//		math.MaxInt32,
//		0,
//		-1,
//		math.MinInt32,
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadInt32()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteInt32(
//	t *testing.T,
//) {
//	xValues := []int32{
//		math.MaxInt32,
//		0,
//		-1,
//		math.MinInt32,
//	}
//	yValues := [][]uint8{
//		{0x7F, 0xFF, 0xFF, 0xFF},
//		{0x00, 0x00, 0x00, 0x00},
//		{0xff, 0xff, 0xff, 0xff},
//		{0x80, 0x00, 0x00, 0x00},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteInt32(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestReadInt64(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
//		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
//		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
//		{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
//	}
//	yValues := []int64{
//		math.MaxInt64,
//		0,
//		-1,
//		math.MinInt64,
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadInt64()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteInt64(
//	t *testing.T,
//) {
//	xValues := []int64{
//		math.MaxInt64,
//		0,
//		-1,
//		math.MinInt64,
//	}
//	yValues := [][]uint8{
//		{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
//		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
//		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
//		{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteInt64(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestReadFloat32(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x7F, 0x7F, 0xFF, 0xFF},
//		{0xFF, 0x7F, 0xFF, 0xFF},
//	}
//	yValues := []float32{
//		math.MaxFloat32,
//		-math.MaxFloat32,
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadFloat32()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteFloat32(
//	t *testing.T,
//) {
//	xValues := []float32{
//		math.MaxFloat32,
//		-math.MaxFloat32,
//	}
//	yValues := [][]uint8{
//		{0x7F, 0x7F, 0xFF, 0xFF},
//		{0xFF, 0x7F, 0xFF, 0xFF},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteFloat32(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestReadFloat64(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x7F, 0xEF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
//		{0xFF, 0xEF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
//	}
//	yValues := []float64{
//		math.MaxFloat64,
//		-math.MaxFloat64,
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadFloat64()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteFloat64(
//	t *testing.T,
//) {
//	xValues := []float64{
//		math.MaxFloat64,
//		-math.MaxFloat64,
//	}
//	yValues := [][]uint8{
//		{0x7F, 0xEF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
//		{0xFF, 0xEF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteFloat64(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestReadString(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x0D, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x21},
//	}
//	yValues := []string{
//		"Hello, World!",
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadString()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteString(
//	t *testing.T,
//) {
//	xValues := []string{
//		"Hello, World!",
//	}
//	yValues := [][]uint8{
//		{0x0D, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x21},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteString(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestReadVarInt(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x00},
//		{0x01},
//		{0x02},
//		{0x7F},
//		{0x80, 0x01},
//		{0xFF, 0x01},
//		{0xDD, 0xC7, 0x01},
//		{0xFF, 0xFF, 0x7F},
//		{0xFF, 0xFF, 0xFF, 0xFF, 0x07},
//		{0xFF, 0xFF, 0xFF, 0xFF, 0x0F},
//		{0x80, 0x80, 0x80, 0x80, 0x08},
//	}
//	yValues := []int32{
//		0,
//		1,
//		2,
//		127,
//		128,
//		255,
//		25565,
//		2097151,
//		math.MaxInt32,
//		-1,
//		math.MinInt32,
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadVarInt()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteVarInt(
//	t *testing.T,
//) {
//	xValues := []int32{
//		0,
//		1,
//		2,
//		127,
//		128,
//		255,
//		25565,
//		2097151,
//		math.MaxInt32,
//		-1,
//		math.MinInt32,
//	}
//	yValues := [][]uint8{
//		{0x00},
//		{0x01},
//		{0x02},
//		{0x7F},
//		{0x80, 0x01},
//		{0xFF, 0x01},
//		{0xDD, 0xC7, 0x01},
//		{0xFF, 0xFF, 0x7F},
//		{0xFF, 0xFF, 0xFF, 0xFF, 0x07},
//		{0xFF, 0xFF, 0xFF, 0xFF, 0x0F},
//		{0x80, 0x80, 0x80, 0x80, 0x08},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteVarInt(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestReadVarLong(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x00},
//		{0x01},
//		{0x02},
//		{0x7F},
//		{0x80, 0x01},
//		{0xFF, 0x01},
//		{0xff, 0xff, 0xff, 0xff, 0x07},
//		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
//		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01},
//		{0x80, 0x80, 0x80, 0x80, 0xf8, 0xff, 0xff, 0xff, 0xff, 0x01},
//		{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01},
//	}
//	yValues := []int64{
//		0,
//		1,
//		2,
//		127,
//		128,
//		255,
//		math.MaxInt32,
//		math.MaxInt64,
//		-1,
//		math.MinInt32,
//		math.MinInt64,
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadVarLong()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteVarLong(
//	t *testing.T,
//) {
//	xValues := []int64{
//		0,
//		1,
//		2,
//		127,
//		128,
//		255,
//		math.MaxInt32,
//		math.MaxInt64,
//		-1,
//		math.MinInt32,
//		math.MinInt64,
//	}
//	yValues := [][]uint8{
//		{0x00},
//		{0x01},
//		{0x02},
//		{0x7F},
//		{0x80, 0x01},
//		{0xFF, 0x01},
//		{0xff, 0xff, 0xff, 0xff, 0x07},
//		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
//		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01},
//		{0x80, 0x80, 0x80, 0x80, 0xf8, 0xff, 0xff, 0xff, 0xff, 0x01},
//		{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteVarLong(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//
//}
//
//func TestReadPosition(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x7F, 0xFF, 0xFF, 0xDF, 0xFF, 0xFF, 0xF7, 0xFF},
//		{0x46, 0x07, 0x63, 0x13, 0xEA, 0x4B, 0x83, 0x3F},
//		{0x80, 0x00, 0x00, 0x20, 0x00, 0x00, 0x08, 0x00},
//	}
//	y0Values := []int{
//		33554431,
//		18357644,
//		-33554432,
//	}
//	y1Values := []int{
//		2047,
//		831,
//		-2048,
//	}
//	y2Values := []int{
//		33554431,
//		20882616,
//		-33554432,
//	}
//	for i, x := range xValues {
//		y0 := y0Values[i]
//		y1 := y1Values[i]
//		y2 := y2Values[i]
//
//		data := NewData(x...)
//		y0Prime, y1Prime, y2Prime := data.ReadPosition()
//
//		if y0 == y0Prime &&
//			y1 == y1Prime &&
//			y2 == y2Prime {
//			continue
//		}
//		t.Errorf(
//			"function value of x (%+v, %+v, %+v) is different than expect (%+v, %+v, %+v)",
//			y0Prime, y1Prime, y2Prime,
//			y0, y1, y2,
//		)
//	}
//}
//
//func TestWritePosition(
//	t *testing.T,
//) {
//	x0Values := []int{
//		33554431,
//		18357644,
//		-33554432,
//	}
//	x1Values := []int{
//		2047,
//		831,
//		-2048,
//	}
//	x2Values := []int{
//		33554431,
//		20882616,
//		-33554432,
//	}
//	yValues := [][]uint8{
//		{0x7F, 0xFF, 0xFF, 0xDF, 0xFF, 0xFF, 0xF7, 0xFF},
//		{0x46, 0x07, 0x63, 0x13, 0xEA, 0x4B, 0x83, 0x3F},
//		{0x80, 0x00, 0x00, 0x20, 0x00, 0x00, 0x08, 0x00},
//	}
//	for i, y := range yValues {
//		x0 := x0Values[i]
//		x1 := x1Values[i]
//		x2 := x2Values[i]
//
//		data := NewData()
//		data.WritePosition(x0, x1, x2)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value %+v of x is different than expect %+v", yPrime, y)
//	}
//
//}
//
//func TestReadAngle(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x00},
//		{0xFE},
//	}
//	yValues := []float32{
//		0,
//		358.5882352941176,
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadAngle()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteAngle(
//	t *testing.T,
//) {
//	xValues := []float32{
//		0,
//		358.5882352941176,
//	}
//	yValues := [][]uint8{
//		{0x00},
//		{0xFE},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteAngle(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestReadUUID(
//	t *testing.T,
//) {
//	xValues := [][]uint8{
//		{0x37, 0x0C, 0xae, 0x47, 0x04, 0xd0, 0x48, 0xa8, 0xbb, 0x99, 0xbf, 0xc5, 0xf7, 0x0f, 0x5f, 0x3b},
//		{0x3a, 0x51, 0x49, 0xdc, 0x55, 0xc7, 0x45, 0xf5, 0x9a, 0xa8, 0x5f, 0x90, 0x9e, 0x1c, 0x55, 0x62},
//	}
//	yValues := []uuid.UUID{
//		uuid.MustParse("370cae47-04d0-48a8-bb99-bfc5f70f5f3b"),
//		uuid.MustParse("3a5149dc-55c7-45f5-9aa8-5f909e1c5562"),
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData(x...)
//		yPrime := data.ReadUUID()
//
//		if y == yPrime {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
//
//func TestWriteUUID(
//	t *testing.T,
//) {
//	xValues := []uuid.UUID{
//		uuid.MustParse("370cae47-04d0-48a8-bb99-bfc5f70f5f3b"),
//		uuid.MustParse("3a5149dc-55c7-45f5-9aa8-5f909e1c5562"),
//	}
//	yValues := [][]uint8{
//		{0x37, 0x0C, 0xae, 0x47, 0x04, 0xd0, 0x48, 0xa8, 0xbb, 0x99, 0xbf, 0xc5, 0xf7, 0x0f, 0x5f, 0x3b},
//		{0x3a, 0x51, 0x49, 0xdc, 0x55, 0xc7, 0x45, 0xf5, 0x9a, 0xa8, 0x5f, 0x90, 0x9e, 0x1c, 0x55, 0x62},
//	}
//	for i, x := range xValues {
//		y := yValues[i]
//
//		data := NewData()
//		data.WriteUUID(x)
//		yPrime := data.arr
//
//		if compare(y, yPrime) == true {
//			continue
//		}
//		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
//	}
//}
