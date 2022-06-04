package vars

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/str"
)

var globalVars []GlobalVar

var globalDefines []string
var preMain []string

type GlobalVar struct {
    Name token.Token
    Type types.Type
}

func (v *GlobalVar) String() string {
    return fmt.Sprintf("{%s %s}", v.Name.Str, v.Type)
}
func (v *GlobalVar) Get() string {
    return fmt.Sprintf("QWORD [%s]", v.Name.Str)
}
func (v *GlobalVar) Gets() (string, string) {
    return fmt.Sprintf("QWORD [%s]", v.Name.Str), fmt.Sprintf("QWORD [%s+8]", v.Name.Str)
}
func (v *GlobalVar) GetType() types.Type {
    return v.Type
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

func isGlobalDec(varname string) bool {
    for _, v := range globalVars {
        if v.Name.Str == varname {
            return true
        }
    }

    return false
}

func declareGlobal(varname token.Token, vartype types.Type) {
    if isGlobalDec(varname.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] a variable with the name \"%s\" is already declared\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    globalVars = append(globalVars, GlobalVar{ Name: varname, Type: vartype })
}

func defGlobalVal(asm *os.File, v *GlobalVar, val token.Token) {
    if v.Name.Str == val.Str {
        fmt.Fprintln(os.Stderr, "[ERROR] cannot define a variable with itself")
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }

    switch v.Type.(type) {
    case types.StrType:
        strIdx := str.Add(val.Str)
        globalDefines = append(globalDefines, fmt.Sprintf("%s: dq str%d, %d\n", v.Name.Str, strIdx, str.GetSize(strIdx)))

    case types.I32Type:
        globalDefines = append(globalDefines, fmt.Sprintf("%s: dq %s\n", v.Name.Str, val.Str))

    case types.BoolType:
        if val.Str == "true" { val.Str = "1" } else { val.Str = "0" }
        globalDefines = append(globalDefines, fmt.Sprintf("%s: dq %s\n", v.Name.Str, val.Str))

    case types.PtrType:
        fmt.Fprintln(os.Stderr, "TODO defGlobalVal PtrType")
        os.Exit(1)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name.Str)
        os.Exit(1)
    }
}

func defGlobalExpr(asm *os.File, v *GlobalVar, reg string) {
    globalDefines = append(globalDefines, fmt.Sprintf("%s: dq 0\n", v.Name.Str))
    preMain = append(preMain, fmt.Sprintf("mov QWORD [%s], %s\n", v.Name.Str, reg))
}
