package cmpTime

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/cmpTime/constVal"
)

var curScope *scope = nil

type scope struct {
    consts map[string]constVal.ConstVal
    parent *scope
}

func startScope() {
    curScope = &scope{ parent: curScope, consts: make(map[string]constVal.ConstVal) }
}

func endScope() {
    curScope = curScope.parent
}

func inConstEnv() bool {
    return curScope != nil
}

func defConst(name string, pos token.Pos, val constVal.ConstVal) {
    if _,ok := curScope.consts[name]; ok {
        fmt.Fprintf(os.Stderr, "[ERROR] %s is already declared in this scope\n", name)
        fmt.Fprintln(os.Stderr, "\t", pos.At())
        os.Exit(1)
    }

    curScope.consts[name] = val
}

func setConst(name string, pos token.Pos, val constVal.ConstVal) {
    cur := curScope
    for cur != nil {
        if _,ok := cur.consts[name]; ok {
            cur.consts[name] = val
            return
        } else {
            cur = cur.parent
        }
    }

    fmt.Fprintf(os.Stderr, "[ERROR] %s is not declared\n", name)
    fmt.Fprintln(os.Stderr, "\t", pos.At())
    os.Exit(1)
}

func getVal(name string, pos token.Pos) (constVal.ConstVal, bool) {
    cur := curScope
    for cur != nil {
        if c,ok := cur.consts[name]; ok {
            return c, true
        } else {
            cur = cur.parent
        }
    }

    return nil, false
}
