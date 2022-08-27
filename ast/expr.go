package ast

import (
    "os"
    "fmt"
    "strconv"
    "gamma/token"
    "gamma/types"
    "gamma/types/str"
    "gamma/asm/x86_64"
    "gamma/asm/x86_64/conditions"
    "gamma/ast/identObj"
    "gamma/ast/identObj/func"
    "gamma/ast/identObj/vars"
    "gamma/ast/identObj/consts"
    "gamma/ast/identObj/struct"
)

type Expr interface {
    Node
    Compile(file *os.File)
    // Assign(file *os.File)
    GetType() types.Type
    typeCheck()
    ConstEval() token.Token
}

type BadExpr struct{}

type FnCall struct {
    F *fn.Func
    Ident Ident
    ParenLPos token.Pos
    Values []Expr
    ParenRPos token.Pos
}

type Lit struct {
    Val token.Token
    Type types.Type
}

type ArrayLit struct {
    Idx int
    Pos token.Pos
    Type types.ArrType
    BraceLPos token.Pos
    Values []Expr
    BraceRPos token.Pos
}

type StructLit struct {
    Idx int
    Pos token.Pos
    StructType types.StructType
    BraceLPos token.Pos
    Fields []FieldLit
    BraceRPos token.Pos
}

type FieldLit struct {
    Name token.Token
    Pos token.Pos
    Value Expr
}

type Indexed struct {
    ArrExpr Expr
    BrackLPos token.Pos
    Indices []Expr
    BrackRPos token.Pos
}

type Field struct {
    Pos token.Pos
    Obj Expr
    DotPos token.Pos
    FieldName token.Token
}

type Ident struct {
    Name string
    Pos token.Pos
    Obj identObj.IdentObj
}

type Unary struct {
    Operator token.Token
    Operand Expr
}

type Binary struct {
    Pos token.Pos
    OperandL Expr
    Operator token.Token
    OperandR Expr
}

type Paren struct {
    ParenLPos token.Pos
    Expr Expr
    ParenRPos token.Pos
}

type XSwitch struct {
    Pos token.Pos
    BraceLPos token.Pos
    Cases []XCase
    BraceRPos token.Pos
}

type XCase struct {
    Cond Expr
    ColonPos token.Pos
    Expr Expr
}


func (e *Lit) Compile(file *os.File) {
    switch e.Val.Type {
    case token.Str:
        strIdx := str.Add(e.Val)

        asm.MovRegVal(file, asm.RegGroup(0), types.Ptr_Size, fmt.Sprintf("_str%d", strIdx))
        asm.MovRegVal(file, asm.RegGroup(1), types.I32_Size, fmt.Sprintf("%d", str.GetSize(strIdx)))

    case token.Boolean:
        if e.Val.Str == "true" { e.Val.Str = "1" } else { e.Val.Str = "0" }
        fallthrough

    default:
        asm.MovRegVal(file, asm.RegGroup(0), e.Type.Size(), e.Val.Str)
    }
}

func (e *StructLit) Compile(file *os.File) {}
func (e *FieldLit) Compile(file *os.File) {}
func (e *ArrayLit) Compile(file *os.File) {}

func (e *Indexed) flatten() Expr {
    // for dim = 1 (no need to flatten)
    if len(e.Indices) == 1 {
        return e.Indices[0]
    }

    res := Binary{ Operator: token.Token{ Str: "+", Type: token.Plus }, OperandR: e.Indices[len(e.Indices)-1] }
    expr := &res.OperandL

    // for dim > 3
    arrType,_ := e.ArrExpr.GetType().(types.ArrType)
    var innerLen uint64 = arrType.Lens[0]
    for i := 1; i < len(e.Indices)-1; i++ {
        b := Binary{
            Operator: token.Token{ Str: "+", Type: token.Plus }, OperandR: &Binary{
                Operator: token.Token{ Str: "*", Type: token.Mul },
                OperandL: e.Indices[len(e.Indices)-1-i],
                OperandR: &Lit{ Val: token.Token{ Str: fmt.Sprint(innerLen), Type: token.Number }, Type: types.I32Type{} },
        } }

        *expr = &b
        expr = &b.OperandL

        innerLen *= arrType.Lens[i]
    }

    *expr = &Binary{
        Operator: token.Token{ Str: "*", Type: token.Mul },
        OperandL: e.Indices[0],
        OperandR: &Lit{ Val: token.Token{ Str: fmt.Sprint(innerLen), Type: token.Number }, Type: types.I32Type{} },
    }

    return &res
}

func (e *Indexed) AddrToRcx(file *os.File) {
    e.typeCheck()

    e.ArrExpr.Compile(file)

    arrType,_ := e.ArrExpr.GetType().(types.ArrType)
    baseTypeSize := uint64(arrType.Ptr.BaseType.Size())

    if len(arrType.Lens) < len(e.Indices){
        fmt.Fprintf(os.Stderr, "[ERROR] dimension of the array is %d but got %d\n", len(arrType.Lens), len(e.Indices))
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    idxExpr := e.flatten()
    val := idxExpr.ConstEval()
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
        idxExpr.Compile(file)

        asm.Mul(file, fmt.Sprint(baseTypeSize), types.Ptr_Size)
        asm.Add(file, asm.GetReg(asm.RegC, types.Ptr_Size), types.Ptr_Size)

        asm.MovRegReg(file, asm.RegC, asm.RegA, types.Ptr_Size)
    }
}

func (e *Indexed) Compile(file *os.File) {
    e.typeCheck()

    arrType,_ := e.ArrExpr.GetType().(types.ArrType)

    e.AddrToRcx(file)

    if arrType.Ptr.BaseType.GetKind() == types.Str {
        asm.MovRegDeref(file, asm.RegGroup(0), asm.GetReg(asm.RegC, types.Ptr_Size), types.Ptr_Size)
        asm.MovRegDeref(file, asm.RegGroup(1), asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, 8), types.I32_Size)
    } else {
        asm.MovRegDeref(file, asm.RegGroup(0), asm.GetReg(asm.RegC, types.Ptr_Size), arrType.Ptr.BaseType.Size())
    }
}

func (e *Field) AddrToRcx(file *os.File) {
    switch o := e.Obj.(type) {
    case *Ident:
        file.WriteString(fmt.Sprintf("lea %s, [%s]\n", asm.GetReg(asm.RegC, types.Ptr_Size), o.Obj.Addr(0)))

    case *Field:
        o.AddrToRcx(file)

    default:
        fmt.Fprintln(os.Stderr, "[ERROR] only ident and field expr supported yet")
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }
}

func (e *Field) Offset() int {
    if t,ok := e.Obj.GetType().(types.StructType); ok {
        if s,ok := identObj.Get(t.Name).(*structDec.Struct); ok {
            if i,b := s.GetFieldNum(e.FieldName.Str); b {
                switch o := e.Obj.(type) {
                case *Ident:
                    return t.GetOffset(uint(i))

                case *Field:
                    return t.GetOffset(uint(i)) + o.Offset()

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

func (e *Field) Compile(file *os.File) {
    if t,ok := e.Obj.GetType().(types.StructType); ok {
        if s,ok := identObj.Get(t.Name).(*structDec.Struct); ok {
            if i,b := s.GetFieldNum(e.FieldName.Str); b {
                switch o := e.Obj.(type) {
                case *Ident:
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

                case *Field:
                    o.AddrToRcx(file)

                    offset := e.Offset()
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

func (e *Ident) Compile(file *os.File) {
    if c,ok := e.Obj.(*consts.Const); ok {
        l := Lit{ Val: c.GetVal(), Type: c.GetType() }
        l.Compile(file)
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
func (e *Paren) Compile(file *os.File) { e.Expr.Compile(file) }
func (e *Unary) Compile(file *os.File) {
    e.typeCheck()

    // compile time evaluation
    if c := e.ConstEval(); c.Type != token.Unknown {
        asm.MovRegVal(file, asm.RegA, e.Operand.GetType().Size(), c.Str)
        return
    }

    e.Operand.Compile(file)

    switch e.Operator.Type {
    case token.Minus:
        size := e.Operand.GetType().Size()
        asm.Neg(file, asm.GetReg(asm.RegA, size), size)

    case token.Mul:
        if _,ok := e.Operand.(*Ident); !ok {
            if _,ok := e.Operand.(*Paren); !ok {
                fmt.Fprintln(os.Stderr, "[ERROR] expected a variable or parentheses after \"*\"")
                fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
                os.Exit(1)
            }
        }

        asm.DerefRax(file, e.GetType().Size())
    }
}
func (e *Binary) Compile(file *os.File) {
    e.typeCheck()

    size := e.OperandL.GetType().Size()
    if sizeR := e.OperandR.GetType().Size(); sizeR > size {
        size = sizeR
    }

    // compile time evaluation (constEval whole expr)
    if c := e.ConstEval(); c.Type != token.Unknown {
        if c.Type == token.Boolean {
            if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
        }

        asm.MovRegVal(file, asm.RegA, size, c.Str)
        return
    }


    // +,-,*,/, <,<=,>,>=,==,!=
    if e.Operator.Type != token.And && e.Operator.Type != token.Or {
        // compile time evaluation (constEval only left expr)
        if c := e.OperandL.ConstEval(); c.Type != token.Unknown {
            if c.Type == token.Boolean {
                if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
            }

            asm.MovRegVal(file, asm.RegA, size, c.Str)
        } else {
            e.OperandL.Compile(file)

            // compile time evaluation (constEval only right expr)
            if c := e.OperandR.ConstEval(); c.Type != token.Unknown {
                if c.Type == token.Boolean {
                    if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
                }

                asm.BinaryOp(file, e.Operator.Type, c.Str, size)
                return
            }
        }

        if ident,ok := e.OperandR.(*Ident); ok {
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

            e.OperandR.Compile(file)
            asm.MovRegReg(file, asm.RegB, asm.RegA, size)

            asm.PopReg(file, asm.RegA)
            asm.BinaryOpReg(file, e.Operator.Type, asm.RegB, size)
        }

    // &&, ||
    } else {
        // compile time evaluation
        if c := e.OperandL.ConstEval(); c.Type != token.Unknown {
            if e.Operator.Type == token.And && c.Str == "false" {
                asm.MovRegVal(file, asm.RegA, size, "0")
                return
            }
            if e.Operator.Type == token.Or && c.Str == "true" {
                asm.MovRegVal(file, asm.RegA, size, "1")
                return
            }

            e.OperandR.Compile(file)
        } else {
            e.OperandL.Compile(file)

            count := cond.LogicalOp(file, e.Operator)
            e.OperandR.Compile(file)
            cond.LogicalOpEnd(file, count)
        }
    }
}

func (e *FnCall) Compile(file *os.File) {
    e.typeCheck()

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
        if v := e.Values[i].ConstEval(); v.Type != token.Unknown {
            fn.PassValStack(file, v, e.F.GetArgs()[i])

        } else if ident,ok := e.Values[i].(*Ident); ok {
            fn.PassVarStack(file, ident.Obj.(vars.Var))

        } else {
            e.Values[i].Compile(file)
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
                file.WriteString("mov rcx, rsp\n")

                if v := e.Values[i].ConstEval(); v.Type != token.Unknown {
                    fn.PassBigStructLit(file, t, v, 0)

                } else if ident,ok := e.Values[i].(*Ident); ok {
                    fn.PassBigStructVar(file, t, ident.Obj.(vars.Var), 0)

                } else {
                    if _,ok := e.Values[i].(*FnCall); ok {
                        file.WriteString(fmt.Sprintf("lea rdi, [rbp-%d]\n", bigArgsSize))
                    }
                    e.Values[i].Compile(file)
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

        if v := e.Values[i].ConstEval(); v.Type != token.Unknown {
            fn.PassVal(file, regIdx, v, t)

        } else if ident,ok := e.Values[i].(*Ident); ok {
            fn.PassVar(file, regIdx, ident.Obj.(vars.Var))

        } else {
            e.Values[i].Compile(file)
            fn.PassReg(file, regIdx, t)
        }
    }

    e.F.Call(file)

    // clear stack -------------------------------------------------
    if bigArgsSize > 0 {
        file.WriteString(fmt.Sprintf("add rsp, %d\n", bigArgsSize))
    }
}

func (e *XCase) Compile(file *os.File, switchCount uint) {
    cond.CaseStart(file)

    if e.Cond == nil {
        cond.CaseBody(file)
        e.Expr.Compile(file)
        return
    }

    // compile time evaluation
    if val := e.Cond.ConstEval(); val.Type != token.Unknown {
        if val.Str == "true" {
            cond.CaseBody(file)
            e.Expr.Compile(file)
            cond.CaseBodyEnd(file, switchCount)
        }

        return
    }

    if i,ok := e.Cond.(*Ident); ok {
        cond.CaseVar(file, i.Obj.Addr(0))
    } else {
        e.Cond.Compile(file)
        cond.CaseExpr(file)
    }

    cond.CaseBody(file)
    e.Expr.Compile(file)
    cond.CaseBodyEnd(file, switchCount)
}

func (e *XSwitch) Compile(file *os.File) {
    e.typeCheck()

    // compile time evaluation
    if c := e.ConstEval(); c.Type != token.Unknown {
        asm.MovRegVal(file, asm.RegA, types.TypeOfVal(c.Str).Size(), c.Str)
        return
    }

    count := cond.StartSwitch()

    for i := 0; i < len(e.Cases)-1; i++ {
        e.Cases[i].Compile(file, count)
    }
    cond.InLastCase()
    e.Cases[len(e.Cases)-1].Compile(file, count)

    cond.EndSwitch(file)
}

func (e *BadExpr) Compile(file *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
}


func (e *BadExpr)   At() string { return "" }
func (e *FnCall)    At() string { return e.Ident.At() }
func (e *Lit)       At() string { return e.Val.At() }
func (e *FieldLit)  At() string { return e.Name.At() }
func (e *StructLit) At() string { return e.Pos.At() }
func (e *ArrayLit)  At() string { return e.Pos.At() }
func (e *Indexed)   At() string { return e.ArrExpr.At() }
func (e *Field)     At() string { return e.Pos.At() }
func (e *Ident)     At() string { return e.Pos.At() }
func (e *Unary)     At() string { return e.Operator.At() }
func (e *Binary)    At() string { return e.OperandL.At() }    // TODO: At() of Operand with higher precedence
func (e *Paren)     At() string { return e.ParenLPos.At() }
func (e *XSwitch)   At() string { return e.Pos.At() }
func (e *XCase)     At() string { return e.ColonPos.At() }

func (e *BadExpr)   End() string { return "" }
func (e *FnCall)    End() string { return e.ParenRPos.At() }
func (e *Lit)       End() string { return e.Val.At() }
func (e *FieldLit)  End() string { return e.Value.End() }
func (e *StructLit) End() string { return e.BraceRPos.At() }
func (e *ArrayLit)  End() string { return e.BraceRPos.At() }
func (e *Indexed)   End() string { return e.BrackRPos.At() }
func (e *Field)     End() string { return e.FieldName.At() }
func (e *Ident)     End() string { return e.Pos.At() }
func (e *Unary)     End() string { return e.Operand.At() }
func (e *Binary)    End() string { return e.OperandR.At() }
func (e *Paren)     End() string { return e.ParenRPos.At() }
func (e *XSwitch)   End() string { return e.BraceRPos.At() }
func (e *XCase)     End() string { return e.Expr.At() }
