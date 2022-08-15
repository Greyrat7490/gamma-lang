package vars

import (
    "os"
    "fmt"
    "strconv"
    "gamma/token"
    "gamma/types"
    "gamma/types/str"
    "gamma/types/struct"
    "gamma/asm/x86_64"
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
    switch t := v.GetType().(type) {
    case types.StrType:
        asm.MovDerefReg(file, v.Addr(0), types.Ptr_Size, asm.RegA)
        asm.MovDerefReg(file, v.Addr(1), types.I32_Size, asm.RegB)

    case types.StructType:
        for i,t := range t.Types {
            asm.MovDerefReg(file, v.Addr(i), t.Size(), uint8(i))
        }

    default:
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

    switch t := t.(type) {
    case types.StrType:
        asm.MovDerefDeref(file, v.Addr(0), other.Addr(0), types.Ptr_Size, asm.RegA)
        asm.MovDerefDeref(file, v.Addr(1), other.Addr(1), types.I32_Size, asm.RegA)

    case types.StructType:
        for i,t := range t.Types {
            asm.MovDerefDeref(file, v.Addr(i), other.Addr(i), t.Size(), asm.RegA)
        }

    case types.I32Type, types.BoolType:
        asm.MovDerefDeref(file, v.Addr(0), other.Addr(0), t.Size(), asm.RegA)

    case types.PtrType, types.ArrType:
        asm.MovDerefVal(file, v.Addr(0), t.Size(), other.Addr(0))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet\n", t)
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

    case types.Arr:
        if idx,err := strconv.ParseUint(val.Str, 10, 64); err == nil {
            asm.MovDerefVal(file, v.Addr(0), types.Ptr_Size, fmt.Sprintf("_arr%d", idx))
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected size of array to be a Number but got %v\n", val)
            fmt.Fprintln(os.Stderr, "\t" + val.At())
            os.Exit(1)
        }

    case types.Struct:
        if t,ok := t.(types.StructType); ok {
            if val.Type != token.Number {
                fmt.Fprintf(os.Stderr, "[ERROR] expected a Number but got %v\n", val)
                fmt.Fprintln(os.Stderr, "\t" + val.At())
                os.Exit(1)
            }

            idx,_ := strconv.ParseUint(val.Str, 10, 64)

            for i,val := range structLit.GetValues(idx) {
                asm.MovDerefVal(file, v.Addr(i), t.Types[i].Size(), val.Str)
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) expected StructType but got %v\n", t)
            os.Exit(1)
        }

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
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet\n", t)
        os.Exit(1)
    }
}


func DerefSetVal(file *os.File, typ types.Type, val token.Token) {
    switch t := typ.(type) {
    case types.StrType:
        strIdx := str.Add(val)

        asm.MovDerefVal(file, "rax", types.Ptr_Size, fmt.Sprintf("_str%d\n", strIdx))
        asm.MovDerefVal(file, fmt.Sprintf("rax+%d", types.Ptr_Size), types.I32_Size, fmt.Sprintf("%d\n", str.GetSize(strIdx)))

    case types.ArrType:
        if idx,err := strconv.ParseUint(val.Str, 10, 64); err == nil {
            asm.MovDerefVal(file, "rax", types.Ptr_Size, fmt.Sprintf("_arr%d", idx))
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected size of array to be a Number but got %v\n", val)
            fmt.Fprintln(os.Stderr, "\t" + val.At())
            os.Exit(1)
        }

    case types.PtrType:
        if val.Type == token.Name {
            file.WriteString(fmt.Sprintf("lea rbx, [%s]\n", val.Str))
            asm.MovDerefReg(file, "rax", t.Size(), asm.RegB)
        } else {
            asm.MovDerefVal(file, "rax", t.Size(), val.Str)
        }

    case types.StructType:
        if val.Type != token.Number {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a Number but got %v\n", val)
            fmt.Fprintln(os.Stderr, "\t" + val.At())
            os.Exit(1)
        }

        idx,_ := strconv.ParseUint(val.Str, 10, 64)

        for i,val := range structLit.GetValues(idx) {
            size := t.Types[i].Size()
            asm.MovDerefVal(file, asm.GetOffsetedReg(asm.RegA, types.Ptr_Size, size), size, val.Str)
        }

    case types.BoolType:
        if val.Str == "true" {
            asm.MovDerefVal(file, "rax", types.Bool_Size, "1")
        } else {
            asm.MovDerefVal(file, "rax", types.Bool_Size, "0")
        }

    case types.I32Type:
        asm.MovDerefVal(file, "rax", types.I32_Size, val.Str)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet\n", t)
    }
}

func DerefSetVar(file *os.File, other Var) {
    switch t := other.GetType().(type) {
    case types.StrType:
        asm.MovDerefDeref(file, asm.GetReg(asm.RegD, types.Ptr_Size), other.Addr(0), types.Ptr_Size, asm.RegA)
        asm.MovDerefDeref(file, asm.GetOffsetedReg(asm.RegD, types.Ptr_Size, types.Ptr_Size), other.Addr(1), types.I32_Size, asm.RegA)
    case types.StructType:
        var offset uint = 0
        for i,t := range t.Types {
            asm.MovDerefDeref(file, asm.GetOffsetedReg(asm.RegD, types.Ptr_Size, offset), other.Addr(i), t.Size(), asm.RegA)
            offset += t.Size()
        }
    default:
        asm.MovDerefDeref(file, asm.GetReg(asm.RegD, types.Ptr_Size), other.Addr(0), other.GetType().Size(), asm.RegA)
    }
}
