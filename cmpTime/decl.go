package cmpTime

import (
    "os"
    "fmt"
    "reflect"
    "gamma/ast"
)

func evalDecl(d ast.Decl) {
    switch d := d.(type) {
    case *ast.DefVar:
        if val := ConstEval(d.Value); val != nil {
            defVar(d.V.GetName(), d.V.Addr(0), d.V.GetType(), d.V.GetPos(), val)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a const value to define %v\n", d.V.GetName())
            fmt.Fprintln(os.Stderr, "\t" + d.At())
            os.Exit(1)
        }

    case *ast.DefConst:
        if val := ConstEval(d.Value); val != nil {
            defConst(d.C.GetName(), d.C.GetPos(), val)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a const value to define %v\n", d.C.GetName())
            fmt.Fprintln(os.Stderr, "\t" + d.At())
            os.Exit(1)
        }

    case *ast.BadDecl:
        fmt.Fprintln(os.Stderr, "[ERROR] bad declaration")
        os.Exit(1)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] evalDecl for %v is not implemente yet\n", reflect.TypeOf(d))
        os.Exit(1)
    }
}
