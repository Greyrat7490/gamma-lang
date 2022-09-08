package gen

import (
    "os"
    "fmt"
    "reflect"
    "strconv"
    "gamma/token"
    "gamma/types"
    "gamma/cmpTime"
    "gamma/types/str"
    "gamma/types/struct"
    "gamma/ast"
    "gamma/ast/identObj"
    "gamma/ast/identObj/func"
    "gamma/ast/identObj/vars"
    "gamma/ast/identObj/struct"
    "gamma/ast/identObj/consts"
    "gamma/gen/asm/x86_64"
    "gamma/gen/asm/x86_64/conditions"
)

func GenExpr(file *os.File, e ast.Expr) {
    switch e := e.(type) {
    case *ast.Lit:
        GenLit(file, e)
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
        if e.Ident.Name == "_syscall" {
            GenSyscall(file, e.Values[0])
        } else {
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

    case *ast.BadExpr:
        fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
        os.Exit(1)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] GenExpr for %v is not implemente yet\n", reflect.TypeOf(e))
        os.Exit(1)
    }
}

func GenLit(file *os.File, e *ast.Lit) {
    switch e.Val.Type {
    case token.Str:
        strIdx := str.Add(e.Val)

        asm.MovRegVal(file, asm.RegGroup(0), types.Ptr_Size, fmt.Sprintf("_str%d", strIdx))
        asm.MovRegVal(file, asm.RegGroup(1), types.I32_Size, fmt.Sprintf("%d", str.GetSize(strIdx)))

    case token.Boolean:
        if e.Val.Str == "true" {
            asm.MovRegVal(file, asm.RegGroup(0), e.Type.Size(), "1")
        } else {
            asm.MovRegVal(file, asm.RegGroup(0), e.Type.Size(), "0")
        }

    default:
        asm.MovRegVal(file, asm.RegGroup(0), e.Type.Size(), e.Val.Str)
    }
}

func GenStructLit(file *os.File, e *ast.StructLit) {
    values := structLit.GetValues(uint64(e.Idx))

    vs := PackValues(e.StructType.Types, values)
    asm.MovRegVal(file, asm.RegGroup(0), types.Ptr_Size, vs[0])
    if len(vs) == 2 {
        asm.MovRegVal(file, asm.RegGroup(1), types.Ptr_Size, vs[1])
    }
}

func GenArrayLit(file *os.File, e *ast.ArrayLit) {
    asm.MovRegVal(file, asm.RegGroup(0), types.Ptr_Size, fmt.Sprintf("_arr%d", e.Idx))
}

func IndexedAddrToReg(file *os.File, e *ast.Indexed, r asm.RegGroup) {
    GenExpr(file, e.ArrExpr)

    baseTypeSize := uint64(e.ArrType.Ptr.BaseType.Size())

    if len(e.ArrType.Lens) < len(e.Indices) {
        fmt.Fprintf(os.Stderr, "[ERROR] dimension of the array is %d but got %d\n", len(e.ArrType.Lens), len(e.Indices))
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    idxExpr := e.Flatten()
    val := cmpTime.ConstEval(idxExpr)
    if val.Type != token.Unknown {
        if val.Type != token.Number {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a Number but got %v\n", val)
            fmt.Fprintln(os.Stderr, "\t" + idxExpr.At())
            os.Exit(1)
        }

        idx,_ := strconv.ParseUint(val.Str, 10, 64)
        file.WriteString(fmt.Sprintf("lea %s, [rax+%d]\n", asm.GetReg(r, types.Ptr_Size), idx * baseTypeSize))
    } else {
        asm.MovRegReg(file, asm.RegC, asm.RegA, types.Ptr_Size)
        GenExpr(file, idxExpr)

        asm.Mul(file, fmt.Sprint(baseTypeSize), types.Ptr_Size, false)
        asm.Add(file, asm.GetReg(asm.RegC, types.Ptr_Size), types.Ptr_Size)

        if r != asm.RegA {
            asm.MovRegReg(file, r, asm.RegA, types.Ptr_Size)
        }
    }
}

func GenIndexed(file *os.File, e *ast.Indexed) {
    IndexedAddrToReg(file, e, asm.RegC)

    switch t := e.ArrType.Ptr.BaseType.(type) {
    case types.StrType:
        asm.MovRegDeref(file, asm.RegGroup(0), asm.GetReg(asm.RegC, types.Ptr_Size), types.Ptr_Size)
        asm.MovRegDeref(file, asm.RegGroup(1), asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, 8), types.I32_Size)

    case types.StructType:
        if t.Size() > uint(8) {
            asm.MovRegDeref(file, asm.RegGroup(0), asm.GetReg(asm.RegC, types.Ptr_Size), types.Ptr_Size)
            asm.MovRegDeref(
                file,
                asm.RegGroup(1),
                asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, int(t.Size() - 8)),
                t.Size() - 8,
            )
        } else {
            asm.MovRegDeref(file, asm.RegGroup(0), asm.GetReg(asm.RegC, types.Ptr_Size), t.Size())
        }

    default:
        asm.MovRegDeref(
            file,
            asm.RegGroup(0),
            asm.GetReg(asm.RegC, types.Ptr_Size),
            e.ArrType.Ptr.BaseType.Size(),
        )
    }
}

func FieldAddrToReg(file *os.File, e *ast.Field, r asm.RegGroup) {
    switch o := e.Obj.(type) {
    case *ast.Ident:
        file.WriteString(fmt.Sprintf("lea %s, [%s]\n", asm.GetReg(r, types.Ptr_Size), o.Obj.Addr(0)))

    case *ast.Field:
        FieldAddrToReg(file, o, r)

    default:
        fmt.Fprintln(os.Stderr, "[ERROR] only ident and field expr supported yet")
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }
}

func FieldToOffset(e *ast.Field) int {
    if s,ok := identObj.Get(e.StructType.Name).(*structDec.Struct); ok {
        if i,b := s.GetFieldNum(e.FieldName.Str); b {
            switch o := e.Obj.(type) {
            case *ast.Ident:
                return e.StructType.GetOffset(uint(i))

            case *ast.Field:
                return e.StructType.GetOffset(uint(i)) + FieldToOffset(o)

            default:
                fmt.Fprintln(os.Stderr, "[ERROR] only ident and field expr supported yet")
                fmt.Fprintln(os.Stderr, "\t" + e.At())
                os.Exit(1)
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] struct %s has no %s field\n", e.StructType.Name, e.FieldName)
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }

    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] struct %s is not declared\n", e.StructType.Name)
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    return 0
}

func GenField(file *os.File, e *ast.Field) {
    if s,ok := identObj.Get(e.StructType.Name).(*structDec.Struct); ok {
        if i,b := s.GetFieldNum(e.FieldName.Str); b {
            switch o := e.Obj.(type) {
            case *ast.Ident:
                switch t := s.GetTypes()[i].(type) {
                case types.StrType:
                    asm.MovRegDeref(file, asm.RegGroup(0), o.Obj.Addr(i), types.Ptr_Size)
                    asm.MovRegDeref(
                        file,
                        asm.RegGroup(1),
                        fmt.Sprintf("%s+%d", o.Obj.Addr(i), int(types.Ptr_Size)),
                        types.I32_Size,
                    )

                case types.StructType:
                    if t.Size() > uint(8) {
                        asm.MovRegDeref(file, asm.RegGroup(0), o.Obj.Addr(i), types.Ptr_Size)
                        asm.MovRegDeref(
                            file,
                            asm.RegGroup(1),
                            fmt.Sprintf("%s+%d", o.Obj.Addr(i), int(types.Ptr_Size)),
                            t.Size() - 8,
                        )
                    } else {
                        asm.MovRegDeref(file, asm.RegGroup(0), o.Obj.Addr(i), t.Size())
                    }

                default:
                    asm.MovRegDeref(file, asm.RegGroup(0), o.Obj.Addr(i), t.Size())
                }

            case *ast.Field:
                FieldAddrToReg(file, o, asm.RegC)

                offset := FieldToOffset(e)
                switch t := s.GetTypes()[i].(type) {
                case types.StrType:
                    asm.MovRegDeref(
                        file,
                        asm.RegGroup(0),
                        asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset),
                        types.Ptr_Size,
                    )
                    asm.MovRegDeref(
                        file,
                        asm.RegGroup(1),
                        asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset + int(types.Ptr_Size)),
                        types.I32_Size,
                    )

                case types.StructType:
                    if t.Size() > uint(8) {
                        asm.MovRegDeref(
                            file,
                            asm.RegGroup(0),
                            asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset),
                            types.Ptr_Size,
                        )
                        asm.MovRegDeref(
                            file,
                            asm.RegGroup(1),
                            asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset + int(types.Ptr_Size)),
                            t.Size() - 8,
                        )
                    } else {
                        asm.MovRegDeref(file, asm.RegGroup(0), asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset), t.Size())
                    }

                default:
                    asm.MovRegDeref(file, asm.RegGroup(0), asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset), t.Size())
                }

            default:
                fmt.Fprintln(os.Stderr, "[ERROR] only ident and field expr supported yet")
                fmt.Fprintln(os.Stderr, "\t" + e.At())
                os.Exit(1)
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] struct %s has no %s field\n", e.StructType.Name, e.FieldName)
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }
    }
}

func GenIdent(file *os.File, e *ast.Ident) {
    if c,ok := e.Obj.(*consts.Const); ok {
        l := ast.Lit{ Val: c.GetVal(), Type: c.GetType() }
        GenLit(file, &l)
        return
    }

    if v,ok := e.Obj.(vars.Var); ok {
        switch t := v.GetType().(type) {
        case types.StrType:
            asm.MovRegDeref(file, asm.RegGroup(0), v.Addr(0), types.Ptr_Size)
            asm.MovRegDeref(file, asm.RegGroup(1), v.Addr(1), types.I32_Size)

        case types.StructType:
            if t.Size() > uint(8) {
                asm.MovRegDeref(file, asm.RegGroup(0), v.Addr(0), types.Ptr_Size)
                asm.MovRegDeref(file, asm.RegGroup(1), v.OffsetedAddr(int(types.Ptr_Size)), t.Size() - 8)
            } else {
                asm.MovRegDeref(file, asm.RegGroup(0), v.Addr(0), t.Size())
            }

        default:
            asm.MovRegDeref(file, asm.RegGroup(0), v.Addr(0), t.Size())
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
    if c := cmpTime.ConstEval(e); c.Type != token.Unknown {
        asm.MovRegVal(file, asm.RegA, e.Operand.GetType().Size(), c.Str)
        return
    }

    GenExpr(file, e.Operand)

    switch e.Operator.Type {
    case token.Minus:
        asm.Neg(file, e.Operand.GetType().Size())

    case token.BitNot:
        asm.Not(file, e.Operand.GetType().Size())

    case token.Mul:
        if _,ok := e.Operand.(*ast.Ident); !ok {
            if _,ok := e.Operand.(*ast.Paren); !ok {
                fmt.Fprintln(os.Stderr, "[ERROR] expected a variable or parentheses after \"*\"")
                fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
                os.Exit(1)
            }
        }

        asm.DerefRax(file, e.GetType().Size())
    }
}

func GenBinary(file *os.File, e *ast.Binary) {
    // compile time evaluation (constEval whole expr)
    if c := cmpTime.ConstEval(e); c.Type != token.Unknown {
        if c.Type == token.Boolean {
            if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
        }

        asm.MovRegVal(file, asm.RegA, e.GetType().Size(), c.Str)
        return
    }


    // +,-,*,/, <,<=,>,>=,==,!=
    if e.Operator.Type != token.And && e.Operator.Type != token.Or {
        // compile time evaluation (constEval only left expr)
        if c := cmpTime.ConstEval(e.OperandL); c.Type != token.Unknown {
            if c.Type == token.Boolean {
                if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
            }

            asm.MovRegVal(file, asm.RegA, e.OperandL.GetType().Size(), c.Str)
        } else {
            GenExpr(file, e.OperandL)

            // compile time evaluation (constEval only right expr)
            if c := cmpTime.ConstEval(e.OperandR); c.Type != token.Unknown {
                if c.Type == token.Boolean {
                    if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
                }

                asm.BinaryOp(file, e.Operator.Type, c.Str, e.OperandR.GetType().Size(), e.GetType().GetKind() == types.Int)
                return
            }
        }

        if ident,ok := e.OperandR.(*ast.Ident); ok {
            if v,ok := ident.Obj.(vars.Var); ok {
                t := v.GetType()

                asm.BinaryOp(file,
                    e.Operator.Type,
                    fmt.Sprintf("%s [%s]", asm.GetWord(t.Size()), v.Addr(0)),
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
        if c := cmpTime.ConstEval(e.OperandL); c.Type != token.Unknown {
            if e.Operator.Type == token.And && c.Str == "false" {
                asm.MovRegVal(file, asm.RegA, types.Bool_Size, "0")
                return
            }
            if e.Operator.Type == token.Or && c.Str == "true" {
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
            s := (t.Size() + 7) & ^uint(7)
            rest += s

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
        if v := cmpTime.ConstEval(e.Values[i]); v.Type != token.Unknown {
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

                if v := cmpTime.ConstEval(e.Values[i]); v.Type != token.Unknown {
                    file.WriteString("mov rcx, rsp\n")
                    PassBigStructLit(file, t, v, 0)

                } else if ident,ok := e.Values[i].(*ast.Ident); ok {
                    file.WriteString("mov rcx, rsp\n")
                    PassBigStructVar(file, t, ident.Obj.(vars.Var), 0)

                } else {
                    file.WriteString("mov rcx, rsp\n")
                    PassBigStructReg(file, "rcx", e.Values[i])
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

        if v := cmpTime.ConstEval(e.Values[i]); v.Type != token.Unknown {
            PassVal(file, regIdx, v, t)

        } else if ident,ok := e.Values[i].(*ast.Ident); ok {
            PassVar(file, regIdx, t, ident.Obj.(vars.Var))

        } else {
            GenExpr(file, e.Values[i])
            PassReg(file, regIdx, t, e.Values[i].GetType().Size())
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

    if val := cmpTime.ConstEval(e.Cond); val.Type != token.Unknown {
        if val.Str == "true" {
            cond.CaseBody(file)
            GenExpr(file, e.Expr)
            cond.CaseBodyEnd(file, switchCount)
        }

        return
    }

    if i,ok := e.Cond.(*ast.Ident); ok {
        cond.CaseVar(file, i.Obj.Addr(0))
    } else {
        GenExpr(file, e.Cond)
        cond.CaseExpr(file)
    }

    cond.CaseBody(file)
    GenExpr(file, e.Expr)
    cond.CaseBodyEnd(file, switchCount)
}

func GenXSwitch(file *os.File, e *ast.XSwitch) {
    if c := cmpTime.ConstEval(e); c.Type != token.Unknown {
        asm.MovRegVal(file, asm.RegA, types.TypeOfVal(c.Str).Size(), c.Str)
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



func DerefSetBigStruct(file *os.File, addr string, e ast.Expr) {
    if !types.IsBigStruct(e.GetType()) {
        fmt.Fprintln(os.Stderr, "[ERROR] expected expr to be a big struct")
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    switch e := e.(type) {
    case *ast.StructLit:
        DerefSetVal(file, addr, e.StructType, token.Token{ Str: fmt.Sprint(e.Idx), Type: token.Number })

    case *ast.Indexed:
        IndexedAddrToReg(file, e, asm.RegA)
        DerefSetDeref(file, addr, e.GetType(), "rax")

    case *ast.Field:
        FieldAddrToReg(file, e, asm.RegA)
        offset := FieldToOffset(e)
        file.WriteString(fmt.Sprintf("lea rax, [rax+%d]\n", offset))
        DerefSetDeref(file, addr, e.GetType(), "rax")

    case *ast.Ident:
        if v,ok := e.Obj.(vars.Var); ok {
            DerefSetVar(file, addr, v)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected identifier %s to be a variable but got %v\n", e.Name, reflect.TypeOf(e.Obj))
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }

    case *ast.Unary:
        GenExpr(file, e.Operand)
        DerefSetDeref(file, addr, e.GetType(), "rax")

    case *ast.FnCall:
        file.WriteString(fmt.Sprintf("lea rdi, [%s]\n", addr))
        GenExpr(file, e)

    case *ast.Paren:
        DerefSetBigStruct(file, addr, e.Expr)

    case *ast.XSwitch:
        bigStructXSwitchToStack(file, addr, e)

    case *ast.BadExpr:
        fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
        os.Exit(1)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] DerefSetBigStruct for %v is not implemente yet\n", reflect.TypeOf(e))
        os.Exit(1)
    }
}

func bigStructXSwitchToStack(file *os.File, addr string, e *ast.XSwitch) {
    if c := cmpTime.ConstEval(e); c.Type != token.Unknown {
        asm.MovRegVal(file, asm.RegA, types.TypeOfVal(c.Str).Size(), c.Str)
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

func bigStructXCaseToStack(file *os.File, addr string, e *ast.XCase, switchCount uint) {
    cond.CaseStart(file)

    if e.Cond == nil {
        cond.CaseBody(file)
        DerefSetBigStruct(file, addr, e.Expr)
        return
    }

    if val := cmpTime.ConstEval(e.Cond); val.Type != token.Unknown {
        if val.Str == "true" {
            cond.CaseBody(file)
            DerefSetBigStruct(file, addr, e.Expr)
            cond.CaseBodyEnd(file, switchCount)
        }

        return
    }

    if i,ok := e.Cond.(*ast.Ident); ok {
        cond.CaseVar(file, i.Obj.Addr(0))
    } else {
        GenExpr(file, e.Cond)
        cond.CaseExpr(file)
    }

    cond.CaseBody(file)
    DerefSetBigStruct(file, addr, e.Expr)
    cond.CaseBodyEnd(file, switchCount)
}

func GenSyscall(file *os.File, val ast.Expr) {
    if v := cmpTime.ConstEval(val); v.Type != token.Unknown {
        asm.MovRegVal(file, asm.RegA, types.Ptr_Size, v.Str)
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] _syscall takes only const")
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }

    file.WriteString("syscall\n")
}
