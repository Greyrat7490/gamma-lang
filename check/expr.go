package check

import (
    "os"
    "fmt"
    "reflect"
    "strconv"
    "gamma/token"
    "gamma/types"
    "gamma/cmpTime"
    "gamma/ast"
    "gamma/ast/identObj/func"
)

func typeCheckExpr(e ast.Expr) {
    switch e := e.(type) {
    case *ast.ArrayLit:
        typeCheckArrayLit(e)
    case *ast.StructLit:
        typeCheckStructLit(e)

    case *ast.Indexed:
        typeCheckIndexed(e)
    case *ast.Field:
        // TODO

    case *ast.Unary:
        typeCheckUnary(e)
    case *ast.Binary:
        typeCheckBinary(e)
    case *ast.Paren:
        typeCheckExpr(e.Expr)

    case *ast.XSwitch:
        typeCheckXSwitch(e)

    case *ast.FnCall:
        typeCheckFnCall(e)

    case *ast.Lit, *ast.Ident:
        // nothing to check

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] typeCheckExpr for %v is not implemente yet\n", reflect.TypeOf(e))
        os.Exit(1)
    }
}

func checkInt(destType types.Type, val ast.Expr) bool {
    t := val.GetType()
    if t.GetKind() != types.Uint && t.GetKind() != types.Int {
        return false
    }

    if v := cmpTime.ConstEval(val); v.Type == token.Number {
        _,err := strconv.ParseInt(v.Str, 0, int(destType.Size()*8))
        if err != nil {
            if e,ok := err.(*strconv.NumError); ok && e.Err == strconv.ErrRange {
                fmt.Fprintf(os.Stderr, "[ERROR] %s does not fit into %v\n", v.Str, destType)
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] %s is not a valid %v\n", v.Str, destType)
            }

            fmt.Fprintln(os.Stderr, "\t" + val.At())
            os.Exit(1)
        }

        return true

    } else {
        return t.GetKind() == types.Int && t.Size() <= destType.Size()
    }
}

func checkUint(destType types.Type, val ast.Expr) bool {
    t := val.GetType()
    if t.GetKind() != types.Uint && t.GetKind() != types.Int {
        return false
    }

    if v := cmpTime.ConstEval(val); v.Type == token.Number {
        _,err := strconv.ParseUint(v.Str, 0, int(destType.Size()*8))
        if err != nil {
            if e,ok := err.(*strconv.NumError); ok && e.Err == strconv.ErrRange {
                fmt.Fprintf(os.Stderr, "[ERROR] %s does not fit into %v\n", v.Str, destType)
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] %s is not a valid %v\n", v.Str, destType)
            }

            fmt.Fprintln(os.Stderr, "\t" + val.At())
            os.Exit(1)
        }

        return true

    } else {
        return t.GetKind() == types.Uint && t.Size() <= destType.Size()
    }
}

func checkTypeExpr(destType types.Type, e ast.Expr) bool {
    typeCheckExpr(e)

    switch destType.GetKind() {
    case types.Int:
        return checkInt(destType, e)

    case types.Uint:
        return checkUint(destType, e)

    default:
        return CheckTypes(destType, e.GetType())
    }
}

func typeCheckIndexed(e *ast.Indexed) {
    if t,ok := e.ArrExpr.GetType().(types.ArrType); !ok {
        fmt.Fprintf(os.Stderr, "[ERROR] you can only index an array but got %v\n", t)
        os.Exit(1)
    } else {
        if len(t.Lens) < len(e.Indices){
            fmt.Fprintf(os.Stderr, "[ERROR] dimension of the array is %d but got %d\n", len(t.Lens), len(e.Indices))
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }
    }
}

func typeCheckUnary(e *ast.Unary) {
    switch e.Operator.Type {
    case token.Mul:
        if _,ok := e.Operand.(*ast.Ident); !ok {
            if _,ok := e.Operand.(*ast.Paren); !ok {
                fmt.Fprintln(os.Stderr, "[ERROR] expected a variable or parentheses after \"*\"")
                fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
                os.Exit(1)
            }
        }

    case token.Amp:
        if _,ok := e.Operand.(*ast.Ident); !ok {
            fmt.Fprintln(os.Stderr, "[ERROR] expected a variable after \"&\"")
            fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
            os.Exit(1)
        }

    case token.Plus, token.Minus, token.BitNot:
        if t := e.Operand.GetType(); t.GetKind() != types.Int && t.GetKind() != types.Uint {
            fmt.Fprintf(os.Stderr, "[ERROR] expected int/uint after +,-,~ unary op but got %v\n", t)
            fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
            os.Exit(1)
        }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected unary op %v\n", e.Operator)
        fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
        os.Exit(1)
    }
}

func typeCheckArrayLit(o *ast.ArrayLit) {
    for _,v := range o.Values {
        t := v.GetType()
        if !checkTypeExpr(t, v) {
            fmt.Fprintf(os.Stderr, "[ERROR] all values in the ArrayLit should be of type %v but got a value of %v\n", o.Type.Ptr.BaseType, t)
            fmt.Fprintln(os.Stderr, "\t" + v.At())
            os.Exit(1)
        }

        if cmpTime.ConstEval(v).Type == token.Unknown {
            fmt.Fprintln(os.Stderr, "[ERROR] all values in the ArrayLit should be const")
            fmt.Fprintln(os.Stderr, "\t" + v.At())
            os.Exit(1)
        }
    }
}

func typeCheckStructLit(o *ast.StructLit) {
    t := o.StructType

    for i,f := range o.Fields {
        if !checkTypeExpr(t.Types[i], f.Value) {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a %v as field %d of struct %s but got %v\n",
                t.Types[i], i, o.StructType.Name, f.GetType())
            fmt.Fprintln(os.Stderr, "\t" + f.At())
            os.Exit(1)
        }
    }
}

func typeCheckBinary(o *ast.Binary) {
    t1 := o.OperandL.GetType()
    t2 := o.OperandR.GetType()

    if o.Operator.Type == token.And || o.Operator.Type == token.Or {
        if t1.GetKind() != types.Bool || t2.GetKind() != types.Bool {
            fmt.Fprintf(os.Stderr, "[ERROR] expected 2 bools for logic op \"%s\" but got %v and %v\n", o.Operator.Str, t1, t2)
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }
    } else {
                                                    // TODO only Uint
        if (t1.GetKind() == types.Ptr && t2.GetKind() == types.Int) ||
           (t2.GetKind() == types.Ptr && t1.GetKind() == types.Int) {
            if o.Operator.Type == token.Plus || o.Operator.Type == token.Minus {
                return
            }

            fmt.Fprintf(os.Stderr, "[ERROR] only +/- operators are allowed for binary ops with %v and %v\n", t1, t2)
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        } else {
            b := false
            if t2.Size() > t1.Size() {
                b = checkTypeExpr(t2, o.OperandL)
            } else {
                b = checkTypeExpr(t1, o.OperandR)
            }

            if !b {
                fmt.Fprintf(os.Stderr, "[ERROR] binary operation has two differente types (left: %v right: %v)\n", t1, t2)
                fmt.Fprintln(os.Stderr, "\t(ptr +/- int is allowed)")
                fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
                os.Exit(1)
            }
        }
    }
}

func typeCheckXSwitch(o *ast.XSwitch) {
    if len(o.Cases) > 1 {
        t1 := o.Cases[0].Expr.GetType()

        for _,c := range o.Cases[1:] {
            t2 := c.Expr.GetType()
            if !CheckTypes(t1, t2) {
                fmt.Fprintln(os.Stderr, "[ERROR] expected every case body to return the same type but got:")
                for i,c := range o.Cases {
                    fmt.Fprintf(os.Stderr, "\tcase%d: %v\n", i, c.Expr.GetType())
                }
                fmt.Fprintln(os.Stderr, "\t" + o.At())

                os.Exit(1)
            }
        }
    }
}

func typeCheckFnCall(o *ast.FnCall) {
    for _,a := range o.Values {
        typeCheckExpr(a)
    }

    if f,ok := o.Ident.Obj.(*fn.Func); ok {
        args := f.GetArgs()

        if len(args) != len(o.Values) {
            fmt.Fprintf(os.Stderr, "[ERROR] expected %d args for function \"%s\" but got %d\n", len(args), f.GetName(), len(o.Values))
            fmt.Fprintf(os.Stderr, "\texpected: %v\n", args)
            fmt.Fprintf(os.Stderr, "\tgot:      %v\n", valuesToTypes(o.Values))
            fmt.Fprintln(os.Stderr, "\t" + o.At())
            os.Exit(1)
        }

        for i, t1 := range args {
            t2 := o.Values[i].GetType()

            if !checkTypeExpr(t1, o.Values[i]) {
                fmt.Fprintf(os.Stderr, "[ERROR] expected %v as arg %d but got %v for function \"%s\"\n", t1, i, t2, f.GetName())
                fmt.Fprintf(os.Stderr, "\texpected: %v\n", args)
                fmt.Fprintf(os.Stderr, "\tgot:      %v\n", valuesToTypes(o.Values))
                fmt.Fprintln(os.Stderr, "\t" + o.At())
                os.Exit(1)
            }
        }
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] expected identObj to be a func (in typecheck.go FnCall)")
        os.Exit(1)
    }
}

func valuesToTypes(values []ast.Expr) (res []types.Type) {
    for _, v := range values {
        res = append(res, v.GetType())
    }

    return res
}
