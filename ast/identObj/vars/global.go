package vars

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/types/str"
    "gorec/asm/x86_64"
    "gorec/asm/x86_64/nasm"
)

type GlobalVar struct {
    decPos token.Pos
    name string
    vartype types.Type
}

func CreateGlobalVar(name token.Token, t types.Type) GlobalVar {
    return GlobalVar{ name: name.Str, decPos: name.Pos, vartype: t }
}

func (v *GlobalVar) SetType(t types.Type) {
    if v.vartype != nil {
        fmt.Println("[ERROR] setting the type of a var again is not allowed")
        os.Exit(1)
    }

    v.vartype = t
}

func (v *GlobalVar) String() string {
    return fmt.Sprintf("{%s %s}", v.name, v.vartype)
}

func (v *GlobalVar) GetName() string {
    return v.name
}

func (v *GlobalVar) GetPos () token.Pos {
    return v.decPos
}

func (v *GlobalVar) GetType() types.Type {
    return v.vartype
}

func (v *GlobalVar) Addr(fieldNum int) string {
    if v.vartype.GetKind() == types.Str {
        if fieldNum == 1 {
            return v.name + "+" + fmt.Sprint(types.Ptr_Size)
        }
    }

    return v.name
}


func (v *GlobalVar) DefVal(file *os.File, val token.Token) {
    switch v.vartype.GetKind() {
    case types.Str:
        strIdx := str.Add(val)
        nasm.AddData(fmt.Sprintf("%s:\n  %s _str%d\n  %s %d",
            v.name, asm.GetDataSize(types.Ptr_Size), strIdx, asm.GetDataSize(types.I32_Size), str.GetSize(strIdx)))

    case types.Arr:
        nasm.AddData(fmt.Sprintf("%s:\n  %s _arr%s\n", v.name, asm.GetDataSize(types.Ptr_Size), val.Str))

    case types.Bool:
        if val.Str == "true" { val.Str = "1" } else { val.Str = "0" }
        fallthrough

    case types.I32, types.Ptr:
        nasm.AddData(fmt.Sprintf("%s:\n  %s %s\n", v.name, asm.GetDataSize(v.vartype.Size()), val.Str))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.name)
        fmt.Fprintln(os.Stderr, "\t" + v.decPos.At())
        os.Exit(1)
    }
}
