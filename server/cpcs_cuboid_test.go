package server

import (
	"testing"
)

func TestIsOverlappingOfCpcsCuboid(
	t *testing.T,
) {
	x0Values := []*CpcsCuboid{
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(0, 0, 0, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(-1, 0, -1, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(-2, 1, -5, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(-2, 1, -5, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
	}
	x1Values := []*CpcsCuboid{
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(1, 1, 1, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(1, 3, 1, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(-1, 1, -4, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(-4, 1, -3, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
	}
	yValues := []bool{
		true,
		false,
		true,
		true,
	}
	for i, y := range yValues {
		x0 := x0Values[i]
		x1 := x1Values[i]

		yPrime := x0.IsOverlapping(x1)
		if yPrime == y {
			continue
		}
		t.Errorf("function value %+v of x0 and x1 is different than expect %+v", yPrime, y)
	}
}

func TestSubOfCpcsCuboid(
	t *testing.T,
) {
	x0Values := []*CpcsCuboid{
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(0, 0, 0, 2, 2, 2)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
	}
	x1Values := []*CpcsCuboid{
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(3, 3, 3, 3, 3, 3)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
	}
	yValues := []*CpcsCuboid{
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(1, 1, 1, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
	}
	for i, y := range yValues {
		x0 := x0Values[i]
		x1 := x1Values[i]

		yPrime := x0.Sub(x1)
		if yPrime.Same(y) {
			continue
		}
		t.Errorf("function value %+v of x0 and x1 is different than expect %+v", yPrime, y)
	}
}

func TestMap0OfCpcsCuboid(
	t *testing.T,
) {
	xValues := []*CpcsCuboid{
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(0, 0, 0, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(-2, 3, -5, 3, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
	}
	yValues := [][]*CpcsColIndices{
		{
			&CpcsColIndices{-1, -1, []int{1, 0}, 2},
			&CpcsColIndices{0, -1, []int{1, 0}, 2},
			&CpcsColIndices{1, -1, []int{1, 0}, 2},
			&CpcsColIndices{-1, 0, []int{1, 0}, 2},
			&CpcsColIndices{0, 0, []int{1, 0}, 2},
			&CpcsColIndices{1, 0, []int{1, 0}, 2},
			&CpcsColIndices{-1, 1, []int{1, 0}, 2},
			&CpcsColIndices{0, 1, []int{1, 0}, 2},
			&CpcsColIndices{1, 1, []int{1, 0}, 2},
		},
		{
			&CpcsColIndices{-5, -6, []int{4, 3, 2}, 3},
			&CpcsColIndices{-4, -6, []int{4, 3, 2}, 3},
			&CpcsColIndices{-3, -6, []int{4, 3, 2}, 3},
			&CpcsColIndices{-2, -6, []int{4, 3, 2}, 3},
			&CpcsColIndices{-1, -6, []int{4, 3, 2}, 3},
			&CpcsColIndices{0, -6, []int{4, 3, 2}, 3},
			&CpcsColIndices{1, -6, []int{4, 3, 2}, 3},
			&CpcsColIndices{-5, -5, []int{4, 3, 2}, 3},
			&CpcsColIndices{-4, -5, []int{4, 3, 2}, 3},
			&CpcsColIndices{-3, -5, []int{4, 3, 2}, 3},
			&CpcsColIndices{-2, -5, []int{4, 3, 2}, 3},
			&CpcsColIndices{-1, -5, []int{4, 3, 2}, 3},
			&CpcsColIndices{0, -5, []int{4, 3, 2}, 3},
			&CpcsColIndices{1, -5, []int{4, 3, 2}, 3},
			&CpcsColIndices{-5, -4, []int{4, 3, 2}, 3},
			&CpcsColIndices{-4, -4, []int{4, 3, 2}, 3},
			&CpcsColIndices{-3, -4, []int{4, 3, 2}, 3},
			&CpcsColIndices{-2, -4, []int{4, 3, 2}, 3},
			&CpcsColIndices{-1, -4, []int{4, 3, 2}, 3},
			&CpcsColIndices{0, -4, []int{4, 3, 2}, 3},
			&CpcsColIndices{1, -4, []int{4, 3, 2}, 3}, // 20
		},
	}

	for i, y := range yValues {
		x := xValues[i]
		yPrime := x.Map0()
		for j, v0 := range y {
			v1 := yPrime[j]
			if v1.Same(v0) == true {
				continue
			}
			t.Errorf(
				"the value %s is different than expect %s at %d-th of %d-th",
				v0, v1, j, i,
			)
			return
		}
	}
}
func TestMap1OfCpcsCuboid(
	t *testing.T,
) {
	x0Values := []*CpcsCuboid{
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(1, 1, 1, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(-3, 2, 2, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(-6, 0, 0, 2, 2, 2)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
	}
	x1Values := []*CpcsCuboid{
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(3, 3, 2, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(-6, 0, 0, 2, 2, 2)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
		func() *CpcsCuboid {
			cuboid, err := NewCpcsCuboid(-3, 2, 2, 1, 1, 1)
			if err != nil {
				panic(err)
			}
			return cuboid
		}(),
	}
	y0Values := [][]*CpcsColIndices{
		{
			&CpcsColIndices{1, 2, []int{2, 1, 0}, 3},
			&CpcsColIndices{0, 2, []int{2, 1, 0}, 3},
			&CpcsColIndices{1, 1, []int{2, 1, 0}, 3},
			&CpcsColIndices{0, 1, []int{2, 1, 0}, 3},
			&CpcsColIndices{2, 0, []int{2, 1, 0}, 3},
			&CpcsColIndices{1, 0, []int{2, 1, 0}, 3},
			&CpcsColIndices{0, 0, []int{2, 1, 0}, 3},
		},
		{
			&CpcsColIndices{-2, 3, []int{3, 2, 1}, 3},
			&CpcsColIndices{-3, 3, []int{3, 2, 1}, 3},
			&CpcsColIndices{-4, 3, []int{3, 2, 1}, 3},
			&CpcsColIndices{-2, 2, []int{3, 2, 1}, 3},
			&CpcsColIndices{-3, 2, []int{3, 2, 1}, 3},
			&CpcsColIndices{-2, 1, []int{3, 2, 1}, 3},
			&CpcsColIndices{-3, 1, []int{3, 2, 1}, 3},
		},
		{
			&CpcsColIndices{-5, 2, []int{2, 1, 0}, 3},
			&CpcsColIndices{-6, 2, []int{2, 1, 0}, 3},
			&CpcsColIndices{-7, 2, []int{2, 1, 0}, 3},
			&CpcsColIndices{-8, 2, []int{2, 1, 0}, 3},
			&CpcsColIndices{-5, 1, []int{2, 1, 0}, 3},
			&CpcsColIndices{-6, 1, []int{2, 1, 0}, 3},
			&CpcsColIndices{-7, 1, []int{2, 1, 0}, 3},
			&CpcsColIndices{-8, 1, []int{2, 1, 0}, 3},
			&CpcsColIndices{-4, 0, []int{2, 1, 0}, 3},
			&CpcsColIndices{-5, 0, []int{2, 1, 0}, 3},
			&CpcsColIndices{-6, 0, []int{2, 1, 0}, 3},
			&CpcsColIndices{-7, 0, []int{2, 1, 0}, 3},
			&CpcsColIndices{-8, 0, []int{2, 1, 0}, 3},
		},
	}
	y1Values := [][]*CpcsColIndices{
		{
			&CpcsColIndices{2, 2, []int{2}, 1},
			&CpcsColIndices{2, 1, []int{2}, 1},
		},
		{
			&CpcsColIndices{-4, 2, []int{2, 1}, 2},
			&CpcsColIndices{-4, 1, []int{2, 1}, 2},
		},
		{
			&CpcsColIndices{-4, 2, []int{2, 1}, 2},
			&CpcsColIndices{-4, 1, []int{2, 1}, 2},
		},
	}
	y2Values := [][]*CpcsColIndices{
		{
			&CpcsColIndices{4, 3, []int{4, 3, 2}, 3},
			&CpcsColIndices{3, 3, []int{4, 3, 2}, 3},
			&CpcsColIndices{2, 3, []int{4, 3, 2}, 3},
			&CpcsColIndices{4, 2, []int{4, 3, 2}, 3},
			&CpcsColIndices{3, 2, []int{4, 3, 2}, 3},
			&CpcsColIndices{4, 1, []int{4, 3, 2}, 3},
			&CpcsColIndices{3, 1, []int{4, 3, 2}, 3},
		},
		{
			&CpcsColIndices{-5, 2, []int{2, 1, 0}, 3},
			&CpcsColIndices{-6, 2, []int{2, 1, 0}, 3},
			&CpcsColIndices{-7, 2, []int{2, 1, 0}, 3},
			&CpcsColIndices{-8, 2, []int{2, 1, 0}, 3},
			&CpcsColIndices{-5, 1, []int{2, 1, 0}, 3},
			&CpcsColIndices{-6, 1, []int{2, 1, 0}, 3},
			&CpcsColIndices{-7, 1, []int{2, 1, 0}, 3},
			&CpcsColIndices{-8, 1, []int{2, 1, 0}, 3},
			&CpcsColIndices{-4, 0, []int{2, 1, 0}, 3},
			&CpcsColIndices{-5, 0, []int{2, 1, 0}, 3},
			&CpcsColIndices{-6, 0, []int{2, 1, 0}, 3},
			&CpcsColIndices{-7, 0, []int{2, 1, 0}, 3},
			&CpcsColIndices{-8, 0, []int{2, 1, 0}, 3},
		},
		{
			&CpcsColIndices{-2, 3, []int{3, 2, 1}, 3},
			&CpcsColIndices{-3, 3, []int{3, 2, 1}, 3},
			&CpcsColIndices{-4, 3, []int{3, 2, 1}, 3},
			&CpcsColIndices{-2, 2, []int{3, 2, 1}, 3},
			&CpcsColIndices{-3, 2, []int{3, 2, 1}, 3},
			&CpcsColIndices{-2, 1, []int{3, 2, 1}, 3},
			&CpcsColIndices{-3, 1, []int{3, 2, 1}, 3},
		},
	}

	n := 3
	for i := 0; i < n; i++ {
		x0, x1 := x0Values[i], x1Values[i]
		y0, y1, y2 := y0Values[i], y1Values[i], y2Values[i]
		y0Prime, y1Prime, y2Prime, err := x0.Map1(x1)
		if err != nil {
			panic(err)
		}
		for j, v0 := range y0 {
			v1 := y0Prime[j]
			if v1.Same(v0) == true {
				continue
			}
			t.Errorf(
				"the element %s of %d-th of y0 is different than expect %s in %d-th case",
				v1, j, v0, i,
			)
			return
		}
		for j, v0 := range y1 {
			v1 := y1Prime[j]
			if v1.Same(v0) == true {
				continue
			}
			t.Errorf(
				"the element %s of %d-th of y1 is different than expect %s in %d-th case",
				v1, j, v0, i,
			)
			return
		}
		for j, v0 := range y2 {
			v1 := y2Prime[j]
			if v1.Same(v0) == true {
				continue
			}
			t.Errorf(
				"the element %s of %d-th of y2 is different than expect %s in %d-th case",
				v1, j, v0, i,
			)
			return
		}
	}
}
