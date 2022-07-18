package vars

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/types/str"
    "gorec/asm/x86_64"
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
    v.SetVal(file, val)
}

func (v *LocalVar) DefExpr(file *os.File) {
    v.SetExpr(file)
}

func (v *LocalVar) DefVar(file *os.File, name token.Token) {
    v.SetVar(file, name)
}

func (v *LocalVar) SetVal(file *os.File, val token.Token) {
    switch v.Type.GetKind() {
    case types.Str:
        strIdx := str.Add(val)

        file.WriteString(asm.MovDerefVal(v.Addr(0), types.Ptr_Size, fmt.Sprintf("_str%d", strIdx)))
        file.WriteString(asm.MovDerefVal(v.Addr(1), types.I32_Size, fmt.Sprint(str.GetSize(strIdx))))

    case types.Ptr:
        if val.Type == token.Name {
            Write(file, fmt.Sprintf("lea rax, [%s]\n", val.Str))
            Write(file, asm.MovDerefReg(v.Addr(0), v.GetType().Size(), asm.RegA))
        } else {
            file.WriteString(asm.MovDerefVal(v.Addr(0), v.Type.Size(), val.Str))
        }

    case types.Bool:
        if val.Str == "true" { val.Str = "1" } else { val.Str = "0" }
        fallthrough

    case types.I32:
        file.WriteString(asm.MovDerefVal(v.Addr(0), v.Type.Size(), val.Str))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name.Str)
        fmt.Fprintln(os.Stderr, "\t" + v.Name.At())
        os.Exit(1)
    }
}

func (v *LocalVar) SetVar(file *os.File, name token.Token) {
    if v.Name.Str == name.Str {
        fmt.Fprintln(os.Stderr, "[WARNING] assigning a variable to itself is redundant")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        return
    }

    other := GetVar(name.Str)
    if other == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    switch v.Type.GetKind() {
    case types.Str:
        file.WriteString(asm.MovDerefDeref(v.Addr(0), other.Addr(0), types.Ptr_Size, asm.RegA))
        file.WriteString(asm.MovDerefDeref(v.Addr(1), other.Addr(1), types.I32_Size, asm.RegA))

    case types.I32, types.Bool:
        file.WriteString(asm.MovDerefDeref(v.Addr(0), other.Addr(0), v.Type.Size(), asm.RegA))

    case types.Ptr:
        file.WriteString(asm.MovDerefVal(v.Addr(0), v.Type.Size(), other.Addr(0)))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name.Str)
        fmt.Fprintln(os.Stderr, "\t" + v.Name.At())
        os.Exit(1)
    }
}

func (v *LocalVar) SetExpr(file *os.File) {
    if v.Type.GetKind() == types.Str {
        Write(file, asm.MovDerefReg(v.Addr(0), types.Ptr_Size, asm.RegA))
        Write(file, asm.MovDerefReg(v.Addr(1), types.I32_Size, asm.RegB))
    } else {
        Write(file, asm.MovDerefReg(v.Addr(0), v.GetType().Size(), asm.RegA))
    }
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

func varInCurScope(name string) bool {
    for _,v := range scopes[len(scopes)-1].vars {
        if v.Name.Str == name {
            return true
        }
    }

    return false
}

func declareLocal(varname token.Token, vartype types.Type) {
    if varInCurScope(varname.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] local var \"%s\" is already declared in this scope\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    if constInCurScope(varname.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] local const \"%s\" is already declared in this scope\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    scopes[len(scopes)-1].vars = append(scopes[len(scopes)-1].vars, LocalVar{
        Name: varname,
        Type: vartype,
        offset: calcOffset(vartype),
    })
}
