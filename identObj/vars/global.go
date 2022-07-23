package vars

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/types/str"
    "gorec/asm/x86_64"
)

var globalDefines []string

type GlobalVar struct {
    name token.Token
    vartype types.Type
}

func CreateGlobalVar(name token.Token, t types.Type) GlobalVar {
    return GlobalVar{ name: name, vartype: t }
}

func (v *GlobalVar) SetType(t types.Type) {
    if v.vartype == nil {
        v.vartype = t
    }
}

func (v *GlobalVar) String() string {
    return fmt.Sprintf("{%s %s}", v.name.Str, v.vartype)
}

func (v *GlobalVar) GetName() token.Token {
    return v.name
}

func (v *GlobalVar) GetType() types.Type {
    return v.vartype
}

func (v *GlobalVar) Addr(fieldNum int) string {
    if v.vartype.GetKind() == types.Str {
        if fieldNum == 1 {
            return v.name.Str + "+" + fmt.Sprint(types.Ptr_Size)
        }
    }

    return v.name.Str
}


func (v *GlobalVar) DefVal(file *os.File, val token.Token) {
    switch v.vartype.GetKind() {
    case types.Str:
        strIdx := str.Add(val)
        globalDefines = append(globalDefines, fmt.Sprintf("%s:\n  %s _str%d\n  %s %d\n",
            v.name.Str, asm.GetDataSize(types.Ptr_Size), strIdx, asm.GetDataSize(types.I32_Size), str.GetSize(strIdx)))

    case types.Bool:
        if val.Str == "true" { val.Str = "1" } else { val.Str = "0" }
        fallthrough

    case types.I32, types.Ptr:
        globalDefines = append(globalDefines, fmt.Sprintf("%s:\n  %s %s\n", v.name.Str, asm.GetDataSize(v.vartype.Size()), val.Str))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.name.Str)
        fmt.Fprintln(os.Stderr, "\t" + v.name.At())
        os.Exit(1)
    }
}

func DefineGlobalVars(file *os.File) {
    for _, s := range globalDefines {
        file.WriteString(s)
    }
}
