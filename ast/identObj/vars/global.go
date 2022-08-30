package vars

import (
    "fmt"
    "gamma/token"
    "gamma/types"
)

type GlobalVar struct {
    decPos token.Pos
    name string
    typ types.Type
}

func CreateGlobalVar(name token.Token, t types.Type) GlobalVar {
    return GlobalVar{ name: name.Str, decPos: name.Pos, typ: t }
}

func (v *GlobalVar) String() string {
    return fmt.Sprintf("{%s %s}", v.name, v.typ)
}

func (v *GlobalVar) GetName() string {
    return v.name
}

func (v *GlobalVar) GetPos () token.Pos {
    return v.decPos
}

func (v *GlobalVar) GetType() types.Type {
    return v.typ
}

func (v *GlobalVar) OffsetedAddr(offset int) string {
    if offset > 0 {
        return fmt.Sprintf("%s+%d", v.name, offset)
    } else if offset < 0 {
        return fmt.Sprintf("%s%d", v.name, offset)
    } else {
        return v.name
    }
}

func (v *GlobalVar) Addr(field uint) string {
    switch t := v.typ.(type) {
    case types.StrType:
        if field == 1 {
            return fmt.Sprintf("%s+%d", v.name, types.Ptr_Size)
        }

    case types.StructType:
        if field != 0 {
            return fmt.Sprintf("%s+%d", v.name, t.GetOffset(field))
        }
    }

    return v.name
}
