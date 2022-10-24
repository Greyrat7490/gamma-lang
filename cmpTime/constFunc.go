package cmpTime

import (
    "gamma/ast"
    "gamma/cmpTime/constVal"
)

var funcs map[string]constFunc = make(map[string]constFunc)

type constFunc struct {
    fn ast.DefFn
    // TODO save result (depending on args)
}

func (c constFunc) eval(args []constVal.ConstVal) constVal.ConstVal {
    startScope()
    initStack(c.fn.F.GetFrameSize())
    defer clearStack()
    defer endScope()

    for i,a := range c.fn.Args {
        defVar(a.V.GetName(), a.V.Addr(0), a.V.GetType(), a.TypePos, args[i])
    }

    return evalBlock(&c.fn.Block)
}

func AddConstFunc(fn ast.DefFn) {
    funcs[fn.F.GetName()] = constFunc{ fn: fn }
}

func EvalFunc(name string, args []constVal.ConstVal) constVal.ConstVal {
    if f,ok := funcs[name]; ok {
        return f.eval(args)
    } else {
        return nil
    }
}
