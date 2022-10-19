package cmpTime

import (
    "gamma/ast"
    "gamma/cmpTime/constVal"
)

var through bool = false

func EvalStmt(s ast.Stmt) constVal.ConstVal {
    switch s := s.(type) {
    case *ast.Ret:
        return ConstEval(s.RetExpr)
    case *ast.Block:
        return evalBlock(s)
    case *ast.If:
        return evalIf(s)
    case *ast.Switch:
        return evalSwitch(s)
    case *ast.Assign:
        evalAssign(s)
        return nil
    case *ast.Through:
        through = true
        return nil
    case *ast.DeclStmt:
        evalDecl(s.Decl)
        return nil
    default:
        return nil
    }
}


func evalBlock(s *ast.Block) constVal.ConstVal {
    return evalStmts(s.Stmts)
}

func evalRet(s *ast.Ret) constVal.ConstVal {
    return ConstEval(s.RetExpr)
}

func evalIf(s *ast.If) constVal.ConstVal {
    if cond,ok := ConstEval(s.Cond).(*constVal.BoolConst); ok {
        if bool(*cond) {
            return evalBlock(&s.Block)
        } else {
            if s.Elif != nil {
                if cond,ok := ConstEval(s.Elif.Cond).(*constVal.BoolConst); ok {
                    if bool(*cond) {
                        return evalBlock(&s.Elif.Block)
                    } else {
                        return evalBlock(&s.Elif.Else.Block)
                    }
                }
            } else {
                return evalBlock(&s.Else.Block)
            }
        }
    }

    return nil
}

func evalSwitch(s *ast.Switch) constVal.ConstVal {
    for i,c := range s.Cases {
        if c.Cond == nil {
            return evalStmts(c.Stmts)
        }

        if cond := ConstEval(c.Cond); cond != nil {
            if val,ok := cond.(*constVal.BoolConst); ok && bool(*val) {
                res := evalStmts(c.Stmts)

                if res == nil && through {
                    through = false
                    return evalStmts(s.Cases[i+1].Stmts)
                }

                return res
            }
        } else {
            return nil
        }
    }

    return nil
}

func evalStmts(stmts []ast.Stmt) constVal.ConstVal {
    for _,s := range stmts {
        if res := EvalStmt(s); res != nil {
            return res
        }
    }

    return nil
}

func evalAssign(s *ast.Assign) {
    if ident,ok := s.Dest.(*ast.Ident); ok {
        if val := ConstEval(s.Value); val != nil {
            setConst(ident.Name, s.Pos, val)
        }
    } else {
        /*
        // TODO only in const funcs
        fmt.Fprintf(os.Stderr, "[ERROR] only assigning to ident is allowed yet (but got %v)\n", reflect.TypeOf(s.Dest))
        fmt.Fprintln(os.Stderr, "\t" + s.At())
        os.Exit(1)
        */
    }
}
