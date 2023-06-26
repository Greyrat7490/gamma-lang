package check

import (
    "os"
    "fmt"
    "reflect"
    "gamma/ast"
    "gamma/ast/identObj"
    "gamma/types"
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

    case *ast.Impl:
        typeCheckImpl(d)

    case *ast.DefStruct, *ast.DefInterface, *ast.DefEnum:
        // nothing to do

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] typeCheckDecl for %v is not implemente yet\n", reflect.TypeOf(d))
        os.Exit(1)
    }
}

func typeCheckDefVar(d *ast.DefVar) {
    t1 := d.V.GetType()
    t2 := d.Value.GetType()

    if !checkTypeExpr(t1, d.Value) {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot define \"%s\" (type: %v) with type %v\n", d.V.GetName(), t1, t2)
        fmt.Fprintln(os.Stderr, "\t" + d.At())
        os.Exit(1)
    }

    typeCheckExpr(d.Value)
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
    if d.FnHead.RetType != nil && !hasRet(&d.Block) {
        fmt.Fprintln(os.Stderr, "[ERROR] missing return")
        fmt.Fprintln(os.Stderr, "\t" + d.At())
        os.Exit(1)
    }

    if d.FnHead.IsConst && d.FnHead.RetType == nil {
        fmt.Fprintln(os.Stderr, "[ERROR] a const func returning nothing has no purpose")
        fmt.Fprintln(os.Stderr, "\t" + d.At())
        os.Exit(1)
    }

    for _,s := range d.Block.Stmts {
        typeCheckStmt(s)
    }
}

func typeCheckImplNoInterface(d *ast.Impl) {
    for _,f := range d.FnDefs {
        typeCheckDefFn(&f)
    }
}

func typeCheckInterfaceImplemented(d *ast.Impl) {
    err := false

    for _,expected := range d.Impl.GetInterfaceFuncs() {
        found := false
        for _,f := range d.FnDefs {
            typeCheckDefFn(&f)
            if f.FnHead.Name.Str == expected.Name {
                if !types.Equal(expected, f.FnHead.F.GetType()) {
                    fmt.Fprintln(os.Stderr, "[ERROR] different function signatures in interface and impl")
                    fmt.Fprintln(os.Stderr, "\texpected: " + expected.String())
                    fmt.Fprintln(os.Stderr, "\tgot:      " + f.FnHead.F.String())
                    fmt.Fprintln(os.Stderr, "\tinterface: " + d.Impl.GetInterfaceFuncPos(expected.Name).At())
                    fmt.Fprintln(os.Stderr, "\timpl:      " + f.At())
                    err = true
                }

                found = true
                break
            }
        }

        if !found {
            fmt.Fprintln(os.Stderr, "[ERROR] missing function definition in impl")
            fmt.Fprintln(os.Stderr, "\texpected: " + expected.String())
            fmt.Fprintln(os.Stderr, "\tinterface: " + d.Impl.GetInterfaceFuncPos(expected.Name).At())
            fmt.Fprintln(os.Stderr, "\timpl:      " + d.At())
            err = true
        }
    }

    if len(d.Impl.GetInterfaceFuncs()) < len(d.FnDefs) {
        fmt.Fprintf(os.Stderr, "[ERROR] too many functions are defined in impl (expected %d got %d)\n", 
            len(d.Impl.GetInterfaceFuncs()), len(d.FnDefs))
        for _,f := range d.FnDefs {
            found := false
            for _,expected := range d.Impl.GetInterfaceFuncs() {
                if f.FnHead.Name.Str == expected.Name {
                    if types.Equal(expected, f.FnHead.F.GetType()) {
                        found = true
                        break
                    }
                }
            }

            if !found { fmt.Fprintln(os.Stderr, "\tgot: " + f.FnHead.F.String()) }
        }
        fmt.Fprintln(os.Stderr, "\tinterface: " + d.Impl.GetInterfacePos().At())
        fmt.Fprintln(os.Stderr, "\timpl:      " + d.At())
        os.Exit(1)
    }

    if err { os.Exit(1) }
}

func typeCheckImpl(d *ast.Impl) {
    if d.Impl.HasInterface() {
        typeCheckInterfaceImplemented(d)
    } else {
        typeCheckImplNoInterface(d)
    }
}

func printFuncs(interfaceFuncs []identObj.Func, implFuncs []ast.DefFn) {
    fmt.Fprintln(os.Stderr, "\tinterface:")
    for _,f := range interfaceFuncs {
        fmt.Fprintln(os.Stderr, "\t\t" + f.String())
    }
    fmt.Fprintln(os.Stderr, "\timpl:")
    for _,f := range implFuncs {
        fmt.Fprintln(os.Stderr, "\t\t" + f.FnHead.F.String())
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
            if !hasRet(c.Stmt) {
                return false
            }
        }
        return true

    case *ast.Ret, *ast.Through:
        return true

    case *ast.ExprStmt:
        if f,ok := s.Expr.(*ast.FnCall); ok {
            return f.Ident.Name == "exit"
        }
        return false

    case *ast.For:
        return hasRet(&s.Block)

    case *ast.While:
        return hasRet(&s.Block)

    default:
        return false
    }
}
