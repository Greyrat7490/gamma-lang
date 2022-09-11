package cmpTime

import (
    "os"
    "fmt"
    "strconv"
    "reflect"
    "gamma/token"
    "gamma/types/array"
    "gamma/types/struct"
    "gamma/gen/asm/x86_64"
    "gamma/ast"
    "gamma/ast/identObj"
    "gamma/ast/identObj/vars"
    "gamma/ast/identObj/consts"
    "gamma/ast/identObj/struct"
)

func ConstEval(e ast.Expr) token.Token {
    switch e := e.(type) {
    case *ast.BasicLit:
        return ConstEvalLit(e)
    case *ast.StrLit:
        return ConstEvalStrLit(e)
    case *ast.FieldLit:
        return ConstEvalFieldLit(e)
    case *ast.ArrayLit:
        return ConstEvalArrayLit(e)
    case *ast.StructLit:
        return ConstEvalStructLit(e)

    case *ast.Indexed:
        return ConstEvalIndexed(e)
    case *ast.Field:
        return ConstEvalField(e)

    case *ast.Ident:
        return ConstEvalIdent(e)

    case *ast.FnCall:
        return ConstEvalFnCall(e)

    case *ast.Unary:
        return ConstEvalUnary(e)
    case *ast.Binary:
        return ConstEvalBinary(e)
    case *ast.Paren:
        return ConstEvalParen(e)

    case *ast.XSwitch:
        return ConstEvalXSwitch(e)

    case *ast.Cast:
        return ConstEval(e.Expr)

    case *ast.BadExpr:
        fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
        os.Exit(1)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] ConstEval for %v is not implemente yet\n", reflect.TypeOf(e))
        os.Exit(1)
    }

    return token.Token{ Type: token.Unknown }
}

func ConstEvalLit(e *ast.BasicLit) token.Token {
    return e.Repr
}

func ConstEvalStrLit(e *ast.StrLit) token.Token {
    return token.Token{ Type: token.Number, Str: fmt.Sprint(e.Idx) }
}

func ConstEvalFieldLit(e *ast.FieldLit) token.Token {
    return ConstEval(e.Value)
}

func ConstEvalStructLit(e *ast.StructLit) token.Token {
    return token.Token{ Type: token.Number, Str: fmt.Sprint(e.Idx) }
}

func ConstEvalArrayLit(e *ast.ArrayLit) token.Token {
    return token.Token{ Type: token.Number, Str: fmt.Sprint(e.Idx) }
}

func ConstEvalIndexed(e *ast.Indexed) token.Token {
    idxExpr := e.Flatten()
    val := ConstEval(idxExpr)
    if val.Type != token.Unknown {
        if val.Type != token.Number {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a Number but got %v\n", val)
            fmt.Fprintln(os.Stderr, "\t" + idxExpr.At())
            os.Exit(1)
        }

        idx,_ := strconv.ParseUint(val.Str, 10, 64)

        arr := ConstEval(e.ArrExpr)
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

func ConstEvalField(e *ast.Field) token.Token {
    if c := ConstEval(e.Obj); c.Type != token.Unknown {
        obj := identObj.Get(e.StructType.Name)
        if s,ok := obj.(*structDec.Struct); ok {
            if i,b := s.GetFieldNum(e.FieldName.Str); b {
                if c.Type != token.Number {
                    fmt.Fprintf(os.Stderr, "[ERROR] expected a Number but got %v\n", c)
                    fmt.Fprintln(os.Stderr, "\t" + c.At())
                    os.Exit(1)
                }

                idx,_ := strconv.ParseUint(c.Str, 10, 64)
                return structLit.GetValues(idx)[i]
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] struct %s has no %s field\n", e.StructType.Name, e.FieldName)
                fmt.Fprintln(os.Stderr, "\t" + e.At())
                os.Exit(1)
            }
        }
    }

    return token.Token{ Type: token.Unknown }
}

func ConstEvalUnary(e *ast.Unary) token.Token {
    val := ConstEval(e.Operand)

    switch e.Operator.Type {
    case token.Minus:
        return token.Token{ Str: e.Operator.Str + val.Str, Type: val.Type, Pos: e.Operator.Pos }

    case token.BitNot:
        return token.Token{ Str: e.Operator.Str + val.Str, Type: val.Type, Pos: e.Operator.Pos }

    case token.Plus:
        return val

    case token.Amp:
        if ident,ok := e.Operand.(*ast.Ident); ok {
            if v,ok := ident.Obj.(vars.Var); ok {
                // global vars are lables with optional offset -> constEval for assembler
                if _,ok := v.(*vars.GlobalVar); ok {
                    return token.Token{ Str: v.Addr(0), Type: token.Name, Pos: e.Operator.Pos }
                // local vars are addr with optional offset -> not constEval for assembler
                // TokenType = Str to indicate
                } else {
                    return token.Token{ Str: v.Addr(0), Type: token.Str, Pos: e.Operator.Pos }
                }
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] expected identObj to be a var (in constEval.go Unary)")
                os.Exit(1)
            }
        }
    }

    return token.Token{ Type: token.Unknown }
}

func ConstEvalFnCall(e *ast.FnCall) token.Token {
    // TODO: in work
    return token.Token{ Type: token.Unknown }
}

func ConstEvalIdent(e *ast.Ident) token.Token {
    if c,ok := e.Obj.(*consts.Const); ok {
        return c.GetVal()
    }

    return token.Token{ Type: token.Unknown }
}

func ConstEvalParen(e *ast.Paren) token.Token {
    return ConstEval(e.Expr)
}

func ConstEvalXSwitch(e *ast.XSwitch) token.Token {
    for _,c := range e.Cases {
        if c.Cond == nil {
            return ConstEval(c.Expr)
        }

        v := ConstEval(c.Cond)

        if v.Type == token.Boolean && v.Str == "1" {
            return ConstEval(c.Expr)
        } else if v.Type == token.Unknown {
            return token.Token{ Type: token.Unknown }
        }
    }

    return token.Token{ Type: token.Unknown }
}

func ConstEvalBinary(e *ast.Binary) token.Token {
    l := ConstEval(e.OperandL)
    r := ConstEval(e.OperandR)

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
