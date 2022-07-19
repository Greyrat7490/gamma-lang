package ast

import (
    "os"
    "fmt"
    "gorec/func"
    "gorec/vars"
    "gorec/token"
    "gorec/types"
    "gorec/types/str"
    "gorec/arithmetic"
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
}

type Unary struct {
    Operator token.Token
    Operand Expr
}

type Binary struct {
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

        vars.Write(file, asm.MovRegVal(asm.RegA, types.Ptr_Size, fmt.Sprintf("_str%d", strIdx)))
        vars.Write(file, asm.MovRegVal(asm.RegB, types.I32_Size, fmt.Sprintf("%d", str.GetSize(strIdx))))
    case token.Boolean:
        if e.Val.Str == "true" { e.Val.Str = "1" } else { e.Val.Str = "0" }
        fallthrough
    default:
        vars.Write(file, asm.MovRegVal(asm.RegA, e.Type.Size(), e.Val.Str))
    }
}
func (e *Ident) Compile(file *os.File) {
    if c := vars.GetConst(e.Ident.Str); c != nil {
        l := Lit{ Val: c.Val, Type: c.Type }
        l.Compile(file)
        return
    }

    v := vars.GetVar(e.Ident.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared)\n", e.Ident.Str)
        fmt.Fprintln(os.Stderr, "\t" + e.Ident.At())
        os.Exit(1)
    }

    vars.Write(file, asm.MovRegDeref(asm.RegA, v.Addr(0), v.GetType().Size()))
}
func (e *Paren) Compile(file *os.File) { e.Expr.Compile(file) }
func (e *Unary) Compile(file *os.File) {
    e.typeCheck()

    // compile time evaluation
    if c := e.constEval(); c.Type != token.Unknown {
        asm.MovRegVal(asm.RegA, types.TypeOfVal(c.Str).Size(), c.Str)
        return
    }

    e.Operand.Compile(file)

    if e.Operator.Type == token.Minus {
        size := e.Operand.GetType().Size()
        vars.Write(file, asm.Neg(asm.GetReg(asm.RegA, size), size))
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

        vars.Write(file, asm.MovRegVal(asm.RegA, size, c.Str))
        return
    }


    // +,-,*,/, <,<=,>,>=,==,!=
    if e.Operator.Type != token.And && e.Operator.Type != token.Or {
        // compile time evaluation (constEval only left expr)
        if c := e.OperandL.constEval(); c.Type != token.Unknown {
            if c.Type == token.Boolean {
                if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
            }

            vars.Write(file, asm.MovRegVal(asm.RegA, size, c.Str))
        } else {
            e.OperandL.Compile(file)
            if u,ok := e.OperandL.(*Unary); ok && u.Operator.Type == token.Mul {
                vars.Write(file, asm.DerefRax(size))
            }

            // compile time evaluation (constEval only right expr)
            if c := e.OperandR.constEval(); c.Type != token.Unknown {
                if c.Type == token.Boolean {
                    if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
                }

                arith.BinaryOp(file, e.Operator.Type, c.Str, size)
                return
            }
        }

        switch opR := e.OperandR.(type) {
        case *Ident:
            v := vars.GetVar(opR.Ident.Str)
            if v == nil {
                fmt.Fprintf(os.Stderr, "[ERROR] variable %s is not declared\n", opR.Ident.Str)
                fmt.Fprintln(os.Stderr, "\t" + opR.Ident.At())
                os.Exit(1)
            }

            if t := v.GetType(); t.Size() < size {
                vars.Write(file, asm.MovRegDeref(asm.RegC, v.Addr(0), t.Size()))
                arith.BinaryOpReg(file, e.Operator.Type, asm.RegC, size)
            } else {
                arith.BinaryOp(file, e.Operator.Type, fmt.Sprintf("%s [%s]", asm.GetWord(t.Size()), v.Addr(0)), size)
            }

        default:
            vars.Write(file, asm.Push(asm.RegA))

            e.OperandR.Compile(file)
            if u,ok := opR.(*Unary); ok && u.Operator.Type == token.Mul {
                vars.Write(file, asm.MovRegDeref(asm.RegB, "rax", size))
            } else {
                vars.Write(file, asm.MovRegReg(asm.RegB, asm.RegA, size))
            }

            vars.Write(file, asm.Pop(asm.RegA))
            arith.BinaryOpReg(file, e.Operator.Type, asm.RegB, size)
        }

    // &&, ||
    } else {
        // compile time evaluation
        if c := e.OperandL.constEval(); c.Type != token.Unknown {
            if e.Operator.Type == token.And && c.Str == "false" {
                vars.Write(file, asm.MovRegVal(asm.RegA, size, "0"))
                return
            }
            if e.Operator.Type == token.Or && c.Str == "true" {
                vars.Write(file, asm.MovRegVal(asm.RegA, size, "1"))
                return
            }

            e.OperandR.Compile(file)
        } else {
            e.OperandL.Compile(file)
            if u,ok := e.OperandL.(*Unary); ok && u.Operator.Type == token.Mul {
                vars.Write(file, asm.DerefRax(size))
            }

            // TODO move to arithmetic and move to asm and rename
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
        } else {
            switch v := val.(type) {
            case *Ident:
                fn.PassVar(file, regIdx, v.Ident)

            case *Unary:
                val.Compile(file)
                if v.Operator.Type == token.Mul {
                    vars.Write(file, asm.DerefRax(val.GetType().Size()))
                }
                fn.PassReg(file, regIdx, val.GetType())

            default:
                val.Compile(file)
                fn.PassReg(file, regIdx, val.GetType())
            }
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
        cond.CaseIdent(file, i.Ident)
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
        asm.MovRegVal(asm.RegA, types.TypeOfVal(c.Str).Size(), c.Str)
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
func (e *Binary)  At() string { return e.OperandL.At() }
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
