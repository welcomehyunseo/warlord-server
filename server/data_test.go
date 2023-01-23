package server

import (
	"math"
	"testing"
)

func TestReadBool(
	t *testing.T,
) {
	xValues := [][]uint8{
		{0x01},
		{0x00},
	}
	yValues := []bool{
		true,
		false,
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData(x...)
		yPrime := data.ReadBool()

		if y == yPrime {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestWriteBool(
	t *testing.T,
) {
	xValues := []bool{
		true,
		false,
	}
	yValues := [][]uint8{
		{0x01},
		{0x00},
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData()
		data.WriteBool(x)
		yPrime := data.buf

		if compare(y, yPrime) == true {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestReadInt8(
	t *testing.T,
) {
	xValues := [][]uint8{
		{0x7F},
		{0x00},
		{0xff},
		{0x80},
	}
	yValues := []int8{
		math.MaxInt8,
		0,
		-1,
		math.MinInt8,
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData(x...)
		yPrime := data.ReadInt8()

		if y == yPrime {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestWriteInt8(
	t *testing.T,
) {
	xValues := []int8{
		math.MaxInt8,
		0,
		-1,
		math.MinInt8,
	}
	yValues := [][]uint8{
		{0x7F},
		{0x00},
		{0xff},
		{0x80},
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData()
		data.WriteInt8(x)
		yPrime := data.buf

		if compare(y, yPrime) == true {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestReadUint8(
	t *testing.T,
) {
	xValues := [][]uint8{
		{0x00},
		{0xff},
	}
	yValues := []uint8{
		0,
		math.MaxUint8,
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData(x...)
		yPrime := data.ReadUint8()

		if y == yPrime {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestWriteUint8(
	t *testing.T,
) {
	xValues := []uint8{
		0,
		math.MaxUint8,
	}
	yValues := [][]uint8{
		{0x00},
		{0xff},
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData()
		data.WriteUint8(x)
		yPrime := data.buf

		if compare(y, yPrime) == true {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestReadInt16(
	t *testing.T,
) {
	xValues := [][]uint8{
		{0x7F, 0xFF},
		{0x00, 0x00},
		{0xff, 0xff},
		{0x80, 0x00},
	}
	yValues := []int16{
		math.MaxInt16,
		0,
		-1,
		math.MinInt16,
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData(x...)
		yPrime := data.ReadInt16()

		if y == yPrime {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestWriteInt16(
	t *testing.T,
) {
	xValues := []int16{
		math.MaxInt16,
		0,
		-1,
		math.MinInt16,
	}
	yValues := [][]uint8{
		{0x7F, 0xFF},
		{0x00, 0x00},
		{0xff, 0xff},
		{0x80, 0x00},
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData()
		data.WriteInt16(x)
		yPrime := data.buf

		if compare(y, yPrime) == true {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestReadUint16(
	t *testing.T,
) {
	xValues := [][]uint8{
		{0x00, 0x00},
		{0xff, 0xff},
	}
	yValues := []uint16{
		0,
		math.MaxUint16,
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData(x...)
		yPrime := data.ReadUint16()

		if y == yPrime {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestWriteUit16(
	t *testing.T,
) {
	xValues := []uint16{
		0,
		math.MaxUint16,
	}
	yValues := [][]uint8{
		{0x00, 0x00},
		{0xff, 0xff},
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData()
		data.WriteUint16(x)
		yPrime := data.buf

		if compare(y, yPrime) == true {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestReadInt32(
	t *testing.T,
) {
	xValues := [][]uint8{
		{0x7F, 0xFF, 0xFF, 0xFF},
		{0x00, 0x00, 0x00, 0x00},
		{0xff, 0xff, 0xff, 0xff},
		{0x80, 0x00, 0x00, 0x00},
	}
	yValues := []int32{
		math.MaxInt32,
		0,
		-1,
		math.MinInt32,
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData(x...)
		yPrime := data.ReadInt32()

		if y == yPrime {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestWriteInt32(
	t *testing.T,
) {
	xValues := []int32{
		math.MaxInt32,
		0,
		-1,
		math.MinInt32,
	}
	yValues := [][]uint8{
		{0x7F, 0xFF, 0xFF, 0xFF},
		{0x00, 0x00, 0x00, 0x00},
		{0xff, 0xff, 0xff, 0xff},
		{0x80, 0x00, 0x00, 0x00},
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData()
		data.WriteInt32(x)
		yPrime := data.buf

		if compare(y, yPrime) == true {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestReadInt64(
	t *testing.T,
) {
	xValues := [][]uint8{
		{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	yValues := []int64{
		math.MaxInt64,
		0,
		-1,
		math.MinInt64,
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData(x...)
		yPrime := data.ReadInt64()

		if y == yPrime {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestWriteInt64(
	t *testing.T,
) {
	xValues := []int64{
		math.MaxInt64,
		0,
		-1,
		math.MinInt64,
	}
	yValues := [][]uint8{
		{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData()
		data.WriteInt64(x)
		yPrime := data.buf

		if compare(y, yPrime) == true {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestReadFloat32(
	t *testing.T,
) {
	xValues := [][]uint8{
		{0x7F, 0x7F, 0xFF, 0xFF},
		{0xFF, 0x7F, 0xFF, 0xFF},
	}
	yValues := []float32{
		math.MaxFloat32,
		-math.MaxFloat32,
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData(x...)
		yPrime := data.ReadFloat32()

		if y == yPrime {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestWriteFloat32(
	t *testing.T,
) {
	xValues := []float32{
		math.MaxFloat32,
		-math.MaxFloat32,
	}
	yValues := [][]uint8{
		{0x7F, 0x7F, 0xFF, 0xFF},
		{0xFF, 0x7F, 0xFF, 0xFF},
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData()
		data.WriteFloat32(x)
		yPrime := data.buf

		if compare(y, yPrime) == true {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestReadFloat64(
	t *testing.T,
) {
	xValues := [][]uint8{
		{0x7F, 0xEF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		{0xFF, 0xEF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	}
	yValues := []float64{
		math.MaxFloat64,
		-math.MaxFloat64,
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData(x...)
		yPrime := data.ReadFloat64()

		if y == yPrime {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}

func TestWriteFloat64(
	t *testing.T,
) {
	xValues := []float64{
		math.MaxFloat64,
		-math.MaxFloat64,
	}
	yValues := [][]uint8{
		{0x7F, 0xEF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		{0xFF, 0xEF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	}
	for i, x := range xValues {
		y := yValues[i]

		data := NewData()
		data.WriteFloat64(x)
		yPrime := data.buf

		if compare(y, yPrime) == true {
			continue
		}
		t.Errorf("function value of x %+v is different than expect %+v", yPrime, y)
	}
}
