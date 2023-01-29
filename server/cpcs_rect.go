package server

type CpcsRect struct {
	cx0 int
	cz0 int
	cx1 int
	cz1 int
}

func NewCpcsRect(
	cx, cz,
	dx, dz int,
) *CpcsRect {
	return &CpcsRect{
		cx + dx, cz + dz, cx - dx, cz - dz,
	}
}

func (cr *CpcsRect) GetCx0() int {
	return cr.cx0
}

func (cr *CpcsRect) GetCz0() int {
	return cr.cz0
}

func (cr *CpcsRect) GetCx1() int {
	return cr.cx1
}

func (cr *CpcsRect) GetCz1() int {
	return cr.cz1
}
