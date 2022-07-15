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

    switch e.Operator.Type {
    case token.Mul:
        if _,ok := e.Operand.(*Ident); !ok {
            if _,ok := e.Operand.(*Paren); !ok {
                fmt.Fprintln(os.Stderr, "[ERROR] expected a variable or parentheses after \"*\"")
                fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
                os.Exit(1)
            }
        }
        e.Operand.Compile(file)

    case token.Amp:
        if i,ok := e.Operand.(*Ident); ok {
            vars.AddrToRax(file, i.Ident)
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] expected a variable after \"&\"")
            fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
            os.Exit(1)
        }

    default:
        e.Operand.Compile(file)
        if e.Operator.Type == token.Minus {
            size := e.Operand.GetType().Size()
            vars.Write(file, asm.Neg(asm.GetReg(asm.RegA, size), size))
        }
    }
}
func (e *Binary) Compile(file *os.File) {
    e.typeCheck()

    e.OperandL.Compile(file)
    if u,ok := e.OperandL.(*Unary); ok && u.Operator.Type == token.Mul {
        vars.Write(file, asm.DerefRax(u.GetType().Size()))
    }

    // +,-,*,/, <,<=,>,>=,==,!=
    if e.Operator.Type != token.And && e.Operator.Type != token.Or {
        size := e.OperandL.GetType().Size()
        if sizeR := e.OperandR.GetType().Size(); sizeR > size {
            size = sizeR
        }

        switch opR := e.OperandR.(type) {
        case *Lit:
            switch opR.Val.Type {
            case token.Str:
                // TODO
                fmt.Fprintln(os.Stderr, "[ERROR] TODO: expr.go compile Binary with Str")
                os.Exit(1)
            case token.Boolean:
                if opR.Val.Str == "true" { opR.Val.Str = "1" } else { opR.Val.Str = "0" }
                fallthrough
            default:
                arith.BinaryOp(file, e.Operator.Type, opR.Val.Str, size)
            }

        case *Ident:
            if c := vars.GetConst(opR.Ident.Str); c != nil {
                arith.BinaryOp(file, e.Operator.Type, c.Val.Str, size)
            } else {
                v := vars.GetVar(opR.Ident.Str)
                if v == nil {
                    fmt.Fprintf(os.Stderr, "[ERROR] variable %s is not declared\n", opR.Ident.Str)
                    fmt.Fprintln(os.Stderr, "\t" + opR.Ident.At())
                    os.Exit(1)
                }

                arith.BinaryOp(file, e.Operator.Type, fmt.Sprintf("%s [%s]", asm.GetWord(v.GetType().Size()), v.Addr(0)), size)
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
        count := cond.LogicalOp(file, e.Operator)
        e.OperandR.Compile(file)
        cond.LogicalOpEnd(file, count)
    }
}

func (e *FnCall) Compile(file *os.File) {
    e.typeCheck()

    regIdx := 0
    for _, val := range e.Values {
        switch v := val.(type) {
        case *Lit:
            fn.PassVal(file, e.Name, regIdx, v.Val)

        case *Ident:
            if c := vars.GetConst(v.Ident.Str); c != nil {
                fn.PassVal(file, e.Name, regIdx, c.Val)
            } else {
                fn.PassVar(file, regIdx, v.Ident)
            }

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

    if i,ok := e.Cond.(*Ident); ok {
        if c := vars.GetConst(i.Ident.Str); c != nil {
            if c.Val.Str == "true" {
                cond.CaseBody(file)
                e.Expr.Compile(file)
                cond.CaseBodyEnd(file, switchCount)
            }
            return
        }

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
