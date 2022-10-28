package cmpTime

import (
    "os"
    "fmt"
    "reflect"
    "gamma/ast"
    "gamma/types"
    "gamma/types/array"
    "gamma/cmpTime/constVal"
)

var through bool = false

func EvalStmt(s ast.Stmt) constVal.ConstVal {
    switch s := s.(type) {
    case *ast.Ret:
        return evalRet(s)
    case *ast.Block:
        return evalBlock(s)
    case *ast.If:
        return evalIf(s)
    case *ast.Switch:
        return evalSwitch(s)
    case *ast.For:
        return evalFor(s)
    case *ast.While:
        return evalWhile(s)
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
        fmt.Fprintf(os.Stderr, "[ERROR] EvalStmt for %v is not implemente yet\n", reflect.TypeOf(s))
        fmt.Fprintln(os.Stderr, "\t" + s.At())
        os.Exit(1)
        return nil
    }
}


func evalBlock(s *ast.Block) constVal.ConstVal {
    return evalStmts(s.Stmts)
}

func evalRet(s *ast.Ret) constVal.ConstVal {
    if c := ConstEval(s.RetExpr); c != nil {
        return c
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] ret expr is not const")
        fmt.Fprintln(os.Stderr, "\t" + s.At())
        os.Exit(1)
        return nil
    }
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
    if val := ConstEval(s.Value); val != nil {
        switch dst := s.Dest.(type) {
        case *ast.Ident:
            setVar(dst.Name, dst.Obj.Addr(0), dst.GetType(), s.Pos, val)

        case *ast.Indexed:
            if ident,ok := dst.ArrExpr.(*ast.Ident); ok {
                if idx,ok := ConstEvalUint(dst.Flatten()); ok {
                    if arrIdx,ok := getVal(ident.Name, ident.Pos).(*constVal.ArrConst); ok {
                        array.SetElem(uint64(*arrIdx), idx, val)
                    }
                } else {
                    if len(dst.Indices) == 1 {
                        fmt.Fprintln(os.Stderr, "[ERROR] cannot const eval index")
                    } else {
                        fmt.Fprintln(os.Stderr, "[ERROR] cannot const eval indices")
                    }
                    fmt.Fprintln(os.Stderr, "\t" + dst.At())
                    os.Exit(1)
                }

            } else {
                if addr,ok := ConstEval(dst.ArrExpr).(*constVal.PtrConst); ok {
                    fmt.Println(addr)
                }
            }

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] assigning to %v is not supported yet\n", reflect.TypeOf(s.Dest))
            fmt.Fprintln(os.Stderr, "\t" + s.At())
            os.Exit(1)
        }
    }
}

func evalFor(s *ast.For) constVal.ConstVal {
    evalDecl(&s.Def)

    assign := ast.Assign{
        Dest: &ast.Ident{ Obj: s.Def.V, Name: s.Def.V.GetName(), Pos: s.Def.V.GetPos() },
        Value: s.Step,
    }

    if s.Def.Type.GetKind() == types.Uint {
        if limit,ok := ConstEvalUint(s.Limit); ok {
            i := uint64(*getVal(s.Def.V.GetName(), s.Def.V.GetPos()).(*constVal.UintConst))

            for i < limit {
                evalBlock(&s.Block)

                evalAssign(&assign)
                i = uint64(*getVal(s.Def.V.GetName(), s.Def.V.GetPos()).(*constVal.UintConst))
            }
        }
    } else if s.Def.Type.GetKind() == types.Int {
        if limit,ok := ConstEvalInt(s.Limit); ok {
            i := int64(*getVal(s.Def.V.GetName(), s.Def.V.GetPos()).(*constVal.IntConst))

            for i < limit {
                evalBlock(&s.Block)

                evalAssign(&assign)
                i = int64(*getVal(s.Def.V.GetName(), s.Def.V.GetPos()).(*constVal.IntConst))
            }
        }
    }

    return nil
}

func evalWhile(s *ast.While) constVal.ConstVal {
    if s.Def != nil {
        evalDecl(s.Def)
    }

    cond := ConstEval(s.Cond).(*constVal.BoolConst)
    for cond != nil && bool(*cond) {
        evalBlock(&s.Block)
        cond = ConstEval(s.Cond).(*constVal.BoolConst)
    }

    return nil
}

func offsetAddr(addr string, offset uint) string {
    offset += getOffset(addr)
    return fmt.Sprintf("rbp-%d", offset)
}
