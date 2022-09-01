package gen

import (
    "os"
    "fmt"
    "reflect"
    "gamma/token"
    "gamma/types"
    "gamma/cmpTime"
    "gamma/ast"
    "gamma/ast/identObj/vars"
    "gamma/gen/asm/x86_64"
    "gamma/gen/asm/x86_64/loops"
    "gamma/gen/asm/x86_64/conditions"
)

func GenStmt(file *os.File, s ast.Stmt) {
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

func GenAssign(file *os.File, s *ast.Assign) {
    t := s.Dest.GetType()

    switch dest := s.Dest.(type) {
    case *ast.Indexed:
        AssignIndexed(file, t, dest, s.Value)

    case *ast.Field:
        AssignField(file, t, dest, s.Value)

    case *ast.Unary:
        AssignDeref(file, t, dest, s.Value)

    case *ast.Ident:
        if v,ok := dest.Obj.(vars.Var); ok {
            AssignVar(file, v, s.Value)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected identifier %s to be a variable but got %v\n", dest.Name, reflect.TypeOf(dest.Obj))
            fmt.Fprintln(os.Stderr, "\t" + dest.At())
            os.Exit(1)
        }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] expected a variable or a dereferenced pointer but got %v\n", reflect.TypeOf(dest))
        fmt.Fprintln(os.Stderr, "\t" + s.Pos.At())
        os.Exit(1)
    }
}

func GenBlock(file *os.File, s *ast.Block) {
    for _,stmt := range s.Stmts {
        GenStmt(file, stmt)
    }
}

func GenIf(file *os.File, s *ast.If) {
    if val := cmpTime.ConstEval(s.Cond); val.Type != token.Unknown {
        if val.Str == "true" {
            GenBlock(file, &s.Block)
        } else if s.Else != nil {
            GenBlock(file, &s.Else.Block)
        }

        return
    }

    hasElse := s.Else != nil || s.Elif != nil

    var count uint = 0
    if ident,ok := s.Cond.(*ast.Ident); ok {
        count = cond.IfVar(file, ident.Obj.Addr(0), hasElse)
    } else {
        GenExpr(file, s.Cond)
        count = cond.IfExpr(file, hasElse)
    }

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

func GenElif(file *os.File, s *ast.Elif) {
    GenIf(file, (*ast.If)(s))
}

func GenElse(file *os.File, s *ast.Else) {
    GenBlock(file, &s.Block)
}

func GenCase(file *os.File, s *ast.Case, switchCount uint) {
    cond.CaseStart(file)

    if s.Cond == nil {
        cond.CaseBody(file)
        for _,s := range s.Stmts {
            GenStmt(file, s)
        }
        return
    }

    if val := cmpTime.ConstEval(s.Cond); val.Type != token.Unknown {
        if val.Str == "true" {
            cond.CaseBody(file)
            for _,s := range s.Stmts {
                GenStmt(file, s)
            }
            cond.CaseBodyEnd(file, switchCount)
        }

        return
    }

    if i,ok := s.Cond.(*ast.Ident); ok {
        cond.CaseVar(file, i.Obj.Addr(0))
    } else {
        GenExpr(file, s.Cond)
        cond.CaseExpr(file)
    }

    cond.CaseBody(file)
    for _,s := range s.Stmts {
        GenStmt(file, s)
    }
    cond.CaseBodyEnd(file, switchCount)
}

func GenSwitch(file *os.File, s *ast.Switch) {
    for _,c := range s.Cases {
        if c.Cond == nil {
            for _,s := range c.Stmts {
                GenStmt(file, s)
            }

            return
        }

        cond := cmpTime.ConstEval(c.Cond)

        if cond.Type == token.Boolean && cond.Str == "true" {
            for _,s := range c.Stmts {
                GenStmt(file, s)
            }

            return
        } else if cond.Type == token.Unknown {
            break
        }
    }


    // TODO: detect unreachable code and throw error
    // * a1 < but case 420 before 86
    // * cases with same cond
    count := cond.StartSwitch()

    for i := 0; i < len(s.Cases)-1; i++ {
        GenCase(file, &s.Cases[i], count)
    }
    cond.InLastCase()
    GenCase(file, &s.Cases[len(s.Cases)-1], count)

    cond.EndSwitch(file)
}

func GenThrough(file *os.File, s *ast.Through) {
    cond.Through(file, s.Pos)
}

func GenWhile(file *os.File, s *ast.While) {
    if s.Def != nil {
        GenDefVar(file, s.Def)
    }

    if c := cmpTime.ConstEval(s.Cond); c.Type != token.Unknown {
        if c.Str == "true" {
            count := loops.WhileStart(file)
            GenBlock(file, &s.Block)
            loops.WhileEnd(file, count)
        }

        return
    }

    count := loops.WhileStart(file)
    if e,ok := s.Cond.(*ast.Ident); ok {
        loops.WhileVar(file, e.Obj.Addr(0))
    } else {
        GenExpr(file, s.Cond)
        loops.WhileExpr(file)
    }

    GenBlock(file, &s.Block)
    loops.WhileEnd(file, count)
}

func GenFor(file *os.File, s *ast.For) {
    GenDefVar(file, &s.Def)

    count := loops.ForStart(file)
    if s.Limit != nil {
        cond := ast.Binary{
            Operator: token.Token{ Type: token.Lss },
            OperandL: &ast.Ident{ Obj: s.Def.V, Name: s.Def.V.GetName(), Pos: s.Def.V.GetPos() },
            OperandR: s.Limit,
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

func GenBreak(file *os.File, s *ast.Break) {
    loops.Break(file)
}

func GenContinue(file *os.File, s *ast.Continue) {
    loops.Continue(file)
}

func GenRet(file *os.File, s *ast.Ret) {
    if s.RetExpr != nil {
        if types.IsBigStruct(s.RetExpr.GetType()) {
            asm.MovRegDeref(file, asm.RegC, fmt.Sprintf("rbp-%d", types.Ptr_Size), types.Ptr_Size)

            t := s.RetExpr.GetType().(types.StructType)

            if val := cmpTime.ConstEval(s.RetExpr); val.Type != token.Unknown {
                RetBigStructLit(file, t, val)
            } else if ident,ok := s.RetExpr.(*ast.Ident); ok {
                RetBigStructVar(file, t, ident.Obj.(vars.Var))
            } else {
                RetBigStructExpr(file, "rcx", s.RetExpr)
            }

            asm.MovRegReg(file, asm.RegA, asm.RegC, types.Ptr_Size)
        } else {
            GenExpr(file, s.RetExpr)
        }
    }
    FnEnd(file)
}
