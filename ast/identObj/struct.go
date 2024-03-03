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
    generic *Generic
}

func CreateStruct(name token.Token, generic *Generic) Struct {
    return Struct{ decPos: name.Pos, name: name.Str, typ: types.CreateEmptyStructType(name.Str), generic: generic }
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

func (s *Struct) resolveRecursiveField_(t types.PtrType) types.Type {
    switch t := t.BaseType.(type) {
    case types.PtrType:
        return s.resolveRecursiveField_(t)

    case types.StructType:
        if t.Name == s.typ.Name {
            return types.PtrType{ BaseType: s.typ }
        }
    }

    return t
}

func (s *Struct) resolveRecursiveField() {
    for i,t := range s.typ.Types {
        if t,ok := t.(types.PtrType); ok {
            s.typ.Types[i] = s.resolveRecursiveField_(t)
        }
    }
}

func (s *Struct) SetFields(fieldNames []string, fieldTypes []types.Type) {
    if s.IsGeneric() {
        s.typ = types.CreateStructType(s.name, fieldTypes, fieldNames, s.generic.Typ.Name)
    } else {
        s.typ = types.CreateStructType(s.name, fieldTypes, fieldNames, "")
    }

    s.resolveRecursiveField()
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

func (s *Struct) GetGeneric() *types.GenericType {
    return &s.generic.Typ
}

func (s *Struct) IsGeneric() bool {
    return s.generic != nil
}
