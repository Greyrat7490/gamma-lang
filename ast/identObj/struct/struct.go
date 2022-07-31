package structDec

import (
	"gamma/token"
	"gamma/types"
)

type Struct struct {
    decPos token.Pos
    name string
    fields []types.Type
}

func CreateStruct(name token.Token, fields []types.Type) Struct {
    return Struct{ decPos: name.Pos, name: name.Str, fields: fields }
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

func (s *Struct) GetType() types.StructType {
    return types.StructType{ Name: s.name, Types: s.fields }
}
