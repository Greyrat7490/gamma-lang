package ast

import (
    "os"
    "fmt"
    "strings"
    "gorec/arithmetic"
    "gorec/func"
    "gorec/token"
    "gorec/types"
    "gorec/vars"
)

type OpExpr interface {
    Op
    Compile(asm *os.File)
    expr()  // to differenciate OpExpr from OpDecl and OpStmt
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

func (o *BadExpr)    expr() {}
func (o *LitExpr)    expr() {}
func (o *IdentExpr)  expr() {}
func (o *OpFnCall)   expr() {}
func (o *UnaryExpr)  expr() {}
func (o *BinaryExpr) expr() {}
func (o *ParenExpr)  expr() {}


func (o *LitExpr)   Compile(asm *os.File) {}
func (o *IdentExpr) Compile(asm *os.File) {}
func (o *ParenExpr) Compile(asm *os.File) { o.Expr.Compile(asm) }
func (o *UnaryExpr) Compile(asm *os.File) {
    if o.Operator.Type == token.Mul {
        switch e := o.Operand.(type) {
        case *IdentExpr:
            vars.SetRax(asm, e.Ident)

        case *ParenExpr:
            o.Operand.Compile(asm)

        default:
            fmt.Fprintln(os.Stderr, "[ERROR] expected a variable or parentheses after \"*\"")
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }
    } else {
        switch e := o.Operand.(type) {
        case *IdentExpr:
            vars.SetRax(asm, e.Ident)

        case *LitExpr:
            vars.Write(asm, fmt.Sprintf("mov rax, %s\n", e.Val.Str))

        default:
            o.Operand.Compile(asm)
        }

        if o.Operator.Type == token.Minus {
            vars.Write(asm, "neg rax\n")
        }
    }
}
func (o *BinaryExpr) Compile(asm *os.File) {
    switch e := o.OperandL.(type) {
    case *LitExpr:
        vars.Write(asm, fmt.Sprintf("mov rax, %s\n", e.Val.Str))
    case *IdentExpr:
        vars.SetRax(asm, e.Ident)
    case *UnaryExpr:
        o.OperandL.Compile(asm)
        if e.Operator.Type == token.Mul {
            vars.Write(asm, "mov rax, QWORD [rax]\n")
        }
    default:
        o.OperandL.Compile(asm)
    }
 

    switch e := o.OperandR.(type) {
    case *LitExpr:
        arith.BinaryOp(asm, o.Operator.Type, e.Val.Str)
    case *IdentExpr: 
        v := vars.GetVar(e.Ident.Str)
        if v == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] variable %s is not declared\n", e.Ident.Str)
            fmt.Fprintln(os.Stderr, "\t" + e.Ident.At())
            os.Exit(1)
        }

        arith.BinaryOp(asm, o.Operator.Type, v.Get())

    default:
        vars.Write(asm, "push rbx\n")
        vars.Write(asm, "mov rbx, rax\n")

        o.OperandR.Compile(asm)
        if u,ok := e.(*UnaryExpr); ok && u.Operator.Type == token.Mul {
            vars.Write(asm, "mov rax, QWORD [rax]\n")
        }
        arith.BinaryOp(asm, o.Operator.Type, "rbx")

        vars.Write(asm, "pop rbx\n")
    }
}

func (o *OpFnCall) Compile(asm *os.File) {
    for i, val := range o.Values {
        switch e := val.(type) {
        case *LitExpr:
            fn.PassVal(asm, o.FnName, i, e.Val)

        case *IdentExpr:
            fn.PassVar(asm, o.FnName, i, e.Ident)

        case *UnaryExpr:
            val.Compile(asm)
            if e.Operator.Type == token.Mul {
                fn.PassReg(asm, o.FnName, i, "QWORD [rax]")
            } else {   
                fn.PassReg(asm, o.FnName, i, "rax")
            }

        default:
            val.Compile(asm)
            fn.PassReg(asm, o.FnName, i, "rax")
        }
    }

    fn.CallFunc(asm, o.FnName)
}

func (o *BadExpr) Compile(asm *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
}


func (o *LitExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(%v)\n", o.Val.Str, o.Type)
}

func (o *IdentExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + o.Ident.Str + "(Name)\n"
}

func (o *OpFnCall) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    res := fmt.Sprintf("%sOP_CALL_FN:\n%s%s\n", s, s2, o.FnName.Str)
    for _, e := range o.Values {
        res += e.Readable(indent+1)
    }

    return res
}

func (o *UnaryExpr) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return fmt.Sprintf("%sOP_UNARY:\n%s%s(%s)\n", s, s2, o.Operator.Str, o.Operator.Type.Readable()) +
        o.Operand.Readable(indent+1)
}

func (o *BinaryExpr) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return s + "OP_BINARY:\n" +
        o.OperandL.Readable(indent+1) +
        s2 + fmt.Sprintf("%s(%s)\n", o.Operator.Str, o.Operator.Type.Readable()) +
        o.OperandR.Readable(indent+1)
}

func (o *ParenExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "PAREN:\n" + o.Expr.Readable(indent+1)
}

func (o *BadExpr) Readable(indent int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
    return ""
}
