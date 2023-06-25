package gen

import (
    "os"
    "fmt"
    "bufio"
    "reflect"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
    "gamma/cmpTime"
    "gamma/cmpTime/constVal"
    "gamma/ast"
    "gamma/ast/identObj"
    "gamma/ast/identObj/vars"
    "gamma/gen/asm/x86_64"
    "gamma/gen/asm/x86_64/loops"
    "gamma/gen/asm/x86_64/conditions"
)

func GenStmt(file *bufio.Writer, s ast.Stmt) {
    switch s := s.(type) {
    case *ast.Assign:
        GenAssign(file, s)

    case *ast.Block:
        GenBlock(file, s)

    case *ast.If:
        GenIf(file, s)
    case *ast.Else:
        GenElse(file, s)
    case *ast.Elif:
        GenElif(file, s)

    case *ast.Switch:
        GenSwitch(file, s)

    case *ast.Through:
        GenThrough(file, s)

    case *ast.For:
        GenFor(file, s)
    case *ast.While:
        GenWhile(file, s)

    case *ast.Break:
        GenBreak(file, s)
    case *ast.Continue:
        GenContinue(file, s)
    case *ast.Ret:
        GenRet(file, s)

    case *ast.DeclStmt:
        GenDecl(file, s.Decl)
    case *ast.ExprStmt:
        GenExpr(file, s.Expr)

    case *ast.Case:
        fmt.Fprintln(os.Stderr, "[ERROR] Cases outside of a switch are not allowed")
        fmt.Fprintln(os.Stderr, "\t" + s.At())
        os.Exit(1)
    case *ast.BadStmt:
        fmt.Fprintln(os.Stderr, "[ERROR] bad statement")
        fmt.Fprintln(os.Stderr, "\t" + s.At())
        os.Exit(1)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] GenStmt for %v is not implemente yet\n", reflect.TypeOf(s))
        os.Exit(1)
    }
}

func GenAssign(file *bufio.Writer, s *ast.Assign) {
    t := s.Dest.GetType()
    var addr addr.Addr

    if ident, ok := s.Dest.(*ast.Ident); ok {
        if v,ok := ident.Obj.(vars.Var); ok {
            addr = v.Addr()
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected identifier %s to be a variable but got %v\n", ident.Name, reflect.TypeOf(ident.Obj))
            fmt.Fprintln(os.Stderr, "\t" + ident.At())
            os.Exit(1)
        }
    } else {
        ExprAddrToReg(file, s.Dest, asm.RegC)
        addr = asm.RegAsAddr(asm.RegC)
    }

    if c := cmpTime.ConstEval(s.Value); c != nil {
        DerefSetVal(file, addr, t, c)
    } else if ident,ok := s.Value.(*ast.Ident); ok {
        if v,ok := ident.Obj.(vars.Var); ok {
            DerefSetVar(file, addr, v)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected identifier %s to be a variable but got %v\n", ident.Name, reflect.TypeOf(ident.Obj))
            fmt.Fprintln(os.Stderr, "\t" + ident.At())
            os.Exit(1)
        }
    } else {
        DerefSetExpr(file, addr, t, s.Value)
    }
}

func GenBlock(file *bufio.Writer, s *ast.Block) {
    for _,stmt := range s.Stmts {
        GenStmt(file, stmt)
    }
}

func GenIfCond(file *bufio.Writer, e ast.Expr, hasElse bool) uint {
    switch e := e.(type) {
    case *ast.Ident:
        return cond.IfVar(file, e.Obj.Addr(), hasElse)

    case *ast.Unwrap:
        idType := e.EnumType.IdType
        id := e.EnumType.GetID(e.ElemName.Str)

        ExprAddrToReg(file, e.SrcExpt, asm.RegD)
        asm.MovRegDeref(file, asm.RegA, asm.RegAsAddr(asm.RegD), idType.Size(), false)
        asm.Eql(file, asm.GetAnyReg(asm.RegA, idType.Size()), fmt.Sprint(id))

        count := cond.IfExpr(file, hasElse)

        if v,ok := e.Obj.(*vars.LocalVar); ok {
            v.SetOffset(identObj.GetStackSize(), false)
            identObj.IncStackSize(v.GetType())
            DerefSetDeref(file, v.Addr(), v.GetType(), asm.RegAsAddr(asm.RegD).Offseted(int64(idType.Size())))
        }

        return count

    default:
        GenExpr(file, e)
        return cond.IfExpr(file, hasElse)
    }
}

func GenIf(file *bufio.Writer, s *ast.If) {
    if val,ok := cmpTime.ConstEval(s.Cond).(*constVal.BoolConst); ok {
        if bool(*val) {
            GenBlock(file, &s.Block)
        } else if s.Else != nil {
            GenBlock(file, &s.Else.Block)
        }

        return
    }

    hasElse := s.Else != nil || s.Elif != nil

    count := GenIfCond(file, s.Cond, hasElse)

    GenBlock(file, &s.Block)

    if hasElse {
        cond.ElseStart(file, count)

        if s.Else != nil {
            GenElse(file, s.Else)
        } else {
            GenElif(file, s.Elif)
        }

        cond.ElseEnd(file, count)
    }

    cond.IfEnd(file, count)
}

func GenElif(file *bufio.Writer, s *ast.Elif) {
    GenIf(file, (*ast.If)(s))
}

func GenElse(file *bufio.Writer, s *ast.Else) {
    GenBlock(file, &s.Block)
}

func GenCase(file *bufio.Writer, s *ast.Case) {
    if s.Cond == nil {
        cond.CaseStart(file)
        cond.CaseBody(file)
        for _,s := range s.Stmts {
            GenStmt(file, s)
        }
        return
    }

    if val,ok := cmpTime.ConstEval(s.Cond).(*constVal.BoolConst); ok {
        if bool(*val) {
            cond.CaseStart(file)
            cond.CaseBody(file)
            for _,s := range s.Stmts {
                GenStmt(file, s)
            }
            cond.CaseBodyEnd(file)
        }

        return
    }

    cond.CaseStart(file)
    if i,ok := s.Cond.(*ast.Ident); ok {
        cond.CaseVar(file, i.Obj.Addr())
    } else {
        GenExpr(file, s.Cond)
        cond.CaseExpr(file)
    }

    cond.CaseBody(file)
    for _,s := range s.Stmts {
        GenStmt(file, s)
    }
    cond.CaseBodyEnd(file)
}

func GenSwitch(file *bufio.Writer, s *ast.Switch) {
    cond.StartSwitch()

    // TODO: detect unreachable code and throw error
    // * a1 < but case 420 before 86
    // * cases with same cond
    for i := 0; i < len(s.Cases)-1; i++ {
        GenCase(file, &s.Cases[i])
    }
    cond.InLastCase()
    GenCase(file, &s.Cases[len(s.Cases)-1])

    cond.EndSwitch(file)
}

func GenThrough(file *bufio.Writer, s *ast.Through) {
    cond.Through(file, s.Pos)
}

func GenWhile(file *bufio.Writer, s *ast.While) {
    if s.Def != nil {
        GenDefVar(file, s.Def)
    }

    if val,ok := cmpTime.ConstEval(s.Cond).(*constVal.BoolConst); ok {
        if bool(*val) {
            count := loops.WhileStart(file)
            GenBlock(file, &s.Block)
            loops.WhileEnd(file, count)
        }

        return
    }

    count := loops.WhileStart(file)
    if e,ok := s.Cond.(*ast.Ident); ok {
        loops.WhileVar(file, e.Obj.Addr())
    } else {
        GenExpr(file, s.Cond)
        loops.WhileExpr(file)
    }

    GenBlock(file, &s.Block)
    loops.WhileEnd(file, count)
}

func GenFor(file *bufio.Writer, s *ast.For) {
    GenDefVar(file, &s.Def)

    count := loops.ForStart(file)
    if s.Limit != nil {
        cond := ast.Binary{
            Operator: token.Token{ Type: token.Lss },
            OperandL: &ast.Ident{ Obj: s.Def.V, Name: s.Def.V.GetName(), Pos: s.Def.V.GetPos() },
            OperandR: s.Limit,
            Type: types.BoolType{},
        }
        GenBinary(file, &cond)
        loops.ForExpr(file)
    }

    GenBlock(file, &s.Block)
    loops.ForBlockEnd(file, count)

    step := ast.Assign{
        Dest: &ast.Ident{ Obj: s.Def.V, Name: s.Def.V.GetName(), Pos: s.Def.V.GetPos() },
        Value: s.Step,
    }
    GenAssign(file, &step)
    loops.ForEnd(file, count)
}

func GenBreak(file *bufio.Writer, s *ast.Break) {
    if loops.InLoop() {
        loops.Break(file)
    } else if cond.InSwitch() {
        cond.Break(file)
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] break can only be used inside of a switch or a loop")
        fmt.Fprintln(os.Stderr, "\t" + s.At())
        os.Exit(1)
    }
}

func GenContinue(file *bufio.Writer, s *ast.Continue) {
    loops.Continue(file)
}

func GenRet(file *bufio.Writer, s *ast.Ret) {
    if s.RetExpr != nil {
        t := s.F.GetRetType()

        if types.IsBigStruct(t) {
            asm.MovRegDeref(file,
                asm.RegC, s.F.GetRetAddr(),
                types.Ptr_Size, false)

            switch t := t.(type) {
            case types.StructType:
                if val := cmpTime.ConstEval(s.RetExpr); val != nil {
                    RetBigStructLit(file, t, *val.(*constVal.StructConst))
                } else if ident,ok := s.RetExpr.(*ast.Ident); ok {
                    RetBigStructVar(file, t, ident.Obj.(vars.Var))
                } else {
                    RetBigStructExpr(file, asm.RegAsAddr(asm.RegC), s.RetExpr)
                }
            case types.VecType, types.EnumType:
                // TODO: Lit and Var
                RetBigStructExpr(file, asm.RegAsAddr(asm.RegC), s.RetExpr)
            default:
                fmt.Fprintln(os.Stderr, "[ERROR] (internal) unreachable GenRet")
                os.Exit(1)
            }

            asm.MovRegReg(file, asm.RegA, asm.RegC, types.Ptr_Size)
        } else {
            GenExpr(file, s.RetExpr)
        }
    }
    FnEnd(file)
}
