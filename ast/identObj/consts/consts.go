package consts

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
)

type Const struct {
    decPos token.Pos
    name string
    typ types.Type
    val token.Token
}

func CreateConst(name token.Token, t types.Type) Const {
    return Const{ name: name.Str, decPos: name.Pos, typ: t }
}

func (c *Const) Define(val token.Token) {
    c.val = val
}

func (c *Const) SetType(t types.Type) {
    if c.typ != nil {
        fmt.Fprintln(os.Stderr, "[ERROR] setting the type of a const again is not allowed")
        os.Exit(1)
    }

    c.typ = t
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

func (c *Const) GetVal() token.Token {
    return c.val
}

func (f *Const) Addr(fieldNum int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] Cannot get the addr of a const (consts are not allocated anywhere)")
    os.Exit(1)
    return ""
}
