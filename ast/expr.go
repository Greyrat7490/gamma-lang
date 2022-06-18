package ast

import (
    "os"
    "fmt"
    "gorec/func"
    "gorec/vars"
    "gorec/token"
    "gorec/types"
    "gorec/arithmetic"
    "gorec/asm/x86_64"
)

type OpExpr interface {
    Op
    Compile(file *os.File)
    typeCheck()
    GetType() types.Type
}

type BadExpr struct{}

type OpFnCall struct {
    FnName token.Token
    Values []OpExpr
}

type LitExpr struct {
    Val token.Token
    Type types.Type
}

type IdentExpr struct {
    Ident token.Token
}

type UnaryExpr struct {
    Operator token.Token
    Operand OpExpr
}

type BinaryExpr struct {
    OperandL OpExpr
    Operator token.Token
    OperandR OpExpr
}

type ParenExpr struct {
    ParenLPos token.Pos
    Expr OpExpr
    ParenRPos token.Pos
}


func (o *LitExpr)   Compile(file *os.File) {}
func (o *IdentExpr) Compile(file *os.File) {}
func (o *ParenExpr) Compile(file *os.File) { o.Expr.Compile(file) }
func (o *UnaryExpr) Compile(file *os.File) {
    switch o.Operator.Type {
    case token.Mul:
        switch e := o.Operand.(type) {
        case *IdentExpr:
            v := vars.GetVar(e.Ident.Str)
            if v == nil {
                fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared)\n", e.Ident.Str)
                fmt.Fprintln(os.Stderr, "\t" + e.Ident.At())
                os.Exit(1)
            }

            vars.Write(file, asm.MovRegDeref(asm.RegA, v.Addr(0), v.GetType().Size()))

        case *ParenExpr:
            o.Operand.Compile(file)

        default:
            fmt.Fprintln(os.Stderr, "[ERROR] expected a variable or parentheses after \"*\"")
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }

    case token.Amp:
        if e,ok := o.Operand.(*IdentExpr); ok {
            vars.AddrToRax(file, e.Ident)
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] expected a variable after \"&\"")
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }

    default:
        size := o.Operand.GetType().Size()

        switch e := o.Operand.(type) {
        case *IdentExpr:
            v := vars.GetVar(e.Ident.Str)
            if v == nil {
                fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared)\n", e.Ident.Str)
                fmt.Fprintln(os.Stderr, "\t" + e.Ident.At())
                os.Exit(1)
            }

            vars.Write(file, asm.MovRegDeref(asm.RegA, v.Addr(0), v.GetType().Size()))

        case *LitExpr:
            vars.Write(file, asm.MovRegVal(asm.RegA, size, e.Val.Str))

        default:
            o.Operand.Compile(file)
        }

        if o.Operator.Type == token.Minus {
            vars.Write(file, asm.Neg(asm.GetReg(asm.RegA, size), size))
        }
    }
}
func (o *BinaryExpr) Compile(file *os.File) {
    size := o.OperandL.GetType().Size()
    if sizeR := o.OperandR.GetType().Size(); sizeR > size {
        size = sizeR
    }

    switch e := o.OperandL.(type) {
    case *LitExpr:
        vars.Write(file, asm.MovRegVal(asm.RegA, size, e.Val.Str))
    case *IdentExpr:
        v := vars.GetVar(e.Ident.Str)
        if v == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared)\n", e.Ident.Str)
            fmt.Fprintln(os.Stderr, "\t" + e.Ident.At())
            os.Exit(1)
        }

        vars.Write(file, asm.MovRegDeref(asm.RegA, v.Addr(0), v.GetType().Size()))

    case *UnaryExpr:
        o.OperandL.Compile(file)
        if e.Operator.Type == token.Mul {
            vars.Write(file, asm.DerefRax(size))
        }

    default:
        o.OperandL.Compile(file)
    }


    switch e := o.OperandR.(type) {
    case *LitExpr:
        arith.BinaryOp(file, o.Operator.Type, e.Val.Str, size)
    case *IdentExpr:
        v := vars.GetVar(e.Ident.Str)
        if v == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] variable %s is not declared\n", e.Ident.Str)
            fmt.Fprintln(os.Stderr, "\t" + e.Ident.At())
            os.Exit(1)
        }

        arith.BinaryOp(file, o.Operator.Type, fmt.Sprintf("%s [%s]", asm.GetWord(v.GetType().Size()), v.Addr(0)), size)

    default:
        vars.Write(file, asm.Push(asm.RegA))

        o.OperandR.Compile(file)
        if u,ok := e.(*UnaryExpr); ok && u.Operator.Type == token.Mul {
            vars.Write(file, asm.MovRegDeref(asm.RegB, "rax", size))
        } else {
            vars.Write(file, asm.MovRegReg(asm.RegB, asm.RegA, size))
        }

        vars.Write(file, asm.Pop(asm.RegA))
        arith.BinaryOpReg(file, o.Operator.Type, asm.RegB, size)
    }
}

func (o *OpFnCall) Compile(file *os.File) {
    regIdx := 0
    for _, val := range o.Values {
        switch e := val.(type) {
        case *LitExpr:
            fn.PassVal(file, o.FnName, regIdx, e.Val)

        case *IdentExpr:
            fn.PassVar(file, regIdx, e.Ident)

        case *UnaryExpr:
            size := val.GetType().Size()

            val.Compile(file)
            if e.Operator.Type == token.Mul {
                vars.Write(file, asm.DerefRax(size))
            }
            fn.PassReg(file, regIdx, size)

        default:
            val.Compile(file)
            fn.PassReg(file, regIdx, val.GetType().Size())
        }

        if val.GetType().GetKind() == types.Str {
            regIdx += 2
        } else {
            regIdx++
        }
    }

    fn.CallFunc(file, o.FnName)
}

func (o *BadExpr) Compile(file *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
}
