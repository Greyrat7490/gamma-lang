package resolver

import (
    "fmt"
    "gamma/ast"
    "gamma/types"
)

var resolvedInfers map[uint64]types.Type = make(map[uint64]types.Type)

func Resolve(a ast.Ast) ast.Ast {
    fmt.Println("[INFO] resolve types...")
    for _,d := range a.Decls {
        resolveForwardDecl(d)
    }

    for _,d := range a.Decls {
        resolveBackwardDecl(d)
        // TODO: move create identObj here (resolve names)
        // TODO: allow calling functions before defining
    }

    return a
}

func getResolvedForwardType(t types.Type) types.Type {
    if inferType,ok := t.(types.InferType); ok {
        if resolvedType := resolvedInfers[inferType.Idx]; resolvedType != nil {
            return resolvedType
        }
    }

    return t
}

func getResolvedBackwardType(t types.Type) types.Type {
    if inferType,ok := t.(types.InferType); ok {
        if resolvedType,ok := resolvedInfers[inferType.Idx]; ok {
            if t,ok := resolvedType.(types.InferType); ok {
                if t.Idx != inferType.Idx {
                    return getResolvedBackwardType(t)
                } else {
                    return t.DefaultType
                }
            }
            return resolvedType
        }
        return inferType.DefaultType
    }

    return t
}

func addResolved(dstType types.Type, t types.Type) {
    if t == nil { return }

    switch t.GetKind() {
    case types.Infer, types.Uint, types.Int, types.Generic:
        if dstType,ok := dstType.(types.InferType); ok {
            resolvedType := resolvedInfers[dstType.Idx]
            if resolvedType == nil || resolvedType.GetKind() == types.Infer {
                resolvedInfers[dstType.Idx] = t
            }
        }
    }
}
