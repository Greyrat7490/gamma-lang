package vars

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
)

var localVarOffset int = 0

type LocalVar struct {
    Name token.Token
    Type types.Type
    offset int
}

func (v *LocalVar) String() string {
    return fmt.Sprintf("{%s %v}", v.Name.Str, v.Type)
}

func (v *LocalVar) GetName() token.Token {
    return v.Name
}

func (v *LocalVar) GetType() types.Type {
    return v.Type
}

func (v *LocalVar) Addr(fieldNum int) string {
    if v.Type.GetKind() == types.Str {
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
    if !InGlobalScope() {
        if vartype.GetKind() == types.Str {
            offset = localVarOffset + types.I32_Size
        } else {
            offset = localVarOffset + vartype.Size()
        }

        localVarOffset += vartype.Size()
    }

    return offset
}
