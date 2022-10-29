package cmpTime

import (
    "os"
    "fmt"
    "reflect"
    "gamma/token"
    "gamma/types"
    "gamma/types/array"
    "gamma/gen/asm/x86_64"
    "gamma/cmpTime/constVal"
    "gamma/ast"
    "gamma/ast/identObj"
    "gamma/ast/identObj/vars"
    "gamma/ast/identObj/consts"
    "gamma/ast/identObj/struct"
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

func ConstEval(e ast.Expr) constVal.ConstVal {
    switch e := e.(type) {
    case *ast.IntLit:
        return (*constVal.IntConst)(&e.Repr)
    case *ast.UintLit:
        return (*constVal.UintConst)(&e.Repr)
    case *ast.BoolLit:
        return (*constVal.BoolConst)(&e.Repr)
    case *ast.CharLit:
        return (*constVal.CharConst)(&e.Repr)
    case *ast.PtrLit:
        return &constVal.PtrConst{ Addr: e.Addr, Local: e.Local }

    case *ast.StrLit:
        return (*constVal.StrConst)(&e.Idx)
    case *ast.ArrayLit:
        return (*constVal.ArrConst)(&e.Idx)
    case *ast.StructLit:
        return ConstEvalStructLit(e)
    case *ast.FieldLit:
        return ConstEval(e.Value)

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

    return nil
}

func ConstEvalIdent(e *ast.Ident) constVal.ConstVal {
    if c,ok := e.Obj.(*consts.Const); ok {
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
        if arrIdx,ok := ConstEval(e.ArrExpr).(*constVal.ArrConst); ok {
            return array.GetValues(uint64(*arrIdx))[idx]
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
    if c := ConstEval(e.Obj); c != nil {
        if strct,ok := c.(*constVal.StructConst); ok {
            obj := identObj.Get(e.StructType.Name)
            if s,ok := obj.(*structDec.Struct); ok {
                if i,ok := s.GetFieldNum(e.FieldName.Str); ok {
                    return strct.Fields[i]
                } else {
                    fmt.Fprintf(os.Stderr, "[ERROR] struct %s has no %s field\n", e.StructType.Name, e.FieldName)
                    fmt.Fprintln(os.Stderr, "\t" + e.At())
                    os.Exit(1)
                }
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a *constVal.StructConst but got %v\n", reflect.TypeOf(c))
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
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
            if ident,ok := e.Operand.(*ast.Ident); ok {
                if t,ok := ident.GetType().(types.PtrType); ok {
                    if addr := getVal(ident.Name, ident.Pos); addr != nil {
                        return getValAddr(addr.(*constVal.PtrConst).Addr, t.BaseType)
                    } else {
                        fmt.Fprintf(os.Stderr, "[ERROR] %s is not declared\n", ident.Name)
                        fmt.Fprintln(os.Stderr, "\t" + e.At())
                        os.Exit(1)
                    }
                } else {
                    fmt.Fprintf(os.Stderr, "[ERROR] expected a pointer type to dereference but got %v\n", ident.GetType())
                    fmt.Fprintln(os.Stderr, "\t" + e.At())
                    os.Exit(1)
                }
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] only dereferencing an identifier is support in const funcs yet")
                fmt.Fprintln(os.Stderr, "\t" + e.At())
                os.Exit(1)
            }
        }
        return nil

    case token.Amp:
        if ident,ok := e.Operand.(*ast.Ident); ok {
            if v,ok := ident.Obj.(vars.Var); ok {
                // global vars are lables with optional offset -> constEval for assembler
                if _,ok := v.(*vars.GlobalVar); ok {
                    return &constVal.PtrConst{ Addr: v.Addr(0), Local: false }
                // local vars are rbp with a const offset -> not constEval for assembler
                } else {
                    return &constVal.PtrConst{ Addr: v.Addr(0), Local: true }
                }
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] expected identObj to be a var (in constEval.go Unary)")
                os.Exit(1)
            }
        }
    }

    return nil
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
            switch r := r.(type) {
            case *constVal.UintConst:
                c := *l
                c.Addr += e.Operator.Str + fmt.Sprint(*r)
                return &c

            case *constVal.IntConst:
                c := *l
                c.Addr += e.Operator.Str + fmt.Sprint(*r)
                return &c
            }

        case *constVal.UintConst:
            switch r := r.(type) {
            case *constVal.UintConst:
                return asm.BinaryOpEvalUints(e.Operator, uint64(*l), uint64(*r))

            case *constVal.IntConst:
                return asm.BinaryOpEvalUints(e.Operator, uint64(*l), uint64(*r))

            case *constVal.PtrConst:
                c := *r
                c.Addr += e.Operator.Str + fmt.Sprint(uint64(*l))
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
                c.Addr += e.Operator.Str + fmt.Sprint(int64(*l))
                return &c
            }

        case *constVal.BoolConst:
            if r, ok := r.(*constVal.BoolConst); ok {
                return asm.BinaryOpEvalBools(e.Operator, bool(*l), bool(*r))
            }
        }
    }

    return nil
}
