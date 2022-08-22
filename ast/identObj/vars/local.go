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
        if fieldNum == 1 {
            offset += int(types.Ptr_Size)
        }

    case types.StructType:
        for i := 0; i < fieldNum; i++ {
            offset += int(t.Types[i].Size())
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
        return int(types.Ptr_Size + frameSize + 7) & ^7
    }

    return -int(frameSize + t.Size())
}
