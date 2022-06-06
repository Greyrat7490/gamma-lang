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
        if ident, ok := o.Operand.(*IdentExpr); ok {
            v := vars.GetVar(ident.Ident.Str)
            if v == nil {
                fmt.Fprintf(os.Stderr, "[ERROR] variable \"%s\" is not declared\n", ident.Ident.Str)
                fmt.Fprintln(os.Stderr, "\t" + ident.Ident.At())
                os.Exit(1)
            }

            if _, ok := v.(*vars.GlobalVar); ok {
                asm.WriteString(fmt.Sprintf("mov rax, QWORD [%s]\n", ident.Ident.Str))
            } else {
                asm.WriteString(fmt.Sprintf("mov rax, %s\n", v.Get()))
            }

            return
        }

        if _, ok := o.Operand.(*ParenExpr); ok {
            o.Operand.Compile(asm)
            return
        }

        fmt.Fprintln(os.Stderr, "[ERROR] expected a variable or parentheses after \"*\"")
        fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
        os.Exit(1)
    }


    if l, ok := o.Operand.(*LitExpr); ok {
        vars.Write(asm, fmt.Sprintf("mov rax, %s\n", l.Val.Str))
    } else if ident, ok := o.Operand.(*IdentExpr); ok {
        if vars.GetVar(ident.Ident.Str) == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] variable \"%s\" is not declared\n", ident.Ident.Str)
            fmt.Fprintln(os.Stderr, "\t" + ident.Ident.At())
            os.Exit(1)
        }

        vars.Write(asm, fmt.Sprintf("mov rax, %s\n", vars.GetVar(ident.Ident.Str).Get()))
    }

    o.Operand.Compile(asm)

    if o.Operator.Type == token.Minus {
        vars.Write(asm, "neg rax\n")
    }
}
func (o *BinaryExpr) Compile(asm *os.File) {
    if l, ok := o.OperandL.(*LitExpr); ok {
        vars.Write(asm, fmt.Sprintf("mov rax, %s\n", l.Val.Str))
    } else if ident, ok := o.OperandL.(*IdentExpr); ok {
        if v := vars.GetVar(ident.Ident.Str); v == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] variable %s is not declared\n", ident.Ident.Str)
            fmt.Fprintln(os.Stderr, "\t" + ident.Ident.At())
            os.Exit(1)
        }

        vars.Write(asm, fmt.Sprintf("mov rax, %s\n", vars.GetVar(ident.Ident.Str).Get()))
    }

    o.OperandL.Compile(asm)
    if u, ok := o.OperandL.(*UnaryExpr); ok {
        if u.Operator.Type == token.Mul {
            vars.Write(asm, "mov rax, QWORD [rax]\n")
        }
    }

    if l, ok := o.OperandR.(*LitExpr); ok {
        arith.BinaryOp(asm, o.Operator.Type, l.Val.Str)
    } else if ident, ok := o.OperandR.(*IdentExpr); ok {
        if v := vars.GetVar(ident.Ident.Str); v == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] variable %s is not declared\n", ident.Ident.Str)
            fmt.Fprintln(os.Stderr, "\t" + ident.Ident.At())
            os.Exit(1)
        }

        arith.BinaryOp(asm, o.Operator.Type, vars.GetVar(ident.Ident.Str).Get())
    } else {
        vars.Write(asm, "push rbx\n")
        vars.Write(asm, "mov rbx, rax\n")
        o.OperandR.Compile(asm)
        if u, ok := o.OperandR.(*UnaryExpr); ok {
            if u.Operator.Type == token.Mul {
                vars.Write(asm, "mov rax, QWORD [rax]\n")
            }
        }

        arith.BinaryOp(asm, o.Operator.Type, "rbx")
        vars.Write(asm, "pop rbx\n")
    }
}

func (o *OpFnCall) Compile(asm *os.File) {
    for i, val := range o.Values {
        if l, ok := val.(*LitExpr); ok {
            fn.PassVal(asm, o.FnName, i, l.Val)
        } else if ident, ok := val.(*IdentExpr); ok {
            fn.PassVar(asm, o.FnName, i, ident.Ident)
        } else if u, ok := val.(*UnaryExpr); ok {
            if u.Operator.Type == token.Mul {
                val.Compile(asm)
                fn.PassReg(asm, o.FnName, i, "QWORD [rax]")
            }
        } else {
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
