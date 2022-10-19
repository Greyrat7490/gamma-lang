package cmpTime

import (
    "gamma/ast"
    "gamma/cmpTime/constVal"
)

var funcs map[string]constFunc = make(map[string]constFunc)

type constFunc struct {
    fn ast.DefFn
}

func (c constFunc) eval() constVal.ConstVal {
    return evalBlock(&c.fn.Block)
}

func AddConstFunc(fn ast.DefFn) {
    funcs[fn.F.GetName()] = constFunc{ fn: fn }
}

func EvalFunc(name string) constVal.ConstVal {
    if f,ok := funcs[name]; ok {
        return f.eval()
    } else {
        return nil
    }
}