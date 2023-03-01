package server

type Cmd struct {
	use         string
	usage, desc string
	name        CmdName
}

func NewCmd(
	use string,
	usage, desc string,
	name CmdName,
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

func (c *Cmd) GetName() CmdName {
	return c.name
}
