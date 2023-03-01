package server

type CmdName string

var (
	HeadCmdForWorld = NewCmd(
		"world",
		"world <sub-command>", "", // TODO: complete description
		"head-cmd/world",
	)

	HeadCmds = []*Cmd{
		HeadCmdForWorld,
	}

	WorldCmdToChange = NewCmd(
		"change",
		"change <index>", "", // TODO: complete description
		"world-cmd/change",
	)

	WorldCmdToTeleport = NewCmd(
		"teleport",
		"teleport <x> <y> <z>", "", // TODO: complete description
		"world-cmd/teleport",
	)

	WorldCmds = []*Cmd{
		WorldCmdToChange,
		WorldCmdToTeleport,
	}
)
