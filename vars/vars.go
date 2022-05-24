package vars

import (
    "fmt"
    "os"
    "gorec/str"
    "gorec/types"
    "gorec/token"
)

var InGlobalScope bool = true

var vars []Var
var globalScope []string
var preMain []string

// TODO: register allocator

type Var struct {
    Name string
    Vartype types.Type
}

func ShowVars() {
    for _, v := range vars {
        fmt.Printf("%s (type:%s)\n", v.Name, v.Vartype.Readable())
    }
}

func GetVar(varname string) *Var {
    for _, v := range vars {
        if v.Name == varname {
            return &v
        }
    }

    return nil
}

func Declare(varname token.Token, vartype types.Type) {
    // maybe implement shadowing later (TODO)
    if GetVar(varname.Str) != nil {
        fmt.Fprintf(os.Stderr, "[ERROR] a variable with the name \"%s\" is already declared\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    vars = append(vars, Var{ Name: varname.Str, Vartype: vartype })
}

func DefineByReg(asm *os.File, varname token.Token, reg string) {
    v := GetVar(varname.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot define var \"%s\" (is not declared)\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    globalScope = append(globalScope, fmt.Sprintf("%s: dq 0\n", v.Name))
    WriteVar(asm, fmt.Sprintf("mov QWORD [%s], %s\n", v.Name, reg))
}

func DefineByVal(asm *os.File, varname token.Token, value token.Token) {
    v := GetVar(varname.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot define var \"%s\" (is not declared)\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    if v.Name == value.Str {
        fmt.Fprintln(os.Stderr, "[ERROR] cannot define a variable with itself")
        fmt.Fprintln(os.Stderr, "\t" + value.At())
        os.Exit(1)
    }

    if value.Type == token.Boolean || value.Type == token.Number || value.Type == token.Str {
        const _ uint = 3 - types.TypesCount
        switch v.Vartype {
        case types.Str:
            strIdx := str.Add(value.Str)
            WriteDefine(asm, fmt.Sprintf("%s: dq str%d, %d\n", v.Name, strIdx, str.GetSize(strIdx)))

        case types.I32:
            WriteDefine(asm, fmt.Sprintf("%s: dq %s\n", v.Name, value.Str))

        case types.Bool:
            if value.Str == "true" {
                WriteDefine(asm, fmt.Sprintf("%s: dq %d\n", v.Name, 1))
            } else {
                WriteDefine(asm, fmt.Sprintf("%s: dq %d\n", v.Name, 0))
            }

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name)
        }
    }
}

func AssignByReg(asm *os.File, destVar token.Token, reg string) {
    v := GetVar(destVar.Str)

    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", destVar.Str)
        fmt.Fprintln(os.Stderr, "\t" + destVar.At())
        os.Exit(1)
    }

    WriteVar(asm, fmt.Sprintf("mov QWORD [%s], %s\n", v.Name, reg))
}

func AssignByVal(asm *os.File, varname token.Token, value token.Token) {
    v := GetVar(varname.Str)

    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot assign var \"%s\" is not declared\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    const _ uint = 3 - types.TypesCount
    switch v.Vartype {
    case types.Str:
        strIdx := str.Add(value.Str)
        WriteVar(asm, fmt.Sprintf("mov QWORD [%s], str%d\n", v.Name, strIdx))
        WriteVar(asm, fmt.Sprintf("mov QWORD [%s+8], %d\n", v.Name, str.GetSize(strIdx)))

    case types.I32, types.Bool:
        WriteVar(asm, fmt.Sprintf("mov QWORD [%s], %s\n", v.Name, value.Str))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name)
        os.Exit(1)
    }
}

func AssignByVar(asm *os.File, varname token.Token, value token.Token) {
    v := GetVar(varname.Str)

    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot assign var \"%s\" is not declared\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    // skip assigning a variable to itself (redundant)
    if v.Name == value.Str {
        return
    }

    // TODO: check if var is defined
    if otherVar := GetVar(value.Str); otherVar != nil {
        WriteVar(asm, fmt.Sprintf("mov QWORD [%s], QWORD [%s]\n", v.Name, otherVar.Name))
    }
}

func Add(v Var) {
    vars = append(vars, v)
}


func WriteVar(asm *os.File, s string) {
    if InGlobalScope {
        preMain = append(preMain, s)
    } else {
        asm.WriteString(s)
    }
}

func WriteDefine(asm *os.File, s string) {
    if InGlobalScope {
        globalScope = append(globalScope, s)
    } else {
        asm.WriteString(s)
    }
}

func DefineGlobalVars(asm *os.File) {
    for _, s := range globalScope {
        asm.WriteString(s)
    }
}

func InitVarWithExpr(asm *os.File) {
    for _, s := range preMain {
        asm.WriteString(s)
    }
}
