package prs

import (
    "os"
    "fmt"
    "reflect"
    "gamma/ast"
    "gamma/token"
    "gamma/types"
)

func getTypeUnary(e *ast.Unary) types.Type {
    if e.Operator.Type == token.Amp {
        return types.PtrType{ BaseType: e.Operand.GetType() }
    }

    if e.Operator.Type == token.Minus {
        t := e.Operand.GetType()
        if t.GetKind() == types.Uint {
            return types.CreateInt(t.Size())
        }
        return t
    }

    if e.Operator.Type == token.Mul {
        if ptr, ok := e.Operand.GetType().(types.PtrType); ok {
            return ptr.BaseType
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] you cannot deref this expr (expected a pointer/address)")
            fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
            os.Exit(1)
        }
    }

    return e.Operand.GetType()
}

func getTypeBinary(e *ast.Binary) types.Type {
    // bool
    if  e.Operator.Type == token.Eql || e.Operator.Type == token.Neq ||
        e.Operator.Type == token.Grt || e.Operator.Type == token.Lss ||
        e.Operator.Type == token.Geq || e.Operator.Type == token.Leq ||
        e.Operator.Type == token.And || e.Operator.Type == token.Or {
        return types.BoolType{}
    }


    t1 := e.OperandL.GetType()
    t2 := e.OperandR.GetType()

    if t1.GetKind() == types.Str && t2.GetKind() == types.Str {
        if e.Operator.Type != token.Plus {
            fmt.Fprintln(os.Stderr, "[ERROR] you can only concat two strs")
            fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
            os.Exit(1)
        }

        return types.StrType{}
    }


    if t1.GetKind() == types.Ptr && t2.GetKind() == types.Ptr {
        if e.Operator.Type != token.Minus {
            fmt.Fprintln(os.Stderr, "[ERROR] you can only subtract two pointer")
            fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
            os.Exit(1)
        }

        return types.CreateUint(types.Ptr_Size)
    }
    if t2.GetKind() == types.Ptr {
        return t2
    }

    return t1
}

func getTypeIndexed(e ast.Expr) types.ArrType {
    t := e.GetType()

    if arrType,ok := t.(types.ArrType); ok {
        return arrType
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] expected an array type but got %v\n", reflect.TypeOf(t))
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
        return types.ArrType{}
    }
}
