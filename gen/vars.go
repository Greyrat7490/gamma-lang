package gen

import (
    "os"
    "fmt"
    "reflect"
    "gamma/token"
    "gamma/types"
    "gamma/types/str"
    "gamma/types/addr"
    "gamma/types/array"
    "gamma/ast"
    "gamma/ast/identObj/vars"
    "gamma/cmpTime"
    "gamma/cmpTime/constVal"
    "gamma/gen/asm/x86_64"
    "gamma/gen/asm/x86_64/nasm"
)

// Define variable ----------------------------------------------------------

func VarDefVal(file *os.File, v vars.Var, val constVal.ConstVal) {
    switch v := v.(type) {
    case *vars.GlobalVar:
        globalVarDefVal(file, v, val)

    case *vars.LocalVar:
        DerefSetVal(file, v.Addr(), v.GetType(), val)

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

    DerefSetExpr(file, v.Addr(), v.GetType(), e)
}

func globalVarDefVal(file *os.File, v *vars.GlobalVar, val constVal.ConstVal) {
    nasm.AddData(fmt.Sprintf("%s:", v.GetName()))

    switch c := val.(type) {
    case *constVal.StrConst:
        defStr(c)

    case *constVal.StructConst:
        defStruct(v.GetType().(types.StructType), c)

    case *constVal.ArrConst, *constVal.PtrConst, *constVal.BoolConst, *constVal.IntConst, *constVal.UintConst:
        defBasic(val.GetVal(), v.GetType().Size())

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] define global var of typ %v is not supported yet\n", v.GetType())
        fmt.Fprintln(os.Stderr, "\t" + v.GetPos().At())
        os.Exit(1)
    }
}

func defStruct(t types.StructType, val *constVal.StructConst) {
    for i,v := range val.Fields {
        switch c := v.(type) {
        case *constVal.StrConst:
            defStr(c)

        case *constVal.StructConst:
            defStruct(t.Types[i].(types.StructType), c)

        case *constVal.ArrConst, *constVal.PtrConst, *constVal.BoolConst, *constVal.IntConst, *constVal.UintConst:
            defBasic(v.GetVal(), t.Types[i].Size())

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet\n", t)
            os.Exit(1)
        }
    }
}

func defBasic(val string, size uint) {
    nasm.AddData(fmt.Sprintf("  %s %s", asm.GetDataSize(size), val))
}

func defStr(val *constVal.StrConst) {
    nasm.AddData(fmt.Sprintf("  %s _str%d\n  %s %d",
        asm.GetDataSize(types.Ptr_Size),
        uint64(*val),
        asm.GetDataSize(types.I32_Size),
        str.GetSize(uint64(*val))))
}


// Assign -------------------------------------------------------------------

func AssignVar(file *os.File, v vars.Var, val ast.Expr) {
    if value := cmpTime.ConstEval(val); value != nil {
        DerefSetVal(file, v.Addr(), v.GetType(), value)

    } else if e,ok := val.(*ast.Ident); ok {
        if other,ok := e.Obj.(vars.Var); ok {
            if v.GetName() == other.GetName() {
                fmt.Fprintln(os.Stderr, "[WARNING] assigning a variable to itself is redundant")
                fmt.Fprintln(os.Stderr, "\t" + v.GetPos().At())
                return
            }

            DerefSetVar(file, v.Addr(), other)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected identifier %s to be a variable but got %v\n", e.Name, reflect.TypeOf(e.Obj))
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }

    } else {
        DerefSetExpr(file, v.Addr(), v.GetType(), val)
    }
}

func AssignDeref(file *os.File, t types.Type, dest *ast.Unary, val ast.Expr) {
    if dest.Operator.Type != token.Mul {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"*\" but got \"%v\"\n", dest.Operator)
        fmt.Fprintln(os.Stderr, "\t" + dest.At())
        os.Exit(1)
    }

    GenExpr(file, dest.Operand)

    if value := cmpTime.ConstEval(val); value != nil {
        DerefSetVal(file, asm.RegAsAddr(asm.RegA), t, value)

    } else if e,ok := val.(*ast.Ident); ok {
        if v,ok := e.Obj.(vars.Var); ok {
            DerefSetVar(file, asm.RegAsAddr(asm.RegA), v)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected identifier %s to be a variable but got %v\n", val, reflect.TypeOf(e.Obj))
            fmt.Fprintln(os.Stderr, "\t" + dest.At())
            os.Exit(1)
        }

    } else {
        asm.MovRegReg(file, asm.RegC, asm.RegA, types.Ptr_Size)
        DerefSetExpr(file, asm.RegAsAddr(asm.RegC), t, val)
    }
}

func AssignField(file *os.File, t types.Type, dest *ast.Field, val ast.Expr) {
    FieldAddrToReg(file, dest, asm.RegC)
    file.WriteString(fmt.Sprintf("lea rcx, [rcx+%d]\n", FieldToOffset(dest)))

    DerefSetExpr(file, asm.RegAsAddr(asm.RegC), t, val)
}

func AssignIndexed(file *os.File, t types.Type, dest *ast.Indexed, val ast.Expr) {
    IndexedAddrToReg(file, dest, asm.RegC)

    DerefSetExpr(file, asm.RegAsAddr(asm.RegC), t, val)
}

func DerefSetVar(file *os.File, addr addr.Addr, other vars.Var) {
    DerefSetDeref(file, addr, other.GetType(), other.Addr())
}

func DerefSetDeref(file *os.File, addr addr.Addr, t types.Type, otherAddr addr.Addr) {
    switch t := t.(type) {
    case types.StrType:
        asm.MovDerefDeref(file, addr, otherAddr, types.Ptr_Size, asm.RegB, false)
        asm.MovDerefDeref(file, addr.Offseted(int64(types.Ptr_Size)), otherAddr.Offseted(int64(types.Ptr_Size)), types.I32_Size, asm.RegB, false)

    case types.StructType:
        for i := 0; i < int(t.Size()/types.Ptr_Size); i++ {
            asm.MovDerefDeref(file, addr, otherAddr, types.Ptr_Size, asm.RegB, false)

            addr.Offset += int64(types.Ptr_Size)
            otherAddr.Offset += int64(types.Ptr_Size)
        }

        if size := t.Size() % types.Ptr_Size; size != 0 {
            asm.MovDerefDeref(file, addr, otherAddr, size, asm.RegB, false)
        }

    case types.IntType:
        asm.MovDerefDeref(file, asm.RegAsAddr(asm.RegA), otherAddr, t.Size(), asm.RegB, true)

    case types.UintType, types.BoolType, types.PtrType, types.ArrType, types.CharType:
        asm.MovDerefDeref(file, asm.RegAsAddr(asm.RegA), otherAddr, t.Size(), asm.RegB, false)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (DerefSetVar)\n", t)
        os.Exit(1)
    }
}

func DerefSetExpr(file *os.File, dst addr.Addr, t types.Type, val ast.Expr) {
    switch t := t.(type) {
    case types.StrType:
        GenExpr(file, val)
        asm.MovDerefReg(file, dst, types.Ptr_Size, asm.RegGroup(0))
        asm.MovDerefReg(file, dst.Offseted(int64(types.Ptr_Size)), types.I32_Size, asm.RegGroup(1))

    case types.StructType:
        if types.IsBigStruct(t) {
            DerefSetBigStruct(file, dst, val)
        } else {
            GenExpr(file, val)
            if t.Size() > uint(8) {
                asm.MovDerefReg(file, dst, types.Ptr_Size, asm.RegGroup(0))
                asm.MovDerefReg(file, dst.Offseted(int64(types.Ptr_Size)), t.Size() - 8, asm.RegGroup(1))
            } else {
                asm.MovDerefReg(file, dst, t.Size(), asm.RegGroup(0))
            }
        }

    case types.IntType, types.UintType, types.BoolType, types.PtrType, types.CharType:
        // TODO only push rcx for return structLit
        asm.PushReg(file, asm.RegC)
        GenExpr(file, val)
        asm.PopReg(file, asm.RegC)
        asm.MovDerefReg(file, dst, t.Size(), asm.RegGroup(0))

    case types.ArrType:
        if lit,ok := val.(*ast.ArrayLit); ok {
            arrAddr := addr.Addr{ BaseAddr: fmt.Sprintf("_arr%d", lit.Idx) }
            for i, v := range array.GetValues(lit.Idx) {
                if v == nil {
                    DerefSetExpr(file, arrAddr.Offseted(int64(i) * int64(t.BaseType.Size())), t.BaseType, lit.Values[i])
                }
            }
            asm.MovDerefVal(file, dst, types.Arr_Size, arrAddr.BaseAddr)
        } else {
            GenExpr(file, val)
            asm.MovDerefReg(file, dst, t.Size(), asm.RegGroup(0))
        }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (DerefSetExpr)\n", t)
        os.Exit(1)
    }
}

func derefSetBigStructLit(file *os.File, t types.StructType, val constVal.StructConst, offset int) {
    addr := asm.RegAsAddr(asm.RegC).Offseted(int64(offset))

    for _,field := range val.Fields {
        switch f := field.(type) {
        case *constVal.StrConst:
            idx := uint64(*f)
            asm.MovDerefVal(file, addr, types.Ptr_Size, fmt.Sprintf("_str%d", idx))
            asm.MovDerefVal(file, addr.Offseted(int64(types.Ptr_Size)), types.I32_Size, fmt.Sprint(str.GetSize(idx)))
        case *constVal.StructConst:
            derefSetBigStructLit(file, t, *f, offset)

        default:
            asm.MovDerefVal(file, addr, t.Size(), field.GetVal())
        }

        addr.Offset += int64(t.Size())
    }
}

func DerefSetVal(file *os.File, addr addr.Addr, typ types.Type, val constVal.ConstVal) {
    switch val := val.(type) {
    case *constVal.StrConst:
        derefSetStrVal(file, addr, 0, val)

    case *constVal.StructConst:
        derefSetStructVal(file, typ.(types.StructType), addr, 0, val)

    case *constVal.ArrConst, *constVal.IntConst, *constVal.UintConst, *constVal.BoolConst, *constVal.CharConst:
        derefSetBasicVal(file, addr, 0, typ.Size(), val.GetVal())

    case *constVal.PtrConst:
        derefSetPtrVal(file, addr, 0, val)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (derefSetStructVal)\n", typ)
        os.Exit(1)
    }
}

func derefSetBasicVal(file *os.File, addr addr.Addr, offset int, size uint, val string) {
    asm.MovDerefVal(file, addr.Offseted(int64(offset)), size, val)
}

func derefSetPtrVal(file *os.File, addr addr.Addr, offset int, val *constVal.PtrConst) {
    asm.MovDerefVal(file, addr.Offseted(int64(offset)), types.Ptr_Size, PtrConstToAddr(file, *val))
}

func derefSetStrVal(file *os.File, addr addr.Addr, offset int, val *constVal.StrConst) {
    asm.MovDerefVal(file, addr.Offseted(int64(offset)), types.Ptr_Size, fmt.Sprintf("_str%d", uint64(*val)))
    asm.MovDerefVal(file, addr.Offseted(int64(offset) + int64(types.Ptr_Size)), types.I32_Size, fmt.Sprint(str.GetSize(uint64(*val))))
}

func derefSetStructVal(file *os.File, t types.StructType, addr addr.Addr, offset int, val *constVal.StructConst) {
    for i,val := range val.Fields {
        switch val := val.(type) {
        case *constVal.StrConst:
            derefSetStrVal(file, addr, offset + t.GetOffset(uint(i)), val)

        case *constVal.StructConst:
            derefSetStructVal(file, t.Types[i].(types.StructType), addr, offset + t.GetOffset(uint(i)), val)

        case *constVal.ArrConst, *constVal.IntConst, *constVal.UintConst, *constVal.BoolConst, *constVal.CharConst:
            derefSetBasicVal(file, addr, offset + t.GetOffset(uint(i)), t.Types[i].Size(), val.GetVal())

        case *constVal.PtrConst:
            derefSetPtrVal(file, addr, offset + t.GetOffset(uint(i)), val)

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (derefSetStructVal)\n", t.Types[i])
            os.Exit(1)
        }
    }
}
