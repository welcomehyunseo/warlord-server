package server

type Cmd struct {
	use         string
	usage, desc string
	name        string
}

func NewCmd(
	use string,
	usage, desc string,
	name string,
) *Cmd {
	return &Cmd{
		use,
		usage, desc,
		name,
	}
}

func (c *Cmd) GetUse() string {
	return c.use
}

func (c *Cmd) GetUsage() string {
	return c.usage
}

func (c *Cmd) GetDesc() string {
	return c.desc
}

func (c *Cmd) GetName() string {
	return c.name
}
