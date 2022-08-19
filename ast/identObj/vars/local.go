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
    isArg bool
    offset int
}

func CreateLocal(name token.Token, t types.Type, frameSize uint, isArg bool, fromStack bool) LocalVar {
    return LocalVar{ name: name.Str, decPos: name.Pos, typ: t, isArg: isArg, offset: calcOffset(t, frameSize, fromStack) }
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

func (v *LocalVar) OffsetedAddr(offset int) string {
    offset = v.offset + offset
    switch t := v.typ.(type) {
    case types.StrType:
        offset -= int(types.Ptr_Size)

    case types.StructType:
        if v.isArg && types.IsBigStruct(v.typ) {
            for i := len(t.Types)-1; i > 0; i-- {
                offset -= int(types.Ptr_Size)
            }
        } else {
            for i := len(t.Types)-1; i > 0; i-- {
                offset -= int(t.Types[i].Size())
            }
        }
    }

    if offset > 0 {
        return fmt.Sprintf("rbp+%d", offset)
    } else if offset < 0 {
        return fmt.Sprintf("rbp%d", offset)
    } else {
        return "rbp"
    }
}

func (v *LocalVar) Addr(fieldNum int) string {
    offset := v.offset

    switch t := v.typ.(type) {
    case types.StrType:
        if fieldNum == 0 {
            offset -= int(types.Ptr_Size)
        }

    case types.StructType:
        if v.isArg && types.IsBigStruct(v.typ) {
            for i := len(t.Types)-1; i > fieldNum; i-- {
                offset -= int(types.Ptr_Size)
            }
        } else {
            for i := len(t.Types)-1; i > fieldNum; i-- {
                offset -= int(t.Types[i].Size())
            }
        }
    }

    if offset > 0 {
        return fmt.Sprintf("rbp+%d", offset)
    } else if offset < 0 {
        return fmt.Sprintf("rbp%d", offset)
    } else {
        return "rbp"
    }
}


func (v *LocalVar) DefVal(file *os.File, val token.Token) {
    VarSetVal(file, v, val)
}

func calcOffset(t types.Type, frameSize uint, fromStack bool) int {
    if fromStack {
        offset := types.Ptr_Size + frameSize
        if t,ok := t.(types.StructType); ok {
            offset += uint(len(t.Types)) * types.Ptr_Size
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a struct but got %v (for calculating offset)", t)
            os.Exit(1)
        }

        return int(offset)
    } else {
        switch t := t.(type) {
        case types.StrType:
            return -int(frameSize + types.I32_Size)

        case types.StructType:
            return -int(frameSize + t.Types[0].Size())

        default:
            return -int(frameSize + t.Size())
        }
    }
}
