package structDec

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/types"
)

type Struct struct {
    decPos token.Pos
    name string
    typ types.StructType
    fieldNames []string
    fieldTypes []types.Type
}

func CreateStruct(name token.Token, fieldNames []string, fieldTypes []types.Type) Struct {
    return Struct{
        decPos: name.Pos,
        name: name.Str,
        typ: types.CreateStructType(name.Str, fieldTypes),
        fieldNames: fieldNames,
        fieldTypes: fieldTypes,
    }
}

func (s *Struct) GetName() string {
    return s.name
}

func (s *Struct) GetPos() token.Pos {
    return s.decPos
}

func (s *Struct) Addr(field uint) string {
    fmt.Fprintln(os.Stderr, "[ERROR] Cannot get the addr of a struct type definition (not allocated anywhere)")
    os.Exit(1)
    return ""
}

func (s *Struct) GetType() types.Type {
    return s.typ
}

func (s *Struct) GetTypes() []types.Type {
    return s.fieldTypes
}

func (s *Struct) GetNames() []string {
    return s.fieldNames
}

func (s *Struct) GetTypeOfField(name string) types.Type {
    for i,f := range s.fieldNames {
        if f == name {
            return s.fieldTypes[i]
        }
    }

    return nil
}

func (s *Struct) GetFieldNum(name string) (uint, bool) {
    for i,f := range s.fieldNames {
        if f == name {
            return uint(i), true
        }
    }

    return 0, false
}

func (s *Struct) GetField(name string) (types.Type, uint) {
    for i,f := range s.fieldNames {
        if f == name {
            return s.fieldTypes[i], uint(i)
        }
    }

    return nil, 0
}
