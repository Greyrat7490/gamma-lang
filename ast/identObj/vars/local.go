package vars

import (
    "fmt"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
)

type LocalVar struct {
    decPos token.Pos
    name string
    typ types.Type
    isArg bool
    addr addr.Addr
}

func CreateLocal(name token.Token, t types.Type, frameSize uint, isArg bool, fromStack bool) LocalVar {
    return LocalVar{ name: name.Str, decPos: name.Pos, typ: t, isArg: isArg, addr: addr.Addr{ BaseAddr: "rbp", Offset: calcOffset(t, frameSize, fromStack) } }
}

func (v *LocalVar) String() string {
    return fmt.Sprintf("{%s %v}", v.name, v.typ)
}

func (v *LocalVar) GetName() string {
    return v.name
}

func (v *LocalVar) GetPos() token.Pos {
    return v.decPos
}

func (v *LocalVar) GetType() types.Type {
    return v.typ
}

func (v *LocalVar) Addr() addr.Addr {
    return v.addr
}

func calcOffset(t types.Type, frameSize uint, fromStack bool) int64 {
    if fromStack {
        return int64(types.Ptr_Size + frameSize + 7) & ^7
    }

    return -int64(frameSize + t.Size())
}
