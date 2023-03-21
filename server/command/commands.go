package command

var (
	HeadCmdForWorld = NewCommand(
		"world",
		"world <sub-command>", "", // TODO: complete description
		"head-cmd/world",
	)

	HeadCmds = []*Command{
		HeadCmdForWorld,
	}

	WorldCmdToGive = NewCommand(
		"give",
		"give ", "", // TODO: complete description
		"world-cmd/give",
	)

	WorldCmds = []*Command{
		WorldCmdToGive,
	}
)
