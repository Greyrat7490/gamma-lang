package check

import (
    "gamma/ast"
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

func CheckTypes(destType types.Type, srcType types.Type) bool {
    switch t := destType.(type) {
    case types.ArrType:
        if t2,ok := srcType.(types.ArrType); ok {
            if CheckTypes(t.Ptr.BaseType, t2.Ptr.BaseType) {
                if len(t.Lens) == len(t2.Lens) {
                    for i,l := range t.Lens {
                        if l != t2.Lens[i] {
                            return false
                        }
                    }

                    return true
                }
            }
        }

    case types.PtrType:
        // allow generic ptr with any other pointer
        if t.BaseType == nil && srcType.GetKind() == types.Ptr {
            return true
        }

        if t2,ok := srcType.(types.PtrType); ok {
            return CheckTypes(t.BaseType, t2.BaseType)
        }

    case types.StructType:
        if t2,ok := srcType.(types.StructType); ok {
            for i,t := range t.Types {
                if !CheckTypes(t, t2.Types[i]) {
                    return false
                }
            }

            return true
        }

    default:
        return destType == srcType
    }

    return false
}
