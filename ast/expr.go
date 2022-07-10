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
}

type BadExpr struct{}

type FnCall struct {
    Name token.Token
    Values []Expr
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
            arith.BinaryOp(file, e.Operator.Type, opR.Val.Str, size)
        case *Ident:
            v := vars.GetVar(opR.Ident.Str)
            if v == nil {
                fmt.Fprintf(os.Stderr, "[ERROR] variable %s is not declared\n", opR.Ident.Str)
                fmt.Fprintln(os.Stderr, "\t" + opR.Ident.At())
                os.Exit(1)
            }

            arith.BinaryOp(file, e.Operator.Type, fmt.Sprintf("%s [%s]", asm.GetWord(v.GetType().Size()), v.Addr(0)), size)

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

        if val.GetType().GetKind() == types.Str {
            regIdx += 2
        } else {
            regIdx++
        }
    }

    fn.CallFunc(file, e.Name)
}

func (e *XCase) Compile(file *os.File, switchCount uint) {
    vars.CreateScope()
    defer vars.RemoveScope()

    if e.Cond == nil {
        cond.Default(file)
        e.Expr.Compile(file)
        return
    }

    cond.CaseStart(file)

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
