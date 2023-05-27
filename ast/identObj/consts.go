package identObj

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
    "gamma/cmpTime/constVal"
)

type Const struct {
    decPos token.Pos
    name string
    typ types.Type
    val constVal.ConstVal
}

func CreateConst(name token.Token, t types.Type, val constVal.ConstVal) Const {
    return Const{ name: name.Str, decPos: name.Pos, typ: t, val: val }
}

func (c *Const) GetName() string {
    return c.name
}

func (c *Const) GetPos() token.Pos {
    return c.decPos
}

func (c *Const) GetType() types.Type {
    return c.typ
}

func (c *Const) GetVal() constVal.ConstVal {
    return c.val
}

func (c *Const) ResolveType(t types.Type, useDefault bool) {
    if useDefault {
        if typ,ok := c.typ.(types.InferType); ok {
            if inferedType,ok := t.(types.InferType); ok {
                c.typ = inferedType.DefaultType
            } else if t != nil {
                c.typ = t
            } else {
                c.typ = typ.DefaultType
            }
        }
    } else {
        if c.typ.GetKind() == types.Infer && t != nil {
            c.typ = t
        }
    }
}

func (c *Const) Addr() addr.Addr {
    fmt.Fprintln(os.Stderr, "[ERROR] Cannot get the addr of a const (consts are not allocated anywhere)")
    os.Exit(1)
    return addr.Addr{}
}
