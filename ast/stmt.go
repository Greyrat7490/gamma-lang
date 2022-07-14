package ast

import (
    "os"
    "fmt"
    "gorec/vars"
    "gorec/loops"
    "gorec/token"
    "gorec/conditions"
    "gorec/asm/x86_64"
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
        dest.Compile(file)

        switch e := s.Value.(type) {
        case *Lit:
            vars.DerefSetVal(file, e.Val, size)

        case *Ident:
            if c := vars.GetConst(e.Ident.Str); c != nil {
                vars.DerefSetVal(file, c.Val, size)
            } else {
                vars.DerefSetVar(file, e.Ident)
            }

        case *Unary:
            file.WriteString(asm.MovRegReg(asm.RegD, asm.RegA, size))
            s.Value.Compile(file)
            if e.Operator.Type == token.Mul {
                file.WriteString(asm.DerefRax(size))
            }
            file.WriteString(asm.MovDerefReg("rdx", size, asm.RegA))

        default:
            file.WriteString(asm.MovRegReg(asm.RegD, asm.RegA, size))
            s.Value.Compile(file)
            file.WriteString(asm.MovDerefReg("rdx", size, asm.RegA))
        }

    case *Ident:
        if c := vars.GetConst(dest.Ident.Str); c != nil {
            fmt.Fprintf(os.Stderr, "[ERROR] you cannot change a const(%s)\n", dest.Ident.Str)
            fmt.Fprintln(os.Stderr, "\t" + s.At())
            os.Exit(1)
        }

        v := vars.GetVar(dest.Ident.Str)
        if v == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] variable %s is not declared\n", dest.Ident.Str)
            fmt.Fprintln(os.Stderr, "\t" + dest.Ident.At())
            os.Exit(1)
        }

        switch e := s.Value.(type) {
        case *Lit:
            v.SetVal(file, e.Val)

        case *Ident:
            v.SetVar(file, e.Ident)

        case *Unary:
            s.Value.Compile(file)
            if e.Operator.Type == token.Mul {
                file.WriteString(asm.DerefRax(size))
            }
            v.SetExpr(file)

        default:
            s.Value.Compile(file)
            v.SetExpr(file)
        }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] expected a variable or a derefenced pointer but got \"%t\"\n", dest)
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

    vars.CreateScope()

    hasElse := s.Else != nil || s.Elif != nil

    var count uint = 0
    switch c := s.Cond.(type) {
    case *Lit:
        if c.Val.Str == "true" {
            s.Block.Compile(file)
            return
        } else if s.Else != nil {
            s.Else.Block.Compile(file)
            return
        }

    case *Ident:
        if con := vars.GetConst(c.Ident.Str); con != nil {
            if con.Val.Str == "true" {
                s.Block.Compile(file)
                return
            } else if s.Else != nil {
                s.Else.Block.Compile(file)
                return
            }
        } else {
            count = cond.IfIdent(file, c.Ident, hasElse)
            s.Block.Compile(file)
        }

    default:
        s.Cond.Compile(file)
        count = cond.IfExpr(file, hasElse)
        s.Block.Compile(file)
    }

    vars.RemoveScope()

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
    vars.CreateScope()
    defer vars.RemoveScope()

    block := Block{ Stmts: s.Stmts }

    if s.Cond == nil {
        cond.Default(file)
        block.Compile(file)
        return
    }

    cond.CaseStart(file)

    if i,ok := s.Cond.(*Ident); ok {
        if c := vars.GetConst(i.Ident.Str); c != nil {
            // TODO
            os.Exit(1)
        }
        cond.CaseIdent(file, i.Ident)
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
    count := cond.StartSwitch()

    // TODO: detect unreachable code and throw error
    // * a1 < but case 420 before 86
    // * cases with same cond

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
    vars.CreateScope()
    s.Block.Compile(file)
    vars.RemoveScope()
}

func (s *While) Compile(file *os.File) {
    vars.CreateScope()
    defer vars.RemoveScope()

    if s.Def != nil {
        s.Def.Compile(file)
    }

    s.typeCheck()

    switch e := s.Cond.(type) {
    case *Lit:
        if e.Val.Str == "true" {
            count := loops.WhileStart(file)
            s.Block.Compile(file)
            loops.WhileEnd(file, count)
        }

    case *Ident:
        if c := vars.GetConst(e.Ident.Str); c != nil {
            if c.Val.Str == "true" {
                count := loops.WhileStart(file)
                s.Block.Compile(file)
                loops.WhileEnd(file, count)
            }
        } else {
            count := loops.WhileStart(file)
            loops.WhileIdent(file, e.Ident)
            s.Block.Compile(file)
            loops.WhileEnd(file, count)
        }


    default:
        count := loops.WhileStart(file)
        s.Cond.Compile(file)
        loops.WhileExpr(file)
        s.Block.Compile(file)
        loops.WhileEnd(file, count)
    }
}

func (s *For) Compile(file *os.File) {
    vars.CreateScope()
    defer vars.RemoveScope()

    s.Def.Compile(file)

    s.typeCheck()

    count := loops.ForStart(file)
    if s.Limit != nil {
        cond := Binary{ Operator: token.Token{ Type: token.Lss }, OperandL: &Ident{ Ident: s.Def.Name }, OperandR: s.Limit }
        cond.Compile(file)
        loops.ForExpr(file)
    }

    s.Block.Compile(file)
    loops.ForBlockEnd(file, count)

    step := Assign{ Dest: &Ident{ Ident: s.Def.Name }, Value: s.Step }
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
