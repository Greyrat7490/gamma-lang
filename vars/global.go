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
    return fmt.Sprintf("%s [%s]", GetWord(v.Type.Size()), v.Name.Str)
}
func (v *GlobalVar) Gets() (string, string) {
    return fmt.Sprintf("%s [%s]",    GetWord(types.Ptr_Size), v.Name.Str),
           fmt.Sprintf("%s [%s+%d]", GetWord(types.I32_Size), v.Name.Str, types.Ptr_Size)
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

    switch v.Type.GetKind() {
    case types.Str:
        strIdx := str.Add(val)
        globalDefines = append(globalDefines, fmt.Sprintf("%s:\n  %s _str%d\n  %s %d\n",
            v.Name.Str, GetDataSize(types.Ptr_Size), strIdx, GetDataSize(types.I32_Size), str.GetSize(strIdx)))

    case types.I32:
        globalDefines = append(globalDefines, fmt.Sprintf("%s:\n  %s %s\n", v.Name.Str, GetDataSize(types.I32_Size), val.Str))

    case types.Bool:
        if val.Str == "true" { val.Str = "1" } else { val.Str = "0" }
        globalDefines = append(globalDefines, fmt.Sprintf("%s:\n  %s %s\n", v.Name.Str, GetDataSize(types.I32_Size), val.Str))

    case types.Ptr:
        fmt.Fprintln(os.Stderr, "TODO defGlobalVal PtrType")
        os.Exit(1)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name.Str)
        os.Exit(1)
    }
}

func defGlobalExpr(asm *os.File, v *GlobalVar, reg RegGroup) {
    size := v.Type.Size()
    globalDefines = append(globalDefines, fmt.Sprintf("%s:\n  %s 0\n", v.Name.Str, GetDataSize(size)))
    preMain = append(preMain, fmt.Sprintf("mov %s [%s], %s\n", GetWord(size), v.Name.Str, GetReg(reg, size)))
}
