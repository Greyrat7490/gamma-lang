package gen

import (
    "os"
    "fmt"
    "bufio"
    "reflect"
    "gamma/types"
    "gamma/types/str"
    "gamma/types/addr"
    "gamma/ast"
    "gamma/ast/identObj/vars"
    "gamma/cmpTime"
    "gamma/cmpTime/constVal"
    "gamma/gen/asm/x86_64"
    "gamma/gen/asm/x86_64/nasm"
)

// Define variable ----------------------------------------------------------

func VarDefVal(file *bufio.Writer, v vars.Var, val constVal.ConstVal) {
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

func VarDefExpr(file *bufio.Writer, v vars.Var, e ast.Expr) {
    if _,ok := v.(*vars.GlobalVar); ok {
        fmt.Fprintln(os.Stderr, "[ERROR] defining a global variable with a non const expr is not allowed")
        fmt.Fprintln(os.Stderr, "\t" + v.GetPos().At())
        os.Exit(1)
    }

    DerefSetExpr(file, v.Addr(), v.GetType(), e)
}

func globalVarDefVal(file *bufio.Writer, v *vars.GlobalVar, val constVal.ConstVal) {
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


// Deref -------------------------------------------------------------------
func DerefSetVar(file *bufio.Writer, addr addr.Addr, other vars.Var) {
    DerefSetDeref(file, addr, other.GetType(), other.Addr())
}

func DerefSetDeref(file *bufio.Writer, addr addr.Addr, t types.Type, otherAddr addr.Addr) {
    t = types.ResolveGeneric(t)

    switch t := t.(type) {
    case types.StrType:
        asm.MovDerefDeref(file, addr, otherAddr, types.Ptr_Size, asm.RegB, false)
        asm.MovDerefDeref(file, addr.Offseted(int64(types.Ptr_Size)), otherAddr.Offseted(int64(types.Ptr_Size)), types.I32_Size, asm.RegB, false)

    case types.StructType, types.EnumType:
        for i := 0; i < int(t.Size()/types.Ptr_Size); i++ {
            asm.MovDerefDeref(file, addr, otherAddr, types.Ptr_Size, asm.RegB, false)
            addr = addr.Offseted(int64(types.Ptr_Size))
            otherAddr = otherAddr.Offseted(int64(types.Ptr_Size))
        }

        if size := t.Size() % types.Ptr_Size; size != 0 {
            for i := 2; i >= 0; i-- {
                sz := uint(1 << i)
                if sz <= size {
                    size -= sz
                    asm.MovDerefDeref(file, addr, otherAddr, sz, asm.RegB, false)
                    addr = addr.Offseted(int64(sz))
                    otherAddr = otherAddr.Offseted(int64(sz))
                }
            }
        }

    case types.VecType:
        asm.MovDerefDeref(file, addr, otherAddr, types.Ptr_Size, asm.RegB, false)
        asm.MovDerefDeref(file,
            addr.Offseted(int64(types.Ptr_Size)),
            otherAddr.Offseted(int64(types.Ptr_Size)),
            types.Ptr_Size, asm.RegB, false)
        asm.MovDerefDeref(file,
            addr.Offseted(int64(2*types.Ptr_Size)),
            otherAddr.Offseted(int64(2*types.Ptr_Size)),
            types.Ptr_Size, asm.RegB, false)

    case types.IntType:
        asm.MovDerefDeref(file, addr, otherAddr, t.Size(), asm.RegB, true)

    case types.UintType, types.BoolType, types.PtrType, types.ArrType, types.CharType:
        asm.MovDerefDeref(file, addr, otherAddr, t.Size(), asm.RegB, false)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (DerefSetVar)\n", t)
        os.Exit(1)
    }
}

func DerefSetExpr(file *bufio.Writer, dst addr.Addr, t types.Type, val ast.Expr) {
    t = types.ResolveGeneric(t)

    switch t := t.(type) {
    case types.StrType:
        GenExpr(file, val)
        asm.MovDerefReg(file, dst, types.Ptr_Size, asm.RegGroup(0))
        asm.MovDerefReg(file, dst.Offseted(int64(types.Ptr_Size)), types.I32_Size, asm.RegGroup(1))

    case types.StructType, types.EnumType:
        if types.IsBigStruct(t) {
            DerefSetBigStruct(file, dst, val)
        } else {
            GenExpr(file, val)
            size := t.Size() 
            reg := asm.RegGroup(0)

            if size > types.Ptr_Size {
                asm.MovDerefReg(file, dst, types.Ptr_Size, reg)
                size -= types.Ptr_Size
                dst = dst.Offseted(int64(types.Ptr_Size))
                reg += 1
            }

            for i := 3; i >= 0; i-- {
                sz := uint(1 << i)
                if sz <= size {
                    size -= sz
                    asm.MovDerefReg(file, dst, sz, reg)
                    if size > 0 {
                        dst = dst.Offseted(int64(sz))
                        asm.Shr(file, fmt.Sprint(sz*8), types.Ptr_Size)
                    }
                }
            }
        }

    case types.IntType, types.UintType, types.BoolType, types.PtrType, types.CharType:
        GenExpr(file, val)
        asm.MovDerefReg(file, dst, t.Size(), asm.RegGroup(0))

    case types.ArrType:
        if lit,ok := val.(*ast.ArrayLit); ok {
            derefSetArrLit(file, dst, t, lit)
        } else {
            GenExpr(file, val)
            asm.MovDerefReg(file, dst, t.Size(), asm.RegGroup(0))
        }

    case types.VecType:
        DerefSetBigStruct(file, dst, val)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (DerefSetExpr)\n", reflect.TypeOf(t))
        os.Exit(1)
    }
}

func getActualArrBaseSize(t types.Type) uint64 {
    if t,ok := t.(types.ArrType); ok {
        return t.Len * getActualArrBaseSize(t.BaseType)
    }

    return uint64(t.Size())
}

func derefSetArrLit(file *bufio.Writer, dst addr.Addr, t types.ArrType, lit *ast.ArrayLit) {
    if lit.Idx != ^uint64(0) {
        arrAddr := addr.Addr{ BaseAddr: fmt.Sprintf("_arr%d", lit.Idx) }
        for i, v := range lit.Values {
            DerefSetExpr(file, arrAddr.Offseted(int64(i) * int64(getActualArrBaseSize(t.BaseType))), t.BaseType, v)
        }
        asm.MovDerefVal(file, dst, lit.Type.Size(), arrAddr.BaseAddr)
    } else {
        for i, v := range lit.Values {
            if cmpTime.ConstEval(v) == nil {
                DerefSetExpr(file, dst.Offseted(int64(i) * int64(getActualArrBaseSize(t.BaseType))), t.BaseType, v)
            }
        }
    }
}

func derefSetBigStructLit(file *bufio.Writer, t types.StructType, val constVal.StructConst, offset int) {
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

func DerefSetVal(file *bufio.Writer, addr addr.Addr, typ types.Type, val constVal.ConstVal) {
    typ = types.ResolveGeneric(typ)

    switch val := val.(type) {
    case *constVal.StrConst:
        derefSetStrVal(file, addr, 0, val)

    case *constVal.StructConst:
        derefSetStructVal(file, typ.(types.StructType), addr, 0, val)

    case *constVal.ArrConst, *constVal.IntConst, *constVal.UintConst, *constVal.BoolConst, *constVal.CharConst:
        derefSetBasicVal(file, addr, 0, typ.Size(), val.GetVal())

    case *constVal.PtrConst:
        derefSetPtrVal(file, addr, 0, val)

    case *constVal.EnumConst:
        derefSetEnumVal(file, typ.(types.EnumType), addr, 0, val)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (DerefSetVal)\n", reflect.TypeOf(val))
        os.Exit(1)
    }
}

func derefSetBasicVal(file *bufio.Writer, addr addr.Addr, offset int, size uint, val string) {
    asm.MovDerefVal(file, addr.Offseted(int64(offset)), size, val)
}

func derefSetPtrVal(file *bufio.Writer, addr addr.Addr, offset int, val *constVal.PtrConst) {
    PtrConstToAddr(file, *val, addr.Offseted(int64(offset)))
}

func derefSetStrVal(file *bufio.Writer, addr addr.Addr, offset int, val *constVal.StrConst) {
    asm.MovDerefVal(file, addr.Offseted(int64(offset)), types.Ptr_Size, fmt.Sprintf("_str%d", uint64(*val)))
    asm.MovDerefVal(file, addr.Offseted(int64(offset) + int64(types.Ptr_Size)), types.I32_Size, fmt.Sprint(str.GetSize(uint64(*val))))
}

func derefSetEnumVal(file *bufio.Writer, t types.EnumType, addr addr.Addr, offset int, val *constVal.EnumConst) {
    asm.MovDerefVal(file, addr.Offseted(int64(offset)), t.IdType.Size(), fmt.Sprint(val.Id))
    if val.Elem != nil {
        DerefSetVal(file, addr.Offseted(int64(offset) + int64(t.IdType.Size())), val.ElemType, val.Elem)
    }
}

func derefSetStructVal(file *bufio.Writer, t types.StructType, addr addr.Addr, offset int, val *constVal.StructConst) {
    for i,val := range val.Fields {
        switch val := val.(type) {
        case *constVal.StrConst:
            derefSetStrVal(file, addr, offset, val)

        case *constVal.StructConst:
            derefSetStructVal(file, t.Types[i].(types.StructType), addr, offset, val)

        case *constVal.ArrConst, *constVal.IntConst, *constVal.UintConst, *constVal.BoolConst, *constVal.CharConst:
            derefSetBasicVal(file, addr, offset, t.Types[i].Size(), val.GetVal())

        case *constVal.PtrConst:
            derefSetPtrVal(file, addr, offset, val)

        case *constVal.EnumConst:
            derefSetEnumVal(file, t.Types[i].(types.EnumType), addr, offset, val)

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (derefSetStructVal)\n", reflect.TypeOf(val))
            os.Exit(1)
        }

        offset += int(t.Types[i].Size())
    }
}
