package ast

import (
    "os"
    "fmt"
    "gorec/vars"
    "gorec/token"
    "gorec/arithmetic"
)

func (e *Lit) constEval() token.Token { return e.Val }
func (e *Unary) constEval() token.Token {
    v := e.Operand.constEval()

    switch e.Operator.Type {
    case token.Minus:
        return token.Token{ Str: e.Operator.Str + v.Str, Type: v.Type, Pos: e.Operator.Pos }

    case token.Plus:
        return v

    case token.Amp:
        if i,ok := e.Operand.(*Ident); ok {
            v := vars.GetVar(i.Ident.Str)
            if v == nil {
                fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", i.Ident.Str)
                fmt.Fprintln(os.Stderr, "\t" + i.Ident.At())
                os.Exit(1)
            }

            return token.Token{ Str: v.Addr(0), Type: token.Name, Pos: e.Operator.Pos }
        }
    }

    return token.Token{ Type: token.Unknown }
}

func (e *BadExpr) constEval() token.Token { return token.Token{ Type: token.Unknown } }
func (e *FnCall) constEval() token.Token { return token.Token{ Type: token.Unknown } }

func (e *Ident) constEval() token.Token {
    c := vars.GetConst(e.Ident.Str)
    if c != nil {
        return c.Val
    }

    return token.Token{ Type: token.Unknown }
}

func (e *Paren) constEval() token.Token { return e.Expr.constEval() }

func (e *XCase) constEval() token.Token {
    if e.Cond == nil {
        return token.Token{ Type: token.Unknown }
    }

    return e.Cond.constEval()
}

func (e *XSwitch) constEval() token.Token {
    for _,c := range e.Cases {
        if c.Cond == nil {
            return c.Expr.constEval()
        }

        v := c.constEval()

        if v.Type == token.Boolean && v.Str == "true" {
            return c.Expr.constEval()
        } else if v.Type == token.Unknown {
            return token.Token{ Type: token.Unknown }
        }
    }

    return token.Token{ Type: token.Unknown }
}

func (e *Binary) constEval() token.Token {
    l := e.OperandL.constEval()
    r := e.OperandR.constEval()

    if l.Type != token.Unknown && r.Type != token.Unknown {
        if l.Type == token.Name {
            l.Str += e.Operator.Str + r.Str
            return l
        }
        if r.Type == token.Name {
            r.Str += e.Operator.Str + l.Str
            return r
        }

        return arith.BinaryOpVals(e.Operator, l, r)
    }

    return token.Token{ Type: token.Unknown }
}
