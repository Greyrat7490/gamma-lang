package gen

import (
	"bufio"
	"fmt"
	"gamma/ast"
	"gamma/ast/identObj/vars"
	"gamma/cmpTime"
	"gamma/gen/asm/x86_64"
	"gamma/gen/asm/x86_64/conditions"
	"gamma/gen/asm/x86_64/loops"
	"gamma/types"
	"gamma/types/addr"
	"os"
	"reflect"
)

func GenDecl(file *bufio.Writer, d ast.Decl) {
    switch d := d.(type) {
    case *ast.Import:
        GenImport(file, d)

    case *ast.DefVar:
        GenDefVar(file, d)

    case *ast.DefFn:
        if d.IsGeneric {
            GenDefGenFn(file, d)
        } else {
            GenDefFn(file, d)
        }

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

func GenDefGenFn(file *bufio.Writer, d *ast.DefFn) {
    genType := d.F.GetGeneric()
    for _,t := range genType.UsedTypes {
        genType.CurUsedType = t
        GenDefFn(file, d)
    }
}


func GenDefFn(file *bufio.Writer, d *ast.DefFn) {
    argsSize := d.F.Scope.ArgsSize()
    innersize := d.F.Scope.SetLocalVarOffsets()
    framesize := argsSize + innersize
    if types.IsBigStruct(d.F.GetRetType()) {
        framesize += types.Ptr_Size
    }

    Define(file, d.F, framesize)

    regIdx := uint(0)
    argsFromStackOffset := uint(8)
    regArgsOffset := innersize

    if types.IsBigStruct(d.F.GetRetType()) {
        regArgsOffset += types.Ptr_Size
        addr := addr.Addr{ BaseAddr: "rbp", Offset: -int64(regArgsOffset) }
        asm.MovDerefReg(file, addr, types.Ptr_Size, asm.RegDi)
        d.F.SetRetAddr(addr)
        regIdx++
    }

    for _,a := range d.Args {
        if v,ok := a.V.(*vars.LocalVar); ok {
            if types.IsBigStruct(v.GetType()) {
                v.SetOffset(argsFromStackOffset, true) 
                argsFromStackOffset += v.GetType().Size()
            }
        }
    }

    for _,a := range d.Args {
        if v,ok := a.V.(*vars.LocalVar); ok {
            if !types.IsBigStruct(v.GetType()) {
                needed := types.RegCount(v.GetType())

                if regIdx + needed <= 6 {
                    v.SetOffset(regArgsOffset, false) 
                    DefArg(file, regIdx, v)
                    regIdx += needed
                    regArgsOffset += v.GetType().Size() 
                } else {
                    v.SetOffset(argsFromStackOffset, true) 
                    argsFromStackOffset += v.GetType().Size()
                }
            }
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] (internal) expected arg to be local var")
            os.Exit(1)
        }
    }

    GenBlock(file, &d.Block)

    if d.F.GetRetType() == nil {
        FnEnd(file);
    }

    cond.ResetCount()
    loops.ResetCount()
}
