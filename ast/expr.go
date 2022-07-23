package ast

import (
    "os"
    "fmt"
    "gorec/func"
    "gorec/vars"
    "gorec/token"
    "gorec/types"
    "gorec/types/str"
    "gorec/conditions"
    "gorec/asm/x86_64"
)

type Expr interface {
    Node
    Compile(file *os.File)
    GetType() types.Type
    typeCheck()
    constEval() token.Token
}

type BadExpr struct{}

type FnCall struct {
    Name token.Token
    ParenLPos token.Pos
    Values []Expr
    ParenRPos token.Pos
}

type Lit struct {
    Val token.Token
    Type types.Type
}

type Ident struct {
    Ident token.Token
    // TODO: IdentObj
    V vars.Var
    C *vars.Const
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

        asm.MovRegVal(file, asm.RegA, types.Ptr_Size, fmt.Sprintf("_str%d", strIdx))
        asm.MovRegVal(file, asm.RegB, types.I32_Size, fmt.Sprintf("%d", str.GetSize(strIdx)))

    case token.Boolean:
        if e.Val.Str == "true" { e.Val.Str = "1" } else { e.Val.Str = "0" }
        fallthrough

    default:
        asm.MovRegVal(file, asm.RegA, e.Type.Size(), e.Val.Str)
    }
}
func (e *Ident) Compile(file *os.File) {
    if e.C != nil {
        l := Lit{ Val: e.C.Val, Type: e.C.Type }
        l.Compile(file)
        return
    }

    if e.V != nil {
        asm.MovRegDeref(file, asm.RegA, e.V.Addr(0), e.V.GetType().Size())
        return
    }

    fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not declared)\n", e.Ident.Str)
    fmt.Fprintln(os.Stderr, "\t" + e.Ident.At())
    os.Exit(1)
}
func (e *Paren) Compile(file *os.File) { e.Expr.Compile(file) }
func (e *Unary) Compile(file *os.File) {
    e.typeCheck()

    // compile time evaluation
    if c := e.constEval(); c.Type != token.Unknown {
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
    if c := e.constEval(); c.Type != token.Unknown {
        if c.Type == token.Boolean {
            if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
        }

        asm.MovRegVal(file, asm.RegA, size, c.Str)
        return
    }


    // +,-,*,/, <,<=,>,>=,==,!=
    if e.Operator.Type != token.And && e.Operator.Type != token.Or {
        // compile time evaluation (constEval only left expr)
        if c := e.OperandL.constEval(); c.Type != token.Unknown {
            if c.Type == token.Boolean {
                if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
            }

            asm.MovRegVal(file, asm.RegA, size, c.Str)
        } else {
            e.OperandL.Compile(file)

            // compile time evaluation (constEval only right expr)
            if c := e.OperandR.constEval(); c.Type != token.Unknown {
                if c.Type == token.Boolean {
                    if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
                }

                asm.BinaryOp(file, e.Operator.Type, c.Str, size)
                return
            }
        }

        if ident,ok := e.OperandR.(*Ident); ok {
            if t := ident.V.GetType(); t.Size() < size {
                asm.MovRegDeref(file, asm.RegC, ident.V.Addr(0), t.Size())
                asm.BinaryOpReg(file, e.Operator.Type, asm.RegC, size)
            } else {
                asm.BinaryOp(file, e.Operator.Type, fmt.Sprintf("%s [%s]", asm.GetWord(t.Size()), ident.V.Addr(0)), size)
            }
        } else {
            asm.Push(file, asm.RegA)

            e.OperandR.Compile(file)
            asm.MovRegReg(file, asm.RegB, asm.RegA, size)

            asm.Pop(file, asm.RegA)
            asm.BinaryOpReg(file, e.Operator.Type, asm.RegB, size)
        }

    // &&, ||
    } else {
        // compile time evaluation
        if c := e.OperandL.constEval(); c.Type != token.Unknown {
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

    regIdx := 0
    for _, val := range e.Values {
        // compile time evaluation:
        if v := val.constEval(); v.Type != token.Unknown {
            fn.PassVal(file, e.Name, regIdx, v, val.GetType())

        } else if v,ok := val.(*Ident); ok {
            fn.PassVar(file, regIdx, v.V)

        } else {
            val.Compile(file)
            fn.PassReg(file, regIdx, val.GetType())
        }

        if val.GetType().GetKind() == types.Str {
            regIdx += 2
        } else {
            regIdx++
        }
    }

    fn.CallFunc(file, e.Name)
}

func (e *XCase) Compile(file *os.File, switchCount uint) {
    cond.CaseStart(file)

    if e.Cond == nil {
        cond.CaseBody(file)
        e.Expr.Compile(file)
        return
    }

    // compile time evaluation
    if val := e.Cond.constEval(); val.Type != token.Unknown {
        if val.Str == "true" {
            cond.CaseBody(file)
            e.Expr.Compile(file)
            cond.CaseBodyEnd(file, switchCount)
        }

        return
    }

    if i,ok := e.Cond.(*Ident); ok {
        cond.CaseVar(file, i.V)
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
    if c := e.constEval(); c.Type != token.Unknown {
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


func (e *BadExpr) At() string { return "" }
func (e *FnCall)  At() string { return e.Name.At() }
func (e *Lit)     At() string { return e.Val.At() }
func (e *Ident)   At() string { return e.Ident.At() }
func (e *Unary)   At() string { return e.Operator.At() }
func (e *Binary)  At() string { return e.OperandL.At() }    // TODO: At() of Operand with higher precedence
func (e *Paren)   At() string { return e.ParenLPos.At() }
func (e *XSwitch) At() string { return e.Pos.At() }
func (e *XCase)   At() string { return e.ColonPos.At() }

func (e *BadExpr) End() string { return "" }
func (e *FnCall)  End() string { return e.ParenRPos.At() }
func (e *Lit)     End() string { return e.Val.At() }
func (e *Ident)   End() string { return e.Ident.At() }
func (e *Unary)   End() string { return e.Operand.At() }
func (e *Binary)  End() string { return e.OperandR.At() }
func (e *Paren)   End() string { return e.ParenRPos.At() }
func (e *XSwitch) End() string { return e.BraceRPos.At() }
func (e *XCase)   End() string { return e.Expr.At() }
