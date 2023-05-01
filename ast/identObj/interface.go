package identObj

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
)

type Interface struct {
    decPos token.Pos
    name string
    scope *Scope
    Funcs []Func
}

func CreateInterface(name token.Token) Interface {
    return Interface{ decPos: name.Pos, name: name.Str, Funcs: make([]Func, 0) }
}

func (i *Interface) GetName() string {
    return i.name
}

func (i *Interface) GetType() types.Type {
    return nil
}

func (i *Interface) GetPos() token.Pos {
    return i.decPos
}

func (i *Interface) Addr() addr.Addr {
    fmt.Fprintln(os.Stderr, "[ERROR] (internal) Cannot get the addr of an interface definition")
    os.Exit(1)
    return addr.Addr{}
}

func (i *Interface) SetFuncs(funcs []Func) {
    i.Funcs = funcs
}
