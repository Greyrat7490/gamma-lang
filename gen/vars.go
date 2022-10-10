package gen

import (
    "os"
    "fmt"
    "reflect"
    "strconv"
    "gamma/token"
    "gamma/types"
    "gamma/types/str"
    "gamma/types/struct"
    "gamma/ast"
    "gamma/ast/identObj/vars"
    "gamma/cmpTime"
    "gamma/gen/asm/x86_64"
    "gamma/gen/asm/x86_64/nasm"
)

// Define variable ----------------------------------------------------------

func VarDefVal(file *os.File, v vars.Var, val token.Token) {
    switch v := v.(type) {
    case *vars.GlobalVar:
        globalVarDefVal(file, v, val)

    case *vars.LocalVar:
        DerefSetVal(file, v.Addr(0), v.GetType(), val)

    default:
        fmt.Fprintln(os.Stderr, "[ERROR] (unreachable) DefVarVal: v is neigther GlobalVar nor LocalVar")
        os.Exit(1)
    }
}

func VarDefExpr(file *os.File, v vars.Var, e ast.Expr) {
    if _,ok := v.(*vars.GlobalVar); ok {
        fmt.Fprintln(os.Stderr, "[ERROR] defining a global variable with a non const expr is not allowed")
        fmt.Fprintln(os.Stderr, "\t" + v.GetPos().At())
        os.Exit(1)
    }

    DerefSetExpr(file, v.Addr(0), v.GetType(), e)
}

func globalVarDefVal(file *os.File, v *vars.GlobalVar, val token.Token) {
    nasm.AddData(fmt.Sprintf("%s:", v.GetName()))

    switch t := v.GetType().(type) {
    case types.StrType:
        defStr(val)
    case types.ArrType:
        defArr(val.Str)
    case types.StructType:
        defStruct(t, val)
    case types.BoolType:
        defBool(val.Str)
    case types.IntType, types.UintType:
        defInt(val.Str, t.Size())
    case types.PtrType:
        defPtr(val.Str)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] define global var of typ %v is not supported yet\n", v.GetType())
        fmt.Fprintln(os.Stderr, "\t" + v.GetPos().At())
        os.Exit(1)
    }
}

func defStruct(t types.StructType, val token.Token) {
    if val.Type != token.Number {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Number but got %v\n", val)
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }

    idx,_ := strconv.ParseUint(val.Str, 10, 64)

    for i,v := range structLit.GetValues(idx) {
        switch t := t.Types[i].(type) {
        case types.StrType:
            defStr(v)
        case types.IntType, types.UintType:
            defInt(v.Str, t.Size())
        case types.BoolType:
            defBool(v.Str)
        case types.PtrType:
            defPtr(v.Str)
        case types.ArrType:
            defArr(v.Str)
        case types.StructType:
            defStruct(t, v)
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet\n", t)
            os.Exit(1)
        }
    }
}

func defInt(val string, size uint) {
    nasm.AddData(fmt.Sprintf("  %s %s", asm.GetDataSize(size), val))
}

func defPtr(val string) {
    nasm.AddData(fmt.Sprintf("  %s %s", asm.GetDataSize(types.Ptr_Size), val))
}

func defBool(val string) {
    nasm.AddData(fmt.Sprintf("  %s %s", asm.GetDataSize(types.Bool_Size), val))
}

func defStr(val token.Token) {
    if idx,err := strconv.Atoi(val.Str); err == nil {
        nasm.AddData(fmt.Sprintf("  %s _str%d\n  %s %d",
            asm.GetDataSize(types.Ptr_Size),
            idx,
            asm.GetDataSize(types.I32_Size),
            str.GetSize(idx)))
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] expected str literal converted to a Number but got %v\n", val)
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }
}

func defArr(val string) {
    nasm.AddData(fmt.Sprintf("  %s _arr%s", asm.GetDataSize(types.Ptr_Size), val))
}


// Assign -------------------------------------------------------------------

func AssignVar(file *os.File, v vars.Var, val ast.Expr) {
    if value := cmpTime.ConstEval(val); value.Type != token.Unknown {
        DerefSetVal(file, v.Addr(0), v.GetType(), value)

    } else if e,ok := val.(*ast.Ident); ok {
        if other,ok := e.Obj.(vars.Var); ok {
            if v.GetName() == other.GetName() {
                fmt.Fprintln(os.Stderr, "[WARNING] assigning a variable to itself is redundant")
                fmt.Fprintln(os.Stderr, "\t" + v.GetPos().At())
                return
            }

            DerefSetVar(file, v.Addr(0), other)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected identifier %s to be a variable but got %v\n", e.Name, reflect.TypeOf(e.Obj))
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }

    } else {
        DerefSetExpr(file, v.Addr(0), v.GetType(), val)
    }
}

func AssignDeref(file *os.File, t types.Type, dest *ast.Unary, val ast.Expr) {
    if dest.Operator.Type != token.Mul {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"*\" but got \"%v\"\n", dest.Operator)
        fmt.Fprintln(os.Stderr, "\t" + dest.At())
        os.Exit(1)
    }

    GenExpr(file, dest.Operand)

    if value := cmpTime.ConstEval(val); value.Type != token.Unknown {
        DerefSetVal(file, asm.GetReg(asm.RegA, types.Ptr_Size), t, value)

    } else if e,ok := val.(*ast.Ident); ok {
        if v,ok := e.Obj.(vars.Var); ok {
            DerefSetVar(file, asm.GetReg(asm.RegA, types.Ptr_Size), v)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected identifier %s to be a variable but got %v\n", val, reflect.TypeOf(e.Obj))
            fmt.Fprintln(os.Stderr, "\t" + dest.At())
            os.Exit(1)
        }

    } else {
        asm.MovRegReg(file, asm.RegC, asm.RegA, types.Ptr_Size)
        DerefSetExpr(file, asm.GetReg(asm.RegC, types.Ptr_Size), t, val)
    }
}

func AssignField(file *os.File, t types.Type, dest *ast.Field, val ast.Expr) {
    FieldAddrToReg(file, dest, asm.RegC)
    offset := FieldToOffset(dest)
    file.WriteString(fmt.Sprintf("lea rcx, [rcx+%d]\n", offset))

    DerefSetExpr(file, asm.GetReg(asm.RegC, types.Ptr_Size), t, val)
}

func AssignIndexed(file *os.File, t types.Type, dest *ast.Indexed, val ast.Expr) {
    IndexedAddrToReg(file, dest, asm.RegC)

    DerefSetExpr(file, asm.GetReg(asm.RegC, types.Ptr_Size), t, val)
}

func DerefSetVar(file *os.File, addr string, other vars.Var) {
    DerefSetDeref(file, addr, other.GetType(), other.Addr(0))
}

func DerefSetDeref(file *os.File, addr string, t types.Type, otherAddr string) {
    switch t := t.(type) {
    case types.StrType:
        asm.MovDerefDeref(file, addr, otherAddr, types.Ptr_Size, asm.RegB, false)
        asm.MovDerefDeref(file, asm.OffsetAddr(addr, int(types.Ptr_Size)), asm.OffsetAddr(otherAddr, int(types.Ptr_Size)), types.I32_Size, asm.RegB, false)

    case types.StructType:
        var offset int = 0
        for i := 0; i < int(t.Size()/types.Ptr_Size); i++ {
            asm.MovDerefDeref(
                file,
                asm.OffsetAddr(addr, offset),
                asm.OffsetAddr(otherAddr, offset),
                types.Ptr_Size,
                asm.RegB,
                false,
            )
            offset += int(types.Ptr_Size)
        }

        if size := t.Size() % types.Ptr_Size; size != 0 {
            asm.MovDerefDeref(
                file,
                asm.OffsetAddr(addr, offset),
                asm.OffsetAddr(otherAddr, offset),
                size,
                asm.RegB,
                false,
            )
        }

    case types.IntType:
        asm.MovDerefDeref(file, asm.GetReg(asm.RegA, types.Ptr_Size), otherAddr, t.Size(), asm.RegB, true)

    case types.UintType, types.BoolType, types.PtrType, types.ArrType, types.CharType:
        asm.MovDerefDeref(file, asm.GetReg(asm.RegA, types.Ptr_Size), otherAddr, t.Size(), asm.RegB, false)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (DerefSetVar)\n", t)
        os.Exit(1)
    }
}

func DerefSetExpr(file *os.File, addr string, t types.Type, val ast.Expr) {
    switch t := t.(type) {
    case types.StrType:
        GenExpr(file, val)
        asm.MovDerefReg(file, addr, types.Ptr_Size, asm.RegGroup(0))
        asm.MovDerefReg(file, asm.OffsetAddr(addr, int(types.Ptr_Size)), types.I32_Size, asm.RegGroup(1))

    case types.StructType:
        if types.IsBigStruct(t) {
            DerefSetBigStruct(file, addr, val)
        } else {
            GenExpr(file, val)
            if t.Size() > uint(8) {
                asm.MovDerefReg(file, addr, types.Ptr_Size, asm.RegGroup(0))
                asm.MovDerefReg(file, asm.OffsetAddr(addr, int(types.Ptr_Size)), t.Size() - 8, asm.RegGroup(1))
            } else {
                asm.MovDerefReg(file, addr, t.Size(), asm.RegGroup(0))
            }
        }

    case types.IntType, types.UintType, types.BoolType, types.PtrType, types.ArrType, types.CharType:
        GenExpr(file, val)
        asm.MovDerefReg(file, addr, t.Size(), asm.RegGroup(0))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (DerefSetExpr)\n", t)
        os.Exit(1)
    }
}

func derefSetBigStructLit(file *os.File, t types.StructType, val token.Token, offset int) {
    if idx,err := strconv.ParseUint(val.Str, 10, 64); err == nil {
        fields := structLit.GetValues(idx)

        for i,t := range t.Types {
            switch t := t.(type) {
            case types.StrType:
                if idx,err := strconv.Atoi(fields[i].Str); err == nil {
                    asm.MovDerefVal(file,
                        asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset),
                        types.Ptr_Size,
                        fmt.Sprintf("_str%d", idx))
                    asm.MovDerefVal(file,
                        asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset + int(types.Ptr_Size)),
                        types.I32_Size,
                        fmt.Sprint(str.GetSize(idx)))
                } else {
                    fmt.Fprintf(os.Stderr, "[ERROR] expected str literal converted to a Number but got %v\n", fields[i])
                    fmt.Fprintln(os.Stderr, "\t" + fields[i].At())
                    os.Exit(1)
                }

            case types.StructType:
                derefSetBigStructLit(file, t, fields[i], offset)

            default:
                asm.MovDerefVal(file, asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset), t.Size(), fields[i].Str)
            }
            offset += int(t.Size())
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] expected struct literal converted to a Number but got %v\n", val)
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }
}

func DerefSetVal(file *os.File, addr string, typ types.Type, val token.Token) {
    switch t := typ.(type) {
    case types.StrType:
        derefSetStrVal(file, addr, 0, val)
    case types.ArrType:
        derefSetArrVal(file, addr, 0, val)
    case types.StructType:
        derefSetStructVal(file, t, addr, 0, val)
    case types.PtrType:
        derefSetPtrVal(file, addr, 0, val)
    case types.BoolType:
        derefSetBoolVal(file, addr, 0, val)
    case types.IntType, types.UintType:
        derefSetIntVal(file, addr, 0, t.Size(), val)
    case types.CharType:
        derefSetCharVal(file, addr, 0, val)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (DerefSetVal)\n", t)
        os.Exit(1)
    }
}

func derefSetIntVal(file *os.File, addr string, offset int, size uint, val token.Token) {
    asm.MovDerefVal(file, asm.OffsetAddr(addr, offset), size, val.Str)
}

func derefSetBoolVal(file *os.File, addr string, offset int, val token.Token) {
    asm.MovDerefVal(file, asm.OffsetAddr(addr, offset), types.Bool_Size, val.Str)
}

func derefSetPtrVal(file *os.File, addr string, offset int, val token.Token) {
    if val.Type == token.Str {
        file.WriteString(fmt.Sprintf("lea rax, [%s]\n", val.Str))
        asm.MovDerefReg(file, asm.OffsetAddr(addr, offset), types.Ptr_Size, asm.RegA)
    } else {
        asm.MovDerefVal(file, asm.OffsetAddr(addr, offset), types.Ptr_Size, val.Str)
    }
}

func derefSetCharVal(file *os.File, addr string, offset int, val token.Token) {
    asm.MovDerefVal(file, asm.OffsetAddr(addr, offset), types.Char_Size, val.Str)
}

func derefSetStrVal(file *os.File, addr string, offset int, val token.Token) {
    if idx,err := strconv.Atoi(val.Str); err == nil {
        asm.MovDerefVal(file, asm.OffsetAddr(addr, offset), types.Ptr_Size, fmt.Sprintf("_str%d", idx))
        asm.MovDerefVal(file, asm.OffsetAddr(addr, offset + int(types.Ptr_Size)), types.I32_Size, fmt.Sprint(str.GetSize(idx)))
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] expected str literal converted to a Number but got %v\n", val)
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }
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
        case types.IntType, types.UintType:
            derefSetIntVal(file, addr, offset + t.GetOffset(uint(i)), typ.Size(), val)
        case types.PtrType:
            derefSetPtrVal(file, addr, offset + t.GetOffset(uint(i)), val)
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (derefSetStructVal)\n", t)
            os.Exit(1)
        }
    }
}
