package vars

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/types/str"
    "gorec/asm/x86_64"
)

type Var interface {
    DefVal(file *os.File, val token.Token)
    DefVar(file *os.File, other Var)
    DefExpr(file *os.File)

    SetVal(file *os.File, val token.Token)
    SetVar(file *os.File, other Var)
    SetExpr(file *os.File)

    Addr(fieldNum int) string
    String() string
    GetName() token.Token
    GetType() types.Type
}


func GetVar(name string) Var {
    scope := curScope

    for scope != nil {
        if v,ok := scope.vars[name]; ok {
            return v
        }

        scope = scope.parent
    }

    return nil
}

func varNameTaken(name string) bool {
    if _,ok := curScope.vars[name]; ok {
        return ok
    }

    return false
}

func DecVar(varname token.Token, vartype types.Type) Var {
    if varname.Str[0] == '_' {
        fmt.Fprintln(os.Stderr, "[ERROR] names starting with \"_\" are reserved for the compiler")
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    if varNameTaken(varname.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is already declared in this scope\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    if constNameTaken(varname.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] const \"%s\" is already declared in this scope\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    var v Var
    if InGlobalScope() {
        v = &GlobalVar{ Name: varname, Type: vartype }
    } else {
        v = &LocalVar{ Name: varname, Type: vartype, offset: calcOffset(vartype) }
    }

    curScope.vars[varname.Str] = v
    return v
}


func DerefSetVal(file *os.File, val token.Token, size int) {
    switch val.Type {
    case token.Str:
        strIdx := str.Add(val)

        file.WriteString(asm.MovDerefVal("rax", types.Ptr_Size, fmt.Sprintf("_str%d\n", strIdx)))
        file.WriteString(asm.MovDerefVal(fmt.Sprintf("rax+%d", types.Ptr_Size), types.I32_Size, fmt.Sprintf("%d\n", str.GetSize(strIdx))))
    case token.Boolean:
        if val.Str == "true" { val.Str = "1" } else { val.Str = "0" }
        fallthrough
    default:
        file.WriteString(asm.MovDerefVal("rax", size, val.Str))
    }
}

func DerefSetVar(file *os.File, other Var) {
    if other.GetType().GetKind() == types.Str {
        file.WriteString(asm.MovDerefDeref("rax", other.Addr(0), types.Ptr_Size, asm.RegB))
        file.WriteString(asm.MovDerefDeref(fmt.Sprintf("rax+%d", types.Ptr_Size), other.Addr(types.Ptr_Size), types.I32_Size, asm.RegB))
    } else {
        file.WriteString(asm.MovDerefDeref("rax", other.Addr(0), other.GetType().Size(), asm.RegB))
    }
}
