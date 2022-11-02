package ast

import (
    "os"
    "fmt"
    "strings"
    "gamma/token"
    "gamma/ast/identObj/func"
)

type Stmt interface {
    Node
    stmt()  // to distinguish Stmt from Decl and Expr
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

type Ret struct {
    F *fn.Func
    Pos token.Pos
    RetExpr Expr    // nil -> return nothing
}

func (o *Assign) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "ASSIGN:\n" +
        o.Dest.Readable(indent+1) +
        o.Value.Readable(indent+1)
}

func (o *Block) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "BLOCK:\n"
    for _, op := range o.Stmts {
        res += op.Readable(indent+1)
    }

    return res
}

func (o *If) Readable(indent int) string {
    s := strings.Repeat("   ", indent) + "IF:\n" +
        o.Cond.Readable(indent+1) +
        o.Block.Readable(indent+1)

    if o.Elif != nil {
        s += o.Elif.Readable(indent)
    } else if o.Else != nil {
        s += o.Else.Readable(indent)
    }

    return s
}

func (o *Elif) Readable(indent int) string {
    s := strings.Repeat("   ", indent) + "ELIF:\n" +
        o.Cond.Readable(indent+1) +
        o.Block.Readable(indent+1)

    if o.Elif != nil {
        s += o.Elif.Readable(indent)
    } else if o.Else != nil {
        s += o.Else.Readable(indent)
    }

    return s
}

func (o *Else) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "ELSE:\n" +
        o.Block.Readable(indent+1)
}

func (o *Case) Readable(indent int) string {
    var s string
    if o.Cond == nil {
        s = strings.Repeat("   ", indent) + "DEFAULT:\n"
    } else {
        s = strings.Repeat("   ", indent) + "CASE:\n" +
            o.Cond.Readable(indent+1)
    }

    for _,stmt := range o.Stmts {
        s += stmt.Readable(indent+1)
    }

    return s
}

func (o *Switch) Readable(indent int) string {
    s := strings.Repeat("   ", indent) + "SWITCH:\n"

    for _, c := range o.Cases {
        s += c.Readable(indent+1)
    }

    return s
}

func (o *Through) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "THROUGH\n"
}

func (o *While) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "WHILE:\n" +
        o.Cond.Readable(indent+1)
    if o.Def != nil {
        res += o.Def.Readable(indent+1)
    }
    res += o.Block.Readable(indent+1)

    return res
}

func (o *For) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "FOR:\n" +
        o.Def.Readable(indent+1)
    if o.Limit != nil {
        res += o.Limit.Readable(indent+1)
    }

    res += o.Step.Readable(indent+1) +
    o.Block.Readable(indent+1)

    return res
}

func (o *Break) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "BREAK\n"
}

func (o *Continue) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "CONTINUE\n"
}

func (o *Ret) Readable(indent int) string {
    if o.RetExpr != nil {
        return strings.Repeat("   ", indent) + "RET:\n" +
            o.RetExpr.Readable(indent+1)
    } else {
        return strings.Repeat("   ", indent) + "RET\n"
    }
}

func (o *ExprStmt) Readable(indent int) string {
    return o.Expr.Readable(indent)
}

func (o *DeclStmt) Readable(indent int) string {
    return o.Decl.Readable(indent)
}

func (o *BadStmt) Readable(indent int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] bad statement")
    os.Exit(1)
    return ""
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
func (s *Ret)      stmt() {}

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
func (s *Ret)      At() string { return s.Pos.At() }

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
func (s *Ret)      End() string { return s.Pos.At() }
