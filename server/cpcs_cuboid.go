package server

import (
	"errors"
	"fmt"
	"sort"
)

var NegativeNumberError = errors.New("the values cy, dx, dy, and dz of cuboid must be greater than or equal zero in the chunk part coordinate system")
var SameError = errors.New("the cpcs cuboids c0 and c1 is same")

type CpcsCuboid struct {
	cx0 int // max
	cy0 int
	cz0 int
	cx1 int // min
	cy1 int
	cz1 int

	cx2 int
	cy2 int
	cz2 int
	dx  int
	dy  int
	dz  int
}

// NewCpcsCuboid returns a new CpcsCuboid with the center point (x, y, z)
// and distances dx, dy, and dz in the chunk part coordinate system.
// The distances are length from center to each point cx0, cy0, cz0, cx1, cy1, and cz1.
// The values of y-axis must be greater than or equal to 0 and lower than 16.
func NewCpcsCuboid(
	cx, cy, cz,
	dx, dy, dz int,
) (*CpcsCuboid, error) {
	if cy < 0 || dx < 0 || dy < 0 || dz < 0 {
		return nil, NegativeNumberError
	}

	cx0, cy0, cz0 := cx+dx, cy+dy, cz+dz
	cx1, cy1, cz1 := cx-dx, cy-dy, cz-dz

	if cx1 > cx0 {
		temp := cx0
		cx0 = cx1
		cx1 = temp
	}
	if cy1 > cy0 {
		temp := cy0
		cy0 = cy1
		cy1 = temp
	}
	if cz1 > cz0 {
		temp := cz0
		cz0 = cz1
		cz1 = temp
	}
	if cy1 < 0 {
		cy1 = 0
	}
	if cy0 > 15 {
		cy0 = 15
	}

	return &CpcsCuboid{
		cx0: cx0,
		cy0: cy0,
		cz0: cz0,
		cx1: cx1,
		cy1: cy1,
		cz1: cz1,
		cx2: cx,
		cy2: cy,
		cz2: cz,
		dx:  dx,
		dy:  dy,
		dz:  dz,
	}, nil
}

func NewCpcsCuboidByPlayerPos(
	x, y, z float64,
	dx, dy, dz int,
) (*CpcsCuboid, error) {
	cx := int(x) / 16
	cy := int(y) / 16
	cz := int(z) / 16
	if cx < 0 {
		cx = cx - 16
	}
	if cy < 0 {
		cy = cy - 16
	}
	if cz < 0 {
		cz = cz - 16
	}

	return NewCpcsCuboid(
		cx, cy, cz,
		dx, dy, dz,
	)
}

func (c0 *CpcsCuboid) Same(
	c1 *CpcsCuboid,
) bool {
	cx0, cy0, cz0, cx1, cy1, cz1 := c0.cx0, c0.cy0, c0.cz0, c0.cx1, c0.cy1, c0.cz1
	cx2, cy2, cz2, cx3, cy3, cz3 := c1.cx0, c1.cy0, c1.cz0, c1.cx1, c1.cy1, c1.cz1

	return cx0 == cx2 && cx1 == cx3 &&
		cy0 == cy2 && cy1 == cy3 &&
		cz0 == cz2 && cz1 == cz3
}

func (c0 *CpcsCuboid) IsOverlapping(
	c1 *CpcsCuboid,
) bool {
	cx0, cy0, cz0, cx1, cy1, cz1 := c0.cx0, c0.cy0, c0.cz0, c0.cx1, c0.cy1, c0.cz1
	cx2, cy2, cz2, cx3, cy3, cz3 := c1.cx0, c1.cy0, c1.cz0, c1.cx1, c1.cy1, c1.cz1

	return cx3 <= cx0 && cx1 <= cx2 &&
		cy3 <= cy0 && cy1 <= cy2 &&
		cz3 <= cz0 && cz1 <= cz2
}

func (c0 *CpcsCuboid) Sub(
	c1 *CpcsCuboid,
) *CpcsCuboid {
	cx0, cy0, cz0, cx1, cy1, cz1 := c0.cx0, c0.cy0, c0.cz0, c0.cx1, c0.cy1, c0.cz1
	cx2, cy2, cz2, cx3, cy3, cz3 := c1.cx0, c1.cy0, c1.cz0, c1.cx1, c1.cy1, c1.cz1

	l0 := []int{cx0, cx1, cx2, cx3}
	l1 := []int{cy0, cy1, cy2, cy3}
	l2 := []int{cz0, cz1, cz2, cz3}
	sort.Ints(l0)
	sort.Ints(l1)
	sort.Ints(l2)

	return &CpcsCuboid{
		cx0: l0[2],
		cy0: l1[2],
		cz0: l2[2],
		cx1: l0[1],
		cy1: l1[1],
		cz1: l2[1],
	}
}

type CpcsColIndices struct {
	cx  int
	cz  int
	cys []int
	i   int
}

func NewCpcsColIndices(
	cx, cz, length int,
) *CpcsColIndices {
	return &CpcsColIndices{
		cx:  cx,
		cz:  cz,
		cys: make([]int, length),
		i:   0,
	}
}

func (c0 *CpcsColIndices) Add(cy int) {
	i := c0.i
	c0.cys[i] = cy
	c0.i++
}

func (c0 *CpcsColIndices) Append(cy int) {
	c0.cys = append(c0.cys, cy)
	c0.i++
}

func (c0 *CpcsColIndices) Same(c1 *CpcsColIndices) bool {
	cys0, cys1 := c0.cys, c1.cys
	l0, l1 := len(cys0), len(cys1)
	if l0 != l1 {
		return false
	}
	for i, cy0 := range cys0 {
		cy1 := c1.cys[i]
		if cy0 != cy1 {
			return false
		}
	}
	return c0.cx == c1.cx && c0.cz == c1.cz
}

func (c0 *CpcsColIndices) GetCx() int {
	return c0.cx
}

func (c0 *CpcsColIndices) GetCz() int {
	return c0.cz
}

func (c0 *CpcsColIndices) GetCys() []int {
	return c0.cys
}

func (c0 *CpcsColIndices) String() string {
	return fmt.Sprintf("{cx: %d, cz: %d, cys: %+v", c0.cx, c0.cz, c0.cys)
}

func (c0 *CpcsCuboid) Map0() []*CpcsColIndices {
	cx0, cy0, cz0, cx1, cy1, cz1 :=
		c0.cx0, c0.cy0, c0.cz0, c0.cx1, c0.cy1, c0.cz1
	l0 := cx0 - cx1 + 1
	l1 := cz0 - cz1 + 1
	length0 := l0 * l1
	arr := make([]*CpcsColIndices, length0)
	length1 := cy0 - cy1 + 1
	for i := cz0; i >= cz1; i-- {
		for j := cx0; j >= cx1; j-- {
			col := NewCpcsColIndices(j, i, length1)
			for k := cy0; k >= cy1; k-- {
				col.Add(k)
			}
			index := (l0 * (i - cz1)) + (j - cx1)
			arr[index] = col
		}
	}
	return arr
}

func (c0 *CpcsCuboid) Map1(
	c1 *CpcsCuboid,
) (
	[]*CpcsColIndices, // c0
	[]*CpcsColIndices, // overlapped
	[]*CpcsColIndices, // c1
	error,
) {
	same := c0.Same(c1)
	if same == true {
		return nil, nil, nil, SameError
	}

	overlap := c0.IsOverlapping(c1)
	if overlap == false {
		arr0 := c0.Map0()
		arr1 := c1.Map0()
		return arr0, nil, arr1, nil
	}

	c2 := c0.Sub(c1)
	cx0, cy0, cz0, cx1, cy1, cz1 :=
		c2.cx0, c2.cy0, c2.cz0, c2.cx1, c2.cy1, c2.cz1
	l0, l1 := cx0-cx1+1, cz0-cz1+1
	l01 := l0 * l1
	arr01 := make([]*CpcsColIndices, l01)

	cx2, cy2, cz2, cx3, cy3, cz3 :=
		c0.cx0, c0.cy0, c0.cz0, c0.cx1, c0.cy1, c0.cz1
	l2, l3 := cx2-cx3+1, cz2-cz3+1
	l23 := l2*l3 - l01
	arr23 := make([]*CpcsColIndices, l23)

	cx4, cy4, cz4, cx5, cy5, cz5 :=
		c1.cx0, c1.cy0, c1.cz0, c1.cx1, c1.cy1, c1.cz1
	l4, l5 := cx4-cx5+1, cz4-cz5+1
	l45 := l4*l5 - l01
	arr45 := make([]*CpcsColIndices, l45)

	l6 := cy2 - cy3 + 1
	index6 := 0
	for i := cz2; i >= cz3; i-- {
		for j := cx2; j >= cx3; j-- {
			if cx1 <= j && j <= cx0 && cz1 <= i && i <= cz0 {
				continue
			}

			col := NewCpcsColIndices(j, i, l6)
			for k := cy2; k >= cy3; k-- {
				col.Add(k)
			}

			arr23[index6] = col
			index6++
		}
	}

	l7 := cy4 - cy5 + 1
	index7 := 0
	for i := cz4; i >= cz5; i-- {
		for j := cx4; j >= cx5; j-- {
			if cx1 <= j && j <= cx0 && cz1 <= i && i <= cz0 {
				continue
			}

			col := NewCpcsColIndices(j, i, l7)
			for k := cy4; k >= cy5; k-- {
				col.Add(k)
			}
			arr45[index7] = col
			index7++
		}
	}

	index8 := 0
	for i := cz0; i >= cz1; i-- {
		for j := cx0; j >= cx1; j-- {
			col := NewCpcsColIndices(j, i, 0)
			for k := cy2; k >= cy3; k-- {
				if k < cy1 || cy0 < k {
					continue
				}

				col.Append(k)
			}
			arr01[index8] = col
			index8++
		}
	}

	return arr23, arr01, arr45, nil
}

func (c0 *CpcsCuboid) GetMax() (int, int, int) {
	return c0.cx0, c0.cy0, c0.cz0
}

func (c0 *CpcsCuboid) GetMin() (int, int, int) {
	return c0.cx1, c0.cy1, c0.cz1
}
