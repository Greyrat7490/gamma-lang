package check

import (
	"gamma/ast"
	"gamma/ast/identObj"
	"gamma/types"
)

func TypeCheckNode(n ast.Node) {
    switch n := n.(type) {
    case ast.Expr:
        typeCheckExpr(n)

    case ast.Stmt:
        typeCheckStmt(n)

    case ast.Decl:
        typeCheckDecl(n)
    }
}

func compatibleBinaryOp(t1 types.Type, t2 types.Type) bool {
    switch t := t1.(type) {
    case types.IntType:
        if t2,ok := t2.(types.IntType); ok {
            return t2.Size() == t.Size()
        }

    case types.UintType:
        if t2,ok := t2.(types.UintType); ok {
            return t2.Size() == t.Size()
        }
        if _,ok := t2.(types.PtrType); ok {
            return t1.Size() == types.Ptr_Size
        }

    case types.PtrType:
        if _,ok := t2.(types.PtrType); ok {
            return true
        }
        return compatible(types.CreateUint(types.Ptr_Size), t2)

    default:
        return compatible(t1, t2)
    }

    return false
}

func compatible(destType types.Type, srcType types.Type) bool {
    interfacesEqFunc := func(name1 string, name2 string)bool {
        if impl,ok := identObj.Get(name2).(identObj.Implementable); ok {
            return impl.HasInterface(name1)
        }

        return false
    }

    return types.EqualCustom(destType, srcType, interfacesEqFunc)
}
