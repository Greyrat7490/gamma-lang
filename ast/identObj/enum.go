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
    generic *types.GenericType
    impls []Impl
}

func CreateEnum(name token.Token) Enum {
    return Enum{ decPos: name.Pos, name: name.Str }
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

func (e *Enum) AddImpl(impl Impl) {
    e.impls = append(e.impls, impl)
    if impl.interface_ != nil {
        e.typ.Interfaces[impl.interface_.name] = impl.interface_.typ
    }
}

func (e *Enum) GetFunc(name string) *Func {
    for _,i := range e.impls {
        f := i.GetFunc(name)
        if f != nil {
            return f
        }
    }

    return nil
}

func (e *Enum) GetFuncNames() []string {
    funcs := []string{}

    for _,i := range e.impls {
        funcs = append(funcs, i.GetInterfaceFuncNames()...)
    }

    return funcs
}

func (e *Enum) GetGeneric() *types.GenericType {
    return e.generic
}

func (e *Enum) IsGeneric() bool {
    return e.generic != nil
}

func (e *Enum) SetGeneric(t *types.GenericType) {
    e.generic = t
}

func (e *Enum) SetElems(idType types.Type, elemNames []string, elemTypes []types.Type) {
    if e.IsGeneric() {
        e.typ = types.CreateEnumType(e.name, idType, elemNames, elemTypes, e.generic.Name)
    } else {
        e.typ = types.CreateEnumType(e.name, idType, elemNames, elemTypes, "")
    }
}
