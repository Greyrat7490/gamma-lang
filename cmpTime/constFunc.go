package cmpTime

import (
    "gamma/ast"
    "gamma/cmpTime/constVal"
)

var funcs map[string]constFunc = make(map[string]constFunc)

type constFunc struct {
    fn ast.DefFn
}

func (c constFunc) eval(args []constVal.ConstVal) constVal.ConstVal {
    startScope()
    defer endScope()

    for i,a := range args {
        defConst(c.fn.Args[i].V.GetName(), c.fn.Args[i].TypePos, a)
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
