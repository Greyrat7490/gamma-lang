package gen

import (
    "os"
    "fmt"
    "bufio"
    "reflect"
    "gamma/ast"
    "gamma/ast/identObj"
    "gamma/ast/identObj/vars"
    "gamma/cmpTime"
    "gamma/types"
    "gamma/types/addr"
    "gamma/gen/asm/x86_64"
    "gamma/gen/asm/x86_64/loops"
    "gamma/gen/asm/x86_64/vtable"
    "gamma/gen/asm/x86_64/conditions"
)

func GenDecl(file *bufio.Writer, d ast.Decl) {
    switch d := d.(type) {
    case *ast.Import:
        GenImport(file, d)

    case *ast.DefVar:
        GenDefVar(file, d)

    case *ast.DefFn:
        if d.FnHead.F.IsGeneric() {
            GenDefGenFn(file, d)
        } else {
            GenDefFn(file, d)
        }

    case *ast.Impl:
        GenImpl(file, d)

    case *ast.DefStruct, *ast.DecVar, *ast.DefConst, *ast.DefInterface, *ast.DefEnum:
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
    if v,ok := d.V.(*vars.LocalVar); ok {
        v.SetOffset(identObj.GetStackSize(), false)
        identObj.IncStackSize(v.GetType())
    }

    if val := cmpTime.ConstEval(d.Value); val != nil {
        VarDefVal(file, d.V, val)
    } else {
        VarDefExpr(file, d.V, d.Value)
    }
}

func GenDefGenFn(file *bufio.Writer, d *ast.DefFn) {
    for _,t := range d.FnHead.F.GetUsedInsetTypes() {
        d.FnHead.F.SetInsetType(t)
        GenDefFn(file, d)
    }
}


func GenDefFn(file *bufio.Writer, d *ast.DefFn) {
    argsSize := d.FnHead.F.Scope.ArgsSize()
    innersize := d.FnHead.F.Scope.GetInnerSize()
    framesize := argsSize + innersize
    if types.IsBigStruct(d.FnHead.F.GetRetType()) {
        framesize += types.Ptr_Size
    }

    Define(file, d.FnHead.F, framesize)

    regIdx := uint(0)
    argsFromStackOffset := uint(8)
    regArgsOffset := innersize

    if types.IsBigStruct(types.ResolveGeneric(d.FnHead.F.GetRetType())) {
        regArgsOffset += types.Ptr_Size
        addr := addr.Addr{ BaseAddr: "rbp", Offset: -int64(regArgsOffset) }
        asm.MovDerefReg(file, addr, types.Ptr_Size, asm.RegDi)
        d.FnHead.F.SetRetAddr(addr)
        regIdx++
    }

    for _,a := range d.FnHead.Args {
        if v,ok := a.V.(*vars.LocalVar); ok {
            t := types.ResolveGeneric(v.GetType())
            if types.IsBigStruct(t) {
                v.SetOffset(argsFromStackOffset, true) 
                argsFromStackOffset += v.GetType().Size()
            }
        }
    }

    for _,a := range d.FnHead.Args {
        if v,ok := a.V.(*vars.LocalVar); ok {
            t := types.ResolveGeneric(v.GetType())
            if !types.IsBigStruct(t) {
                needed := types.RegCount(t)

                if regIdx + needed <= 6 {
                    v.SetOffset(regArgsOffset, false) 
                    DefArg(file, regIdx, v, t)
                    regIdx += needed
                    regArgsOffset += t.Size() 
                } else {
                    v.SetOffset(argsFromStackOffset, true) 
                    argsFromStackOffset += t.Size()
                }
            }
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] (internal) expected arg to be local var")
            os.Exit(1)
        }
    }

    GenBlock(file, &d.Block)

    if d.FnHead.F.GetRetType() == nil {
        FnEnd(file);
    }

    cond.ResetCount()
    loops.ResetCount()
    identObj.ResetStackSize()
}

func createVTable(file *bufio.Writer, d *ast.Impl) {
    vtable.Create(d.Impl.GetDstType(), d.Impl.GetInterfaceType(), d.Impl.GetVTableFuncNames())
}

func GenImpl(file *bufio.Writer, d *ast.Impl) {
    if d.Impl.IsGeneric() {
        generic := d.Impl.GetGeneric()
        for _,t := range generic.UsedInsetTypes {
            d.Impl.SetInsetType(t)
            genImpl(file, d)
        }
    } else {
        genImpl(file, d)
    }
}

func genImpl(file *bufio.Writer, d *ast.Impl) {
    createVTable(file, d)
    for _,f := range d.FnDefs {
        GenDefFn(file, &f)
    }
}
