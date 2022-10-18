package cmpTime

import (
    "gamma/ast"
    "gamma/cmpTime/constVal"
)

func EvalStmt(s ast.Stmt) constVal.ConstVal {
    switch s := s.(type) {
    case *ast.Ret:
        return ConstEval(s.RetExpr)
    default:
        return nil
    }
}

func evalRet(ret *ast.Ret) constVal.ConstVal {
    return ConstEval(ret.RetExpr)
}
