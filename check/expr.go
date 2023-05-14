package check

import (
    "os"
    "fmt"
    "reflect"
    "gamma/token"
    "gamma/types"
    "gamma/cmpTime"
    "gamma/ast"
    "gamma/ast/identObj"
)

func typeCheckExpr(e ast.Expr) {
    switch e := e.(type) {
    case *ast.ArrayLit:
        typeCheckArrayLit(e)
    case *ast.VectorLit:
        typeCheckVecLit(e)
    case *ast.StructLit:
        typeCheckStructLit(e)

    case *ast.Indexed:
        typeCheckIndexed(e)
    case *ast.Field:
        typeCheckField(e)

    case *ast.Unary:
        typeCheckUnary(e)
    case *ast.Binary:
        typeCheckBinary(e)
    case *ast.Paren:
        typeCheckExpr(e.Expr)

    case *ast.XSwitch:
        typeCheckXSwitch(e)

    case *ast.FnCall:
        if e.Ident.Name == "fmt" {
            typeCheckFmtCall(e)
        } else {
            typeCheckFnCall(e)
        }

    case *ast.Cast:
        typeCheckCast(e)

    case *ast.IntLit, *ast.UintLit, *ast.CharLit, *ast.BoolLit, *ast.PtrLit, *ast.StrLit, *ast.Ident:
        // nothing to check

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] typeCheckExpr for %v is not implemente yet\n", reflect.TypeOf(e))
        os.Exit(1)
    }
}

func checkInt(destType types.Type, val ast.Expr) bool {
    t := val.GetType()
    if t == nil || (t.GetKind() != types.Uint && t.GetKind() != types.Int) {
        return false
    }

    if t.GetKind() == types.Int && t.Size() <= destType.Size() {
        return true
    } else {
        if v,ok := cmpTime.ConstEvalInt(val); ok {
            if types.MinSizeInt(v) > destType.Size() {
                fmt.Fprintf(os.Stderr, "[ERROR] %d does not fit into %v\n", v, destType)
                fmt.Fprintln(os.Stderr, "\t" + val.At())
                os.Exit(1)
            }

            return true
        }
    }

    return false
}

func checkUint(destType types.Type, val ast.Expr) bool {
    t := val.GetType()
    if t == nil || (t.GetKind() != types.Uint && t.GetKind() != types.Int) {
        return false
    }

    if t.GetKind() == types.Uint && t.Size() <= destType.Size() {
        return true
    } else {
        if v,ok := cmpTime.ConstEvalUint(val); ok {
            if types.MinSizeUint(v) > destType.Size() {
                fmt.Fprintf(os.Stderr, "[ERROR] %d does not fit into %v\n", v, destType)
                fmt.Fprintln(os.Stderr, "\t" + val.At())
                os.Exit(1)
            }

            return true
        }
    }

    return false
}

func checkTypeExpr(destType types.Type, e ast.Expr) bool {
    typeCheckExpr(e)

    switch destType.GetKind() {
    case types.Int:
        return checkInt(destType, e)

    case types.Uint:
        return checkUint(destType, e)

    default:
        return TypesEqual(destType, e.GetType())
    }
}

func typeCheckIndexed(e *ast.Indexed) {
    typeCheckExpr(e.ArrExpr)

    switch t := e.ArrType.(type) {
    case types.ArrType:
        switch e.Index.GetType().GetKind() {
        case types.Uint, types.Int:
            if c,ok := cmpTime.ConstEvalUint(e.Index); ok {
                if c >= t.Len || c < 0 {
                    fmt.Fprintf(os.Stderr, "[ERROR] index %d is out of bounds [%d]\n", c, t.Len)
                    fmt.Fprintf(os.Stderr, "\tarray type: %v\n", e.ArrType)
                    fmt.Fprintln(os.Stderr, "\t" + e.Index.At())
                    os.Exit(1)
                }
            }
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] expected an int/uint as index but got %v\n", e.Index.GetType())
            fmt.Fprintln(os.Stderr, "\t" + e.Index.At())
            os.Exit(1)
        }
    case types.VecType:
        switch e.Index.GetType().GetKind() {
        case types.Uint, types.Int:
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] expected an int/uint as index but got %v\n", e.Index.GetType())
            fmt.Fprintln(os.Stderr, "\t" + e.Index.At())
            os.Exit(1)
        }
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] you cannot index %v", t)
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }
}

func typeCheckField(e *ast.Field) {
    typeCheckExpr(e.Obj)

    switch e.Obj.GetType().GetKind() {
    case types.Arr:
        if e.FieldName.Str != "len" {
            fmt.Fprintf(os.Stderr, "[ERROR] array has no field \"%s\" (only len)\n", e.FieldName.Str)
            fmt.Fprintln(os.Stderr, "\t" + e.FieldName.At())
            os.Exit(1)
        }
    case types.Vec:
        if e.FieldName.Str != "len" && e.FieldName.Str != "cap" {
            fmt.Fprintf(os.Stderr, "[ERROR] vec has no field \"%s\" (only len and cap)\n", e.FieldName.Str)
            fmt.Fprintln(os.Stderr, "\t" + e.FieldName.At())
            os.Exit(1)
        }
    case types.Str:
        if e.FieldName.Str != "len" {
            fmt.Fprintf(os.Stderr, "[ERROR] str has no field \"%s\" (only len)\n", e.FieldName.Str)
            fmt.Fprintln(os.Stderr, "\t" + e.FieldName.At())
            os.Exit(1)
        }
    default:
        if e.StructType.GetFieldNum(e.FieldName.Str) == -1 {
            fmt.Fprintf(os.Stderr, "[ERROR] struct %s has no %s field\n", e.StructType.Name, e.FieldName.Str)
            fmt.Fprintf(os.Stderr, "\tfields: %v\n", e.StructType.GetFields())
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }
    }
}

func typeCheckUnary(e *ast.Unary) {
    switch e.Operator.Type {
    case token.Mul:
        if u,ok := e.Operand.(*ast.Unary); ok {
            if u.Operator.Type != token.Mul {
                fmt.Fprintf(os.Stderr, "[ERROR] expected another \"*\" but got %s\n", u.Operator.Str)
                fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
                os.Exit(1)
            }
            return
        }
        if _,ok := e.Operand.(*ast.Ident); ok {
            return
        }
        if _,ok := e.Operand.(*ast.Paren); ok {
            return
        }

        fmt.Fprintln(os.Stderr, "[ERROR] expected a variable or parentheses after \"*\"")
        fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
        os.Exit(1)

    case token.Amp:
        if _,ok := e.Operand.(*ast.Ident); !ok {
            if _,ok := e.Operand.(*ast.Field); !ok {
                fmt.Fprintln(os.Stderr, "[ERROR] expected an ident or field after \"&\"")
                fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
                os.Exit(1)
            }
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
            fmt.Fprintf(os.Stderr, "[ERROR] all values in the ArrayLit should be of type %v but got a value of %v\n", o.Type.BaseType, t)
            fmt.Fprintln(os.Stderr, "\t" + v.At())
            os.Exit(1)
        }
    }
}

func typeCheckVecLit(e *ast.VectorLit) {
    if e.Cap != nil {
        if !checkTypeExpr(types.CreateUint(types.U64_Size), e.Cap) {
            fmt.Fprintf(os.Stderr, "[ERROR] expected an u64 as cap for the vector but got %v\n", e.Cap.GetType())
            fmt.Fprintln(os.Stderr, "\t" + e.Cap.At())
            os.Exit(1)
        }
    }

    if e.Len != nil {
        if !checkTypeExpr(types.CreateUint(types.U64_Size), e.Len) {
            fmt.Fprintf(os.Stderr, "[ERROR] expected an u64 as len for the vector but got %v\n", e.Len.GetType())
            fmt.Fprintln(os.Stderr, "\t" + e.Len.At())
            os.Exit(1)
        }
    }
}

func typeCheckStructLit(o *ast.StructLit) {
    for i,f := range o.Fields {
        if !checkTypeExpr(o.StructType.Types[i], f.Value) {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a %v as field %d of struct %s but got %v\n",
                o.StructType.Types[i], i, o.StructType.Name, f.GetType())
            fmt.Fprintf(os.Stderr, "\texpected: %v\n", o.StructType.Types)
            fmt.Fprintf(os.Stderr, "\tgot:      %v\n", fieldsToTypes(o.Fields))
            fmt.Fprintln(os.Stderr, "\t" + f.End())
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
    } else if !TypesEqual(t1, t2) {
        // allow ptr + u64 / u64 + ptr
        if (t1.GetKind() == types.Ptr && t2.GetKind() == types.Uint) || 
            (t2.GetKind() == types.Ptr && t1.GetKind() == types.Uint) {
            if t1.Size() == t2.Size() {
                if o.Operator.Type != token.Plus && o.Operator.Type != token.Minus {
                    fmt.Fprintln(os.Stderr, "[ERROR] you can only add or subtract a pointer with an u64")
                    fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
                    os.Exit(1)
                }
                return
            }
        }

        return 

        fmt.Fprintf(os.Stderr, "[ERROR] binary operation %s has two incompatible types (left: %v right: %v)\n",
            o.Operator.Str, t1, t2)
        fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
        os.Exit(1)
    }
}

func typeCheckXCase(s *ast.XCase) {
    if s.Cond != nil {
        typeCheckExpr(s.Cond)
        if t := s.Cond.GetType(); t.GetKind() != types.Bool {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a condition of type bool but got \"%v\"\n", t)
            fmt.Fprintln(os.Stderr, "\t" + s.ColonPos.At())
            os.Exit(1)
        }
    }

    typeCheckExpr(s.Expr)
}

func typeCheckXSwitch(o *ast.XSwitch) {
    if len(o.Cases) <= 0 {
        fmt.Fprintln(os.Stderr, "[ERROR] empty XSwitch")
        fmt.Fprintln(os.Stderr, "\t" + o.At())
        os.Exit(1)
    }

    t1 := o.Cases[0].Expr.GetType()
    typeCheckXCase(&o.Cases[0])

    for _,c := range o.Cases[1:] {
        typeCheckXCase(&c)

        t2 := c.Expr.GetType()
        if !TypesEqual(t1, t2) {
            fmt.Fprintln(os.Stderr, "[ERROR] expected every case body to return the same type but got:")
            for i,c := range o.Cases {
                fmt.Fprintf(os.Stderr, "\tcase%d: %v\n", i, c.Expr.GetType())
            }
            fmt.Fprintln(os.Stderr, "\t" + o.At())
            os.Exit(1)
        }
    }
}

func typeCheckFnCall(o *ast.FnCall) {
    if o.GenericUsedType != nil {
        o.F.GetGeneric().CurUsedType = o.GenericUsedType
    } else if o.F.IsGeneric() {
        fmt.Fprintf(os.Stderr, "[ERROR] function %s is generic but got no generic typ passed\n", o.F.GetName())
        fmt.Fprintln(os.Stderr, "\t" + o.At())
        os.Exit(1)
    }

    for _,a := range o.Values {
        typeCheckExpr(a)
    }

    if f,ok := o.Ident.Obj.(*identObj.Func); ok {
        if len(f.GetArgs()) != len(o.Values) {
            fmt.Fprintf(os.Stderr, "[ERROR] expected %d args for function \"%s\" but got %d\n", len(f.GetArgs()), f.GetName(), len(o.Values))
            fmt.Fprintf(os.Stderr, "\texpected: %v\n", f.GetArgs())
            fmt.Fprintf(os.Stderr, "\tgot:      %v\n", valuesToTypes(o.Values))
            fmt.Fprintln(os.Stderr, "\t" + o.At())
            os.Exit(1)
        }

        for i, t1 := range f.GetArgs() {
            if !checkTypeExpr(t1, o.Values[i]) {
                fmt.Fprintf(os.Stderr, "[ERROR] expected %v as arg %d but got %v for function \"%s\"\n", t1, i, o.Values[i].GetType(), f.GetName())
                fmt.Fprintf(os.Stderr, "\texpected: %v\n", f.GetArgs())
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

func typeCheckFmtCall(o *ast.FnCall) {
    for _,a := range o.Values {
        typeCheckExpr(a)
    }

    if len(o.Values) < 2 {
        if len(o.Values) == 1 {
            fmt.Fprintln(os.Stderr, "[ERROR] fmt got no arguments to format (only format string)")
            fmt.Fprintln(os.Stderr, "\t" + o.ParenRPos.At())
            os.Exit(1)
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] fmt got no arguments (missing format string and args to format)")
            fmt.Fprintln(os.Stderr, "\t" + o.ParenLPos.At())
            os.Exit(1)
        }
    }

    if fmtStr,ok := o.Values[0].(*ast.StrLit); ok {
        if len(fmtStr.Val.Str) < 4 {
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not a valid format string (missing {})\n", fmtStr.Val)
            fmt.Fprintln(os.Stderr, "\t" + fmtStr.At())
            os.Exit(1)
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] expected string literal as format string but got %s (%s)\n", fmtStr.Val, reflect.TypeOf(fmtStr))
        fmt.Fprintln(os.Stderr, "\t" + fmtStr.At())
        os.Exit(1)
    }
}

func valuesToTypes(values []ast.Expr) (res []types.Type) {
    for _, v := range values {
        res = append(res, v.GetType())
    }

    return res
}

func fieldsToTypes(fields []ast.FieldLit) (res []types.Type) {
    for _, f := range fields {
        res = append(res, f.GetType())
    }

    return res
}

func typeCheckCast(e *ast.Cast) {
    t := e.Expr.GetType()

    switch e.DestType.GetKind() {
    case types.Bool, types.Int, types.Uint, types.Char:
        switch t.GetKind() {
        case types.Bool, types.Uint, types.Int, types.Char:
            return
        case types.Ptr:
            if e.DestType.GetKind() != types.Uint || e.DestType.Size() != types.Ptr_Size {
                fmt.Fprintf(os.Stderr, "[ERROR] you can cast a pointer only into an u64 (got %v)\n", t)
                fmt.Fprintln(os.Stderr, "\t" + e.Expr.At())
                os.Exit(1)
            }

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] cannot cast %v into %v\n", t, e.DestType)
            fmt.Fprintln(os.Stderr, "\t" + e.Expr.At())
            os.Exit(1)
        }


    case types.Ptr:
        switch t.GetKind() {
        case types.Str:
            dstType := e.DestType.(types.PtrType).BaseType

            if dstType.GetKind() != types.Char {
                fmt.Fprintf(os.Stderr, "[ERROR] you can only cast a string into *char (got %v)\n", t)
                fmt.Fprintln(os.Stderr, "\t" + e.Expr.At())
                os.Exit(1)
            }

        case types.Arr:
            dstTyp := e.DestType.(types.PtrType).BaseType
            srcTyp := t.(types.ArrType).BaseType

            if dstTyp.GetKind() != srcTyp.GetKind() {
                fmt.Fprintf(os.Stderr, "[ERROR] you can only cast an array into a pointer with the same baseType (got %v)\n", t)
                fmt.Fprintln(os.Stderr, "\t" + e.Expr.At())
                os.Exit(1)
            }

        case types.Vec:
            dstTyp := e.DestType.(types.PtrType).BaseType
            srcTyp := t.(types.VecType).BaseType

            if dstTyp.GetKind() != srcTyp.GetKind() {
                fmt.Fprintf(os.Stderr, "[ERROR] you can only cast a vector into a pointer with the same baseType (got %v)\n", t)
                fmt.Fprintln(os.Stderr, "\t" + e.Expr.At())
                os.Exit(1)
            }

        case types.Int, types.Uint:
            if !checkTypeExpr(types.CreateUint(types.Ptr_Size), e.Expr) {
                fmt.Fprintf(os.Stderr, "[ERROR] you can only cast an u64 into a pointer (got %v)\n", t)
                fmt.Fprintln(os.Stderr, "\t" + e.Expr.At())
                os.Exit(1)
            }

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] cannot cast %v into %v\n", t, e.DestType)
            fmt.Fprintln(os.Stderr, "\t" + e.Expr.At())
            os.Exit(1)
        }

    case types.Arr:
        if t.GetKind() != types.Ptr {
            fmt.Fprintf(os.Stderr, "[ERROR] you can only cast a pointer into an array (got %v)\n", t)
            fmt.Fprintln(os.Stderr, "\t" + e.Expr.At())
            os.Exit(1)
        }

    case types.Struct:
        fmt.Fprintf(os.Stderr, "[ERROR] casting to a struct(%v) is not allowed\n", e.DestType)
        fmt.Fprintln(os.Stderr, "\t" + e.AsPos.At())
        os.Exit(1)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] typeCheckCast for %v is not implemente yet\n", e.DestType)
        os.Exit(1)
    }
}
