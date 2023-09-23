package resolver

import (
    "os"
    "fmt"
    "reflect"
    "gamma/ast"
    "gamma/types"
)

func resolveForwardDecl(d ast.Decl) {
    switch d := d.(type) {
    case *ast.DefVar:
        if types.IsResolvable(d.Type) {
            t := d.Value.GetType()
            addResolved(d.Type, t)
            d.Type = getResolvedForwardType(d.Type)
            d.V.ResolveType(d.Type)
        } else {
            addResolved(d.Value.GetType(), d.Type)
        }

        resolveForwardExpr(d.Value, d.Type)

    case *ast.DefConst:
        if types.IsResolvable(d.Type) {
            t := d.Value.GetType()
            addResolved(d.Type, t)
            d.Type = getResolvedForwardType(d.Type)
            d.C.ResolveType(d.Type)
        } else {
            addResolved(d.Value.GetType(), d.Type)
        }

        resolveForwardExpr(d.Value, d.Type)

    case *ast.DefFn:
        resolveForwardStmt(&d.Block)

    case *ast.Impl:
        for _,d := range d.FnDefs {
            resolveForwardDecl(&d)
        }

    case *ast.Import:
        for _,d := range d.Decls {
            resolveForwardDecl(d)
        }

    case *ast.DefInterface, *ast.DefStruct, *ast.DefEnum:
        // nothing to do

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] addUnresolvedDecl for %v is not implemente yet\n", reflect.TypeOf(d))
        os.Exit(1)
    }
}

func resolveBackwardDecl(d ast.Decl) {
    switch d := d.(type) {
    case *ast.DefVar:
        d.Type = getResolvedBackwardType(d.Type)
        resolveBackwardExpr(d.Value)
        d.V.ResolveType(d.Type)

    case *ast.DefConst:
        d.Type = getResolvedBackwardType(d.Type)
        resolveBackwardExpr(d.Value)
        d.C.ResolveType(d.Type)

    case *ast.DefFn:
        resolveBackwardStmt(&d.Block)

    case *ast.Impl:
        for _,d := range d.FnDefs {
            resolveBackwardDecl(&d)
        }

    case *ast.Import:
        for _,d := range d.Decls {
            resolveBackwardDecl(d)
        }

    case *ast.DefInterface, *ast.DefStruct, *ast.DefEnum:
        // nothing to do

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] resolveInferDecl for %v is not implemente yet\n", reflect.TypeOf(d))
        os.Exit(1)
    }
}
