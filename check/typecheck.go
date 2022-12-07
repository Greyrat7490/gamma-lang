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
    if srcType == nil {
        return false
    }

    switch t := destType.(type) {
    case types.VecType:
        if t2,ok := srcType.(types.VecType); ok {
            return CheckTypes(t.BaseType, t2.BaseType)
        }

    case types.ArrType:
        if t2,ok := srcType.(types.ArrType); ok {
            if CheckTypes(t.BaseType, t2.BaseType) {
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

    case types.IntType:
        if t2,ok := srcType.(types.IntType); ok {
            return t2.Size() <= destType.Size()
        }

    case types.StrType:
        return destType.GetKind() == srcType.GetKind()

    default:
        return destType == srcType
    }

    return false
}
