package vars

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/types"
)

type LocalVar struct {
    decPos token.Pos
    name string
    typ types.Type
    offset int
}

func CreateLocal(name token.Token, t types.Type, frameSize int) LocalVar {
    return LocalVar{ name: name.Str, decPos: name.Pos, typ: t, offset: calcOffset(t, frameSize) }
}

func (v *LocalVar) SetType(t types.Type) {
    if v.typ != nil {
        fmt.Println("[ERROR] setting the type of a var again is not allowed")
        os.Exit(1)
    }

    v.typ = t
}

func (v *LocalVar) String() string {
    return fmt.Sprintf("{%s %v}", v.name, v.typ)
}

func (v *LocalVar) GetName() string {
    return v.name
}

func (v *LocalVar) GetPos() token.Pos {
    return v.decPos
}

func (v *LocalVar) GetType() types.Type {
    return v.typ
}

func (v *LocalVar) Addr(fieldNum int) string {
    switch t := v.typ.(type) {
    case types.StrType:
        offset := v.offset
        if fieldNum == 0 {
            offset += types.Ptr_Size
        }

        return fmt.Sprintf("rbp-%d", offset)

    case types.StructType:
        offset := v.offset
        for i := len(t.Types)-1; i > fieldNum; i-- {
            offset += t.Types[i].Size()
        }

        return fmt.Sprintf("rbp-%d", offset)

    default:
        return fmt.Sprintf("rbp-%d", v.offset)
    }
}


func (v *LocalVar) DefVal(file *os.File, val token.Token) {
    VarSetVal(file, v, val)
}

func calcOffset(vartype types.Type, frameSize int) int {
    switch t := vartype.(type) {
    case types.StrType:
        return frameSize + types.I32_Size

    case types.StructType:
        return frameSize + t.Types[0].Size()

    default:
        return frameSize + vartype.Size()
    }
}
