package consts

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/types"
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

func (c *Const) Addr(field uint) string {
    fmt.Fprintln(os.Stderr, "[ERROR] Cannot get the addr of a const (consts are not allocated anywhere)")
    os.Exit(1)
    return ""
}
