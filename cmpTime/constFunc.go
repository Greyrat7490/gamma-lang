package cmpTime

import (
    "os"
    "fmt"
    "gamma/ast"
    "gamma/token"
    "gamma/cmpTime/constVal"
)

var funcs map[string]constFunc = make(map[string]constFunc)

type constFunc struct {
    fn ast.DefFn
    // TODO save result (depending on args)
}

func (c constFunc) eval(args []constVal.ConstVal) constVal.ConstVal {
    framesize := c.fn.FnHead.F.Scope.ArgsSize() + c.fn.FnHead.F.Scope.GetInnerSize()
    startScope(framesize)
    defer endScope()

    for i,a := range c.fn.FnHead.Args {
        defVar(a.V.GetName(), a.V.Addr(), a.V.GetType(), a.TypePos, args[i])
    }

    return evalBlock(&c.fn.Block)
}

func AddConstFunc(fn ast.DefFn) {
    funcs[fn.FnHead.F.GetName()] = constFunc{ fn: fn }
}

func EvalFunc(name string, pos token.Pos, args []constVal.ConstVal) constVal.ConstVal {
    if f,ok := funcs[name]; ok {
        return f.eval(args)
    }

    if inConstEnv() {
        fmt.Fprintf(os.Stderr, "[ERROR] %s is not a const func\n", name)
        fmt.Fprintln(os.Stderr, "\t" + pos.At())
        os.Exit(1)
    }

    return nil
}
