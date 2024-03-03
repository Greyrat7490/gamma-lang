package resolver

import (
	"fmt"
	"gamma/ast"
	"gamma/ast/identObj"
	"gamma/ast/identObj/vars"
	"gamma/token"
	"gamma/types"
	"os"
	"reflect"
)

func resolveForwardExpr(e ast.Expr, t types.Type) {
    if e == nil { return }

    switch e := e.(type) {
    case *ast.ArrayLit:
        for _,val := range e.Values {
            resolveForwardExpr(val, e.Type.BaseType)
        }

    case *ast.VectorLit:
        resolveForwardExpr(e.Cap, types.CreateUint(types.Ptr_Size))
        resolveForwardExpr(e.Len, types.CreateUint(types.Ptr_Size))

    case *ast.StructLit:
        for i,field := range e.Fields {
            t := e.StructType.GetType(field.Name.Str)
            if t == nil {
                t = e.StructType.Types[i]
            }
            addResolved(field.Value.GetType(), t)
            resolveForwardExpr(field.Value, t)
        }

    case *ast.EnumLit:
        if e.Content != nil {
            resolveForwardExpr(e.Content.Expr, e.ContentType)
        }

    case *ast.Indexed:
        resolveForwardExpr(e.ArrExpr, e.ArrType)
        resolveForwardExpr(e.Index, types.CreateUint(types.Ptr_Size))
        addResolved(e.Type, t)
        e.Type = getResolvedForwardType(e.Type)

    case *ast.Unary:
        switch e.Operator.Type {
        case token.Amp:
            if t,ok := t.(types.PtrType); ok {
                resolveForwardExpr(e.Operand, t.BaseType)
            }
        case token.Mul:
            resolveForwardExpr(e.Operand, types.PtrType{ BaseType: t })
        default:
            resolveForwardExpr(e.Operand, t)
        }

        addResolved(e.Type, t)
        e.Type = getResolvedForwardType(e.Type)

    case *ast.Binary:
        if e.GetType().GetKind() == types.Ptr {
            t = types.CreateUint(types.Ptr_Size)
        } else {
            if !types.IsResolvable(e.OperandL.GetType()) {
                t = e.OperandL.GetType()
            } else if !types.IsResolvable(e.OperandR.GetType()) || e.GetType().GetKind() == types.Bool {
                t = e.OperandR.GetType()
            }
        }

        addResolved(e.Type, t)
        e.Type = getResolvedForwardType(e.Type)
        resolveForwardExpr(e.OperandL, t)
        resolveForwardExpr(e.OperandR, t)

    case *ast.Paren:
        resolveForwardExpr(e.Expr, t)

    case *ast.XSwitch:
        if t != nil && t.GetKind() == types.Infer {
            for _,c := range e.Cases {
                if c.GetType().GetKind() != types.Infer && (t.GetKind() == types.Infer || types.Equal(c.GetType(), t)) {
                    t = c.GetType()
                }
            }
        }

        addResolved(e.Type, t)
        e.Type = getResolvedForwardType(e.Type)
        for _,c := range e.Cases {
            resolveForwardExpr(c.Cond, nil)
            resolveForwardExpr(c.Expr, e.Type)
        }

    case *ast.FnCall:
        resolveFuncIdent(e)

        if e.F.GetName() == "fmt" {
            for _,arg := range e.Values {
                resolveForwardExpr(arg, nil)
            }
        } else {
                                                                        // resolving from interface not supported
            if e.FnSrc != nil && types.IsResolvable(e.F.GetSrcObj()) && e.FnSrc.GetKind() != types.Interface {
                fnSrcObj := e.F.GetSrcObj()
                addResolved(fnSrcObj, e.FnSrc)
                e.F.ResolveFnSrc(getResolvedForwardType(fnSrcObj))
            }

            for i,arg := range e.Values {
                if i < len(e.F.GetArgs()) {
                    t := types.ReplaceGeneric(e.F.GetArgs()[i], e.InsetType) 
                    addResolved(arg.GetType(), t)
                    resolveForwardExpr(arg, t)
                }
            }

            if e.F.IsGeneric() {
                for i,insetType := range e.F.GetUsedInsetTypes() {
                    e.F.GetUsedInsetTypes()[i] = getResolvedForwardType(insetType)
                }
                e.InsetType = getResolvedForwardType(e.InsetType)
            }
        }

    case *ast.Ident:
        if e.Obj == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] %s is not defined\n", e.Name)
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }
        
        if e.GetType() == nil { return }

        addResolved(e.GetType(), t)
        switch o := e.Obj.(type) {
        case *identObj.Const:
            o.ResolveType(getResolvedForwardType(e.GetType()))
        case vars.Var:
            o.ResolveType(getResolvedForwardType(e.GetType()))
        }

    case *ast.Cast:
        if e.DestType.GetKind() == types.Ptr {
            t = types.CreateUint(types.Ptr_Size)
        }
        resolveForwardExpr(e.Expr, t)

    case *ast.IntLit:
        addResolved(e.Type, t)
        e.Type = getResolvedForwardType(e.Type)

    case *ast.CharLit, *ast.BoolLit, *ast.PtrLit, *ast.StrLit, *ast.Field, *ast.Unwrap:
        // nothing to do

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] resolveInferExpr for %v is not implemente yet\n", reflect.TypeOf(e))
        os.Exit(1)
    }
}

func resolveBackwardExpr(e ast.Expr) {
    if e == nil { return }

    switch e := e.(type) {
    case *ast.ArrayLit:
        for _,e := range e.Values {
            resolveBackwardExpr(e)
        }

    case *ast.VectorLit:
        resolveBackwardExpr(e.Cap)
        resolveBackwardExpr(e.Len)

    case *ast.StructLit:
        for _,e := range e.Fields {
            resolveBackwardExpr(e.Value)
        }

    case *ast.EnumLit:
        if e.Content != nil {
            resolveBackwardExpr(e.Content.Expr)
        }

    case *ast.Indexed:
        resolveBackwardExpr(e.ArrExpr)
        resolveBackwardExpr(e.Index)
        e.Type = getResolvedBackwardType(e.GetType())

    case *ast.Unary:
        resolveBackwardExpr(e.Operand)
        e.Type = getResolvedBackwardType(e.GetType())

    case *ast.Binary:
        e.Type = getResolvedBackwardType(e.Type)
        resolveBackwardExpr(e.OperandL)
        resolveBackwardExpr(e.OperandR)

    case *ast.Paren:
        resolveBackwardExpr(e.Expr)

    case *ast.XSwitch:
        e.Type = getResolvedBackwardType(e.Type)
        for _,e := range e.Cases {
            resolveBackwardExpr(e.Cond)
            resolveBackwardExpr(e.Expr)
        }

    case *ast.FnCall:
        if e.FnSrc != nil && types.IsResolvable(e.F.GetSrcObj()) {
            e.F.ResolveFnSrc(getResolvedBackwardType(e.F.GetSrcObj()))
        }

        for _,e := range e.Values {
            resolveBackwardExpr(e)
        }

        if e.F.IsGeneric() {
            for i,insetType := range e.F.GetUsedInsetTypes() {
                e.F.GetUsedInsetTypes()[i] = getResolvedBackwardType(insetType)
            }
            e.InsetType = getResolvedBackwardType(e.InsetType)

            e.F.RmDuplInsetTypes()
        }

    case *ast.Cast:
        resolveBackwardExpr(e.Expr)

    case *ast.IntLit:
        e.Type = getResolvedBackwardType(e.GetType())

    case *ast.Field:
        resolveBackwardExpr(e.Obj)

    case *ast.CharLit, *ast.BoolLit, *ast.PtrLit, *ast.StrLit, *ast.Ident, *ast.Unwrap:
        // nothing to do

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] resolveInferExpr for %v is not implemente yet\n", reflect.TypeOf(e))
        os.Exit(1)
    }
}

func resolveFuncIdent(e *ast.FnCall) {
    if e.F.IsUnresolved() {
        if obj := identObj.Get(e.Ident.Name); obj != nil {
            if f,ok := obj.(*identObj.Func); ok {
                addResolved(e.F.GetRetType(), f.GetRetType())

                e.F = f
                e.Ident.Obj = obj
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] %s is not a function\n", e.Ident.Name)
                fmt.Fprintln(os.Stderr, "\t" + e.At())
                os.Exit(1)
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] %s is not defined\n", e.Ident.Name)
            fmt.Fprintln(os.Stderr, "\t" + e.At())
            os.Exit(1)
        }
    }
}
