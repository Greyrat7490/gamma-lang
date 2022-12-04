package gen

import (
    "os"
    "fmt"
    "bufio"
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

var preserveRegC bool = false

// TODO vister pattern
func ExprAddrToReg(file *bufio.Writer, e ast.Expr, reg asm.RegGroup) {
    switch e := e.(type) {
    case *ast.Ident:
        asm.Lea(file, reg, e.Obj.Addr().String(), types.Ptr_Size)

    case *ast.Indexed:
        IndexedAddrToReg(file, e, reg)

    case *ast.Field:
        FieldAddrToReg(file, e, reg)

    case *ast.Unary:
        UnaryAddrToReg(file, e, reg)

    case *ast.XSwitch:
        XSwitchAddrToReg(file, e, reg)

    case *ast.Paren:
        ExprAddrToReg(file, e.Expr, reg)

    case *ast.Cast:
        ExprAddrToReg(file, e.Expr, reg)

    case *ast.IntLit, *ast.UintLit, *ast.CharLit, *ast.BoolLit, *ast.PtrLit, *ast.StrLit, *ast.ArrayLit, *ast.StructLit, *ast.Binary, *ast.FnCall:
        fmt.Fprintf(os.Stderr, "[ERROR] cannot get address from %v\n", reflect.TypeOf(e))
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    case *ast.BadExpr:
        fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] ExprAddrToReg for %v is not implemente yet\n", reflect.TypeOf(e))
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }
}

// TODO to a specific reg
func GenExpr(file *bufio.Writer, e ast.Expr) {
    if preserveRegC {
        asm.PushReg(file, asm.RegC)
    }

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

    if preserveRegC {
        asm.PopReg(file, asm.RegC)
    }
}

func GenIntLit(file *bufio.Writer, size uint, e *ast.IntLit) {
    asm.MovRegVal(file, asm.RegA, e.Type.Size(), fmt.Sprint(e.Repr))
}

func GenUintLit(file *bufio.Writer, size uint, e *ast.UintLit) {
    asm.MovRegVal(file, asm.RegA, e.Type.Size(), fmt.Sprint(e.Repr))
}

func GenCharLit(file *bufio.Writer, e *ast.CharLit) {
    asm.MovRegVal(file, asm.RegA, types.Char_Size, fmt.Sprint(e.Repr))
}

func GenPtrLit(file *bufio.Writer, e *ast.PtrLit) {
    if e.Local {
        file.WriteString(fmt.Sprintf("lea %s, [%s]\n", asm.GetReg(asm.RegA, types.Ptr_Size), e.Addr))
    } else {
        asm.MovRegVal(file, asm.RegA, types.Ptr_Size, e.Addr.String())
    }
}

func GenBoolLit(file *bufio.Writer, e *ast.BoolLit) {
    if e.Repr {
        asm.MovRegVal(file, asm.RegA, types.Bool_Size, "1")
    } else {
        asm.MovRegVal(file, asm.RegA, types.Bool_Size, "0")
    }
}

func GenStrLit(file *bufio.Writer, e *ast.StrLit) {
    asm.MovRegVal(file, asm.RegA, types.Ptr_Size, fmt.Sprintf("_str%d", e.Idx))
    asm.MovRegVal(file, asm.RegD, types.I32_Size, fmt.Sprintf("%d", str.GetSize(e.Idx)))
}

func GenStructLit(file *bufio.Writer, e *ast.StructLit) {
    if types.IsBigStruct(e.StructType) {
        fmt.Fprintf(os.Stderr, "[ERROR] (internal) called GenStructLit with a big struct type %v\n", e.StructType)
        os.Exit(1)
    }

    if c,ok := cmpTime.ConstEvalStructLit(e).(*constVal.StructConst); ok {
        vs := PackValues(e.StructType.Types, c.Fields)
        asm.MovRegVal(file, asm.RegA, types.Ptr_Size, vs[0])
        if len(vs) == 2 {
            asm.MovRegVal(file, asm.RegD, e.StructType.Size() - 8, vs[1])
        }
    } else {
        PackFields(file, e.StructType, e.Fields)
    }
}

func GenArrayLit(file *bufio.Writer, e *ast.ArrayLit) {
    asm.MovRegVal(file, asm.RegA, types.Ptr_Size, fmt.Sprintf("_arr%d", e.Idx))
}

func indexedBaseAddrToReg(file *bufio.Writer, e *ast.Indexed) {
    if indexed,ok := e.ArrExpr.(*ast.Indexed); ok {
        indexedBaseAddrToReg(file, indexed)
    } else {
        GenExpr(file, e.ArrExpr)
    }
}

func IndexedAddrToReg(file *bufio.Writer, e *ast.Indexed, r asm.RegGroup) {
    baseTypeSize := uint64(e.Type.Size())
    idxExpr := e.Index

    switch e.ArrType.GetKind() {
    case types.Arr:
        indexedBaseAddrToReg(file, e)
        idxExpr = e.Flatten()

    case types.Vec:
        GenExpr(file, e.ArrExpr)
    }

    if idx,ok := cmpTime.ConstEvalUint(idxExpr); ok {
        file.WriteString(fmt.Sprintf("lea %s, [rax+%d]\n", asm.GetReg(r, types.Ptr_Size), idx * baseTypeSize))
    } else {
        asm.MovRegReg(file, asm.RegD, asm.RegA, types.Ptr_Size)
        GenExpr(file, idxExpr)

        asm.Lea(file, asm.RegA, fmt.Sprintf("%s*%d", asm.GetReg(asm.RegA, types.Ptr_Size), baseTypeSize), types.Ptr_Size)
        asm.Add(file, asm.GetReg(asm.RegD, types.Ptr_Size), types.Ptr_Size)

        if r != asm.RegA {
            asm.MovRegReg(file, r, asm.RegA, types.Ptr_Size)
        }
    }
}

func GenIndexed(file *bufio.Writer, e *ast.Indexed) {
    IndexedAddrToReg(file, e, asm.RegA)
    addr := asm.RegAsAddr(asm.RegA)

    switch t := e.Type.(type) {
    case types.StrType:
        asm.MovRegDeref(file, asm.RegD, addr.Offseted(int64(types.Ptr_Size)), types.U32_Size, false)
        asm.MovRegDeref(file, asm.RegA, addr, types.Ptr_Size, false)

    case types.StructType:
        if t.Size() > uint(8) {
            asm.MovRegDeref(file, asm.RegD, addr.Offseted(int64(t.Size() - 8)), t.Size() - 8, false)
            asm.MovRegDeref(file, asm.RegA, addr, types.Ptr_Size, false)
        } else {
            asm.MovRegDeref(file, asm.RegA, addr, t.Size(), false)
        }

    case types.IntType: 
        asm.MovRegDeref(file, asm.RegA, addr, t.Size(), true)

    default:
        asm.MovRegDeref(file, asm.RegA, addr, t.Size(), false)
    }
}

func FieldAddrToReg(file *bufio.Writer, e *ast.Field, r asm.RegGroup) {
    addr := fieldBaseAddr(e)
    offset := int64(fieldToOffset(e))
    asm.Lea(file, r, addr.Offseted(offset).String(), types.Ptr_Size)
}

func fieldBaseAddr(e *ast.Field) addr.Addr {
    switch o := e.Obj.(type) {
    case *ast.Ident:
        return o.Obj.Addr()

    case *ast.Field:
        return fieldBaseAddr(o)

    default:
        fmt.Fprintln(os.Stderr, "[ERROR] only ident and field expr supported yet")
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
        return addr.Addr{}
    }
}

func fieldToOffset(f *ast.Field) int {
    switch f.Obj.GetType().(type) {
    case types.ArrType:
        return 0
    case types.VecType:
        switch f.FieldName.Str {
        case "cap":
            return int(types.Ptr_Size)
        case "len":
            return int(types.Ptr_Size + types.U64_Size)
        default:
            return 0
        }
    default:
        switch o := f.Obj.(type) {
        case *ast.Ident:
            return f.StructType.GetOffset(f.FieldName.Str)

        case *ast.Field:
            return f.StructType.GetOffset(f.FieldName.Str) + fieldToOffset(o)

        default:
            fmt.Fprintln(os.Stderr, "[ERROR] only ident and field expr supported yet")
            fmt.Fprintln(os.Stderr, "\t" + f.At())
            os.Exit(1)
        }
    }

    return 0
}

func GenField(file *bufio.Writer, e *ast.Field) {
    switch t := e.Obj.GetType().(type) {
    case types.ArrType:
        asm.MovRegVal(file, asm.RegA, types.Ptr_Size, fmt.Sprint(t.Lens[0]))

    case types.VecType:
        FieldAddrToReg(file, e, asm.RegA)
        asm.MovRegDeref(file, asm.RegA, asm.RegAsAddr(asm.RegA), types.U64_Size, false)

    case types.StructType:
        FieldAddrToReg(file, e, asm.RegA)
        addr := asm.RegAsAddr(asm.RegA)

        switch t := e.Type.(type) {
        case types.StrType:
            asm.MovRegDeref(file, asm.RegD, addr.Offseted(int64(types.Ptr_Size)), types.U32_Size, false)
            asm.MovRegDeref(file, asm.RegA, addr, types.Ptr_Size, false)

        case types.StructType:
            if t.Size() > uint(8) {
                asm.MovRegDeref(file, asm.RegD, addr.Offseted(int64(types.Ptr_Size)), t.Size() - 8, false)
                asm.MovRegDeref(file, asm.RegA, addr, types.Ptr_Size, false)
            } else {
                asm.MovRegDeref(file, asm.RegA, addr, t.Size(), false)
            }

        case types.IntType:
            asm.MovRegDeref(file, asm.RegA, addr, t.Size(), true)

        default:
            if t.Size() > types.Ptr_Size {
                asm.MovRegDeref(file, asm.RegA, addr, types.Ptr_Size, false)
            } else {
                asm.MovRegDeref(file, asm.RegA, addr, t.Size(), false)
            }
        }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v has no fields\n", t)
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }
}

func GenIdent(file *bufio.Writer, e *ast.Ident) {
    if c,ok := e.Obj.(*consts.Const); ok {
        GenConstVal(file, e.GetType(), c.GetVal())
        return
    }

    if v,ok := e.Obj.(vars.Var); ok {
        switch t := v.GetType().(type) {
        case types.StrType:
            asm.MovRegDeref(file, asm.RegA, v.Addr(), types.Ptr_Size, false)
            asm.MovRegDeref(file, asm.RegD, v.Addr().Offseted(int64(types.Ptr_Size)), types.I32_Size, false)

        case types.StructType:
            if t.Size() > uint(8) {
                asm.MovRegDeref(file, asm.RegA, v.Addr(), types.Ptr_Size, false)
                asm.MovRegDeref(file, asm.RegD, v.Addr().Offseted(int64(types.Ptr_Size)), t.Size() - 8, false)
            } else {
                asm.MovRegDeref(file, asm.RegA, v.Addr(), t.Size(), false)
            }

        case types.VecType:
            asm.MovRegDeref(file, asm.RegA, v.Addr(), types.Ptr_Size, false)

        case types.IntType:
            asm.MovRegDeref(file, asm.RegA, v.Addr(), t.Size(), false)

        default:
            asm.MovRegDeref(file, asm.RegA, v.Addr(), t.Size(), false)
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

func GenParen(file *bufio.Writer, e *ast.Paren) {
    GenExpr(file, e.Expr)
}

func UnaryAddrToReg(file *bufio.Writer, e *ast.Unary, reg asm.RegGroup) {
    if e.Operator.Type != token.Mul {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"*\" but got \"%v\"\n", e.Operator)
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    GenExpr(file, e.Operand)
    asm.MovRegReg(file, reg, asm.RegA, types.Ptr_Size)
}

func GenUnary(file *bufio.Writer, e *ast.Unary) {
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
        if t.Size() > 8 {
            asm.DerefRax(file, types.Ptr_Size, false)
        } else {
            asm.DerefRax(file, t.Size(), t.GetKind() == types.Int)
        }
    }
}

func GenBinary(file *bufio.Writer, e *ast.Binary) {
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

func GenFnCall(file *bufio.Writer, e *ast.FnCall) {
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
        asm.SubSp(file, int64(rest))
        bigArgsSize += rest
    }

    // pass big struct args -----------------------------------------
    for i := len(e.F.GetArgs())-1; i >= 0; i-- {
        t := e.F.GetArgs()[i]
        if types.IsBigStruct(t) {
            size := (t.Size() + 7) & ^uint(7)
            bigArgsSize += size

            asm.SubSp(file, int64(size))
            asm.MovRegReg(file, asm.RegC, asm.RegSp, types.Ptr_Size)

            switch t := t.(type) {
            case types.StructType:
                if v := cmpTime.ConstEval(e.Values[i]); v != nil {
                    PassBigStructLit(file, t, *v.(*constVal.StructConst))

                } else if ident,ok := e.Values[i].(*ast.Ident); ok {
                    PassBigStructVar(file, t, ident.Obj.(vars.Var), 0)

                } else {
                    PassBigStructReg(file, asm.RegAsAddr(asm.RegC), e.Values[i])
                }
            case types.VecType:
                PassBigStructReg(file, asm.RegAsAddr(asm.RegC), e.Values[i])

            default:
                fmt.Fprintln(os.Stderr, "[ERROR] (internal) unreachable GenFnCall")
                os.Exit(1)
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
        asm.AddSp(file, int64(bigArgsSize))
    }
}

func GenXCase(file *bufio.Writer, e *ast.XCase, switchCount uint) {
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

func GenXSwitch(file *bufio.Writer, e *ast.XSwitch) {
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



func DerefSetBigStruct(file *bufio.Writer, address addr.Addr, e ast.Expr) {
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

    case *ast.VectorLit:
        if e.Len != nil {
            if c := cmpTime.ConstEval(e.Len); c != nil {
                DerefSetVal(file, address.Offseted(int64(2*types.Ptr_Size)), types.CreateUint(types.U64_Size), c)
            } else {
                DerefSetExpr(file, address.Offseted(int64(2*types.Ptr_Size)), types.CreateUint(types.U64_Size), e.Len)
            }
        } else {
            asm.MovDerefVal(file, address.Offseted(int64(2*types.Ptr_Size)), types.U64_Size, "0")
        }

        if e.Cap == nil {
            asm.MovRegDeref(file, asm.RegA, address.Offseted(int64(2*types.Ptr_Size)), types.U64_Size, false)
            asm.Lea(file, asm.RegA, fmt.Sprintf("%s*%d", asm.GetReg(asm.RegA, types.Ptr_Size), e.Type.BaseType.Size()), types.Ptr_Size)
        } else {
            GenExpr(file, e.Cap)
            if c := cmpTime.ConstEval(e.Cap); c != nil {
                asm.MovDerefVal(file, address.Offseted(int64(types.Ptr_Size)), types.U64_Size, c.GetVal())
                asm.MovRegVal(file, asm.RegA, types.U64_Size, fmt.Sprintf("%s*%d", c.GetVal(), e.Type.BaseType.Size()))
            } else {
                asm.MovDerefReg(file, address.Offseted(int64(types.Ptr_Size)), types.U64_Size, asm.RegA)
                asm.Lea(file, asm.RegA, fmt.Sprintf("%s*%d", asm.GetReg(asm.RegA, types.Ptr_Size), e.Type.BaseType.Size()), types.Ptr_Size)
            }
        }

        file.WriteString("call _alloc_vec\n")
        asm.MovDerefReg(file, address, types.Ptr_Size, asm.RegA)

    case *ast.Indexed:
        IndexedAddrToReg(file, e, asm.RegA)
        DerefSetDeref(file, address, e.GetType(), asm.RegAsAddr(asm.RegA))

    case *ast.Field:
        FieldAddrToReg(file, e, asm.RegA)
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
        asm.Lea(file, asm.RegDi, address.String(), types.Ptr_Size)
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

func XCaseAddrToReg(file *bufio.Writer, e *ast.XCase, switchCount uint, reg asm.RegGroup) {
    cond.CaseStart(file)

    if e.Cond == nil {
        cond.CaseBody(file)
        ExprAddrToReg(file, e.Expr, reg)
        return
    }

    if val,ok := cmpTime.ConstEval(e.Cond).(*constVal.BoolConst); ok {
        if bool(*val) {
            cond.CaseBody(file)
            ExprAddrToReg(file, e.Expr, reg)
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
    ExprAddrToReg(file, e.Expr, reg)
    cond.CaseBodyEnd(file, switchCount)
}

func XSwitchAddrToReg(file *bufio.Writer, e *ast.XSwitch, reg asm.RegGroup) {
    count := cond.StartSwitch()

    for i := 0; i < len(e.Cases)-1; i++ {
        XCaseAddrToReg(file, &e.Cases[i], count, reg)
    }
    cond.InLastCase()
    XCaseAddrToReg(file, &e.Cases[len(e.Cases)-1], count, reg)

    cond.EndSwitch(file)
}

func bigStructXSwitchToStack(file *bufio.Writer, addr addr.Addr, e *ast.XSwitch) {
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

func bigStructXCaseToStack(file *bufio.Writer, addr addr.Addr, e *ast.XCase, switchCount uint) {
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

func GenSyscall(file *bufio.Writer, val ast.Expr) {
    if v := cmpTime.ConstEval(val); v != nil {
        asm.MovRegVal(file, asm.RegA, types.Ptr_Size, v.GetVal())
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] _syscall takes only const")
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }

    file.WriteString("syscall\n")
}

func GenInlineAsm(file *bufio.Writer, val ast.Expr) {
    if str,ok := val.(*ast.StrLit); ok {
        file.WriteString(str.Val.Str[1:len(str.Val.Str)-1] + "\n")
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] _asm takes only a string literal")
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }
}

func GenConstVal(file *bufio.Writer, t types.Type, val constVal.ConstVal) {
    switch c := val.(type) {
    case *constVal.StrConst:
        asm.MovRegVal(file, asm.RegA, types.Ptr_Size, fmt.Sprintf("_str%d", uint64(*c)))
        asm.MovRegVal(file, asm.RegD, types.I32_Size, fmt.Sprintf("%d", str.GetSize(uint64(*c))))

    case *constVal.PtrConst:
        asm.MovRegVal(file, asm.RegA, types.Ptr_Size, PtrConstToAddr(file, *c))

    default:
        asm.MovRegVal(file, asm.RegA, t.Size(), c.GetVal())
    }
}

func PtrConstToAddr(file *bufio.Writer, c constVal.PtrConst) string {
    if c.Local {
        asm.Lea(file, asm.RegA, c.Addr.String(), types.Ptr_Size)
        return asm.GetReg(asm.RegA, types.Ptr_Size)
    } else {
        return c.Addr.String()
    }
}
