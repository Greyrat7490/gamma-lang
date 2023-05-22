package resolver

import (
    "os"
    "fmt"
    "reflect"
    "gamma/ast"
    "gamma/types"
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
        for _,field := range e.Fields {
            resolveForwardExpr(field.Value, e.StructType.GetType(field.Name.Str))
        }

    case *ast.Indexed:
        resolveForwardExpr(e.Index, types.CreateUint(types.Ptr_Size))
        addResolved(e.Type, t)
        e.Type = getResolvedForwardType(e.GetType())

    case *ast.Unary:
        resolveForwardExpr(e.Operand, t)
        addResolved(e.Type, t)
        e.Type = getResolvedForwardType(e.GetType())

    case *ast.Binary:
        if t == nil || t.GetKind() == types.Infer {
            if e.OperandL.GetType().GetKind() != types.Infer {
                t = e.OperandL.GetType()
            } else if e.OperandR.GetType().GetKind() != types.Infer {
                t = e.OperandR.GetType()
            }
        }
        resolveForwardExpr(e.OperandL, t)
        resolveForwardExpr(e.OperandR, t)
        addResolved(e.Type, t)
        e.Type = getResolvedForwardType(e.Type)

    case *ast.Paren:
        resolveForwardExpr(e.Expr, t)

    case *ast.XSwitch:
        for _,c := range e.Cases {
            resolveForwardExpr(c.Cond, nil)
            resolveForwardExpr(c.Expr, e.Type)
        }

    case *ast.FnCall:
        if e.F.GetName() == "fmt" {
            for _,arg := range e.Values {
                resolveForwardExpr(arg, nil)
            }
        } else {
            for i,arg := range e.Values {
                addResolved(arg.GetType(), e.F.GetArgs()[i])
                resolveForwardExpr(arg, e.F.GetArgs()[i])
            }
        }

    case *ast.Ident:
        addResolved(e.GetType(), t)

    case *ast.Cast:
        resolveForwardExpr(e.Expr, t)

    case *ast.IntLit:
        addResolved(e.Type, t)
        e.Type = getResolvedForwardType(t)

    case *ast.CharLit, *ast.BoolLit, *ast.PtrLit, *ast.StrLit, *ast.Field:
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

    case *ast.Indexed:
        resolveBackwardExpr(e.Index)
        e.Type = getResolvedBackwardType(e.GetType())

    case *ast.Unary:
        resolveBackwardExpr(e.Operand)
        e.Type = getResolvedBackwardType(e.GetType())

    case *ast.Binary:
        resolveBackwardExpr(e.OperandL)
        resolveBackwardExpr(e.OperandR)
        e.Type = getResolvedBackwardType(e.GetType())

    case *ast.Paren:
        resolveBackwardExpr(e.Expr)

    case *ast.XSwitch:
        for _,e := range e.Cases {
            resolveBackwardExpr(e.Cond)
            resolveBackwardExpr(e.Expr)
        }
        e.Type = getResolvedBackwardType(e.GetType())

    case *ast.FnCall:
        for _,e := range e.Values {
            resolveBackwardExpr(e)
        }

    case *ast.Cast:
        resolveBackwardExpr(e.Expr)

    case *ast.IntLit:
        e.Type = getResolvedBackwardType(e.GetType())

    case *ast.CharLit, *ast.BoolLit, *ast.PtrLit, *ast.StrLit, *ast.Field, *ast.Ident:
        // nothing to do

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] resolveInferExpr for %v is not implemente yet\n", reflect.TypeOf(e))
        os.Exit(1)
    }
}
