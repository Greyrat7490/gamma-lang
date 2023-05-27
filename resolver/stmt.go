package resolver

import (
    "os"
    "fmt"
    "reflect"
    "gamma/ast"
    "gamma/types"
)

func resolveForwardStmt(s ast.Stmt) {
    switch s := s.(type) {
    case *ast.Assign:
        t := s.Dest.GetType()
        if s.Dest.GetType().GetKind() == types.Infer {
            t = s.Value.GetType()
        }

        addResolved(s.Value.GetType(), t)
        resolveForwardExpr(s.Value, t)
        resolveForwardExpr(s.Dest, t)

    case *ast.Block:
        for _,s := range s.Stmts {
            resolveForwardStmt(s)
        }

    case *ast.If:
        resolveForwardExpr(s.Cond, nil)
        resolveForwardStmt(&s.Block)
        if s.Else != nil { resolveForwardStmt(s.Else) }
        if s.Elif != nil { resolveForwardStmt(s.Elif) }

    case *ast.Else:
        resolveForwardStmt(&s.Block)

    case *ast.Elif:
        resolveForwardStmt((*ast.If)(s))

    case *ast.Switch:
        for _,s := range s.Cases {
            resolveForwardStmt(&s)
        }
    case *ast.Case:
        resolveForwardExpr(s.Cond, nil)
        for _,s := range s.Stmts {
            resolveForwardStmt(s)
        }

    case *ast.For:
        resolveForwardDecl(&s.Def)
        resolveForwardExpr(s.Limit, s.Def.Type)
        resolveForwardExpr(s.Step, s.Def.Type)
        resolveForwardStmt(&s.Block)

    case *ast.While:
        resolveForwardExpr(s.Cond, nil)
        if s.Def != nil { resolveForwardDecl(s.Def) }
        resolveForwardStmt(&s.Block)

    case *ast.Ret:
        resolveForwardExpr(s.RetExpr, s.F.GetRetType())

    case *ast.DeclStmt:
        resolveForwardDecl(s.Decl)

    case *ast.ExprStmt:
        resolveForwardExpr(s.Expr, nil)

    case *ast.Through, *ast.Break, *ast.Continue:
        // nothing to do

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] addUnresolvedStmt for %v is not implemente yet\n", reflect.TypeOf(s))
        os.Exit(1)
    }
}

func resolveBackwardStmt(s ast.Stmt) {
    switch s := s.(type) {
    case *ast.Assign:
        resolveBackwardExpr(s.Value)
        resolveBackwardExpr(s.Dest)

    case *ast.Block:
        for _,s := range s.Stmts {
            resolveBackwardStmt(s)
        }

    case *ast.If:
        resolveBackwardExpr(s.Cond)
        resolveBackwardStmt(&s.Block)
        if s.Elif != nil { resolveBackwardStmt(s.Elif) }
        if s.Else != nil { resolveBackwardStmt(s.Else) }

    case *ast.Else:
        resolveBackwardStmt(&s.Block)

    case *ast.Elif:
        resolveBackwardStmt((*ast.If)(s))

    case *ast.Switch:
        for _,s := range s.Cases {
            resolveBackwardStmt(&s)
        }
    case *ast.Case:
        resolveBackwardExpr(s.Cond)
        for _,s := range s.Stmts {
            resolveBackwardStmt(s)
        }

    case *ast.For:
        resolveBackwardDecl(&s.Def)
        resolveBackwardExpr(s.Limit)
        resolveBackwardExpr(s.Step)
        resolveBackwardStmt(&s.Block)

    case *ast.While:
        resolveBackwardExpr(s.Cond)
        if s.Def != nil { resolveBackwardDecl(s.Def) }
        resolveBackwardStmt(&s.Block)

    case *ast.Ret:
        resolveBackwardExpr(s.RetExpr)

    case *ast.DeclStmt:
        resolveBackwardDecl(s.Decl)

    case *ast.ExprStmt:
        resolveBackwardExpr(s.Expr)

    case *ast.Through, *ast.Break, *ast.Continue:
        // nothing to do

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] resolveInferStmt for %v is not implemente yet\n", reflect.TypeOf(s))
        os.Exit(1)
    }
}
