package gen

import (
    "os"
    "fmt"
    "reflect"
    "gamma/token"
    "gamma/types"
    "gamma/cmpTime"
    "gamma/ast"
    "gamma/ast/identObj/func"
    "gamma/ast/identObj/vars"
    "gamma/asm/x86_64"
)

func GenDecl(file *os.File, d ast.Decl) {
    switch d := d.(type) {
    case *ast.DefVar:
        GenDefVar(file, d)

    case *ast.DefConst:
        GenDefConst(file, d)

    case *ast.DefFn:
        GenDefFn(file, d)

    case *ast.DefStruct, *ast.DecVar:
        // nothing to generate

    case *ast.BadDecl:
        fmt.Fprintln(os.Stderr, "[ERROR] bad declaration")
        os.Exit(1)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] GenDecl for %v is not implemente yet\n", reflect.TypeOf(d))
        os.Exit(1)
    }
}

func GenDefVar(file *os.File, d *ast.DefVar) {
    if val := cmpTime.ConstEval(d.Value); val.Type != token.Unknown {
        d.V.DefVal(file, val)
        return
    }

    if _,ok := d.V.(*vars.GlobalVar); ok {
        fmt.Fprintln(os.Stderr, "[ERROR] defining a global variable with a non const expr is not allowed")
        fmt.Fprintln(os.Stderr, "\t" + d.Value.At())
        os.Exit(1)
    }

    if c,ok := d.Value.(*ast.FnCall); ok {
        if types.IsBigStruct(c.F.GetRetType()) {
            file.WriteString(fmt.Sprintf("lea rdi, [%s]\n", d.V.Addr(0)))
        }
    }

    GenExpr(file, d.Value)
    if !types.IsBigStruct(d.Value.GetType()) {
        vars.VarSetExpr(file, d.V)
    }
}

func GenDefConst(file *os.File, d *ast.DefConst) {
    val := cmpTime.ConstEval(d.Value)

    if val.Type == token.Unknown {
        fmt.Fprintln(os.Stderr, "[ERROR] cannot evaluate expr at compile time (not const)")
        fmt.Fprintln(os.Stderr, "\t" + d.Value.At())
        os.Exit(1)
    }

    d.C.Define(val)
}

func GenDefFn(file *os.File, d *ast.DefFn) {
    d.F.Define(file)

    regIdx := uint(0)

    if types.IsBigStruct(d.F.GetRetType()) {
        asm.MovDerefReg(file, fmt.Sprintf("rbp-%d", types.Ptr_Size), types.Ptr_Size, asm.RegDi)
        regIdx++
    }

    for _,a := range d.Args {
        if !types.IsBigStruct(a.V.GetType()) {
            i := types.RegCount(a.Type)

            if regIdx+i <= 6 {
                fn.DefArg(file, regIdx, a.V)
                regIdx += i
            }
        }
    }

    GenBlock(file, &d.Block)

    fn.End(file);
}
