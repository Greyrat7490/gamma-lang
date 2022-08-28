package check

import (
    "os"
    "fmt"
    "reflect"
    "gamma/ast"
    "gamma/types"
)

func typeCheckStmt(s ast.Stmt) {
    switch s := s.(type) {
    case *ast.Assign:
        typeCheckAssign(s)
    case *ast.Block:
        typeCheckBlock(s)

    case *ast.If:
        typeCheckIf(s)
    case *ast.Else:
        typeCheckElse(s)
    case *ast.Elif:
        typeCheckElif(s)

    case *ast.Switch:
        typeCheckSwitch(s)
    case *ast.Case:
        typeCheckCase(s)

    case *ast.For:
        typeCheckFor(s)
    case *ast.While:
        typeCheckWhile(s)

    case *ast.Ret:
        typeCheckRet(s)

    case *ast.DeclStmt:
        typeCheckDecl(s.Decl)
    case *ast.ExprStmt:
        typeCheckExpr(s.Expr)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] typeCheckStmt for %v is not implemente yet\n", reflect.TypeOf(s))
        os.Exit(1)
    }
}

func typeCheckAssign(s *ast.Assign) {
    t1 := s.Dest.GetType()
    t2 := s.Value.GetType()

    if !CheckTypes(t1, t2) {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot assign a type: %v with type: %v\n", t1, t2)
        fmt.Fprintln(os.Stderr, "\t" + s.Pos.At())
        os.Exit(1)
    }
}

func typeCheckBlock(s *ast.Block) {
    for _,s := range s.Stmts {
        typeCheckStmt(s)
    }
}

func typeCheckIf(s *ast.If) {
    if t := s.Cond.GetType(); t.GetKind() != types.Bool {
        fmt.Fprintf(os.Stderr, "[ERROR] expected an bool as if condition but got %v\n", t)
        fmt.Fprintln(os.Stderr, "\t" + s.Pos.At())
        os.Exit(1)
    }
}

func typeCheckElse(s *ast.Else) {
    typeCheckBlock(&s.Block)
}

func typeCheckElif(s *ast.Elif) {
    typeCheckStmt((*ast.If)(s))
}

func typeCheckSwitch(s *ast.Switch) {
    for _,c := range s.Cases {
        typeCheckCase(&c)
    }
}

func typeCheckFor(s *ast.For) {
    t := s.Def.Type

    if s.Limit != nil {
        if t2 := s.Limit.GetType(); !CheckTypes(t, t2) {
            fmt.Fprintf(os.Stderr, "[ERROR] expected %v as for iterator limit type but got %v\n", t, t2)
            fmt.Fprintln(os.Stderr, "\t" + s.ForPos.At())
            os.Exit(1)
        }
    }

    if t2 := s.Step.GetType(); !CheckTypes(t, t2) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected %v as for iterator step type but got %v\n", t, t2)
        fmt.Fprintln(os.Stderr, "\t" + s.ForPos.At())
        os.Exit(1)
    }
}

func typeCheckWhile(s *ast.While) {
    if t := s.Cond.GetType(); t.GetKind() != types.Bool {
        fmt.Fprintf(os.Stderr, "[ERROR] expected an bool as while condition but got %v\n", t)
        fmt.Fprintln(os.Stderr, "\t" + s.WhilePos.At())
        os.Exit(1)
    }
}

func typeCheckRet(s *ast.Ret) {
    t1 := s.F.GetRetType()

    if t1 == nil {
        if s.RetExpr != nil {
            fmt.Fprintf(os.Stderr, "[ERROR] expected nothing to return but got %v\n", s.RetExpr.GetType())
            fmt.Fprintln(os.Stderr, "\t" + s.At())
            os.Exit(1)
        }
    } else {
        if !CheckTypes(s.RetExpr.GetType(), t1) {
            fmt.Fprintf(os.Stderr, "[ERROR] expected to return %v but got %v\n", t1, s.RetExpr.GetType())
            fmt.Fprintln(os.Stderr, "\t" + s.At())
            os.Exit(1)
        }
    }
}

func typeCheckCase(s *ast.Case) {
    // skip default case
    if s.Cond == nil { return }

    typeCheckExpr(s.Cond)
    if t := s.Cond.GetType(); t.GetKind() != types.Bool {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a condition of type bool but got \"%v\"\n", t)
        fmt.Fprintln(os.Stderr, "\t" + s.ColonPos.At())
        os.Exit(1)
    }

    for _,s := range s.Stmts {
        typeCheckStmt(s)
    }
}
