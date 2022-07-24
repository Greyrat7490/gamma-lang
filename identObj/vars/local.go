package vars

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
)

var localVarOffset int = 0

type LocalVar struct {
    name token.Token
    vartype types.Type
    offset int
}

func CreateLocal(name token.Token, t types.Type) LocalVar {
    return LocalVar{ name: name, vartype: t, offset: calcOffset(t) }
}

func (v *LocalVar) SetType(t types.Type) {
    if v.vartype == nil {
        v.vartype = t
    }
}

func (v *LocalVar) String() string {
    return fmt.Sprintf("{%s %v}", v.name.Str, v.vartype)
}

func (v *LocalVar) GetName() token.Token {
    return v.name
}

func (v *LocalVar) GetType() types.Type {
    return v.vartype
}

func (v *LocalVar) Addr(fieldNum int) string {
    if v.vartype.GetKind() == types.Str {
        offset := v.offset
        if fieldNum == 0 {
            offset += types.Ptr_Size
        }

        return fmt.Sprintf("rbp-%d", offset)
    }

    return fmt.Sprintf("rbp-%d", v.offset)
}


func (v *LocalVar) DefVal(file *os.File, val token.Token) {
    VarSetVal(file, v, val)
}


func calcOffset(vartype types.Type) (offset int) {
    if vartype.GetKind() == types.Str {
        offset = localVarOffset + types.I32_Size
    } else {
        offset = localVarOffset + vartype.Size()
    }

    localVarOffset += vartype.Size()

    return offset
}

func ResetLocalVarOffset() {
    localVarOffset = 0
}

func (v *LocalVar) Identobj() {}
