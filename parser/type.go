package prs

import (
    "os"
    "fmt"
    "reflect"
    "gamma/token"
    "gamma/types"
    "gamma/cmpTime"
    "gamma/ast"
    "gamma/ast/identObj"
    "gamma/ast/identObj/struct"
)

func GetTypeUnary(e *ast.Unary) types.Type {
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

func GetTypeBinary(e *ast.Binary) types.Type {
    // bool
    if  e.Operator.Type == token.Eql || e.Operator.Type == token.Neq ||
        e.Operator.Type == token.Grt || e.Operator.Type == token.Lss ||
        e.Operator.Type == token.Geq || e.Operator.Type == token.Leq ||
        e.Operator.Type == token.And || e.Operator.Type == token.Or {
        return types.BoolType{}
    }


    t1 := e.OperandL.GetType()
    t2 := e.OperandR.GetType()

    // ptr
    if t1.GetKind() == types.Ptr {
        // check for cases like ptr1 - ptr2
        switch t2.GetKind() {
        case types.Ptr:
            if e.Operator.Type != token.Minus {
                fmt.Fprintln(os.Stderr, "[ERROR] you can only subtract two pointer")
                fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
                os.Exit(1)
            }

            return types.CreateUint(types.Ptr_Size)
        // check for cases like &v1 + u64
        case types.Uint:
            if t2.Size() == types.Ptr_Size {
                if e.Operator.Type != token.Plus && e.Operator.Type != token.Minus {
                    fmt.Fprintln(os.Stderr, "[ERROR] you can only add or subtract a pointer by a u64")
                    fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
                    os.Exit(1)
                }

                return t1
            }
        // check for cases like &v1 + const
        default:
            if v2 := cmpTime.ConstEval(e.OperandR); v2.Type != token.Unknown {
                if types.MinSizeUint(v2.Str) <= t1.Size() {
                    return t1
                }
            }
        }

        fmt.Fprintf(os.Stderr, "[ERROR] expected an u64 or pointer for this binary operation but got %v\n", t2)
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    if t2.GetKind() == types.Ptr {
        switch t1.GetKind() {
        // check for cases like u64 + &v1
        case types.Uint:
            if t1.Size() == types.Ptr_Size {
                if e.Operator.Type != token.Plus && e.Operator.Type != token.Minus {
                    fmt.Fprintln(os.Stderr, "[ERROR] you can only add or subtract a pointer by a u64")
                    fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
                    os.Exit(1)
                }

                return t1
            }
        // check for cases like const + &v1
        default:
            if v1 := cmpTime.ConstEval(e.OperandL); v1.Type != token.Unknown {
                if types.MinSizeUint(v1.Str) <= t2.Size() {
                    return t2
                }
            }
        }

        fmt.Fprintf(os.Stderr, "[ERROR] expected an u64 or pointer for this binary operation but got %v\n", t2)
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }


    // uint / int
    if (t1.GetKind() == types.Uint || t1.GetKind() == types.Int) &&
        (t2.GetKind() == types.Uint || t2.GetKind() == types.Int) {

        // both uint or int of same size
        if t1 == t2 {
            return t1
        }

        v1 := cmpTime.ConstEval(e.OperandL)
        v2 := cmpTime.ConstEval(e.OperandR)

        // var + const
        if v1.Type == token.Unknown {
            if v2.Type != token.Unknown {
                if t1.GetKind() == types.Int {
                    if types.MinSizeInt(v2.Str) <= t1.Size() {
                        return t1
                    }
                } else if t1.GetKind() == types.Uint {
                    if types.MinSizeUint(v2.Str) <= t1.Size() {
                        return t1
                    }
                }
            }
        // const + var
        } else {
            if v2.Type == token.Unknown {
                if t2.GetKind() == types.Int {
                    if types.MinSizeInt(v1.Str) <= t2.Size() {
                        return t2
                    }
                } else if t2.GetKind() == types.Uint {
                    if types.MinSizeUint(v1.Str) <= t2.Size() {
                        return t2
                    }
                }

            // const + const
            } else {
                if t1.GetKind() == types.Int {
                    if types.MinSizeInt(v2.Str) <= t1.Size() {
                        return t1
                    }
                } else if t2.GetKind() == types.Int {
                    if types.MinSizeInt(v1.Str) <= t2.Size() {
                        return t2
                    }
                } else {
                    if t1.Size() < t2.Size() {
                        return t2
                    } else {
                        return t1
                    }
                }
            }
        }
    }

    fmt.Fprintf(os.Stderr,
        "[ERROR] binary operation (%s) has two incompatible types (left: %v right: %v)\n",
        e.Operator.Str, t1, t2)
    fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
    os.Exit(1)
    return nil
}

func GetTypeIndexed(e *ast.Indexed) types.ArrType {
    t := e.ArrExpr.GetType()

    if arrType,ok := t.(types.ArrType); ok {
        return arrType
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] expected an array type but got %v\n", reflect.TypeOf(t))
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
        return types.ArrType{}
    }
}

func GetTypesField(e *ast.Field) (types.StructType, types.Type) {
    if sType,ok := e.Obj.GetType().(types.StructType); ok {
        if s,ok := identObj.Get(sType.Name).(*structDec.Struct); ok {
            return sType, s.GetTypeOfField(e.FieldName.Str)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] struct %s is not declared\n", sType.Name)
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a struct type but got %v\n", reflect.TypeOf(e.Obj.GetType()))
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    return types.StructType{}, nil
}
