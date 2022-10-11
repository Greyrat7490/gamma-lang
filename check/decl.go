package check

import (
    "os"
    "fmt"
    "reflect"
    "gamma/ast"
)

func typeCheckDecl(d ast.Decl) {
    switch d := d.(type) {
    case *ast.DefVar:
        typeCheckDefVar(d)

    case *ast.DefConst:
        typeCheckDefConst(d)

    case *ast.Import:
        typeCheckImport(d)

    case *ast.DefFn:
        typeCheckDefFn(d)

    case *ast.DefStruct:
        // nothing to do

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] typeCheckDecl for %v is not implemente yet\n", reflect.TypeOf(d))
        os.Exit(1)
    }
}

func typeCheckDefVar(d *ast.DefVar) {
    typeCheckExpr(d.Value)

    t1 := d.V.GetType()
    t2 := d.Value.GetType()

    if !checkTypeExpr(t1, d.Value) {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot define \"%s\" (type: %v) with type %v\n", d.V.GetName(), t1, t2)
        fmt.Fprintln(os.Stderr, "\t" + d.At())
        os.Exit(1)
    }
}

func typeCheckDefConst(d *ast.DefConst) {
    typeCheckExpr(d.Value)

    t2 := d.Value.GetType()
    if !checkTypeExpr(d.Type, d.Value) {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot define \"%s\" (type: %v) with type %v\n", d.C.GetName(), d.Type, t2)
        fmt.Fprintln(os.Stderr, "\t" + d.At())
        os.Exit(1)
    }
}

func typeCheckImport(d *ast.Import) {
    for _,d := range d.Decls {
        typeCheckDecl(d)
    }
}

func typeCheckDefFn(d *ast.DefFn) {
    if d.RetType != nil && !hasRet(&d.Block) {
        fmt.Fprintln(os.Stderr, "[ERROR] missing return")
        fmt.Fprintln(os.Stderr, "\t" + d.At())
        os.Exit(1)
    }

    for _,s := range d.Block.Stmts {
        typeCheckStmt(s)
    }
}

func hasRet(s ast.Stmt) bool {
    switch s := s.(type) {
    case *ast.Block:
        for _,s := range s.Stmts {
            if hasRet(s) {
                return true
            }
        }
        return false

    case *ast.If:
        if hasRet(&s.Block) {
            if s.Elif != nil {
                return hasRet((*ast.If)(s.Elif))
            } else if s.Else != nil {
                return hasRet(&s.Else.Block)
            }
        }
        return false

    case *ast.Switch:
        for _,c := range s.Cases {
            has := false
            for _,s := range c.Stmts {
                if hasRet(s) {
                    has = true
                    break
                }
            }
            if !has {
                return false
            }
        }
        return true

    case *ast.Ret:
        return true

    case *ast.For:
        return hasRet(&s.Block)

    case *ast.While:
        return hasRet(&s.Block)

    default:
        return false
    }
}
