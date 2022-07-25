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
    SetType(t types.Type)
    Addr(fieldNum int) string

    GetType() types.Type
    GetName() string
    GetPos() token.Pos
    String() string
}

func VarSetExpr(file *os.File, v Var) {
    if v.GetType().GetKind() == types.Str {
        asm.MovDerefReg(file, v.Addr(0), types.Ptr_Size, asm.RegA)
        asm.MovDerefReg(file, v.Addr(1), types.I32_Size, asm.RegB)
    } else {
        asm.MovDerefReg(file, v.Addr(0), v.GetType().Size(), asm.RegA)
    }
}

func VarSetVar(file *os.File, v Var, other Var) {
    if v.GetName() == other.GetName() {
        fmt.Fprintln(os.Stderr, "[WARNING] assigning a variable to itself is redundant")
        fmt.Fprintln(os.Stderr, "\t" + v.GetPos().At())
        return
    }

    t := v.GetType()

    switch t.GetKind() {
    case types.Str:
        asm.MovDerefDeref(file, v.Addr(0), other.Addr(0), types.Ptr_Size, asm.RegA)
        asm.MovDerefDeref(file, v.Addr(1), other.Addr(1), types.I32_Size, asm.RegA)

    case types.I32, types.Bool:
        asm.MovDerefDeref(file, v.Addr(0), other.Addr(0), t.Size(), asm.RegA)

    case types.Ptr:
        asm.MovDerefVal(file, v.Addr(0), t.Size(), other.Addr(0))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.GetName())
        fmt.Fprintln(os.Stderr, "\t" + v.GetPos().At())
        os.Exit(1)
    }
}

func VarSetVal(file *os.File, v Var, val token.Token) {
    t := v.GetType()
    switch t.GetKind() {
    case types.Str:
        strIdx := str.Add(val)

        asm.MovDerefVal(file, v.Addr(0), types.Ptr_Size, fmt.Sprintf("_str%d", strIdx))
        asm.MovDerefVal(file, v.Addr(1), types.I32_Size, fmt.Sprint(str.GetSize(strIdx)))

    case types.Ptr:
        if val.Type == token.Name {
            file.WriteString(fmt.Sprintf("lea rax, [%s]\n", val.Str))
            asm.MovDerefReg(file, v.Addr(0), v.GetType().Size(), asm.RegA)
        } else {
            asm.MovDerefVal(file, v.Addr(0), t.Size(), val.Str)
        }

    case types.Bool:
        if val.Str == "true" { val.Str = "1" } else { val.Str = "0" }
        fallthrough

    case types.I32:
        asm.MovDerefVal(file, v.Addr(0), t.Size(), val.Str)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.GetName())
        fmt.Fprintln(os.Stderr, "\t" + v.GetPos().At())
        os.Exit(1)
    }
}


func DerefSetVal(file *os.File, val token.Token, size int) {
    switch val.Type {
    case token.Str:
        strIdx := str.Add(val)

        asm.MovDerefVal(file, "rax", types.Ptr_Size, fmt.Sprintf("_str%d\n", strIdx))
        asm.MovDerefVal(file, fmt.Sprintf("rax+%d", types.Ptr_Size), types.I32_Size, fmt.Sprintf("%d\n", str.GetSize(strIdx)))
    case token.Boolean:
        if val.Str == "true" { val.Str = "1" } else { val.Str = "0" }
        fallthrough

    default:
        asm.MovDerefVal(file, "rax", size, val.Str)
    }
}

func DerefSetVar(file *os.File, other Var) {
    if other.GetType().GetKind() == types.Str {
        asm.MovDerefDeref(file, "rax", other.Addr(0), types.Ptr_Size, asm.RegB)
        asm.MovDerefDeref(file, fmt.Sprintf("rax+%d", types.Ptr_Size), other.Addr(types.Ptr_Size), types.I32_Size, asm.RegB)
    } else {
        asm.MovDerefDeref(file, "rax", other.Addr(0), other.GetType().Size(), asm.RegB)
    }
}
