package ast

import (
    "os"
    "fmt"
    "strconv"
    "gorec/token"
    "gorec/types"
    "gorec/types/array"
    "gorec/asm/x86_64"
    "gorec/ast/identObj/vars"
    "gorec/ast/identObj/consts"
)

func (e *Lit)      ConstEval() token.Token { return e.Val }
func (e *ArrayLit) ConstEval() token.Token {
    return token.Token{ Type: token.Number, Str: fmt.Sprint(e.Idx) }
}
func (e *Indexed) ConstEval() token.Token {
    arrType,_ := e.ArrExpr.GetType().(types.ArrType)
    lens := arrType.GetLens()

    if len(e.Indices) > len(lens) {
        fmt.Fprintf(os.Stderr, "[ERROR] dimension of the array is %d but got %d\n", len(lens), len(e.Indices))
        os.Exit(1)
    }

    idxExpr := e.flattenIndex()
    val := idxExpr.ConstEval()
    if val.Type != token.Unknown {
        if val.Type != token.Number {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a Number but got %v\n", val)
            fmt.Fprintln(os.Stderr, "\t" + idxExpr.At())
            os.Exit(1)
        }

        idx,_ := strconv.ParseUint(val.Str, 10, 64)

        arr := e.ArrExpr.ConstEval()
        if arr.Type != token.Unknown {
            if arr.Type != token.Number {
                fmt.Fprintf(os.Stderr, "[ERROR] expected a Number but got %v\n", val)
                fmt.Fprintln(os.Stderr, "\t" + idxExpr.At())
                os.Exit(1)
            }

            arrIdx,_ := strconv.Atoi(arr.Str)
            return array.GetValues(arrIdx)[idx]
        }
    }

    return token.Token{ Type: token.Unknown }
}

func (e *Unary) ConstEval() token.Token {
    val := e.Operand.ConstEval()

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

func (e *BadExpr) ConstEval() token.Token { return token.Token{ Type: token.Unknown } }
func (e *FnCall) ConstEval() token.Token { return token.Token{ Type: token.Unknown } }

func (e *Ident) ConstEval() token.Token {
    if c,ok := e.Obj.(*consts.Const); ok {
        return c.GetVal()
    }

    return token.Token{ Type: token.Unknown }
}

func (e *Paren) ConstEval() token.Token { return e.Expr.ConstEval() }

func (e *XCase) ConstEval() token.Token {
    if e.Cond == nil {
        return token.Token{ Type: token.Unknown }
    }

    return e.Cond.ConstEval()
}

func (e *XSwitch) ConstEval() token.Token {
    for _,c := range e.Cases {
        if c.Cond == nil {
            return c.Expr.ConstEval()
        }

        v := c.ConstEval()

        if v.Type == token.Boolean && v.Str == "true" {
            return c.Expr.ConstEval()
        } else if v.Type == token.Unknown {
            return token.Token{ Type: token.Unknown }
        }
    }

    return token.Token{ Type: token.Unknown }
}

func (e *Binary) ConstEval() token.Token {
    l := e.OperandL.ConstEval()
    r := e.OperandR.ConstEval()

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
