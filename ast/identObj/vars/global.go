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
    addr addr.Addr
    typ types.Type
}

func CreateGlobalVar(name token.Token, t types.Type) GlobalVar {
    return GlobalVar{ name: name.Str, decPos: name.Pos, typ: t, addr: addr.Addr{ BaseAddr: name.Str } }
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

func (v *GlobalVar) ResolveType(t types.Type) {
    if types.IsResolvable(v.typ) {
        v.typ = t
    }
}

func (v *GlobalVar) SetAddr(addr addr.Addr) {
    v.addr = addr
}

func (v *GlobalVar) Addr() addr.Addr {
    return v.addr
}
