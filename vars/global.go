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
    Name token.Token
    Type types.Type
}


func (v *GlobalVar) String() string {
    return fmt.Sprintf("{%s %s}", v.Name.Str, v.Type)
}

func (v *GlobalVar) GetName() token.Token {
    return v.Name
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
    switch v.Type.GetKind() {
    case types.Str:
        strIdx := str.Add(val)
        globalDefines = append(globalDefines, fmt.Sprintf("%s:\n  %s _str%d\n  %s %d\n",
            v.Name.Str, asm.GetDataSize(types.Ptr_Size), strIdx, asm.GetDataSize(types.I32_Size), str.GetSize(strIdx)))

    case types.Bool:
        if val.Str == "true" { val.Str = "1" } else { val.Str = "0" }
        fallthrough

    case types.I32, types.Ptr:
        globalDefines = append(globalDefines, fmt.Sprintf("%s:\n  %s %s\n", v.Name.Str, asm.GetDataSize(v.Type.Size()), val.Str))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name.Str)
        fmt.Fprintln(os.Stderr, "\t" + v.Name.At())
        os.Exit(1)
    }
}

func (v *GlobalVar) DefVar(file *os.File, other Var) {
    fmt.Fprintln(os.Stderr, "[ERROR] you cannot define global vars with another var")
    fmt.Fprintln(os.Stderr, "\t" + v.Name.At())
    os.Exit(1)
}

func (v *GlobalVar) DefExpr(file *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] defining a global variable with a non const expr is not allowed")
    fmt.Fprintln(os.Stderr, "\t" + v.Name.At())
    os.Exit(1)
}

func (v *GlobalVar) SetVal(file *os.File, val token.Token) {
    switch v.Type.GetKind() {
    case types.Str:
        strIdx := str.Add(val)

        file.WriteString(asm.MovDerefVal(v.Addr(0), types.Ptr_Size, fmt.Sprintf("_str%d", strIdx)))
        file.WriteString(asm.MovDerefVal(v.Addr(1), types.I32_Size, fmt.Sprint(str.GetSize(strIdx))))

    case types.Bool:
        if val.Str == "true" { val.Str = "1" } else { val.Str = "0" }
        fallthrough

    case types.I32, types.Ptr:
        file.WriteString(asm.MovDerefVal(v.Addr(0), v.Type.Size(), val.Str))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name.Str)
        fmt.Fprintln(os.Stderr, "\t" + v.Name.At())
        os.Exit(1)
    }
}

func (v *GlobalVar) SetVar(file *os.File, other Var) {
    if v.Name.Str == other.GetName().Str {
        fmt.Fprintln(os.Stderr, "[WARNING] assigning a variable to itself is redundant")
        fmt.Fprintln(os.Stderr, "\t" + other.GetName().At()) // TODO correct position
        return
    }

    switch v.Type.GetKind() {
    case types.Str:
        file.WriteString(asm.MovDerefDeref(v.Addr(0), other.Addr(0), types.Ptr_Size, asm.RegA))
        file.WriteString(asm.MovDerefDeref(v.Addr(1), other.Addr(1), types.I32_Size, asm.RegA))

    case types.I32, types.Bool:
        file.WriteString(asm.MovDerefDeref(v.Addr(0), other.Addr(0), v.Type.Size(), asm.RegA))

    case types.Ptr:
        file.WriteString(asm.MovDerefVal(v.Addr(0), v.Type.Size(), other.GetName().Str))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", other.GetName().Str)
        os.Exit(1)
    }
}

func (v *GlobalVar) SetExpr(file *os.File) {
    if v.Type.GetKind() == types.Str {
        file.WriteString(asm.MovDerefReg(v.Addr(0), types.Ptr_Size, asm.RegA))
        file.WriteString(asm.MovDerefReg(v.Addr(1), types.I32_Size, asm.RegB))
    } else {
        file.WriteString(asm.MovDerefReg(v.Addr(0), v.GetType().Size(), asm.RegA))
    }
}


func DefineGlobalVars(file *os.File) {
    for _, s := range globalDefines {
        file.WriteString(s)
    }
}
