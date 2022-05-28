package vars

import (
    "fmt"
    "os"
    "gorec/str"
    "gorec/types"
    "gorec/token"
)


var vars []Var

var globalDefines []string
var preMain []string

type Var struct {
    Name token.Token
    Type types.Type
    IsLocal bool
    Offset int
}

func (v Var) String() string {
    return fmt.Sprintf("{%s %s local:%t}", v.Name.Str, v.Type.Readable(), v.IsLocal)
}

func ShowVars() {
    for _, v := range vars {
        fmt.Printf("%v\n", v)
    }
}

func Add(v Var) {
    vars = append(vars, v)
}

func GetVar(varname string) *Var {
    for _, v := range vars {
        if v.Name.Str == varname {
            return &v
        }
    }

    return nil
}

func GetLastVar() *Var {
    return &vars[len(vars)-1]
}

func GetVarIdx(varname string) int {
    for i, v := range vars {
        if v.Name.Str == varname {
            return i
        }
    }

    return -1
}

func (v *Var) Get() string {
    if v.IsLocal {
        return fmt.Sprintf("QWORD [rbp-%d]", v.Offset)
    }

    return fmt.Sprintf("QWORD [%s]", v.Name.Str)
}

func (v *Var) Gets() (string, string) {
    if v.IsLocal {
        return fmt.Sprintf("QWORD [rbp-%d]", v.Offset), fmt.Sprintf("QWORD [rbp-%d]", v.Offset+8)
    }

    return fmt.Sprintf("QWORD [%s]", v.Name.Str), fmt.Sprintf("QWORD [%s+8]", v.Name.Str)
}


var localVarOffset int = 0
func calcOffset(vartype types.Type) (offset int) {
    if !InGlobalScope() {
        if vartype == types.Str {
            offset = localVarOffset + 8
        } else {
            offset = localVarOffset + vartype.Size()
        }

        localVarOffset += vartype.Size()
    }

    return offset
}

func removeLocalVars() {
    localVarOffset = 0

    if len(vars) == scopes[curScope].localStartIdx || len(vars) == 0 {
        return
    } else if scopes[curScope].localStartIdx == 0 {
        vars = nil
    }

    vars = vars[:scopes[curScope].localStartIdx]
}


func Write(asm *os.File, s string) {
    if InGlobalScope() {
        preMain = append(preMain, s)
    } else {
        asm.WriteString(s)
    }
}

func WriteDefine(asm *os.File, v *Var, val string) {
    if InGlobalScope() {
        globalDefines = append(globalDefines, fmt.Sprintf("%s: dq %s\n", v.Name.Str, val))
    } else {
        asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], %s\n", v.Offset, val))
    }
}

func DefineGlobalVars(asm *os.File) {
    for _, s := range globalDefines {
        asm.WriteString(s)
    }
}

func InitVarWithExpr(asm *os.File) {
    for _, s := range preMain {
        asm.WriteString(s)
    }
}


func Declare(varname token.Token, vartype types.Type) {
    // maybe implement shadowing later (TODO)
    if GetVar(varname.Str) != nil {
        fmt.Fprintf(os.Stderr, "[ERROR] a variable with the name \"%s\" is already declared\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    vars = append(vars, Var{ Name: varname, Type: vartype, IsLocal: !InGlobalScope(), Offset: calcOffset(vartype) })
}

func DefineByReg(asm *os.File, varname token.Token, reg string) {
    v := GetVar(varname.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot define var \"%s\" (is not declared)\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    if InGlobalScope() {
        globalDefines = append(globalDefines, fmt.Sprintf("%s: dq 0\n", v.Name.Str))
        preMain = append(preMain, fmt.Sprintf("mov QWORD [%s], %s\n", v.Name.Str, reg))
    } else {
        asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], %s\n", v.Offset, reg))
    }
}

func DefineByVal(asm *os.File, varname token.Token, value token.Token) {
    v := GetVar(varname.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot define var \"%s\" (is not declared)\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    if v.Name.Str == value.Str {
        fmt.Fprintln(os.Stderr, "[ERROR] cannot define a variable with itself")
        fmt.Fprintln(os.Stderr, "\t" + value.At())
        os.Exit(1)
    }

    if value.Type == token.Boolean || value.Type == token.Number || value.Type == token.Str {
        const _ uint = 3 - types.TypesCount
        switch v.Type {
        case types.Str:
            strIdx := str.Add(value.Str)
            if InGlobalScope() {
                globalDefines = append(globalDefines, fmt.Sprintf("%s: dq str%d, %d\n", v.Name.Str, strIdx, str.GetSize(strIdx)))
            } else {
                asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], str%d\n", v.Offset, strIdx))
                asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], %d\n", v.Offset+8, str.GetSize(strIdx)))
            }

        case types.I32:
            WriteDefine(asm, v, value.Str)

        case types.Bool:
            if value.Str == "true" {
                WriteDefine(asm, v, "1")
            } else {
                WriteDefine(asm, v, "0")
            }

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name.Str)
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

    asm.WriteString(fmt.Sprintf("mov %s, %s\n", v.Get(), reg))
}

func AssignByVal(asm *os.File, varname token.Token, value token.Token) {
    v := GetVar(varname.Str)

    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot assign var \"%s\" is not declared\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    const _ uint = 3 - types.TypesCount
    switch v.Type {
    case types.Str:
        strIdx := str.Add(value.Str)
        s1, s2 := v.Gets()
        asm.WriteString(fmt.Sprintf("mov %s, str%d\n", s1, strIdx))
        asm.WriteString(fmt.Sprintf("mov %s, %d\n",    s2, str.GetSize(strIdx)))

    case types.I32, types.Bool:
        asm.WriteString(fmt.Sprintf("mov %s, %s\n", v.Get(), value.Str))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name.Str)
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
    if v.Name.Str == value.Str {
        return
    }

    // TODO: check if var is defined
    if otherVar := GetVar(value.Str); otherVar != nil {
        const _ uint = 3 - types.TypesCount
        switch v.Type {
        case types.Str:
            vS1, vS2 := v.Gets()
            otherS1, otherS2 := otherVar.Gets()
            asm.WriteString(fmt.Sprintf("mov %s, %s\n", vS1, otherS1))
            asm.WriteString(fmt.Sprintf("mov %s, %s\n", vS2, otherS2))

        case types.I32, types.Bool:
            asm.WriteString(fmt.Sprintf("mov %s, %s\n", v.Get(), otherVar.Get()))

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name.Str)
            os.Exit(1)
        }
    }
}
