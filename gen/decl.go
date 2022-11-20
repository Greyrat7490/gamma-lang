package gen

import (
    "os"
    "fmt"
    "bufio"
    "reflect"
    "gamma/ast"
    "gamma/cmpTime"
    "gamma/gen/asm/x86_64"
    "gamma/gen/asm/x86_64/conditions"
    "gamma/gen/asm/x86_64/loops"
    "gamma/types"
    "gamma/types/addr"
)

func GenDecl(file *bufio.Writer, d ast.Decl) {
    switch d := d.(type) {
    case *ast.Import:
        GenImport(file, d)

    case *ast.DefVar:
        GenDefVar(file, d)

    case *ast.DefFn:
        GenDefFn(file, d)

    case *ast.DefStruct, *ast.DecVar, *ast.DefConst:
        // nothing to generate

    case *ast.BadDecl:
        fmt.Fprintln(os.Stderr, "[ERROR] bad declaration")
        os.Exit(1)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] GenDecl for %v is not implemente yet\n", reflect.TypeOf(d))
        os.Exit(1)
    }
}

func GenImport(file *bufio.Writer, d *ast.Import) {
    for _, d := range d.Decls {
        GenDecl(file, d)
    }
}

func GenDefVar(file *bufio.Writer, d *ast.DefVar) {
    if val := cmpTime.ConstEval(d.Value); val != nil {
        VarDefVal(file, d.V, val)
    } else {
        VarDefExpr(file, d.V, d.Value)
    }
}

func GenDefFn(file *bufio.Writer, d *ast.DefFn) {
    Define(file, d.F)

    regIdx := uint(0)

    if types.IsBigStruct(d.F.GetRetType()) {
        asm.MovDerefReg(file, addr.Addr{ BaseAddr: "rbp", Offset: -int64(types.Ptr_Size) }, types.Ptr_Size, asm.RegDi)
        regIdx++
    }

    for _,a := range d.Args {
        if !types.IsBigStruct(a.V.GetType()) {
            i := types.RegCount(a.Type)

            if regIdx+i <= 6 {
                DefArg(file, regIdx, a.V)
                regIdx += i
            }
        }
    }

    GenBlock(file, &d.Block)

    if d.F.GetRetType() == nil {
        FnEnd(file);
    }

    cond.ResetCount()
    loops.ResetCount()
}
