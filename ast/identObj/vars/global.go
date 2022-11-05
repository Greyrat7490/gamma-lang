package vars

import (
    "fmt"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
)

type GlobalVar struct {
    decPos token.Pos
    name string
    typ types.Type
}

func CreateGlobalVar(name token.Token, t types.Type) GlobalVar {
    return GlobalVar{ name: name.Str, decPos: name.Pos, typ: t }
}

func (v *GlobalVar) String() string {
    return fmt.Sprintf("{%s %s}", v.name, v.typ)
}

func (v *GlobalVar) GetName() string {
    return v.name
}

func (v *GlobalVar) GetPos() token.Pos {
    return v.decPos
}

func (v *GlobalVar) GetType() types.Type {
    return v.typ
}

func (v *GlobalVar) Addr() addr.Addr {
    return addr.Addr{ BaseAddr: v.name }
}
