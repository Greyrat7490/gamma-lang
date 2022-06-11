package vars

import (
    "os"
    "fmt"
    "gorec/str"
    "gorec/types"
    "gorec/token"
)

type Var interface {
    Get()     string
    Gets()    (string, string)
    String()  string
    GetType() types.Type
}


func ShowVars() {
    for _, v := range globalVars {
        fmt.Printf("%v\n", v)
    }
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

func Write(asm *os.File, s string) {
    if InGlobalScope() {
        preMain = append(preMain, s)
    } else {
        asm.WriteString(s)
    }
}

func ValToRax(asm *os.File, name token.Token) {
    v := GetVar(name.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] variable \"%s\" is not declared\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    Write(asm, fmt.Sprintf("mov rax, %s\n", v.Get()))
}

func AddrToRax(asm *os.File, name token.Token) {
    v := GetVar(name.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] variable \"%s\" is not declared\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    var s string
    if v.GetType().GetKind() == types.Str {
        s,_ = v.Gets()
    } else {
        s = v.Get()
    }

    Write(asm, fmt.Sprintf("lea rax, %s\n", s))
}

func Declare(varname token.Token, vartype types.Type) {
    if InGlobalScope() {
        declareGlobal(varname, vartype)
    } else {
        declareLocal(varname, vartype)
    }
}

func DefWithExpr(asm *os.File, varname token.Token, reg string) {
    v := GetVar(varname.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot define var \"%s\" (is not declared)\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    if v, ok := v.(*GlobalVar); ok {
        defGlobalExpr(asm, v, reg)
        return
    }

    if v, ok := v.(*LocalVar); ok {
        defLocalExpr(asm, v, reg)
        return
    }
}

func DefWithVal(asm *os.File, varname token.Token, value token.Token) {
    v := GetVar(varname.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot define var \"%s\" (is not declared)\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    if !(value.Type == token.Boolean || value.Type == token.Number || value.Type == token.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a literale\n", value.Str)
        fmt.Fprintln(os.Stderr, "\t" + value.At())
        os.Exit(1)
    }

    if v, ok := v.(*GlobalVar); ok {
        defGlobalVal(asm, v, value)
        return
    }

    if v, ok := v.(*LocalVar); ok {
        defLocalVal(asm, v, value.Str)
        return
    }
}

func DefPtrWithVar(asm *os.File, name token.Token, otherName token.Token) {
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
        AddrToRax(asm, otherName)
        Write(asm, fmt.Sprintf("mov %s, rax\n", v.Get()))
        return
    }
}

func VarSetExpr(asm *os.File, destVar token.Token, reg string) {
    v := GetVar(destVar.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", destVar.Str)
        fmt.Fprintln(os.Stderr, "\t" + destVar.At())
        os.Exit(1)
    }

    Write(asm, fmt.Sprintf("mov %s, %s\n", v.Get(), reg))
}

func DerefSetVal(asm *os.File, value token.Token) {
    if value.Type == token.Str {
        strIdx := str.Add(value.Str)

        asm.WriteString(fmt.Sprintf("mov QWORD [rax], str%d\n", strIdx))
        asm.WriteString(fmt.Sprintf("mov QWORD [rax+8], %d\n", str.GetSize(strIdx)))
    } else {
        asm.WriteString(fmt.Sprintf("mov QWORD [rax], %s\n", value.Str))
    }
}

func DerefSetVar(asm *os.File, name token.Token) {
    v := GetVar(name.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot assign var \"%s\" is not declared\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if v.GetType().GetKind() == types.Str {
        s1, s2 := v.Gets()
        asm.WriteString(fmt.Sprintf("mov rbx, %s\n", s1))
        asm.WriteString("mov QWORD [rax], rbx\n")

        asm.WriteString(fmt.Sprintf("mov rbx, %s\n", s2))
        asm.WriteString("mov QWORD [rax+8], rbx\n")
    } else {
        s := v.Get()
        asm.WriteString(fmt.Sprintf("mov rbx, %s\n", s))
        asm.WriteString("mov QWORD [rax], rbx\n")
    }
}

func VarSetVal(asm *os.File, name token.Token, value token.Token) {
    v := GetVar(name.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot assign var \"%s\" is not declared\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if v.GetType().GetKind() == types.Str {
        strIdx := str.Add(value.Str)
        s1, s2 := v.Gets()

        Write(asm, fmt.Sprintf("mov %s, str%d\n", s1, strIdx))
        Write(asm, fmt.Sprintf("mov %s, %d\n",    s2, str.GetSize(strIdx)))
    } else {
        Write(asm, fmt.Sprintf("mov %s, %s\n", v.Get(), value.Str))
    }
}

func VarSetVar(asm *os.File, name token.Token, otherName token.Token) {
    v := GetVar(name.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot assign var \"%s\" is not declared\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    // skip assigning a variable to itself (redundant)
    if name.Str == otherName.Str { return }

    // TODO: check if var is defined
    if otherVar := GetVar(otherName.Str); otherVar != nil {
        switch v.GetType().GetKind() {
        case types.Str:
            vS1, vS2 := v.Gets()
            otherS1, otherS2 := otherVar.Gets()
            asm.WriteString(fmt.Sprintf("mov %s, %s\n", vS1, otherS1))
            asm.WriteString(fmt.Sprintf("mov %s, %s\n", vS2, otherS2))

        case types.I32, types.Bool:
            asm.WriteString(fmt.Sprintf("mov %s, %s\n", v.Get(), otherVar.Get()))

        case types.Ptr:
            if _, ok := otherVar.(*GlobalVar); ok {
                asm.WriteString(fmt.Sprintf("mov %s, %s\n", v.Get(), otherName.Str))
            } else {
                asm.WriteString(fmt.Sprintf("lea rax, %s\n", otherVar.Get()))
                asm.WriteString(fmt.Sprintf("mov %s, rax\n", v.Get()))
            }

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", name.Str)
            os.Exit(1)
        }
    }
}
