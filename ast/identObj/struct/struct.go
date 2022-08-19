package structDec

import (
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

func (s *Struct) Addr(fieldNum int) string {
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

func (s *Struct) GetFieldNum(name string) int {
    for i,f := range s.fieldNames {
        if f == name {
            return i
        }
    }

    return -1
}
