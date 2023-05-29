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
    impls []Impl
}

func CreateEnum(name token.Token, idType types.Type, elemNames []string, elemTypes []types.Type) Enum {
    return Enum{
        decPos: name.Pos,
        name: name.Str,
        typ: types.CreateEnumType(name.Str, idType, elemNames, elemTypes),
    }
}

func (s *Enum) GetName() string {
    return s.name
}

func (s *Enum) GetPos() token.Pos {
    return s.decPos
}

func (s *Enum) Addr() addr.Addr {
    fmt.Fprintln(os.Stderr, "[ERROR] Cannot get the addr of an enum type definition (not allocated anywhere)")
    os.Exit(1)
    return addr.Addr{}
}

func (s *Enum) GetType() types.Type {
    return s.typ
}
