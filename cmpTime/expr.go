package cmpTime

import (
    "os"
    "fmt"
    "reflect"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
    "gamma/gen/asm/x86_64"
    "gamma/cmpTime/constVal"
    "gamma/ast"
    "gamma/ast/identObj"
    "gamma/ast/identObj/vars"
)

func ConstEvalInt(e ast.Expr) (int64, bool) {
    if c := ConstEval(e); c != nil {
        switch c := c.(type) {
        case *constVal.IntConst:
            return int64(*c), true
        case *constVal.UintConst:
            return int64(*c), true
        }
    }

    return 0, false
}

func ConstEvalUint(e ast.Expr) (uint64, bool) {
    if c := ConstEval(e); c != nil {
        switch c := c.(type) {
        case *constVal.IntConst:
            if int64(*c) >= 0 {
                return uint64(*c), true
            }
        case *constVal.UintConst:
            return uint64(*c), true
        }
    }

    return 0, false
}

func ConstEvalArrWithNils(e ast.Expr) constVal.ConstVal {
    return constEval(e, true)
}

func ConstEval(e ast.Expr) constVal.ConstVal {
    return constEval(e, false)
}

func constEval(e ast.Expr, arrWithNils bool) constVal.ConstVal {
    switch e := e.(type) {
    case *ast.IntLit:
        if e.Type.GetKind() == types.Int {
            i := int64(e.Repr)
            return (*constVal.IntConst)(&i)
        } else {
            return (*constVal.UintConst)(&e.Repr)
        }
    case *ast.BoolLit:
        return (*constVal.BoolConst)(&e.Repr)
    case *ast.CharLit:
        return (*constVal.CharConst)(&e.Repr)
    case *ast.PtrLit:
        return &constVal.PtrConst{ Addr: e.Addr, Local: e.Local }

    case *ast.StrLit:
        return (*constVal.StrConst)(&e.Idx)
    case *ast.ArrayLit:
        elems := make([]constVal.ConstVal, 0, len(e.Values))
        if arrWithNils {
            for _,v := range e.Values {
                elems = append(elems, ConstEvalArrWithNils(v)) 
            }
        } else {
            for _,v := range e.Values {
                c := ConstEval(v)
                if c == nil {
                    return nil
                }
                elems = append(elems, c) 
            }
        }

        return &constVal.ArrConst{ Idx: e.Idx, Elems: elems, Type: e.Type }
    case *ast.VectorLit:
        return nil
    case *ast.StructLit:
        return ConstEvalStructLit(e)
    case *ast.FieldLit:
        return ConstEval(e.Value)

    case *ast.Indexed:
        return ConstEvalIndexed(e)
    case *ast.Field:
        return ConstEvalField(e)

    case *ast.EnumLit:
        var elem constVal.ConstVal = nil
        if e.Content != nil {
            elem = ConstEval(e.Content.Expr)
            if elem == nil { return nil }
        }
        return &constVal.EnumConst{ Id: e.Type.GetID(e.ElemName.Str), Type: e.Type, ElemType: e.Type.GetType(e.ElemName.Str), Elem: elem } 

    case *ast.Unwrap:
        return ConstEvalUnwrap(e)

    case *ast.Ident:
        return ConstEvalIdent(e)

    case *ast.FnCall:
        if e.Ident.Name == "sizeof" {
            return ConstEvalSizeof(e)
        } else {
            return ConstEvalFnCall(e)
        }

    case *ast.Unary:
        return ConstEvalUnary(e)
    case *ast.Binary:
        return ConstEvalBinary(e)
    case *ast.Paren:
        return ConstEvalParen(e)

    case *ast.XSwitch:
        return ConstEvalXSwitch(e)

    case *ast.Cast:
        c := ConstEval(e.Expr)
        if i,ok := c.(*constVal.IntConst); ok && e.DestType.GetKind() == types.Uint {
            switch e.DestType.Size() {
            case 1:
                c := constVal.UintConst(uint8(*i))
                return &c
            case 2:
                c := constVal.UintConst(uint16(*i))
                return &c
            case 4:
                c := constVal.UintConst(uint32(*i))
                return &c
            default:
                c := constVal.UintConst(*i)
                return &c
            }
        }
        return c

    case *ast.BadExpr:
        fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
        os.Exit(1)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] ConstEval for %v is not implemente yet\n", reflect.TypeOf(e))
        os.Exit(1)
    }

    return nil
}

func ConstEvalUnwrap(e *ast.Unwrap) constVal.ConstVal {
    if c,ok := ConstEval(e.SrcExpr).(*constVal.EnumConst); ok {
        if inConstEnv() {
            if v,ok := e.Obj.(*vars.LocalVar); ok {
                v.SetOffset(identObj.GetStackSize(), false)
                identObj.IncStackSize(v.GetType())
                defVar(e.Obj.GetName(), e.Obj.Addr(), e.Obj.GetType(), e.Obj.GetPos(), c.Elem)
            }
        }

        res := c.Id == e.EnumType.GetID(e.ElemName.Str)
        return (*constVal.BoolConst)(&res)
    }

    return nil
}

func ConstEvalIdent(e *ast.Ident) constVal.ConstVal {
    if c,ok := e.Obj.(*identObj.Const); ok {
        return c.GetVal()
    }

    if inConstEnv() {
        if val := getVal(e.Name, e.Pos); val != nil {
            return val
        }

        fmt.Fprintf(os.Stderr, "[ERROR] %s is not declared\n", e.Name)
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    return nil
}

func ConstEvalIndexed(e *ast.Indexed) constVal.ConstVal {
    idxExpr := e.Flatten()
    if idx,ok := ConstEvalUint(idxExpr); ok {
        if arr,ok := ConstEval(e.ArrExpr).(*constVal.ArrConst); ok {
            return arr.Elems[idx]
        }
    }

    return nil
}

func ConstEvalStructLit(e *ast.StructLit) constVal.ConstVal {
    res := make([]constVal.ConstVal, len(e.Fields))

    for i,v := range e.Fields {
        c := ConstEval(v.Value)
        if c == nil {
            return nil
        }
        res[i] = c
    }

    return &constVal.StructConst{ Fields: res }
}

func ConstEvalField(e *ast.Field) constVal.ConstVal {
    if t,ok := e.Obj.GetType().(types.ArrType); ok {
        l := t.Len
        return (*constVal.UintConst)(&l)
    } else {
        if c := ConstEval(e.Obj); c != nil {
            if strct,ok := c.(*constVal.StructConst); ok {
                if i := e.StructType.GetFieldNum(e.FieldName.Str); i != -1 {
                    return strct.Fields[i]
                } else {
                    fmt.Fprintf(os.Stderr, "[ERROR] struct %s has no %s field\n", e.StructType.Name, e.FieldName)
                    fmt.Fprintf(os.Stderr, "\tfields: %v\n", e.StructType.GetFields())
                    fmt.Fprintln(os.Stderr, "\t" + e.At())
                    os.Exit(1)
                }
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] expected a *constVal.StructConst but got %v\n", reflect.TypeOf(c))
                fmt.Fprintln(os.Stderr, "\t" + e.At())
                os.Exit(1)
            }
        }
    }

    return nil
}

func ConstEvalUnary(e *ast.Unary) constVal.ConstVal {
    val := ConstEval(e.Operand)

    switch e.Operator.Type {
    case token.Minus:
        switch v := val.(type) {
        case *constVal.IntConst:
            c := constVal.IntConst(-int64(*v))
            return &c
        case *constVal.UintConst:
            c := constVal.UintConst(-uint64(*v))
            return &c
        }

    case token.BitNot:
        switch v := val.(type) {
        case *constVal.IntConst:
            c := (constVal.IntConst)(^int64(*v))
            return &c
        case *constVal.UintConst:
            c := (constVal.UintConst)(^uint64(*v))
            return &c
        }

    case token.Plus:
        return val

    case token.Mul:
        if inConstEnv() {
            if t,ok := e.Operand.GetType().(types.PtrType); ok {
                if ptr,ok := ConstEval(e.Operand).(*constVal.PtrConst); ok {
                    return getValAddr(ptr.Addr, t.BaseType)
                }
            }

            fmt.Fprintf(os.Stderr, "[ERROR] expected a pointer type to dereference but got %v\n", e.Operand.GetType())
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }
        return nil

    case token.Amp:
        ptrConst := constVal.PtrConst{}

        switch op := e.Operand.(type) {
        case *ast.Field:
            ptrConst.Addr, ptrConst.Local = fieldAddr(op, int64(op.ToOffset()))
            if ptrConst.Addr.BaseAddr == "" {
                return nil
            }

        case *ast.Ident:
        if v,ok := op.Obj.(vars.Var); ok {
            // global vars are lables with optional offset -> constEval for assembler
            // local vars are rbp with a const offset -> not constEval for assembler
            _,ptrConst.Local = v.(*vars.LocalVar)
            ptrConst.Addr = v.Addr()
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] expected identObj to be a var (in constEval.go Unary &)")
            os.Exit(1)
        }

        default:
            return nil
        }

        return &ptrConst 
    }

    return nil
}

func fieldAddr(field *ast.Field, offset int64) (a addr.Addr, isLocal bool) {
    switch obj := field.Obj.(type) {
    case *ast.Field:
        return fieldAddr(obj, offset)

    case *ast.Ident:
        if v,ok := obj.Obj.(vars.Var); ok {
            _,isLocal := v.(*vars.LocalVar)
            return v.Addr().Offseted(offset), isLocal 
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] expected identObj to be a var (in constEval.go Unary &)")
            os.Exit(1)
        }
    }

    return addr.Addr{}, false
}

func ConstEvalFnCall(e *ast.FnCall) constVal.ConstVal {
    args := make([]constVal.ConstVal, len(e.Values))
    for i,val := range e.Values {
        if c := ConstEval(val); c != nil {
            args[i] = c
        } else {
            return nil
        }
    }

    return EvalFunc(e.F.GetName(), e.Ident.Pos, args)
}

func ConstEvalSizeof(e *ast.FnCall) constVal.ConstVal {
    c := constVal.UintConst(e.GenericUsedType.Size())
    return &c
}

func ConstEvalParen(e *ast.Paren) constVal.ConstVal {
    return ConstEval(e.Expr)
}

func ConstEvalXSwitch(e *ast.XSwitch) constVal.ConstVal {
    for _,c := range e.Cases {
        if c.Cond == nil {
            return ConstEval(c.Expr)
        }

        v := ConstEval(c.Cond)

        if b,ok := v.(*constVal.BoolConst); ok && bool(*b) {
            return ConstEval(c.Expr)
        } else if v == nil {
            return nil
        }
    }

    return nil
}

func ConstEvalBinary(e *ast.Binary) constVal.ConstVal {
    l := ConstEval(e.OperandL)
    r := ConstEval(e.OperandR)

    if l != nil && r != nil {
        switch l := l.(type) {
        case *constVal.PtrConst:
            var offset int64 = 0
            switch r := r.(type) {
            case *constVal.UintConst:
                offset = int64(*r)
            case *constVal.IntConst:
                offset = int64(*r)
            }

            c := *l
            if e.Operator.Type == token.Plus {
                c.Addr.Offset += offset
            } else {
                c.Addr.Offset -= offset
            }
            return &c

        case *constVal.UintConst:
            switch r := r.(type) {
            case *constVal.UintConst:
                return asm.BinaryOpEvalUints(e.Operator, uint64(*l), uint64(*r))

            case *constVal.IntConst:
                return asm.BinaryOpEvalUints(e.Operator, uint64(*l), uint64(*r))

            case *constVal.PtrConst:
                c := *r
                if e.Operator.Type == token.Plus {
                    c.Addr.Offset += int64(*l)
                } else {
                    c.Addr.Offset -= int64(*l)
                }
                return &c
            }

        case *constVal.IntConst:
            switch r := r.(type) {
            case *constVal.IntConst:
                return asm.BinaryOpEvalInts(e.Operator, int64(*l), int64(*r))

            case *constVal.UintConst:
                return asm.BinaryOpEvalUints(e.Operator, uint64(*l), uint64(*r))

            case *constVal.PtrConst:
                c := *r
                if e.Operator.Type == token.Plus {
                    c.Addr.Offset += int64(*l)
                } else {
                    c.Addr.Offset -= int64(*l)
                }
                return &c
            }

        case *constVal.BoolConst:
            if r, ok := r.(*constVal.BoolConst); ok {
                return asm.BinaryOpEvalBools(e.Operator, bool(*l), bool(*r))
            }
        case *constVal.StrConst:
            if r, ok := r.(*constVal.StrConst); ok {
                return asm.BinaryOpEvalStrs(e.Operator, uint64(*l), uint64(*r))
            }
        }
    }

    return nil
}
