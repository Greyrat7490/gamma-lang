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
    Addr(field uint) string
    OffsetedAddr(offset int) string

    GetType() types.Type
    GetName() string
    GetPos() token.Pos
    String() string
}

func VarSetExpr(file *os.File, v Var) {
    switch t := v.GetType().(type) {
    case types.StrType:
        asm.MovDerefReg(file, v.Addr(0), types.Ptr_Size, asm.RegGroup(0))
        asm.MovDerefReg(file, v.Addr(1), types.I32_Size, asm.RegGroup(1))

    case types.StructType:
        if t.Size() > uint(8) {
            asm.MovDerefReg(file, v.Addr(0), types.Ptr_Size, asm.RegGroup(0))
            asm.MovDerefReg(file, v.OffsetedAddr(int(types.Ptr_Size)), t.Types[1].Size(), asm.RegGroup(1))
        } else {
            asm.MovDerefReg(file, v.Addr(0), t.Size(), asm.RegGroup(0))
        }

    default:
        asm.MovDerefReg(file, v.Addr(0), v.GetType().Size(), asm.RegGroup(0))
    }
}

func VarSetVar(file *os.File, v Var, other Var) {
    if v.GetName() == other.GetName() {
        fmt.Fprintln(os.Stderr, "[WARNING] assigning a variable to itself is redundant")
        fmt.Fprintln(os.Stderr, "\t" + v.GetPos().At())
        return
    }

    switch t := v.GetType().(type) {
    case types.StrType:
        asm.MovDerefDeref(file, v.Addr(0), other.Addr(0), types.Ptr_Size, asm.RegA)
        asm.MovDerefDeref(file, v.Addr(1), other.Addr(1), types.I32_Size, asm.RegA)

    case types.StructType:
        for i,t := range t.Types {
            asm.MovDerefDeref(file, v.Addr(uint(i)), other.Addr(uint(i)), t.Size(), asm.RegA)
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
    switch t := v.GetType().(type) {
    case types.StrType:
        derefSetStrVal(file, v.Addr(0), 0, val)
    case types.ArrType:
        derefSetArrVal(file, v.Addr(0), 0, val)
    case types.StructType:
        derefSetStructVal(file, t, v.Addr(0), 0, val)
    case types.PtrType:
        derefSetPtrVal(file, v.Addr(0), 0, val)
    case types.BoolType:
        derefSetBoolVal(file, v.Addr(0), 0, val)
    case types.I32Type:
        derefSetIntVal(file, v.Addr(0), 0, val)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet\n", t)
        os.Exit(1)
    }
}


func DerefSetVal(file *os.File, typ types.Type, val token.Token) {
    switch t := typ.(type) {
    case types.StrType:
        derefSetStrVal(file, "rax", 0, val)
    case types.ArrType:
        derefSetArrVal(file, "rax", 0, val)
    case types.StructType:
        derefSetStructVal(file, t, "rax", 0, val)
    case types.PtrType:
        derefSetPtrVal(file, "rax", 0, val)
    case types.BoolType:
        derefSetBoolVal(file, "rax", 0, val)
    case types.I32Type:
        derefSetIntVal(file, "rax", 0, val)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet\n", t)
        os.Exit(1)
    }
}

func DerefSetVar(file *os.File, other Var) {
    switch t := other.GetType().(type) {
    case types.StrType:
        asm.MovDerefDeref(file, asm.GetReg(asm.RegA, types.Ptr_Size), other.Addr(0), types.Ptr_Size, asm.RegB)
        asm.MovDerefDeref(file, asm.GetOffsetedReg(asm.RegA, types.Ptr_Size, int(types.Ptr_Size)), other.Addr(1), types.I32_Size, asm.RegB)
    case types.StructType:
        var offset int = 0
        for i,t := range t.Types {
            asm.MovDerefDeref(file, asm.GetOffsetedReg(asm.RegA, types.Ptr_Size, offset), other.Addr(uint(i)), t.Size(), asm.RegB)
            offset += int(t.Size())
        }
    default:
        asm.MovDerefDeref(file, asm.GetReg(asm.RegA, types.Ptr_Size), other.Addr(0), other.GetType().Size(), asm.RegB)
    }
}

func derefSetIntVal(file *os.File, addr string, offset int, val token.Token) {
    asm.MovDerefVal(file, asm.OffsetAddr(addr, offset), types.I32_Size, val.Str)
}

func derefSetBoolVal(file *os.File, addr string, offset int, val token.Token) {
    if val.Str == "true" {
        asm.MovDerefVal(file, asm.OffsetAddr(addr, offset), types.Bool_Size, "1")
    } else {
        asm.MovDerefVal(file, asm.OffsetAddr(addr, offset), types.Bool_Size, "0")
    }
}

func derefSetPtrVal(file *os.File, addr string, offset int, val token.Token) {
    if val.Type == token.Name {
        file.WriteString(fmt.Sprintf("lea rax, [%s]\n", val.Str))
        asm.MovDerefReg(file, asm.OffsetAddr(addr, offset), types.Ptr_Size, asm.RegA)
    } else {
        asm.MovDerefVal(file, asm.OffsetAddr(addr, offset), types.Ptr_Size, val.Str)
    }
}

func derefSetStrVal(file *os.File, addr string, offset int, val token.Token) {
    strIdx := str.Add(val)

    asm.MovDerefVal(file, asm.OffsetAddr(addr, offset), types.Ptr_Size, fmt.Sprintf("_str%d", strIdx))
    asm.MovDerefVal(file, asm.OffsetAddr(addr, offset + int(types.Ptr_Size)), types.I32_Size, fmt.Sprint(str.GetSize(strIdx)))
}

func derefSetArrVal(file *os.File, addr string, offset int, val token.Token) {
    if idx,err := strconv.ParseUint(val.Str, 10, 64); err == nil {
        asm.MovDerefVal(file, addr, types.Ptr_Size, fmt.Sprintf("_arr%d", idx))
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] expected size of array to be a Number but got %v\n", val)
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }
}

func derefSetStructVal(file *os.File, t types.StructType, addr string, offset int, val token.Token) {
    if val.Type != token.Number {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Number but got %v\n", val)
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }

    idx,_ := strconv.ParseUint(val.Str, 10, 64)

    for i,val := range structLit.GetValues(idx) {
        switch typ := t.Types[i].(type) {
        case types.StrType:
            derefSetStrVal(file, addr, offset + t.GetOffset(uint(i)), val)
        case types.StructType:
            derefSetStructVal(file, typ, addr, offset + t.GetOffset(uint(i)), val)
        case types.ArrType:
            derefSetArrVal(file, addr, offset + t.GetOffset(uint(i)), val)
        case types.BoolType:
            derefSetBoolVal(file, addr, offset + t.GetOffset(uint(i)), val)
        case types.I32Type:
            derefSetIntVal(file, addr, offset + t.GetOffset(uint(i)), val)
        case types.PtrType:
            derefSetPtrVal(file, addr, offset + t.GetOffset(uint(i)), val)
        }
    }
}

