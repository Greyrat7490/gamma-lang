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
    DefVar(file *os.File, name token.Token)
    DefExpr(file *os.File)

    SetVal(file *os.File, val token.Token)
    SetVar(file *os.File, name token.Token)
    SetExpr(file *os.File)

    Addr(fieldNum int) string
    String()  string
    GetType() types.Type
}


func GetVar(name string) Var {
    for i := len(scopes)-1; i >= 0; i-- {
        for _, v := range scopes[i].vars {
            if v.Name.Str == name {
                return &v
            }
        }
    }

    for _, v := range globalVars {
        if v.Name.Str == name {
            return &v
        }
    }

    return nil
}

func Write(file *os.File, s string) {
    if InGlobalScope() {
        preMain = append(preMain, s)
    } else {
        file.WriteString(s)
    }
}

func AddrToRax(file *os.File, name token.Token) {
    v := GetVar(name.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] variable \"%s\" is not declared\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if _,ok := v.(*GlobalVar); ok {
        Write(file, fmt.Sprintf("mov rax, %s\n", v.Addr(0)))
    } else {
        Write(file, fmt.Sprintf("lea rax, [%s]\n", v.Addr(0)))
    }
}

func DecVar(varname token.Token, vartype types.Type) {
    if varname.Str[0] == '_' {
        fmt.Fprintln(os.Stderr, "[ERROR] variable names starting with \"_\" are reserved for the compiler")
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    if InGlobalScope() {
        declareGlobal(varname, vartype)
    } else {
        declareLocal(varname, vartype)
    }
}


func DefPtrWithVar(file *os.File, name token.Token, otherName token.Token) {
    v := GetVar(name.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot define var \"%s\" (is not declared)\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if _, ok := v.(*GlobalVar); ok {
        globalDefines = append(globalDefines, fmt.Sprintf("%s: dq %s\n", name.Str, otherName.Str))
        return
    }

    if v, ok := v.(*LocalVar); ok {
        AddrToRax(file, otherName)
        v.DefExpr(file)
        return
    }
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

func DerefSetVar(file *os.File, name token.Token) {
    other := GetVar(name.Str)

    if other == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] variable \"%s\" is not declared\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if other.GetType().GetKind() == types.Str {
        file.WriteString(asm.MovDerefDeref("rax", other.Addr(0), types.Ptr_Size, asm.RegB))
        file.WriteString(asm.MovDerefDeref(fmt.Sprintf("rax+%d", types.Ptr_Size), other.Addr(types.Ptr_Size), types.I32_Size, asm.RegB))
    } else {
        file.WriteString(asm.MovDerefDeref("rax", other.Addr(0), other.GetType().Size(), asm.RegB))
    }
}
