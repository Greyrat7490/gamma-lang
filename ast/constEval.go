package ast

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/asm/x86_64"
    "gorec/identObj/vars"
    "gorec/identObj/consts"
)

func (e *Lit) constEval() token.Token { return e.Val }
func (e *Unary) constEval() token.Token {
    val := e.Operand.constEval()

    switch e.Operator.Type {
    case token.Minus:
        return token.Token{ Str: e.Operator.Str + val.Str, Type: val.Type, Pos: e.Operator.Pos }

    case token.Plus:
        return val

    case token.Amp:
        if ident,ok := e.Operand.(*Ident); ok {
            if v,ok := ident.Obj.(vars.Var); ok {
                return token.Token{ Str: v.Addr(0), Type: token.Name, Pos: e.Operator.Pos }
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] expected identObj to be a var (in constEval.go Unary)")
                os.Exit(1)
            }
        }
    }

    return token.Token{ Type: token.Unknown }
}

func (e *BadExpr) constEval() token.Token { return token.Token{ Type: token.Unknown } }
func (e *FnCall) constEval() token.Token { return token.Token{ Type: token.Unknown } }

func (e *Ident) constEval() token.Token {
    if c,ok := e.Obj.(*consts.Const); ok {
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

        return asm.BinaryOpVals(e.Operator, l, r)
    }

    return token.Token{ Type: token.Unknown }
}
