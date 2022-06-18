package vars

import (
    "fmt"
    "gorec/asm/x86_64"
    "gorec/token"
    "gorec/types"
    "gorec/types/str"
    "os"
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
    switch v.Type.GetKind() {
    case types.Str:
        strIdx := str.Add(val)
        file.WriteString(asm.MovLocVarVal(v.offset+types.Ptr_Size, types.Ptr_Size, fmt.Sprintf("_str%d", strIdx)))
        file.WriteString(asm.MovLocVarVal(v.offset, types.I32_Size, fmt.Sprintf("%d", str.GetSize(strIdx))))

    case types.Bool:
        if val.Str == "true" { val.Str = "1" } else { val.Str = "0" }
        fallthrough

    case types.I32:
        file.WriteString(asm.MovLocVarVal(v.offset, v.GetType().Size(), val.Str))

    case types.Ptr:
        fmt.Fprintln(os.Stderr, "TODO defLocalVal PtrType")
        os.Exit(1)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name.Str)
        os.Exit(1)
    }
}

func (v *LocalVar) DefExpr(file *os.File) {
    file.WriteString(asm.MovLocVarReg(v.offset, v.GetType().Size(), asm.RegA))
}

func (v *LocalVar) DefVar(file *os.File, name token.Token) {
    v.SetVar(file, name)
}

func (v *LocalVar) SetVal(file *os.File, val token.Token) {
    if v.Type.GetKind() == types.Str {
        strIdx := str.Add(val)

        Write(file, asm.MovDerefVal(v.Addr(0), types.Ptr_Size, fmt.Sprintf("_str%d", strIdx)))
        Write(file, asm.MovDerefVal(v.Addr(1), types.I32_Size, fmt.Sprint(str.GetSize(strIdx))))
    } else {
        Write(file, asm.MovDerefVal(v.Addr(0), v.Type.Size(), val.Str))
    }
}

func (v *LocalVar) SetVar(file *os.File, name token.Token) {
    if v.Name.Str == name.Str { return } // redundant

    if other := GetVar(name.Str); other != nil {
        switch v.Type.GetKind() {
        case types.Str:
            file.WriteString(asm.MovDerefDeref(v.Addr(0), other.Addr(0), types.Ptr_Size, asm.RegA))
            file.WriteString(asm.MovDerefDeref(v.Addr(1), other.Addr(1), types.I32_Size, asm.RegA))

        case types.I32, types.Bool:
            file.WriteString(asm.MovDerefDeref(v.Addr(0), other.Addr(0), v.Type.Size(), asm.RegA))

        case types.Ptr:
            file.WriteString(asm.MovDerefVal(v.Addr(0), v.Type.Size(), other.Addr(0)))

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", name.Str)
            os.Exit(1)
        }
    }
}

func (v *LocalVar) SetExpr(file *os.File) {
    Write(file, asm.MovDerefReg(v.Addr(0), v.GetType().Size(), asm.RegA))
}


func GetLastOffset() int {
    vars := scopes[len(scopes)-1].vars
    return vars[len(vars)-1].offset
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

func inCurScope(name string) bool {
    for _,v := range scopes[len(scopes)-1].vars {
        if v.Name.Str == name {
            return true
        }
    }

    return false
}

func declareLocal(varname token.Token, vartype types.Type) {
    if inCurScope(varname.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] local var \"%s\" is already declared in this scope\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    scopes[len(scopes)-1].vars = append(scopes[len(scopes)-1].vars, LocalVar{
        Name: varname,
        Type: vartype,
        offset: calcOffset(vartype),
    })
}
