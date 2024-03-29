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

    if e.Operator.Type == token.Mul {
        if ptr, ok := e.Operand.GetType().(types.PtrType); ok {
            return ptr.BaseType
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a pointer to deref but got %s\n", e.Operand.GetType())
            fmt.Fprintln(os.Stderr, "\t" + e.Operand.At())
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

    if t1 == nil {
        fmt.Fprintln(os.Stderr, "[ERROR] left operand has no type")
        fmt.Fprintln(os.Stderr, "\t" + e.OperandL.At())
        os.Exit(1)
    }

    if t2 == nil {
        fmt.Fprintln(os.Stderr, "[ERROR] right operand has no type")
        fmt.Fprintln(os.Stderr, "\t" + e.OperandR.At())
        os.Exit(1)
    }

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
    if t1.GetKind() == types.Infer {
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
