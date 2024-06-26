package identObj

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
)

type Enum struct {
    decPos token.Pos
    name string
    typ types.EnumType
    generic *Generic
}

func CreateEnum(name token.Token, generic *Generic) Enum {
    return Enum{ decPos: name.Pos, name: name.Str, generic: generic }
}

func (e *Enum) GetName() string {
    return e.name
}

func (e *Enum) GetPos() token.Pos {
    return e.decPos
}

func (e *Enum) Addr() addr.Addr {
    fmt.Fprintln(os.Stderr, "[ERROR] Cannot get the addr of an enum type definition (not allocated anywhere)")
    os.Exit(1)
    return addr.Addr{}
}

func (e *Enum) GetType() types.Type {
    return e.typ
}

func (e *Enum) GetGeneric() *Generic {
    return e.generic
}

func (e *Enum) IsGeneric() bool {
    return e.generic != nil
}

func (e *Enum) SetElems(idType types.Type, elemNames []string, elemTypes []types.Type) {
    if e.IsGeneric() {
        e.typ = types.CreateEnumType(e.name, idType, elemNames, elemTypes, e.generic.Typ.Name)
    } else {
        e.typ = types.CreateEnumType(e.name, idType, elemNames, elemTypes, "")
    }
}

func (e *Enum) HasElem(name string) bool {
    return e.typ.HasElem(name)
}
