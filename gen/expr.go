package gen

import (
    "os"
    "fmt"
    "reflect"
    "strconv"
    "gamma/token"
    "gamma/types"
    "gamma/types/str"
    "gamma/cmpTime"
    "gamma/ast"
    "gamma/ast/identObj"
    "gamma/ast/identObj/consts"
    "gamma/ast/identObj/func"
    "gamma/ast/identObj/struct"
    "gamma/ast/identObj/vars"
    "gamma/gen/asm/x86_64"
    "gamma/gen/asm/x86_64/conditions"
)

func GenExpr(file *os.File, e ast.Expr) {
    switch e := e.(type) {
    case *ast.Lit:
        GenLit(file, e)
    case *ast.FieldLit:
        // TODO
    case *ast.ArrayLit:
        // TODO
    case *ast.StructLit:
        // TODO

    case *ast.Indexed:
        GenIndexed(file, e)
    case *ast.Field:
        GenField(file, e)

    case *ast.Ident:
        GenIdent(file, e)

    case *ast.FnCall:
        GenFnCall(file, e)

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

func IndexedAddrToRcx(file *os.File, e *ast.Indexed) {
    GenExpr(file, e.ArrExpr)

    arrType,_ := e.ArrExpr.GetType().(types.ArrType)
    baseTypeSize := uint64(arrType.Ptr.BaseType.Size())

    if len(arrType.Lens) < len(e.Indices){
        fmt.Fprintf(os.Stderr, "[ERROR] dimension of the array is %d but got %d\n", len(arrType.Lens), len(e.Indices))
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
        file.WriteString(fmt.Sprintf("lea rcx, [rax+%d]\n", idx * baseTypeSize))
    } else {
        asm.MovRegReg(file, asm.RegC, asm.RegA, types.Ptr_Size)
        GenExpr(file, idxExpr)

        asm.Mul(file, fmt.Sprint(baseTypeSize), types.Ptr_Size)
        asm.Add(file, asm.GetReg(asm.RegC, types.Ptr_Size), types.Ptr_Size)

        asm.MovRegReg(file, asm.RegC, asm.RegA, types.Ptr_Size)
    }
}

func GenIndexed(file *os.File, e *ast.Indexed) {
    arrType,_ := e.ArrExpr.GetType().(types.ArrType)

    IndexedAddrToRcx(file, e)

    if arrType.Ptr.BaseType.GetKind() == types.Str {
        asm.MovRegDeref(file, asm.RegGroup(0), asm.GetReg(asm.RegC, types.Ptr_Size), types.Ptr_Size)
        asm.MovRegDeref(file, asm.RegGroup(1), asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, 8), types.I32_Size)
    } else {
        asm.MovRegDeref(file, asm.RegGroup(0), asm.GetReg(asm.RegC, types.Ptr_Size), arrType.Ptr.BaseType.Size())
    }
}

func FieldAddrToRcx(file *os.File, e *ast.Field) {
    switch o := e.Obj.(type) {
    case *ast.Ident:
        file.WriteString(fmt.Sprintf("lea %s, [%s]\n", asm.GetReg(asm.RegC, types.Ptr_Size), o.Obj.Addr(0)))

    case *ast.Field:
        FieldAddrToRcx(file, o)

    default:
        fmt.Fprintln(os.Stderr, "[ERROR] only ident and field expr supported yet")
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }
}

func FieldToOffset(e *ast.Field) int {
    if t,ok := e.Obj.GetType().(types.StructType); ok {
        if s,ok := identObj.Get(t.Name).(*structDec.Struct); ok {
            if i,b := s.GetFieldNum(e.FieldName.Str); b {
                switch o := e.Obj.(type) {
                case *ast.Ident:
                    return t.GetOffset(uint(i))

                case *ast.Field:
                    return t.GetOffset(uint(i)) + FieldToOffset(o)

                default:
                    fmt.Fprintln(os.Stderr, "[ERROR] only ident and field expr supported yet")
                    fmt.Fprintln(os.Stderr, "\t" + e.At())
                    os.Exit(1)
                }
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] struct %s has no %s field\n", t.Name, e.FieldName)
                fmt.Fprintln(os.Stderr, "\t" + e.At())
                os.Exit(1)
            }

        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] struct %s is not declared\n", t.Name)
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }
    }

    fmt.Fprintf(os.Stderr, "[ERROR] expected struct but got %v\n", e.Obj.GetType())
    fmt.Fprintln(os.Stderr, "\t" + e.At())
    os.Exit(1)

    return 0
}

func GenField(file *os.File, e *ast.Field) {
    if t,ok := e.Obj.GetType().(types.StructType); ok {
        if s,ok := identObj.Get(t.Name).(*structDec.Struct); ok {
            if i,b := s.GetFieldNum(e.FieldName.Str); b {
                switch o := e.Obj.(type) {
                case *ast.Ident:
                    switch t := s.GetTypes()[i].(type) {
                    case types.StrType:
                        asm.MovRegDeref(file, asm.RegGroup(0), o.Obj.Addr(i), types.Ptr_Size)
                        asm.MovRegDeref(file, asm.RegGroup(1), fmt.Sprintf("%s+%d", o.Obj.Addr(i), int(types.Ptr_Size)), types.I32_Size)

                    case types.StructType:
                        if t.Size() > uint(8) {
                            asm.MovRegDeref(file, asm.RegGroup(0), o.Obj.Addr(i), types.Ptr_Size)
                            asm.MovRegDeref(file, asm.RegGroup(1), fmt.Sprintf("%s+%d", o.Obj.Addr(i), int(types.Ptr_Size)), t.Types[len(t.Types)-1].Size())
                        } else {
                            asm.MovRegDeref(file, asm.RegGroup(0), o.Obj.Addr(i), t.Size())
                        }

                    default:
                        asm.MovRegDeref(file, asm.RegGroup(0), o.Obj.Addr(i), t.Size())
                    }

                case *ast.Field:
                    FieldAddrToRcx(file, o)

                    offset := FieldToOffset(e)
                    switch t := s.GetTypes()[i].(type) {
                    case types.StrType:
                        asm.MovRegDeref(file, asm.RegGroup(0), asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset), types.Ptr_Size)
                        asm.MovRegDeref(file, asm.RegGroup(1), asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset + int(types.Ptr_Size)), types.I32_Size)

                    case types.StructType:
                        if t.Size() > uint(8) {
                            asm.MovRegDeref(file, asm.RegGroup(0), asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset), types.Ptr_Size)
                            asm.MovRegDeref(file, asm.RegGroup(1), asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset + int(types.Ptr_Size)), t.Types[len(t.Types)-1].Size())
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
                fmt.Fprintf(os.Stderr, "[ERROR] struct %s has no %s field\n", t.Name, e.FieldName)
                fmt.Fprintln(os.Stderr, "\t" + e.At())
                os.Exit(1)
            }
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] expected struct but got %v\n", e.Obj.GetType())
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
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
                asm.MovRegDeref(file, asm.RegGroup(1), v.OffsetedAddr(int(types.Ptr_Size)), t.Types[1].Size())
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
        size := e.Operand.GetType().Size()
        asm.Neg(file, asm.GetReg(asm.RegA, size), size)

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
    size := e.OperandL.GetType().Size()
    if sizeR := e.OperandR.GetType().Size(); sizeR > size {
        size = sizeR
    }

    // compile time evaluation (constEval whole expr)
    if c := cmpTime.ConstEval(e); c.Type != token.Unknown {
        if c.Type == token.Boolean {
            if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
        }

        asm.MovRegVal(file, asm.RegA, size, c.Str)
        return
    }


    // +,-,*,/, <,<=,>,>=,==,!=
    if e.Operator.Type != token.And && e.Operator.Type != token.Or {
        // compile time evaluation (constEval only left expr)
        if c := cmpTime.ConstEval(e.OperandL); c.Type != token.Unknown {
            if c.Type == token.Boolean {
                if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
            }

            asm.MovRegVal(file, asm.RegA, size, c.Str)
        } else {
            GenExpr(file, e.OperandL)

            // compile time evaluation (constEval only right expr)
            if c := cmpTime.ConstEval(e.OperandR); c.Type != token.Unknown {
                if c.Type == token.Boolean {
                    if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
                }

                asm.BinaryOp(file, e.Operator.Type, c.Str, size)
                return
            }
        }

        if ident,ok := e.OperandR.(*ast.Ident); ok {
            if v,ok := ident.Obj.(vars.Var); ok {
                if t := v.GetType(); t.Size() < size {
                    asm.MovRegDeref(file, asm.RegB, v.Addr(0), t.Size())
                    asm.BinaryOpReg(file, e.Operator.Type, asm.RegB, size)
                } else {
                    asm.BinaryOp(file, e.Operator.Type, fmt.Sprintf("%s [%s]", asm.GetWord(t.Size()), v.Addr(0)), size)
                }
            }
        } else {
            asm.PushReg(file, asm.RegA)

            GenExpr(file, e.OperandR)
            asm.MovRegReg(file, asm.RegB, asm.RegA, size)

            asm.PopReg(file, asm.RegA)
            asm.BinaryOpReg(file, e.Operator.Type, asm.RegB, size)
        }

    // &&, ||
    } else {
        // compile time evaluation
        if c := cmpTime.ConstEval(e.OperandL); c.Type != token.Unknown {
            if e.Operator.Type == token.And && c.Str == "false" {
                asm.MovRegVal(file, asm.RegA, size, "0")
                return
            }
            if e.Operator.Type == token.Or && c.Str == "true" {
                asm.MovRegVal(file, asm.RegA, size, "1")
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
            fn.PassValStack(file, v, e.F.GetArgs()[i])

        } else if ident,ok := e.Values[i].(*ast.Ident); ok {
            fn.PassVarStack(file, ident.Obj.(vars.Var))

        } else {
            GenExpr(file, e.Values[i])
            fn.PassRegStack(file, e.F.GetArgs()[i])
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
                    fn.PassBigStructLit(file, t, v, 0)

                } else if ident,ok := e.Values[i].(*ast.Ident); ok {
                    file.WriteString("mov rcx, rsp\n")
                    fn.PassBigStructVar(file, t, ident.Obj.(vars.Var), 0)

                } else {
                    if _,ok := e.Values[i].(*ast.FnCall); ok {
                        file.WriteString(fmt.Sprintf("lea rdi, [rbp-%d]\n", bigArgsSize))
                    }

                    GenExpr(file, e.Values[i])

                    file.WriteString("mov rcx, rsp\n")
                    fn.PassBigStructReg(file, t)
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
            fn.PassVal(file, regIdx, v, t)

        } else if ident,ok := e.Values[i].(*ast.Ident); ok {
            fn.PassVar(file, regIdx, ident.Obj.(vars.Var))

        } else {
            GenExpr(file, e.Values[i])
            fn.PassReg(file, regIdx, t)
        }
    }

    e.F.Call(file)

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
