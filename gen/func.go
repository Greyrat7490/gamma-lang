package gen

import (
    "os"
    "fmt"
    "bufio"
    "reflect"
    "gamma/types"
    "gamma/types/str"
    "gamma/types/addr"
    "gamma/types/array"
    "gamma/ast"
    "gamma/ast/identObj"
    "gamma/ast/identObj/vars"
    "gamma/cmpTime"
    "gamma/cmpTime/constVal"
    "gamma/gen/asm/x86_64"
    "gamma/gen/asm/x86_64/vtable"
)

/*
System V AMD64 ABI calling convention
  * [x] int:    rdi, rsi, rdx, rcx, r8, r9
  * [ ] float:  xmm0 - xmm7
  * [x] more on stack (right to left)

  * [x] struct:         use int/float fields
  * [x] big struct:     on stack (right to left)
    * [x] bigger than 16Byte or unaligned fields (more than 2 regs needed)

  * [x] return value:
    * [x] int: rax, rdx
    * [ ] float: xmm0, xmm1
    * [x] big struct: stack (addr in rdi)
  * [x] caller cleans stack
  * [x] callee reserves space

  * [x] stack always 16bit aligned

Name mangling
  * [x] normal function: <function name>
  * [x] generic function: <function name>$<type name>
  * [x] interface/impl function: <struct name>.<function name>$<type name>
*/

var regs []asm.RegGroup = []asm.RegGroup{ asm.RegDi, asm.RegSi, asm.RegD, asm.RegC, asm.RegR8, asm.RegR9 }

func Define(file *bufio.Writer, f *identObj.Func, frameSize uint) {
    file.WriteString(f.GetMangledName() + ":\n")
    asm.PushReg(file, asm.RegBp)
    asm.MovRegReg(file, asm.RegBp, asm.RegSp, types.Ptr_Size)
    if frameSize > 0 {
        // framesize has to be the multiple of 16byte
        asm.SubSp(file, int64(frameSize + 15) & ^15)
    }
}

func FnEnd(file *bufio.Writer) {
    file.WriteString("leave\n")
    file.WriteString("ret\n")
}

func CallFn(file *bufio.Writer, f *identObj.Func) {
    file.WriteString("call " + f.GetMangledName() + "\n")
}

func CallVTableFn(file *bufio.Writer, offset uint) {
    addr := addr.Addr{ BaseAddr: asm.GetReg(asm.RegD, types.Ptr_Size), Offset: int64(offset) }
    file.WriteString(fmt.Sprintf("call QWORD [%s]\n", addr))
}

func DefArg(file *bufio.Writer, regIdx uint, v vars.Var, t types.Type) {
    switch t := t.(type) {
    case types.StrType:
        asm.MovDerefReg(file, v.Addr(), types.Ptr_Size, regs[regIdx])
        asm.MovDerefReg(file, v.Addr().Offseted(int64(types.Ptr_Size)), types.I32_Size, regs[regIdx+1])

    case types.StructType, types.EnumType:
        if t.Size() > uint(8) {
            asm.MovDerefReg(file, v.Addr(), types.Ptr_Size, regs[regIdx])
            asm.MovDerefReg(file, v.Addr().Offseted(int64(types.Ptr_Size)), t.Size() - types.Ptr_Size, regs[regIdx+1])
        } else {
            asm.MovDerefReg(file, v.Addr(), t.Size(), regs[regIdx])
        }

    case types.InterfaceType:
        asm.MovDerefReg(file, v.Addr(), types.Ptr_Size, regs[regIdx])
        asm.MovDerefReg(file, v.Addr().Offseted(int64(types.Ptr_Size)), t.Size() - types.Ptr_Size, regs[regIdx+1])

    case *types.GenericType:
        DefArg(file, regIdx, v, t.CurUsedType)

    default:
        asm.MovDerefReg(file, v.Addr(), t.Size(), regs[regIdx])
    }
}


func PassVal(file *bufio.Writer, regIdx uint, value constVal.ConstVal, valtype types.Type) {
    switch v := value.(type) {
    case *constVal.StrConst:
        asm.MovRegVal(file, regs[regIdx], types.Ptr_Size, fmt.Sprintf("_str%d", uint64(*v)))
        if valtype.GetKind() == types.Str { // check for *char cast
            asm.MovRegVal(file, regs[regIdx+1], types.I32_Size, fmt.Sprint(str.GetSize(uint64(*v))))
        }

    case *constVal.StructConst:
        t := valtype.(types.StructType)

        if len(t.Types) == 1 && t.Types[0].GetKind() != types.Str {
            asm.MovRegVal(file, regs[regIdx], t.Size(), v.Fields[0].GetVal())
        } else {
            vs := PackValues(t.Types, v.Fields)
            asm.MovRegVal(file, regs[regIdx], types.Ptr_Size, vs[0])
            if len(vs) == 2 {
                asm.MovRegVal(file, regs[regIdx+1], t.Size() - types.Ptr_Size, vs[1])
            }
        }

    case *constVal.EnumConst:
        if v.Elem == nil {
            asm.MovRegVal(file, regs[regIdx], v.Type.IdType.Size(), fmt.Sprint(v.Id))
        } else {
            idConst := constVal.UintConst(v.Id)
            vs := PackValues([]types.Type{ v.Type.IdType, v.ElemType }, []constVal.ConstVal{ &idConst, v.Elem })
            asm.MovRegVal(file, regs[regIdx], types.Ptr_Size, vs[0])
            if len(vs) == 2 {
                asm.MovRegVal(file, regs[regIdx+1], v.Type.Size() - types.Ptr_Size, vs[1])
            }
        }

    case *constVal.ArrConst, *constVal.IntConst, *constVal.UintConst, *constVal.CharConst, *constVal.BoolConst:
        asm.MovRegVal(file, regs[regIdx], valtype.Size(), value.GetVal())

    case *constVal.PtrConst:
        PtrConstToReg(file, *v, regs[regIdx])

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] cannot pass %v yet\n", reflect.TypeOf(value))
        os.Exit(1)
    }
}

func PassVar(file *bufio.Writer, regIdx uint, t types.Type, otherVar vars.Var) {
    switch t := t.(type) {
    case types.StrType:
        asm.MovRegDeref(file, regs[regIdx],   otherVar.Addr(), types.Ptr_Size, false)
        asm.MovRegDeref(file, regs[regIdx+1], otherVar.Addr().Offseted(int64(types.Ptr_Size)), types.U32_Size, false)

    case types.StructType, types.EnumType:
        if t.Size() > uint(8) {
            asm.MovRegDeref(file, regs[regIdx],   otherVar.Addr(), types.Ptr_Size, false)
            asm.MovRegDeref(file, regs[regIdx+1], otherVar.Addr().Offseted(int64(types.Ptr_Size)), t.Size() - 8, false)
        } else {
            asm.MovRegDeref(file, regs[regIdx],   otherVar.Addr(), t.Size(), false)
        }

    case types.InterfaceType:
        asm.MovRegDeref(file, regs[regIdx], otherVar.Addr(), types.Ptr_Size, false)

        switch otherVar.GetType().GetKind() {
        case types.Interface:
            asm.MovRegDeref(file, regs[regIdx+1], otherVar.Addr().Offseted(int64(types.Ptr_Size)), types.Ptr_Size, false)
        case types.Struct, types.Enum:
            asm.MovRegVal(file, regs[regIdx+1], t.Size() - 8, vtable.GetVTableName(otherVar.GetType().String(), t.String()))
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] (internal) expected %v to be an interface or implementable but got %v\n", otherVar.GetName(), otherVar.GetType())
            os.Exit(1)
        }

    case types.IntType:
        asm.MovRegDerefExtend(file, regs[regIdx], t.Size(), otherVar.Addr(), otherVar.GetType().Size(), true)

    case *types.GenericType:
        PassVar(file, regIdx, t.CurUsedType, otherVar)

    default:
        asm.MovRegDerefExtend(file, regs[regIdx], t.Size(), otherVar.Addr(), otherVar.GetType().Size(), false)
    }
}

func PassExpr(file *bufio.Writer, regIdx uint, argType types.Type, srcSize uint, expr ast.Expr) {
    switch t := argType.(type) {
    case types.StrType:
        GenExpr(file, expr)
        asm.MovRegReg(file, regs[regIdx],   asm.RegGroup(0), types.Ptr_Size)
        asm.MovRegReg(file, regs[regIdx+1], asm.RegGroup(1), types.U32_Size)

    case types.StructType, types.EnumType:
        GenExpr(file, expr)
        if t.Size() > uint(8) {
            asm.MovRegReg(file, regs[regIdx], asm.RegGroup(0), types.Ptr_Size)
            asm.MovRegReg(file, regs[regIdx+1], asm.RegGroup(1), t.Size() - 8)
        } else {
            asm.MovRegReg(file, regs[regIdx], asm.RegGroup(0), t.Size())
        }

    case types.InterfaceType:
        GenExpr(file, expr)
        asm.MovRegReg(file, regs[regIdx], asm.RegGroup(0), types.Ptr_Size)
        asm.MovRegReg(file, regs[regIdx+1], asm.RegGroup(1), t.Size() - 8)

    case types.IntType:
        GenExpr(file, expr)
        asm.MovRegRegExtend(file, regs[regIdx], t.Size(), asm.RegGroup(0), srcSize, true)

    case types.ArrType:
        if lit,ok := expr.(*ast.ArrayLit); ok {
            arrAddr := addr.Addr{ BaseAddr: fmt.Sprintf("_arr%d", lit.Idx) }
            for i, v := range array.GetValues(lit.Idx) {
                if v == nil {
                    DerefSetExpr(file, arrAddr.Offseted(int64(i) * int64(t.BaseType.Size())), t.BaseType, lit.Values[i])
                }
            }
            asm.MovRegVal(file, regs[regIdx], srcSize, fmt.Sprint(arrAddr))
        } else {
            GenExpr(file, expr)
            asm.MovRegRegExtend(file, regs[regIdx], t.Size(), asm.RegGroup(0), srcSize, false)
        }

    case *types.GenericType:
        PassExpr(file, regIdx, t.CurUsedType, srcSize, expr)

    default:
        GenExpr(file, expr)
        asm.MovRegRegExtend(file, regs[regIdx], t.Size(), asm.RegGroup(0), srcSize, false)
    }
}

func PassValStack(file *bufio.Writer, value constVal.ConstVal, valtype types.Type) {
    switch v := value.(type) {
    case *constVal.StrConst:
        asm.PushVal(file, fmt.Sprint(str.GetSize(uint64(*v))))
        asm.PushVal(file, fmt.Sprintf("_str%d", uint64(*v)))

    case *constVal.StructConst:
        t := valtype.(types.StructType)

        if len(t.Types) == 1 && t.Types[0].GetKind() != types.Str {
            asm.PushVal(file, v.Fields[0].GetVal())
        } else {
            vs := PackValues(t.Types, v.Fields)
            asm.PushVal(file, vs[0])
            if len(vs) == 2 {
                asm.PushVal(file, vs[1])
            }
        }

    case *constVal.PtrConst:
        if v.Local {
            asm.Lea(file, asm.RegA, v.Addr.String(), types.Ptr_Size)
            asm.PushReg(file, asm.RegA)
        } else {
            asm.PushVal(file, v.GetVal())
        }

    default:
        asm.PushVal(file, v.GetVal())
    }
}

func PassVarStack(file *bufio.Writer, otherVar vars.Var) {
    switch t := otherVar.GetType().(type) {
    case types.StrType:
        asm.PushDeref(file, otherVar.Addr().Offseted(int64(types.Ptr_Size)))
        asm.PushDeref(file, otherVar.Addr())

    case types.StructType:
        if t.Size() > uint(8) {
            asm.PushDeref(file, otherVar.Addr().Offseted(int64(types.Ptr_Size)))
        }
        asm.PushDeref(file, otherVar.Addr())

    case types.InterfaceType:
        asm.PushVal(file, otherVar.GetType().String())
        asm.PushDeref(file, otherVar.Addr())

    default:
        asm.PushDeref(file, otherVar.Addr())
    }
}

func PassRegStack(file *bufio.Writer, argType types.Type) {
    switch t := argType.(type) {
    case types.StrType, types.InterfaceType:
        asm.PushReg(file, asm.RegGroup(1))
        asm.PushReg(file, asm.RegGroup(0))

    case types.StructType, types.EnumType:
        if t.Size() > uint(8) {
            asm.PushReg(file, asm.RegGroup(1))
        }
        asm.PushReg(file, asm.RegGroup(0))

    default:
        asm.PushReg(file, asm.RegGroup(0))
    }
}

func passBigStructLit(file *bufio.Writer, t types.StructType, value constVal.StructConst, dstAddr addr.Addr) {
    for i,f := range value.Fields {
        passConst(file, t.Types[i], f, dstAddr)
        dstAddr.Offset += int64(t.Types[i].Size())
    }
}

func passConst(file *bufio.Writer, t types.Type, value constVal.ConstVal, dstAddr addr.Addr) {
    switch value := value.(type) {
    case *constVal.StrConst:
        asm.MovDerefVal(file, dstAddr, types.Ptr_Size, fmt.Sprintf("_str%d", uint64(*value)))
        asm.MovDerefVal(file, dstAddr.Offseted(int64(types.Ptr_Size)), types.I32_Size, fmt.Sprint(str.GetSize(uint64(*value))))

    case *constVal.StructConst:
        passBigStructLit(file, t.(types.StructType), *value, dstAddr)
        
    case *constVal.EnumConst:
        passEnumLit(file, *value, dstAddr)

    default:
        asm.MovDerefVal(file, dstAddr, t.Size(), value.GetVal())
    }
}

func passEnumLit(file *bufio.Writer, value constVal.EnumConst, dstAddr addr.Addr) {
    asm.MovDerefVal(file, dstAddr, value.Type.IdType.Size(), fmt.Sprint(value.Id))
    if value.Elem != nil {
        passConst(file, value.ElemType, value.Elem, dstAddr.Offseted(int64(value.Type.IdType.Size())))
    }
}

func PassBigStructLit(file *bufio.Writer, t types.StructType, value constVal.StructConst) {
    passBigStructLit(file, t, value, asm.RegAsAddr(asm.RegC))
}

func PassBigStructVar(file *bufio.Writer, t types.StructType, v vars.Var, offset int64) {
    dstAddr := asm.RegAsAddr(asm.RegC).Offseted(offset)
    srcAddr := v.Addr().Offseted(offset)
    DerefSetDeref(file, dstAddr, t, srcAddr)
}

func PassBigStructReg(file *bufio.Writer, addr addr.Addr, e ast.Expr) {
    DerefSetBigStruct(file, addr, e)
}

func RetBigStructLit(file *bufio.Writer, t types.StructType, val constVal.StructConst) {
    PassBigStructLit(file, t, val)
}

func RetBigStructVar(file *bufio.Writer, t types.StructType, v vars.Var) {
    PassBigStructVar(file, t, v, 0)
}

func RetBigStructExpr(file *bufio.Writer, address addr.Addr, e ast.Expr) {
    if types.IsBigStruct(e.GetType()) {
        switch lit := e.(type) {
        case *ast.StructLit:
            a := addr.Addr{ BaseAddr: address.BaseAddr, Offset: address.Offset }
            for i,f := range lit.Fields {
                if c := cmpTime.ConstEval(f.Value); c != nil {
                    DerefSetVal(file, a, lit.StructType.Types[i], c)
                } else {
                    asm.UseReg(asm.RegC)
                    DerefSetExpr(file, a, lit.StructType.Types[i], f.Value)
                    asm.FreeReg(asm.RegC)
                }
                a.Offset += int64(lit.StructType.Types[i].Size())
            }
        case *ast.VectorLit:
            if lit.Len != nil {
                if c := cmpTime.ConstEval(lit.Len); c != nil {
                    DerefSetVal(file, address.Offseted(int64(2*types.Ptr_Size)), types.CreateUint(types.U64_Size), c)
                } else {
                    asm.UseReg(asm.RegC)
                    DerefSetExpr(file, address.Offseted(int64(2*types.Ptr_Size)), types.CreateUint(types.U64_Size), lit.Len)
                    asm.FreeReg(asm.RegC)
                }
            } else {
                asm.MovDerefVal(file, address.Offseted(int64(2*types.Ptr_Size)), types.U64_Size, "0")
            }

            baseTypeSize := uint64(lit.Type.BaseType.Size())

            if lit.Cap == nil {
                asm.MovRegDeref(file, asm.RegDi, address.Offseted(int64(2*types.Ptr_Size)), types.U64_Size, false)
                asm.Lea(file, asm.RegDi, fmt.Sprintf("%s*%d", asm.GetReg(asm.RegDi, types.Ptr_Size), baseTypeSize), types.Ptr_Size)
            } else {
                if c := cmpTime.ConstEval(lit.Cap); c != nil {
                    asm.MovDerefVal(file, address.Offseted(int64(types.Ptr_Size)), types.U64_Size, c.GetVal())
                    asm.MovRegVal(file, asm.RegDi, types.U64_Size, fmt.Sprintf("%s*%d", c.GetVal(), baseTypeSize))
                } else {
                    asm.UseReg(asm.RegC)
                    GenExpr(file, lit.Cap)
                    asm.FreeReg(asm.RegC)
                    asm.MovDerefReg(file, address.Offseted(int64(types.Ptr_Size)), types.U64_Size, asm.RegGroup(0))
                    asm.Lea(file, asm.RegDi, fmt.Sprintf("%s*%d", asm.GetReg(asm.RegA, types.Ptr_Size), baseTypeSize), types.Ptr_Size)
                }
            }

            asm.SaveReg(file, asm.RegC)
            file.WriteString("call malloc\n")
            asm.RestoreReg(file, asm.RegC)
            asm.MovDerefReg(file, address, types.Ptr_Size, asm.RegGroup(0))

        default:
            PassBigStructReg(file, address, e)
        }
    } else {
        PassBigStructReg(file, address, e)
    }
}



func PackValues(types []types.Type, values []constVal.ConstVal) []string {
    return packValues(types, values, nil, 0)
}

func PackFields(file *bufio.Writer, typ types.StructType, fields []ast.FieldLit) {
    if len(fields) > 1 {
        addr := addr.Addr{ BaseAddr: "rsp", Offset: -8 }
        if typ.Size() > 8 {
            addr.Offset = -16
        }

        for i,f := range fields {
            t := typ.Types[i]

            switch t.GetKind() {
            case types.Str:
                GenExpr(file, f.Value)
                asm.MovDerefReg(file, addr.Offseted(int64(types.Ptr_Size)), types.Ptr_Size, asm.RegA)
                asm.MovDerefReg(file, addr, types.U32_Size, asm.RegD)
                addr.Offset += int64(types.U32_Size)

            case types.Struct:
                GenExpr(file, f.Value)
                if t.Size() > 8 {
                    asm.MovDerefReg(file, addr.Offseted(int64(types.Ptr_Size)), t.Size(), asm.RegA)
                    asm.MovDerefReg(file, addr, t.Size() - 8, asm.RegD)
                    addr.Offset += int64(t.Size()) - 8
                } else {
                    asm.MovDerefReg(file, addr, t.Size(), asm.RegA)
                    addr.Offset += int64(t.Size())
                }

            default:
                GenExpr(file, f.Value)
                asm.MovDerefReg(file, addr, t.Size(), asm.RegA)
                addr.Offset += int64(t.Size())
            }
        }

        addr.Offset = -8
        asm.MovRegDeref(file, asm.RegGroup(0), addr, types.Ptr_Size, false)
        if typ.Size() > 8 {
            addr.Offset = -16
            asm.MovRegDeref(file, asm.RegGroup(1), addr, types.Ptr_Size, false)
        }
    } else {
        GenExpr(file, fields[0].Value)
    }
}

func packValues(valtypes []types.Type, values []constVal.ConstVal, packed []string, offset uint) []string {
    for i,v := range values {
        switch v := v.(type) {
        case *constVal.StrConst:
            packed = append(packed, fmt.Sprintf("_str%d", uint64(*v)))
            if valtypes[i].GetKind() == types.Str { // check for *char cast
                packed = append(packed, fmt.Sprint(str.GetSize(uint64(*v))))
                offset += types.I32_Size
            }

        case *constVal.StructConst:
            packed = packValues(valtypes[i].(types.StructType).Types, v.Fields, packed, offset)
            offset += valtypes[i].Size() % 8

        case *constVal.ArrConst, *constVal.PtrConst:
            packed = pack(packed, v.GetVal(), offset, valtypes[i])
            continue

        case *constVal.IntConst, *constVal.UintConst, *constVal.BoolConst, *constVal.CharConst:
            packed = pack(packed, v.GetVal(), offset, valtypes[i])
            offset += valtypes[i].Size()

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] cannot pack value of type %v yet\n", reflect.TypeOf(valtypes[i]))
            os.Exit(1)
        }

        if offset > 8 {
            offset -= 8
        }
    }

    return packed
}

func pack(packed []string, newVal string, offset uint, t types.Type) []string {
    if packed == nil || t.Size() + offset > 8 {
        return append(packed, newVal)
    }

    packed[len(packed)-1] = fmt.Sprintf("(%s<<%d)|(%s&%d)", newVal, offset*8, packed[len(packed)-1], (1 << (offset*8))-1)
    return packed
}
