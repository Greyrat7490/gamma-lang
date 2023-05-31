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

func (v *GlobalVar) ResolveType(t types.Type, useDefault bool) {
    if useDefault {
        if typ,ok := v.typ.(types.InferType); ok {
            if inferedType,ok := t.(types.InferType); ok {
                v.typ = inferedType.DefaultType
            } else if t != nil {
                v.typ = t
            } else {
                v.typ = typ.DefaultType
            }
        }
    } else {
        if v.typ.GetKind() == types.Infer && t != nil {
            v.typ = t
        }
    }
}

func (v *GlobalVar) SetAddr(addr addr.Addr) {
    v.addr = addr
}

func (v *GlobalVar) Addr() addr.Addr {
    return v.addr
}
