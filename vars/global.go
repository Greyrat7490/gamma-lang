package vars

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/types/str"
    "gorec/asm/x86_64"
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

func (v *GlobalVar) GetType() types.Type {
    return v.Type
}

func (v *GlobalVar) Addr(fieldNum int) string {
    if v.Type.GetKind() == types.Str {
        if fieldNum == 1 {
            return v.Name.Str + "+" + fmt.Sprint(types.Ptr_Size)
        }
    }

    return v.Name.Str
}


func (v *GlobalVar) DefVal(file *os.File, val token.Token) {
    if v.Name.Str == val.Str {
        fmt.Fprintln(os.Stderr, "[ERROR] cannot define a variable with itself")
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }

    switch v.Type.GetKind() {
    case types.Str:
        strIdx := str.Add(val)
        globalDefines = append(globalDefines, fmt.Sprintf("%s:\n  %s _str%d\n  %s %d\n",
            v.Name.Str, asm.GetDataSize(types.Ptr_Size), strIdx, asm.GetDataSize(types.I32_Size), str.GetSize(strIdx)))

    case types.I32:
        globalDefines = append(globalDefines, fmt.Sprintf("%s:\n  %s %s\n", v.Name.Str, asm.GetDataSize(types.I32_Size), val.Str))

    case types.Bool:
        if val.Str == "true" { val.Str = "1" } else { val.Str = "0" }
        globalDefines = append(globalDefines, fmt.Sprintf("%s:\n  %s %s\n", v.Name.Str, asm.GetDataSize(types.Bool_Size), val.Str))

    case types.Ptr:
        fmt.Fprintln(os.Stderr, "TODO defGlobalVal PtrType")
        os.Exit(1)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name.Str)
        os.Exit(1)
    }
}

func (v *GlobalVar) DefVar(file *os.File, name token.Token) {
    fmt.Fprintln(os.Stderr, "[ERROR] you cannot define global vars with another var")
    fmt.Fprintln(os.Stderr, "\t" + v.Name.At())
    os.Exit(1)
}

// TODO: remove if const evaluation is implemented
func (v *GlobalVar) DefExpr(file *os.File) {
    if v.GetType().GetKind() == types.Str {
        // should never get reached
    } else {
        size := v.Type.Size()
        globalDefines = append(globalDefines, fmt.Sprintf("%s:\n  %s 0\n", v.Name.Str, asm.GetDataSize(size)))
        preMain = append(preMain, asm.MovDerefReg(v.Addr(0), size, asm.RegA))
    }
}

func (v *GlobalVar) SetVal(file *os.File, val token.Token) {
    if v.Type.GetKind() == types.Str {
        strIdx := str.Add(val)

        Write(file, asm.MovDerefVal(v.Addr(0), types.Ptr_Size, fmt.Sprintf("_str%d", strIdx)))
        Write(file, asm.MovDerefVal(v.Addr(1), types.I32_Size, fmt.Sprint(str.GetSize(strIdx))))
    } else {
        Write(file, asm.MovDerefVal(v.Addr(0), v.Type.Size(), val.Str))
    }
}

func (v *GlobalVar) SetVar(file *os.File, name token.Token) {
    if v.Name.Str == name.Str { 
        fmt.Fprintln(os.Stderr, "[WARNING] assigning a variable to itself is redundant")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        return
    }

    if other := GetVar(name.Str); other != nil {
        switch v.Type.GetKind() {
        case types.Str:
            file.WriteString(asm.MovDerefDeref(v.Addr(0), other.Addr(0), types.Ptr_Size, asm.RegA))
            file.WriteString(asm.MovDerefDeref(v.Addr(1), other.Addr(1), types.I32_Size, asm.RegA))

        case types.I32, types.Bool:
            file.WriteString(asm.MovDerefDeref(v.Addr(0), other.Addr(0), v.Type.Size(), asm.RegA))

        case types.Ptr:
            file.WriteString(asm.MovDerefVal(v.Addr(0), v.Type.Size(), name.Str))

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", name.Str)
            os.Exit(1)
        }
    }
}

func (v *GlobalVar) SetExpr(file *os.File) {
    if v.Type.GetKind() == types.Str {
        Write(file, asm.MovDerefReg(v.Addr(0), types.Ptr_Size, asm.RegA))
        Write(file, asm.MovDerefReg(v.Addr(1), types.I32_Size, asm.RegB))
    } else {
        Write(file, asm.MovDerefReg(v.Addr(0), v.GetType().Size(), asm.RegA))
    }
}


func DefineGlobalVars(file *os.File) {
    for _, s := range globalDefines {
        file.WriteString(s)
    }
}

func InitVarWithExpr(file *os.File) {
    for _, s := range preMain {
        file.WriteString(s)
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
