package command

//
//import "errors"
//
//func makeIndices(
//	cmds []*Command,
//) map[string]int {
//	indices := make(map[string]int)
//	for i, cmd := range cmds {
//		use := cmd.GetUse()
//		indices[use] = i
//	}
//
//	return indices
//}
//
//type commandLoader struct {
//	name string
//
//	cmds    []*Command
//	indices map[string]int
//
//	childs map[string]*commandLoader
//}
//
//func newCommandLoader(
//	name string,
//	cmds []*Command,
//	arr []*commandLoader,
//) *commandLoader {
//	indices := makeIndices(cmds)
//
//	childs := make(map[string]*commandLoader)
//	for _, ld := range arr {
//		childs[ld.name] = ld
//	}
//
//	return &commandLoader{
//		name,
//
//		cmds,
//		indices,
//
//		childs,
//	}
//}
//
//func (ld *commandLoader) Load(
//	chars []string,
//) (
//	[]string,
//	string,
//	error,
//) {
//	if len(chars) == 0 ||
//		chars == nil {
//		return nil, nil, errors.New("it is empty chars to load in commandLoader")
//	}
//
//	use := chars[0]
//	i, ok := ld.indices[use]
//	if ok == false {
//		return nil, nil, errors.New("it is unknown command to load in commandLoader")
//	}
//
//	args := chars[1:]
//	cmd := ld.get(i)
//	return args, cmd, nil
//}
//
//func (ld *commandLoader) get(
//	i int,
//) *Command {
//	return ld.cmds[i]
//}
//
//func (ld *commandLoader) GetName() string {
//	return ld.name
//}
//
//type HeadCommandLoader struct {
//	*commandLoader
//}
//
//func NewHeadCommandLoader() *HeadCommandLoader {
//	return &HeadCommandLoader{
//		newCommandLoader(
//			HeadCmds,
//		),
//	}
//}
//
//type WorldCommandLoader struct {
//	*commandLoader
//}
//
//func NewWorldCommandLoader() *WorldCommandLoader {
//	return &WorldCommandLoader{
//		newCommandLoader(
//			WorldCmds,
//		),
//	}
//}
