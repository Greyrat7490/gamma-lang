package identObj

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
)

type Struct struct {
    decPos token.Pos
    name string
    typ types.StructType
    impls []Impl
}

func CreateStruct(name token.Token, fieldNames []string, fieldTypes []types.Type) Struct {
    return Struct{
        decPos: name.Pos,
        name: name.Str,
        typ: types.CreateStructType(name.Str, fieldTypes, fieldNames),
    }
}

func (s *Struct) GetName() string {
    return s.name
}

func (s *Struct) GetPos() token.Pos {
    return s.decPos
}

func (s *Struct) Addr() addr.Addr {
    fmt.Fprintln(os.Stderr, "[ERROR] Cannot get the addr of a struct type definition (not allocated anywhere)")
    os.Exit(1)
    return addr.Addr{}
}

func (s *Struct) GetType() types.Type {
    return s.typ
}

func (s *Struct) GetFieldType(name string) types.Type {
    return s.typ.GetType(name)
}

func (s *Struct) GetFieldNames() []string {
    return s.typ.GetFields()
}

func (s *Struct) GetFuncNames() []string {
    funcs := []string{}

    for _,i := range s.impls {
        funcs = append(funcs, i.GetInterfaceFuncNames()...)
    }

    return funcs
}

func (s *Struct) AddImpl(impl Impl) {
    s.impls = append(s.impls, impl)
    if impl.interface_ != nil {
        s.typ.Interfaces[impl.interface_.name] = impl.interface_.typ
    }
}

func (s *Struct) GetInterfaceType(interfaceName string) (hasImpl bool, typ types.InterfaceType) {
    for _,impl := range s.impls {
        if impl.interface_.name == interfaceName {
            return true, impl.interface_.typ
        }
    }

    return
}

func (s *Struct) GetFunc(name string) *Func {
    for _,i := range s.impls {
        f := i.GetFunc(name)
        if f != nil {
            return f
        }
    }

    return nil
}
