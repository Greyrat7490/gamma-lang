package gen

import (
    "os"
    "fmt"
    "reflect"
    "gamma/token"
    "gamma/types"
    "gamma/types/str"
    "gamma/types/addr"
    "gamma/cmpTime"
    "gamma/cmpTime/constVal"
    "gamma/ast"
    "gamma/ast/identObj/func"
    "gamma/ast/identObj/vars"
    "gamma/ast/identObj/consts"
    "gamma/gen/asm/x86_64"
    "gamma/gen/asm/x86_64/conditions"
)

func GenExpr(file *os.File, e ast.Expr) {
    switch e := e.(type) {
    case *ast.IntLit:
        GenIntLit(file, e.Type.Size(), e)
    case *ast.UintLit:
        GenUintLit(file, e.Type.Size(), e)
    case *ast.CharLit:
        GenCharLit(file, e)
    case *ast.BoolLit:
        GenBoolLit(file, e)
    case *ast.PtrLit:
        GenPtrLit(file, e)

    case *ast.StrLit:
        GenStrLit(file, e)
    case *ast.ArrayLit:
        GenArrayLit(file, e)
    case *ast.StructLit:
        GenStructLit(file, e)

    case *ast.Indexed:
        GenIndexed(file, e)
    case *ast.Field:
        GenField(file, e)

    case *ast.Ident:
        GenIdent(file, e)

    case *ast.FnCall:
        switch e.Ident.Name {
        case "_syscall":
            GenSyscall(file, e.Values[0])
        case "_asm":
            GenInlineAsm(file, e.Values[0])
        default:
            GenFnCall(file, e)
        }

    case *ast.Unary:
        GenUnary(file, e)
    case *ast.Binary:
        GenBinary(file, e)
    case *ast.Paren:
        GenParen(file, e)

    case *ast.XSwitch:
        GenXSwitch(file, e)

    case *ast.Cast:
        GenExpr(file, e.Expr)

    case *ast.BadExpr:
        fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
        os.Exit(1)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] GenExpr for %v is not implemente yet\n", reflect.TypeOf(e))
        os.Exit(1)
    }
}

func GenIntLit(file *os.File, size uint, e *ast.IntLit) {
    asm.MovRegVal(file, asm.RegGroup(0), e.Type.Size(), fmt.Sprint(e.Repr))
}

func GenUintLit(file *os.File, size uint, e *ast.UintLit) {
    asm.MovRegVal(file, asm.RegGroup(0), e.Type.Size(), fmt.Sprint(e.Repr))
}

func GenCharLit(file *os.File, e *ast.CharLit) {
    asm.MovRegVal(file, asm.RegGroup(0), types.Char_Size, fmt.Sprint(e.Repr))
}

func GenPtrLit(file *os.File, e *ast.PtrLit) {
    if e.Local {
        file.WriteString(fmt.Sprintf("lea %s, [%s]\n", asm.GetReg(asm.RegA, types.Ptr_Size), e.Addr))
        asm.MovRegReg(file, asm.RegGroup(0), asm.RegA, types.Ptr_Size)
    } else {
        asm.MovRegVal(file, asm.RegGroup(0), types.Ptr_Size, e.Addr.String())
    }
}

func GenBoolLit(file *os.File, e *ast.BoolLit) {
    if e.Repr {
        asm.MovRegVal(file, asm.RegGroup(0), types.Bool_Size, "1")
    } else {
        asm.MovRegVal(file, asm.RegGroup(0), types.Bool_Size, "0")
    }
}

func GenStrLit(file *os.File, e *ast.StrLit) {
    asm.MovRegVal(file, asm.RegGroup(0), types.Ptr_Size, fmt.Sprintf("_str%d", e.Idx))
    asm.MovRegVal(file, asm.RegGroup(1), types.I32_Size, fmt.Sprintf("%d", str.GetSize(e.Idx)))
}

func GenStructLit(file *os.File, e *ast.StructLit) {
    if types.IsBigStruct(e.StructType) {
        fmt.Fprintf(os.Stderr, "[ERROR] (internal) called GenStructLit with a big struct type %v\n", e.StructType)
        os.Exit(1)
    }

    if c,ok := cmpTime.ConstEvalStructLit(e).(*constVal.StructConst); ok {
        vs := PackValues(e.StructType.Types, c.Fields)
        asm.MovRegVal(file, asm.RegGroup(0), types.Ptr_Size, vs[0])
        if len(vs) == 2 {
            asm.MovRegVal(file, asm.RegGroup(1), e.StructType.Size() - 8, vs[1])
        }
    } else {
        PackFields(file, e.StructType, e.Fields)
    }
}

func GenArrayLit(file *os.File, e *ast.ArrayLit) {
    asm.MovRegVal(file, asm.RegGroup(0), types.Ptr_Size, fmt.Sprintf("_arr%d", e.Idx))
}

func indexedBaseAddrToReg(file *os.File, e *ast.Indexed) {
    if indexed,ok := e.ArrExpr.(*ast.Indexed); ok {
        indexedBaseAddrToReg(file, indexed)
    } else {
        GenExpr(file, e.ArrExpr)
    }
}

func IndexedAddrToReg(file *os.File, e *ast.Indexed, r asm.RegGroup) {
    indexedBaseAddrToReg(file, e)

    baseTypeSize := uint64(e.ArrType.BaseType.Size())

    idxExpr := e.Flatten()
    if idx,ok := cmpTime.ConstEvalUint(idxExpr); ok {
        file.WriteString(fmt.Sprintf("lea %s, [rax+%d]\n", asm.GetReg(r, types.Ptr_Size), idx * baseTypeSize))
    } else {
        asm.MovRegReg(file, asm.RegD, asm.RegA, types.Ptr_Size)
        GenExpr(file, idxExpr)

        asm.Mul(file, fmt.Sprint(baseTypeSize), types.Ptr_Size, false)
        asm.Add(file, asm.GetReg(asm.RegD, types.Ptr_Size), types.Ptr_Size)

        if r != asm.RegA {
            asm.MovRegReg(file, r, asm.RegA, types.Ptr_Size)
        }
    }
}

func GenIndexed(file *os.File, e *ast.Indexed) {
    IndexedAddrToReg(file, e, asm.RegC)
    addr := asm.RegAsAddr(asm.RegC)

    switch t := e.ArrType.BaseType.(type) {
    case types.StrType:
        asm.MovRegDeref(file, asm.RegGroup(0), addr, types.Ptr_Size, false)
        asm.MovRegDeref(file, asm.RegGroup(1), addr.Offseted(int64(types.Ptr_Size)), types.U32_Size, false)

    case types.StructType:
        if t.Size() > uint(8) {
            asm.MovRegDeref(file, asm.RegGroup(0), addr, types.Ptr_Size, false)
            asm.MovRegDeref(
                file,
                asm.RegGroup(1),
                addr.Offseted(int64(t.Size() - 8)),
                t.Size() - 8,
                false,
            )
        } else {
            asm.MovRegDeref(file, asm.RegGroup(0), addr, t.Size(), false)
        }

    case types.IntType:
        asm.MovRegDeref(
            file,
            asm.RegGroup(0),
            addr,
            e.ArrType.BaseType.Size(),
            true,
        )

    default:
        asm.MovRegDeref(
            file,
            asm.RegGroup(0),
            addr,
            e.ArrType.BaseType.Size(),
            false,
        )
    }
}

func FieldAddrToReg(file *os.File, e *ast.Field, r asm.RegGroup) {
    switch o := e.Obj.(type) {
    case *ast.Ident:
        file.WriteString(fmt.Sprintf("lea %s, [%s]\n", asm.GetReg(r, types.Ptr_Size), o.Obj.Addr()))

    case *ast.Field:
        FieldAddrToReg(file, o, r)

    default:
        fmt.Fprintln(os.Stderr, "[ERROR] only ident and field expr supported yet")
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }
}

func FieldToOffset(f *ast.Field) int {
    switch o := f.Obj.(type) {
    case *ast.Ident:
        return f.StructType.GetOffset(f.FieldName.Str)

    case *ast.Field:
        return f.StructType.GetOffset(f.FieldName.Str) + FieldToOffset(o)

    default:
        fmt.Fprintln(os.Stderr, "[ERROR] only ident and field expr supported yet")
        fmt.Fprintln(os.Stderr, "\t" + f.At())
        os.Exit(1)
    }

    return 0
}

func GenField(file *os.File, e *ast.Field) {
    if t,ok := e.Obj.GetType().(types.ArrType); ok {
        asm.MovRegVal(file, asm.RegGroup(0), types.Ptr_Size, fmt.Sprint(t.Lens[0]))
    } else {
        switch o := e.Obj.(type) {
        case *ast.Ident:
            offset := int64(e.StructType.GetOffset(e.FieldName.Str))

            switch t := e.Type.(type) {
            case types.StrType:
                asm.MovRegDeref(file, asm.RegGroup(0), o.Obj.Addr().Offseted(offset), types.Ptr_Size, false)
                asm.MovRegDeref(
                    file,
                    asm.RegGroup(1),
                    o.Obj.Addr().Offseted(offset + int64(types.Ptr_Size)),
                    types.U32_Size,
                    false,
                )

            case types.StructType:
                if t.Size() > uint(8) {
                    asm.MovRegDeref(file, asm.RegGroup(0), o.Obj.Addr().Offseted(offset), types.Ptr_Size, false)
                    asm.MovRegDeref(
                        file,
                        asm.RegGroup(1),
                        o.Obj.Addr().Offseted(offset + int64(types.Ptr_Size)),
                        t.Size() - 8,
                        false,
                    )
                } else {
                    asm.MovRegDeref(file, asm.RegGroup(0), o.Obj.Addr().Offseted(offset), t.Size(), false)
                }

            case types.IntType:
                asm.MovRegDeref(file, asm.RegGroup(0), o.Obj.Addr().Offseted(offset), t.Size(), true)

            default:
                asm.MovRegDeref(file, asm.RegGroup(0), o.Obj.Addr().Offseted(offset), t.Size(), false)
            }

        case *ast.Field:
            FieldAddrToReg(file, o, asm.RegC)
            addr := asm.RegAsAddr(asm.RegC)

            offset := FieldToOffset(e)
            switch t := e.Type.(type) {
            case types.StrType:
                asm.MovRegDeref(
                    file,
                    asm.RegGroup(0),
                    addr.Offseted(int64(offset)),
                    types.Ptr_Size,
                    false,
                )
                asm.MovRegDeref(
                    file,
                    asm.RegGroup(1),
                    addr.Offseted(int64(offset + int(types.Ptr_Size))),
                    types.I32_Size,
                    false,
                )

            case types.StructType:
                if t.Size() > uint(8) {
                    asm.MovRegDeref(
                        file,
                        asm.RegGroup(0),
                        addr.Offseted(int64(offset)),
                        types.Ptr_Size,
                        false,
                    )
                    asm.MovRegDeref(
                        file,
                        asm.RegGroup(1),
                        addr.Offseted(int64(offset + int(types.Ptr_Size))),
                        t.Size() - 8,
                        false,
                    )
                } else {
                    asm.MovRegDeref(file, asm.RegGroup(0), addr.Offseted(int64(offset)), t.Size(), false)
                }

            case types.IntType:
                asm.MovRegDeref(file, asm.RegGroup(0), addr.Offseted(int64(offset)), t.Size(), true)

            default:
                asm.MovRegDeref(file, asm.RegGroup(0), addr.Offseted(int64(offset)), t.Size(), false)
            }

        default:
            fmt.Fprintln(os.Stderr, "[ERROR] only ident and field expr supported yet")
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }
    }
}

func GenIdent(file *os.File, e *ast.Ident) {
    if c,ok := e.Obj.(*consts.Const); ok {
        GenConstVal(file, e.GetType(), c.GetVal())
        return
    }

    if v,ok := e.Obj.(vars.Var); ok {
        switch t := v.GetType().(type) {
        case types.StrType:
            asm.MovRegDeref(file, asm.RegGroup(0), v.Addr(), types.Ptr_Size, false)
            asm.MovRegDeref(file, asm.RegGroup(1), v.Addr().Offseted(int64(types.Ptr_Size)), types.I32_Size, false)

        case types.StructType:
            if t.Size() > uint(8) {
                asm.MovRegDeref(file, asm.RegGroup(0), v.Addr(), types.Ptr_Size, false)
                asm.MovRegDeref(file, asm.RegGroup(1), v.Addr().Offseted(int64(types.Ptr_Size)), t.Size() - 8, false)
            } else {
                asm.MovRegDeref(file, asm.RegGroup(0), v.Addr(), t.Size(), false)
            }

        case types.IntType:
            asm.MovRegDeref(file, asm.RegGroup(0), v.Addr(), t.Size(), false)

        default:
            asm.MovRegDeref(file, asm.RegGroup(0), v.Addr(), t.Size(), false)
        }
        return
    }

    if _,ok := e.Obj.(*fn.Func); ok {
        fmt.Fprintf(os.Stderr, "[ERROR] TODO: expr.go compile Ident for functions\n")
        os.Exit(1)
        return
    }

    fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not declared\n", e.Name)
    fmt.Fprintln(os.Stderr, "\t" + e.Pos.At())
    os.Exit(1)
}

func GenParen(file *os.File, e *ast.Paren) {
    GenExpr(file, e.Expr)
}

func GenUnary(file *os.File, e *ast.Unary) {
    if c := cmpTime.ConstEval(e); c != nil {
        asm.MovRegVal(file, asm.RegA, e.Operand.GetType().Size(), c.GetVal())
        return
    }

    GenExpr(file, e.Operand)

    switch e.Operator.Type {
    case token.Minus:
        asm.Neg(file, e.Operand.GetType().Size())

    case token.BitNot:
        asm.Not(file, e.Operand.GetType().Size())

    case token.Mul:
        t := e.GetType()
        asm.DerefRax(file, t.Size(), t.GetKind() == types.Int)
    }
}

func GenBinary(file *os.File, e *ast.Binary) {
    // compile time evaluation (constEval whole expr)
    if c := cmpTime.ConstEval(e); c != nil {
        asm.MovRegVal(file, asm.RegA, e.GetType().Size(), c.GetVal())
        return
    }


    if e.Operator.Type != token.And && e.Operator.Type != token.Or {
        // compile time evaluation (constEval only left expr)
        if c := cmpTime.ConstEval(e.OperandL); c != nil {
            asm.MovRegVal(file, asm.RegA, e.OperandR.GetType().Size(), c.GetVal())
        } else {
            GenExpr(file, e.OperandL)

            // compile time evaluation (constEval only right expr)
            if c := cmpTime.ConstEval(e.OperandR); c != nil {
                asm.BinaryOp(file, e.Operator.Type, c.GetVal(), e.OperandL.GetType().Size(), e.GetType().GetKind() == types.Int)
                return
            }
        }

        if ident,ok := e.OperandR.(*ast.Ident); ok {
            if v,ok := ident.Obj.(vars.Var); ok {
                t := v.GetType()

                asm.BinaryOp(file,
                    e.Operator.Type,
                    fmt.Sprintf("%s [%s]", asm.GetWord(t.Size()), v.Addr().String()),
                    t.Size(),
                    t.GetKind() == types.Int)
            }
        } else {
            asm.PushReg(file, asm.RegA)

            GenExpr(file, e.OperandR)
            asm.MovRegReg(file, asm.RegB, asm.RegA, e.OperandR.GetType().Size())

            asm.PopReg(file, asm.RegA)
            asm.BinaryOpReg(file, e.Operator.Type, asm.RegB, e.OperandR.GetType().Size(), e.GetType().GetKind() == types.Int)
        }

    // &&, ||
    } else {
        // compile time evaluation
        if b,ok := cmpTime.ConstEval(e.OperandL).(*constVal.BoolConst); ok {
            if e.Operator.Type == token.And && !bool(*b) {
                asm.MovRegVal(file, asm.RegA, types.Bool_Size, "0")
                return
            }
            if e.Operator.Type == token.Or && bool(*b) {
                asm.MovRegVal(file, asm.RegA, types.Bool_Size, "1")
                return
            }

            GenExpr(file, e.OperandR)
        } else {
            GenExpr(file, e.OperandL)

            count := cond.LogicalOp(file, e.Operator)
            GenExpr(file, e.OperandR)
            cond.LogicalOpEnd(file, count)
        }
    }
}

func GenFnCall(file *os.File, e *ast.FnCall) {
    regIdx := uint(0)
    if types.IsBigStruct(e.F.GetRetType()) { // rdi contains addr to return big struct to
        regIdx++
    }

    // get start of args on stack, calc big args stack size % 16 ----
    rest := uint(0)
    stackArgsStart := len(e.F.GetArgs())
    for i,t := range e.F.GetArgs() {
        if types.IsBigStruct(t) {
            rest += (t.Size() + 7) & ^uint(7)

        } else if stackArgsStart == len(e.F.GetArgs()) {
            needed := types.RegCount(t)
            if regIdx + needed > 6 {
                stackArgsStart = i
                if rest != 0 {
                    rest += 8
                }
            } else {
                regIdx += needed
            }
        }
    }
    rest %= 16

    // pass args on stack -------------------------------------------
    for i := len(e.F.GetArgs())-1; i >= stackArgsStart; i-- {
        if v := cmpTime.ConstEval(e.Values[i]); v != nil {
            PassValStack(file, v, e.F.GetArgs()[i])

        } else if ident,ok := e.Values[i].(*ast.Ident); ok {
            PassVarStack(file, ident.Obj.(vars.Var))

        } else {
            GenExpr(file, e.Values[i])
            PassRegStack(file, e.F.GetArgs()[i])
        }
    }

    // align stack (16byte) -----------------------------------------
    bigArgsSize := uint(0)
    if rest != 0 {
        file.WriteString(fmt.Sprintf("sub rsp, %d\n", rest))
        bigArgsSize += rest
    }

    // pass big struct args -----------------------------------------
    for i := len(e.F.GetArgs())-1; i >= 0; i-- {
        if t,ok := e.F.GetArgs()[i].(types.StructType); ok {
            if types.IsBigStruct(t) {
                size := (t.Size() + 7) & ^uint(7)
                bigArgsSize += size

                file.WriteString(fmt.Sprintf("sub rsp, %d\n", size))

                if v := cmpTime.ConstEval(e.Values[i]); v != nil {
                    file.WriteString("mov rcx, rsp\n")
                    PassBigStructLit(file, t, *v.(*constVal.StructConst))

                } else if ident,ok := e.Values[i].(*ast.Ident); ok {
                    file.WriteString("mov rcx, rsp\n")
                    PassBigStructVar(file, t, ident.Obj.(vars.Var), 0)

                } else {
                    file.WriteString("mov rcx, rsp\n")
                    PassBigStructReg(file, asm.RegAsAddr(asm.RegC), e.Values[i])
                }
            }
        }
    }

    // pass args with regs -----------------------------------------
    for i := stackArgsStart-1; i >= 0; i-- {
        t := e.F.GetArgs()[i]
        if types.IsBigStruct(t) {
            continue
        }

        regIdx -= types.RegCount(t)

        if v := cmpTime.ConstEval(e.Values[i]); v != nil {
            PassVal(file, regIdx, v, t)

        } else if ident,ok := e.Values[i].(*ast.Ident); ok {
            PassVar(file, regIdx, t, ident.Obj.(vars.Var))

        } else {
            PassExpr(file, regIdx, t, e.Values[i].GetType().Size(), e.Values[i])
        }
    }

    CallFn(file, e.F)

    // clear stack -------------------------------------------------
    if bigArgsSize > 0 {
        file.WriteString(fmt.Sprintf("add rsp, %d\n", bigArgsSize))
    }
}

func GenXCase(file *os.File, e *ast.XCase, switchCount uint) {
    cond.CaseStart(file)

    if e.Cond == nil {
        cond.CaseBody(file)
        GenExpr(file, e.Expr)
        return
    }

    if val,ok := cmpTime.ConstEval(e.Cond).(*constVal.BoolConst); ok {
        if bool(*val) {
            cond.CaseBody(file)
            GenExpr(file, e.Expr)
            cond.CaseBodyEnd(file, switchCount)
        }

        return
    }

    if i,ok := e.Cond.(*ast.Ident); ok {
        cond.CaseVar(file, i.Obj.Addr())
    } else {
        GenExpr(file, e.Cond)
        cond.CaseExpr(file)
    }

    cond.CaseBody(file)
    GenExpr(file, e.Expr)
    cond.CaseBodyEnd(file, switchCount)
}

func GenXSwitch(file *os.File, e *ast.XSwitch) {
    if c := cmpTime.ConstEval(e); c != nil {
        GenConstVal(file, e.GetType(), c)
        return
    }

    count := cond.StartSwitch()

    for i := 0; i < len(e.Cases)-1; i++ {
        GenXCase(file, &e.Cases[i], count)
    }
    cond.InLastCase()
    GenXCase(file, &e.Cases[len(e.Cases)-1], count)

    cond.EndSwitch(file)
}



func DerefSetBigStruct(file *os.File, address addr.Addr, e ast.Expr) {
    if !types.IsBigStruct(e.GetType()) {
        fmt.Fprintln(os.Stderr, "[ERROR] expected expr to be a big struct")
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    switch e := e.(type) {
    case *ast.StructLit:
        a := addr.Addr{ BaseAddr: address.BaseAddr, Offset: address.Offset }
        for i,f := range e.Fields {
            if c := cmpTime.ConstEval(f.Value); c != nil {
                DerefSetVal(file, a, e.StructType.Types[i], c)
            } else {
                DerefSetExpr(file, a, e.StructType.Types[i], f.Value)
            }
            a.Offset += int64(e.StructType.Types[i].Size())
        }

    case *ast.Indexed:
        IndexedAddrToReg(file, e, asm.RegA)
        DerefSetDeref(file, address, e.GetType(), asm.RegAsAddr(asm.RegA))

    case *ast.Field:
        FieldAddrToReg(file, e, asm.RegA)
        file.WriteString(fmt.Sprintf("lea rax, [rax+%d]\n", FieldToOffset(e)))
        DerefSetDeref(file, address, e.GetType(), asm.RegAsAddr(asm.RegA))

    case *ast.Ident:
        if v,ok := e.Obj.(vars.Var); ok {
            DerefSetVar(file, address, v)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected identifier %s to be a variable but got %v\n", e.Name, reflect.TypeOf(e.Obj))
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }

    case *ast.Unary:
        GenExpr(file, e.Operand)
        DerefSetDeref(file, address, e.GetType(), asm.RegAsAddr(asm.RegA))

    case *ast.FnCall:
        file.WriteString(fmt.Sprintf("lea rdi, [%s]\n", address))
        GenExpr(file, e)

    case *ast.Paren:
        DerefSetBigStruct(file, address, e.Expr)

    case *ast.XSwitch:
        bigStructXSwitchToStack(file, address, e)

    case *ast.BadExpr:
        fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
        os.Exit(1)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] DerefSetBigStruct for %v is not implemente yet\n", reflect.TypeOf(e))
        os.Exit(1)
    }
}

func bigStructXSwitchToStack(file *os.File, addr addr.Addr, e *ast.XSwitch) {
    if c := cmpTime.ConstEval(e); c != nil {
        GenConstVal(file, e.GetType(), c)
        return
    }

    count := cond.StartSwitch()

    for i := 0; i < len(e.Cases)-1; i++ {
        bigStructXCaseToStack(file, addr, &e.Cases[i], count)
    }
    cond.InLastCase()
    bigStructXCaseToStack(file, addr, &e.Cases[len(e.Cases)-1], count)

    cond.EndSwitch(file)
}

func bigStructXCaseToStack(file *os.File, addr addr.Addr, e *ast.XCase, switchCount uint) {
    cond.CaseStart(file)

    if e.Cond == nil {
        cond.CaseBody(file)
        DerefSetBigStruct(file, addr, e.Expr)
        return
    }

    if val,ok := cmpTime.ConstEval(e.Cond).(*constVal.BoolConst); ok {
        if bool(*val) {
            cond.CaseBody(file)
            DerefSetBigStruct(file, addr, e.Expr)
            cond.CaseBodyEnd(file, switchCount)
        }

        return
    }

    if i,ok := e.Cond.(*ast.Ident); ok {
        cond.CaseVar(file, i.Obj.Addr())
    } else {
        GenExpr(file, e.Cond)
        cond.CaseExpr(file)
    }

    cond.CaseBody(file)
    DerefSetBigStruct(file, addr, e.Expr)
    cond.CaseBodyEnd(file, switchCount)
}

func GenSyscall(file *os.File, val ast.Expr) {
    if v := cmpTime.ConstEval(val); v != nil {
        asm.MovRegVal(file, asm.RegA, types.Ptr_Size, v.GetVal())
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] _syscall takes only const")
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }

    file.WriteString("syscall\n")
}

func GenInlineAsm(file *os.File, val ast.Expr) {
    if str,ok := val.(*ast.StrLit); ok {
        file.WriteString(str.Val.Str[1:len(str.Val.Str)-1] + "\n")
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] _asm takes only a string literal")
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }
}

func GenConstVal(file *os.File, t types.Type, val constVal.ConstVal) {
    switch c := val.(type) {
    case *constVal.StrConst:
        asm.MovRegVal(file, asm.RegGroup(0), types.Ptr_Size, fmt.Sprintf("_str%d", uint64(*c)))
        asm.MovRegVal(file, asm.RegGroup(1), types.I32_Size, fmt.Sprintf("%d", str.GetSize(uint64(*c))))

    case *constVal.PtrConst:
        asm.MovRegVal(file, asm.RegGroup(0), types.Ptr_Size, PtrConstToAddr(file, *c))

    default:
        asm.MovRegVal(file, asm.RegGroup(0), t.Size(), c.GetVal())
    }
}

func PtrConstToAddr(file *os.File, c constVal.PtrConst) string {
    if c.Local {
        file.WriteString(fmt.Sprintf("lea %s, [%s]\n", asm.GetReg(asm.RegA, types.Ptr_Size), c.Addr))
        return asm.GetReg(asm.RegA, types.Ptr_Size)
    } else {
        return c.Addr.String()
    }
}
