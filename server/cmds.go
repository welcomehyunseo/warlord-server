package server

var (
	HeadCmdForWorld = NewCmd(
		"world",
		"world <sub-command>", "", // TODO: complete description
		"head-cmd/world",
	)

	HeadCmds = []*Cmd{
		HeadCmdForWorld,
	}

	WorldCmdToSummon = NewCmd(
		"summon",
		"summon", "", // TODO: complete description
		"world-cmd/summon",
	)

	WorldCmds = []*Cmd{
		WorldCmdToSummon,
	}
)
