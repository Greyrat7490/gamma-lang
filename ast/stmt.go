package ast

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/asm/x86_64"
    "gorec/asm/x86_64/loops"
    "gorec/asm/x86_64/conditions"
    "gorec/ast/identObj/vars"
)

type Stmt interface {
    Node
    Compile(file *os.File)
    stmt()  // to distinguish Stmt from Decl
}

type BadStmt struct {}

type DeclStmt struct {
    Decl Decl
}

type ExprStmt struct {
    Expr Expr
}

type Assign struct {
    Pos token.Pos
    Dest Expr
    Value Expr
}

type Block struct {
    BraceLPos token.Pos
    Stmts []Stmt
    BraceRPos token.Pos
}

type If struct {
    Pos token.Pos
    Cond Expr
    Block Block
    Else *Else // only one of these is set
    Elif *Elif
}

type Elif If

type Else struct {
    ElsePos token.Pos
    Block Block
}

type Switch struct {
    BraceLPos token.Pos
    Cases []Case
    BraceRPos token.Pos
}

type Case struct {
    Cond Expr         // nil -> default
    ColonPos token.Pos
    Stmts []Stmt
}

type Through struct {
    Pos token.Pos
}

type While struct {
    WhilePos token.Pos
    Cond Expr
    Def *DefVar       // nil -> no iterator
    Block Block
}

type For struct {
    ForPos token.Pos
    Def DefVar
    Limit Expr
    Step Expr
    Block Block
}

type Break struct {
    Pos token.Pos
}

type Continue struct {
    Pos token.Pos
}


func (s *Assign) Compile(file *os.File) {
    s.typeCheck()

    size := s.Dest.GetType().Size()

    switch dest := s.Dest.(type) {
    case *Unary:
        dest.Operand.Compile(file)

        // compile time evaluation
        if val := s.Value.ConstEval(); val.Type != token.Unknown {
            vars.DerefSetVal(file, val, size)
            return
        }

        if e,ok := s.Value.(*Ident); ok {
            vars.DerefSetVar(file, e.Obj.(vars.Var))
            return
        }

        asm.MovRegReg(file, asm.RegD, asm.RegA, types.Ptr_Size)
        s.Value.Compile(file)

        if s.Dest.GetType().GetKind() == types.Str {
            asm.MovDerefReg(file, asm.GetReg(asm.RegD, types.Ptr_Size), types.Ptr_Size, asm.RegA)
            asm.MovDerefReg(file, asm.GetOffsetedReg(asm.RegD, types.Ptr_Size, types.Ptr_Size), types.I32_Size, asm.RegB)
        } else {
            asm.MovDerefReg(file, asm.GetReg(asm.RegD, types.Ptr_Size), size, asm.RegA)
        }

    case *Indexed:
        addr := dest.CompileToAddr(file)

        file.WriteString(fmt.Sprintf("lea rdx, [%s]\n", addr))
        s.Value.Compile(file)

        if s.Dest.GetType().GetKind() == types.Str {
            asm.MovDerefReg(file, asm.GetReg(asm.RegD, types.Ptr_Size), types.Ptr_Size, asm.RegA)
            asm.MovDerefReg(file, asm.GetOffsetedReg(asm.RegD, types.Ptr_Size, types.Ptr_Size), types.I32_Size, asm.RegB)
        } else {
            asm.MovDerefReg(file, asm.GetReg(asm.RegD, types.Ptr_Size), size, asm.RegA)
        }

    case *Ident:
        // compile time evaluation
        if val := s.Value.ConstEval(); val.Type != token.Unknown {
            vars.VarSetVal(file, dest.Obj.(vars.Var), val)
            return
        }

        if e,ok := s.Value.(*Ident); ok {
            vars.VarSetVar(file, dest.Obj.(vars.Var), e.Obj.(vars.Var))
            return
        }

        s.Value.Compile(file)
        vars.VarSetExpr(file, dest.Obj.(vars.Var))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] expected a variable or a dereferenced pointer but got \"%t\"\n", dest)
        fmt.Fprintln(os.Stderr, "\t" + s.Pos.At())
        os.Exit(1)
    }
}

func (s *Block) Compile(file *os.File) {
    for _, stmt := range s.Stmts {
        stmt.Compile(file)
    }
}

func (s *If) Compile(file *os.File) {
    s.typeCheck()

    // compile time evaluation
    if val := s.Cond.ConstEval(); val.Type != token.Unknown {
        if val.Str == "true" {
            s.Block.Compile(file)
        } else if s.Else != nil {
            s.Else.Block.Compile(file)
        }

        return
    }

    hasElse := s.Else != nil || s.Elif != nil

    var count uint = 0
    if ident,ok := s.Cond.(*Ident); ok {
        count = cond.IfVar(file, ident.Obj.Addr(0), hasElse)
    } else {
        s.Cond.Compile(file)
        count = cond.IfExpr(file, hasElse)
    }

    s.Block.Compile(file)

    if hasElse {
        cond.ElseStart(file, count)

        if s.Else != nil {
            s.Else.Compile(file)
        } else {
            s.Elif.Compile(file)
        }

        cond.ElseEnd(file, count)
    }

    cond.IfEnd(file, count)
}

func (s *Elif) Compile(file *os.File) {
    (*If)(s).Compile(file)
}

func (s *Case) Compile(file *os.File, switchCount uint) {
    block := Block{ Stmts: s.Stmts }

    cond.CaseStart(file)

    if s.Cond == nil {
        cond.CaseBody(file)
        block.Compile(file)
        return
    }

    // compile time evaluation
    if val := s.Cond.ConstEval(); val.Type != token.Unknown {
        if val.Str == "true" {
            cond.CaseBody(file)
            block.Compile(file)
            cond.CaseBodyEnd(file, switchCount)
        }

        return
    }

    if i,ok := s.Cond.(*Ident); ok {
        cond.CaseVar(file, i.Obj.Addr(0))
    } else {
        s.Cond.Compile(file)
        cond.CaseExpr(file)
    }

    cond.CaseBody(file)
    block.Compile(file)
    cond.CaseBodyEnd(file, switchCount)
}

func (s *Switch) Compile(file *os.File) {
    s.typeCheck()

    // compile time evaluation
    for _,c := range s.Cases {
        if c.Cond == nil {
            for _,s := range c.Stmts {
                s.Compile(file)
            }

            return
        }

        cond := c.Cond.ConstEval()

        if cond.Type == token.Boolean && cond.Str == "true" {
            for _,s := range c.Stmts {
                s.Compile(file)
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
        s.Cases[i].Compile(file, count)
    }
    cond.InLastCase()
    s.Cases[len(s.Cases)-1].Compile(file, count)

    cond.EndSwitch(file)
}

func (s *Through) Compile(file *os.File) {
    cond.Through(file, s.Pos)
}

func (s *Else) Compile(file *os.File) {
    s.Block.Compile(file)
}

func (s *While) Compile(file *os.File) {
    if s.Def != nil {
        s.Def.Compile(file)
    }

    s.typeCheck()

    // compile time evaluation
    if c := s.Cond.ConstEval(); c.Type != token.Unknown {
        if c.Str == "true" {
            count := loops.WhileStart(file)
            s.Block.Compile(file)
            loops.WhileEnd(file, count)
        }

        return
    }

    count := loops.WhileStart(file)
    if e,ok := s.Cond.(*Ident); ok {
        loops.WhileVar(file, e.Obj.Addr(0))
    } else {
        s.Cond.Compile(file)
        loops.WhileExpr(file)
    }

    s.Block.Compile(file)
    loops.WhileEnd(file, count)
}

func (s *For) Compile(file *os.File) {
    s.Def.Compile(file)

    s.typeCheck()

    count := loops.ForStart(file)
    if s.Limit != nil {
        cond := Binary{
            Operator: token.Token{ Type: token.Lss },
            OperandL: &Ident{ Obj: s.Def.V, Name: s.Def.V.GetName(), Pos: s.Def.V.GetPos() },
            OperandR: s.Limit,
        }
        cond.Compile(file)
        loops.ForExpr(file)
    }

    s.Block.Compile(file)
    loops.ForBlockEnd(file, count)

    step := Assign{
        Dest: &Ident{ Obj: s.Def.V, Name: s.Def.V.GetName(), Pos: s.Def.V.GetPos() },
        Value: s.Step,
    }
    step.Compile(file)
    loops.ForEnd(file, count)
}

func (s *Break) Compile(file *os.File) {
    loops.Break(file)
}

func (s *Continue) Compile(file *os.File) {
    loops.Continue(file)
}

func (s *ExprStmt) Compile(file *os.File) {
    s.Expr.Compile(file)
}

func (s *DeclStmt) Compile(file *os.File) {
    s.Decl.Compile(file)
}

func (s *BadStmt) Compile(file *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad statement")
    os.Exit(1)
}


func (s *BadStmt)  stmt() {}
func (s *DeclStmt) stmt() {}
func (s *ExprStmt) stmt() {}
func (s *Block)    stmt() {}
func (s *Assign)   stmt() {}
func (s *If)       stmt() {}
func (s *Else)     stmt() {}
func (s *Elif)     stmt() {}
func (s *Switch)   stmt() {}
func (s *Through)  stmt() {}
func (s *Case)     stmt() {}
func (s *For)      stmt() {}
func (s *While)    stmt() {}
func (s *Break)    stmt() {}
func (s *Continue) stmt() {}

func (s *BadStmt)  At() string { return "" }
func (s *DeclStmt) At() string { return s.Decl.At() }
func (s *ExprStmt) At() string { return s.Expr.At() }
func (s *Block)    At() string { return s.BraceLPos.At() }
func (s *Assign)   At() string { return s.Pos.At() }
func (s *If)       At() string { return s.Pos.At() }
func (s *Else)     At() string { return s.ElsePos.At() }
func (s *Elif)     At() string { return s.Pos.At() }
func (s *Switch)   At() string { return s.BraceLPos.At() }
func (s *Through)  At() string { return s.Pos.At() }
func (s *Case)     At() string { return s.ColonPos.At() }
func (s *For)      At() string { return s.ForPos.At() }
func (s *While)    At() string { return s.WhilePos.At() }
func (s *Break)    At() string { return s.Pos.At() }
func (s *Continue) At() string { return s.Pos.At() }

func (s *BadStmt)  End() string { return "" }
func (s *DeclStmt) End() string { return s.Decl.End() }
func (s *ExprStmt) End() string { return s.Expr.End() }
func (s *Block)    End() string { return s.BraceRPos.At() }
func (s *Assign)   End() string { return s.Value.End() }
func (s *If)       End() string { return s.Block.End() }
func (s *Else)     End() string { return s.Block.End() }
func (s *Elif)     End() string { return s.Block.End() }
func (s *Switch)   End() string { return s.BraceRPos.At() }
func (s *Through)  End() string { return s.Pos.At() }
func (s *Case)     End() string { return s.Stmts[len(s.Stmts)-1].End() }
func (s *For)      End() string { return s.Block.End() }
func (s *While)    End() string { return s.Block.End() }
func (s *Break)    End() string { return s.Pos.At() }
func (s *Continue) End() string { return s.Pos.At() }
