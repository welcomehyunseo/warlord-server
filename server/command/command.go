package command

type Command struct {
	use         string
	usage, desc string
	name        string
}

func NewCommand(
	use string,
	usage, desc string,
	name string,
) *Command {
	return &Command{
		use,
		usage, desc,
		name,
	}
}

func (c *Command) GetUse() string {
	return c.use
}

func (c *Command) GetUsage() string {
	return c.usage
}

func (c *Command) GetDescription() string {
	return c.desc
}

func (c *Command) GetName() string {
	return c.name
}
