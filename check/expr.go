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

    case *ast.EnumLit:
        typeCheckEnumLit(e)
    case *ast.Unwrap:
        typeCheckUnwrap(e)

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

    case *ast.IntLit, *ast.CharLit, *ast.BoolLit, *ast.PtrLit, *ast.StrLit, *ast.Ident:
        // nothing to check

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] typeCheckExpr for %v is not implemente yet\n", reflect.TypeOf(e))
        os.Exit(1)
    }
}


func checkTypeExpr(destType types.Type, e ast.Expr) bool {
    typeCheckExpr(e)
    return types.Equal(destType, e.GetType())
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

func typeCheckEnumLit(e *ast.EnumLit) {
    if !e.Type.HasElem(e.ElemName.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] enum %v has no %s field\n", e.Type, e.ElemName.Str)
        fmt.Fprintf(os.Stderr, "\telems: %v\n", e.Type.GetElems())
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    if e.ContentType == nil {
        if e.Content != nil {
            fmt.Fprintf(os.Stderr, "[ERROR] enum %s::%s did not expect any content\n", e.Type.Name, e.ElemName.Str)
            fmt.Fprintln(os.Stderr, "\t" + e.Content.At())
            os.Exit(1)
        }
    } else {
         if e.Content == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] missing enum content for %s::%s (expects type %s)\n", e.Type.Name, e.ElemName.Str, e.ContentType)
            fmt.Fprintln(os.Stderr, "\t" + e.ElemName.At())
            os.Exit(1)
        }

        if !checkTypeExpr(e.ContentType, e.Content) {
            fmt.Fprintf(os.Stderr, "[ERROR] enum %s::%s expected content of type %s\n", e.Type.Name, e.ElemName.Str, e.ContentType)
            fmt.Fprintln(os.Stderr, "\t" + e.ElemName.At())
            os.Exit(1)
        }
    }
}

func typeCheckUnwrap(e *ast.Unwrap) {
    t := e.EnumType.GetType(e.ElemName.Str)
    if t != nil {
        if !e.UnusedObj && e.Obj == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] missing identifier (enum %s::%s expects an identifier for type %s)\n", e.EnumType, e.ElemName.Str, t)
            fmt.Fprintln(os.Stderr, "\t" + e.ElemName.At())
            os.Exit(1)
        }
    } else {
        if e.Obj != nil {
            fmt.Fprintf(os.Stderr, "[ERROR] enum %s::%s has no type but got identifier %s\n", e.EnumType, e.ElemName.Str, e.Obj.GetName())
            fmt.Fprintln(os.Stderr, "\t" + e.Obj.GetPos().At())
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

    case token.Minus:
        t := e.Operand.GetType()
        if !types.Equal(types.CreateInt(types.Ptr_Size), t) {
            // TODO print actual flexable type
            fmt.Fprintf(os.Stderr, "[ERROR] expected an int after - unary op but got %v\n", t)
            fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
            os.Exit(1)
        }

    case token.Plus, token.BitNot:
        if t := e.Operand.GetType(); t.GetKind() != types.Int && t.GetKind() != types.Uint {
            fmt.Fprintf(os.Stderr, "[ERROR] expected an int/uint after %s unary op but got %v\n", t, e.Operator.Str)
            fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
            os.Exit(1)
        }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected unary op %v\n", e.Operator)
        fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
        os.Exit(1)
    }

    typeCheckExpr(e.Operand)
}

func typeCheckArrayLit(o *ast.ArrayLit) {
    for _,v := range o.Values {
        if !checkTypeExpr(o.Type.BaseType, v) {
            fmt.Fprintf(os.Stderr, "[ERROR] all values in the ArrayLit should be of type %v but got a value of %v\n", o.Type.BaseType, v.GetType())
            fmt.Fprintln(os.Stderr, "\t" + v.At())
            os.Exit(1)
        }

        typeCheckExpr(v)
    }

    if uint64(len(o.Values)) != 0 && uint64(len(o.Values)) != o.Type.Len {
        if uint64(len(o.Values)) > o.Type.Len {
            fmt.Fprintf(os.Stderr, "[ERROR] too big array literal (expected len %d, but got %d)\n", o.Type.Len, len(o.Values))
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] too small array literal (expected len %d, but got %d)\n", o.Type.Len, len(o.Values))
        }
        fmt.Fprintf(os.Stderr, "\tarray type: %v\n", o.Type)
        fmt.Fprintln(os.Stderr, "\t" + o.At())
        os.Exit(1)
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

func typeCheckBinary(e *ast.Binary) {
    t1 := e.OperandL.GetType()
    t2 := e.OperandR.GetType()

    if e.Operator.Type == token.And || e.Operator.Type == token.Or {
        if t1.GetKind() != types.Bool || t2.GetKind() != types.Bool {
            fmt.Fprintf(os.Stderr, "[ERROR] expected 2 bools for logic op \"%s\" but got %v and %v\n", e.Operator.Str, t1, t2)
            fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
            os.Exit(1)
        }

    } else {
        // allow ptr + u64 / u64 + ptr / ptr - u64
        if e.Type.GetKind() == types.Ptr {
            if e.Operator.Type != token.Plus && e.Operator.Type != token.Minus {
                fmt.Fprintln(os.Stderr, "[ERROR] you can only add or subtract a pointer with an u64")
                fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
                os.Exit(1)
            }

            if t2.GetKind() == types.Ptr && e.Operator.Type == token.Minus {
                fmt.Fprintln(os.Stderr, "[ERROR] you can only subtract a pointer with an u64 (not the other way around)")
                fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
                os.Exit(1)
            }
        }

        if !types.EqualBinary(t1, t2) {
            fmt.Fprintf(os.Stderr, "[ERROR] binary operation %s has two incompatible types (left: %v right: %v)\n",
                e.Operator.Str, t1, t2)
            fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
            os.Exit(1)
        }
    }

    typeCheckExpr(e.OperandL)
    typeCheckExpr(e.OperandR)
}

func typeCheckXCase(s *ast.XCase) {
    if s.Cond != nil {
        if t := s.Cond.GetType(); t.GetKind() != types.Bool {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a condition of type bool but got \"%v\"\n", t)
            fmt.Fprintln(os.Stderr, "\t" + s.ColonPos.At())
            os.Exit(1)
        }
        typeCheckExpr(s.Cond)
    }

    typeCheckExpr(s.Expr)
}

func typeCheckXSwitch(o *ast.XSwitch) {
    if len(o.Cases) <= 0 {
        fmt.Fprintln(os.Stderr, "[ERROR] empty XSwitch")
        fmt.Fprintln(os.Stderr, "\t" + o.At())
        os.Exit(1)
    }

    for _,c := range o.Cases {
        t := c.Expr.GetType()
        if !types.Equal(o.Type, t) {
            fmt.Fprintln(os.Stderr, "[ERROR] expected every case body to return the same type but got:")
            for i,c := range o.Cases {
                fmt.Fprintf(os.Stderr, "\tcase%d: %v\n", i, c.Expr.GetType())
            }
            fmt.Fprintln(os.Stderr, "\t" + o.At())
            os.Exit(1)
        }
    }

    for _,c := range o.Cases {
        typeCheckXCase(&c)
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
        case types.Ptr:
            if e.DestType.GetKind() != types.Uint || e.DestType.Size() != types.Ptr_Size {
                fmt.Fprintf(os.Stderr, "[ERROR] you can cast a pointer only into an u64 (got %v)\n", t)
                fmt.Fprintln(os.Stderr, "\t" + e.Expr.At())
                os.Exit(1)
            }

        case types.Bool, types.Uint, types.Int, types.Char:

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
            if !types.Equal(types.CreateUint(types.Ptr_Size), t) {
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

    typeCheckExpr(e.Expr)
}
