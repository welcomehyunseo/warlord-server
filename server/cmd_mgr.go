package server

import "errors"

func makeIndices(
	cmds []*Cmd,
) map[string]int {
	indices := make(map[string]int)
	for i, cmd := range cmds {
		use := cmd.GetUse()
		indices[use] = i
	}

	return indices
}

type CmdMgr interface {
	Distribute(
		chars []string,
	) (
		[]string,
		*Cmd,
		error,
	)
}

type cmdMgr struct {
	cmds    []*Cmd
	indices map[string]int
}

func newCmdMgr(
	cmds []*Cmd,
) *cmdMgr {
	indices := makeIndices(cmds)
	return &cmdMgr{
		cmds,
		indices,
	}
}

func (mgr *cmdMgr) Distribute(
	chars []string,
) (
	[]string,
	*Cmd,
	error,
) {
	if len(chars) == 0 ||
		chars == nil {
		return nil, nil, errors.New("it is empty command")
	}

	use := chars[0]
	indices := mgr.indices
	i, ok := indices[use]
	if ok == false {
		return nil, nil, errors.New("it is unknown command")
	}

	args := chars[1:]
	cmd := mgr.get(i)
	return args, cmd, nil
}

func (mgr *cmdMgr) get(
	i int,
) *Cmd {
	return mgr.cmds[i]
}

type HeadCmdMgr struct {
	*cmdMgr
}

func NewHeadCmdMgr() *HeadCmdMgr {
	return &HeadCmdMgr{
		newCmdMgr(
			HeadCmds,
		),
	}
}

type WorldCmdMgr struct {
	*cmdMgr
}

func NewWorldCmdMgr() *WorldCmdMgr {
	return &WorldCmdMgr{
		newCmdMgr(
			WorldCmds,
		),
	}
}
