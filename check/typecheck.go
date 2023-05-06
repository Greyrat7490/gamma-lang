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

func TypesEqual(destType types.Type, srcType types.Type) bool {
    srcType = types.ReplaceGeneric(srcType)
    destType = types.ReplaceGeneric(destType)

    if t,ok := destType.(types.InterfaceType); ok {
        if t2,ok := srcType.(types.StructType); ok {
            if s,ok := identObj.Get(t2.Name).(*identObj.Struct); ok {
                return s.HasImpl(t.Name)
            }
        }
    }

    return types.Equal(destType, srcType)
}
