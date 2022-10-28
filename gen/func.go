package gen

import (
    "os"
    "fmt"
    "reflect"
    "gamma/types"
    "gamma/types/str"
    "gamma/ast"
    "gamma/ast/identObj/func"
    "gamma/ast/identObj/vars"
    "gamma/cmpTime/constVal"
    "gamma/gen/asm/x86_64"
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
*/

var regs []asm.RegGroup = []asm.RegGroup{ asm.RegDi, asm.RegSi, asm.RegD, asm.RegC, asm.RegR8, asm.RegR9 }

func Define(file *os.File, f *fn.Func) {
    file.WriteString(f.GetName() + ":\n")
    file.WriteString("push rbp\nmov rbp, rsp\n")
    if f.GetFrameSize() > 0 {
        file.WriteString(fmt.Sprintf("sub rsp, %d\n", f.GetFrameSize()))
    }
}

func FnEnd(file *os.File) {
    file.WriteString("leave\n")
    file.WriteString("ret\n")
}

func CallFn(file *os.File, f *fn.Func) {
    file.WriteString("call " + f.GetName() + "\n")
}

func DefArg(file *os.File, regIdx uint, v vars.Var) {
    switch t := v.GetType().(type) {
    case types.StrType:
        asm.MovDerefReg(file, v.Addr(0), types.Ptr_Size, regs[regIdx])
        asm.MovDerefReg(file, v.Addr(1), types.I32_Size, regs[regIdx+1])

    case types.StructType:
        if t.Size() > uint(8) {
            asm.MovDerefReg(file, v.Addr(0), types.Ptr_Size, regs[regIdx])
            asm.MovDerefReg(file, v.OffsetedAddr(int(types.Ptr_Size)), t.Size() - 8, regs[regIdx+1])
        } else {
            asm.MovDerefReg(file, v.Addr(0), t.Size(), regs[regIdx])
        }

    default:
        asm.MovDerefReg(file, v.Addr(0), t.Size(), regs[regIdx])
    }
}


func PassVal(file *os.File, regIdx uint, value constVal.ConstVal, valtype types.Type) {
    switch v := value.(type) {
    case *constVal.StrConst:
        asm.MovRegVal(file, regs[regIdx],   types.Ptr_Size, fmt.Sprintf("_str%d", uint64(*v)))
        asm.MovRegVal(file, regs[regIdx+1], types.I32_Size, fmt.Sprint(str.GetSize(uint64(*v))))

    case *constVal.StructConst:
        t := valtype.(types.StructType)

        if len(t.Types) == 1 && t.Types[0].GetKind() != types.Str {
            asm.MovRegVal(file, regs[regIdx], t.Size(), v.Fields[0].GetVal())
        } else {
            vs := PackValues(t.Types, v.Fields)
            asm.MovRegVal(file, regs[regIdx], types.Ptr_Size, vs[0])
            if len(vs) == 2 {
                asm.MovRegVal(file, regs[regIdx+1], t.Size() - 8, vs[1])
            }
        }

    case *constVal.ArrConst, *constVal.IntConst, *constVal.UintConst, *constVal.CharConst, *constVal.BoolConst:
        asm.MovRegVal(file, regs[regIdx], valtype.Size(), value.GetVal())

    case *constVal.PtrConst:
        asm.MovRegVal(file, regs[regIdx], valtype.Size(), PtrConstToAddr(file, *v))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] cannot pass %v yet\n", reflect.TypeOf(value))
        os.Exit(1)
    }
}

func PassVar(file *os.File, regIdx uint, t types.Type, otherVar vars.Var) {
    switch t := t.(type) {
    case types.StrType:
        asm.MovRegDeref(file, regs[regIdx],   otherVar.Addr(0), types.Ptr_Size, false)
        asm.MovRegDeref(file, regs[regIdx+1], otherVar.Addr(1), types.U32_Size, false)

    case types.StructType:
        if t.Size() > uint(8) {
            asm.MovRegDeref(file, regs[regIdx],   otherVar.Addr(0), types.Ptr_Size, false)
            asm.MovRegDeref(file, regs[regIdx+1], otherVar.OffsetedAddr(int(types.Ptr_Size)), t.Size() - 8, false)
        } else {
            asm.MovRegDeref(file, regs[regIdx],   otherVar.Addr(0), t.Size(), false)
        }

    case types.IntType:
        asm.MovRegDerefExtend(file, regs[regIdx], t.Size(), otherVar.Addr(0), otherVar.GetType().Size(), true)

    default:
        asm.MovRegDerefExtend(file, regs[regIdx], t.Size(), otherVar.Addr(0), otherVar.GetType().Size(), false)
    }
}

func PassReg(file *os.File, regIdx uint, argType types.Type, regSize uint) {
    switch t := argType.(type) {
    case types.StrType:
        asm.MovRegReg(file, regs[regIdx],   asm.RegGroup(0), types.Ptr_Size)
        asm.MovRegReg(file, regs[regIdx+1], asm.RegGroup(1), types.U32_Size)

    case types.StructType:
        if t.Size() > uint(8) {
            asm.MovRegReg(file, regs[regIdx], asm.RegGroup(0), types.Ptr_Size)
            asm.MovRegReg(file, regs[regIdx+1], asm.RegGroup(1), t.Size() - 8)
        } else {
            asm.MovRegReg(file, regs[regIdx], asm.RegGroup(0), t.Size())
        }

    case types.IntType:
        asm.MovRegRegExtend(file, regs[regIdx], t.Size(), asm.RegGroup(0), regSize, true)

    default:
        asm.MovRegRegExtend(file, regs[regIdx], t.Size(), asm.RegGroup(0), regSize, false)
    }
}

func PassValStack(file *os.File, value constVal.ConstVal, valtype types.Type) {
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
            file.WriteString(fmt.Sprintf("lea %s, [%s]\n", asm.GetReg(asm.RegA, types.Ptr_Size), v.Addr))
            asm.PushReg(file, asm.RegA)
        } else {
            asm.PushVal(file, v.GetVal())
        }

    default:
        asm.PushVal(file, v.GetVal())
    }
}

func PassVarStack(file *os.File, otherVar vars.Var) {
    switch t := otherVar.GetType().(type) {
    case types.StrType:
        asm.PushDeref(file, otherVar.Addr(1))
        asm.PushDeref(file, otherVar.Addr(0))

    case types.StructType:
        if t.Size() > uint(8) {
            asm.PushDeref(file, otherVar.Addr(1))
        }
        asm.PushDeref(file, otherVar.Addr(0))

    default:
        asm.PushDeref(file, otherVar.Addr(0))
    }
}

func PassRegStack(file *os.File, argType types.Type) {
    switch t := argType.(type) {
    case types.StrType:
        asm.PushReg(file, asm.RegGroup(1))
        asm.PushReg(file, asm.RegGroup(0))

    case types.StructType:
        if t.Size() > uint(8) {
            asm.PushReg(file, asm.RegGroup(1))
        }
        asm.PushReg(file, asm.RegGroup(0))

    default:
        asm.PushReg(file, asm.RegGroup(0))
    }
}

func PassBigStructLit(file *os.File, t types.StructType, value constVal.StructConst, offset int) {
    for i,f := range value.Fields {
        switch f := f.(type) {
        case *constVal.StrConst:
            asm.MovDerefVal(file,
                asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset),
                types.Ptr_Size,
                fmt.Sprintf("_str%d", uint64(*f)))
            asm.MovDerefVal(file,
                asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset + int(types.Ptr_Size)),
                types.I32_Size,
                fmt.Sprint(str.GetSize(uint64(*f))))

        case *constVal.StructConst:
            PassBigStructLit(file, t.Types[i].(types.StructType), *f, offset)

        default:
            asm.MovDerefVal(file, asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset), t.Types[i].Size(), f.GetVal())
        }
        offset += int(t.Types[i].Size())
    }
}

func PassBigStructVar(file *os.File, t types.StructType, v vars.Var, offset int) {
    for i := 0; i < int(t.Size()/types.Ptr_Size); i++ {
        asm.MovDerefDeref(
            file,
            asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset),
            v.OffsetedAddr(offset),
            types.Ptr_Size,
            asm.RegA,
            false,
        )
        offset += int(types.Ptr_Size)
    }

    if size := t.Size() % types.Ptr_Size; size != 0 {
        asm.MovDerefDeref(
            file,
            asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset),
            v.OffsetedAddr(offset),
            size,
            asm.RegA,
            false,
        )
    }
}

func PassBigStructReg(file *os.File, addr string, e ast.Expr) {
    DerefSetBigStruct(file, addr, e)
}

func RetBigStructLit(file *os.File, t types.StructType, val constVal.StructConst) {
    PassBigStructLit(file, t, val, 0)
}

func RetBigStructVar(file *os.File, t types.StructType, v vars.Var) {
    PassBigStructVar(file, t, v, 0)
}

func RetBigStructExpr(file *os.File, addr string, e ast.Expr) {
    PassBigStructReg(file, addr, e)
}



func PackValues(types []types.Type, values []constVal.ConstVal) []string {
    return packValues(types, values, nil, 0)
}

func PackFields(file *os.File, typ types.StructType, fields []ast.FieldLit) {
    if len(fields) > 1 {
        offset := -8
        for i,f := range fields {
            t := typ.Types[i]

            switch t.GetKind() {
            case types.Str:
                GenExpr(file, f.Value)
                asm.MovDerefReg(file, asm.OffsetAddr("rsp", offset), types.Ptr_Size, asm.RegA)
                asm.MovDerefReg(file, asm.OffsetAddr("rsp", offset - int(types.Ptr_Size)), types.U32_Size, asm.RegD)
                offset -= int(types.U32_Size)

            case types.Struct:
                GenExpr(file, f.Value)
                if t.Size() > 8 {
                    asm.MovDerefReg(file, asm.OffsetAddr("rsp", offset), t.Size(), asm.RegA)
                    asm.MovDerefReg(file, asm.OffsetAddr("rsp", offset - int(types.Ptr_Size)), t.Size() - 8, asm.RegD)
                    offset -= int(t.Size()) - 8
                } else {
                    asm.MovDerefReg(file, asm.OffsetAddr("rsp", offset), t.Size(), asm.RegA)
                    offset -= int(t.Size())
                }

            default:
                GenExpr(file, f.Value)
                asm.MovDerefReg(file, asm.OffsetAddr("rsp", offset), t.Size(), asm.RegA)
                offset += int(t.Size())
            }
        }

        asm.MovRegDeref(file, asm.RegGroup(0), "rsp-8", types.Ptr_Size, false)
        if typ.Size() > 8 {
            asm.MovRegDeref(file, asm.RegGroup(1), "rsp-16", types.Ptr_Size, false)
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
            packed = append(packed, fmt.Sprint(str.GetSize(uint64(*v))))
            offset += types.I32_Size

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

    packed[len(packed)-1] = fmt.Sprintf("(%s<<%d)+%s", newVal, offset*8, packed[len(packed)-1])
    return packed
}
