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

type OpStmt interface {
    Op
    Compile(file *os.File)
    stmt()  // to differenciate OpStmt from OpDecl
}

type BadStmt struct {}

type OpDeclStmt struct {
    Decl OpDecl
}

type OpExprStmt struct {
    Expr OpExpr
}

type OpAssignVar struct {
    Pos token.Pos
    Dest OpExpr
    Value OpExpr
}

type OpBlock struct {
    BraceLPos token.Pos
    Stmts []OpStmt
    BraceRPos token.Pos
}

type IfStmt struct {
    Pos token.Pos
    Cond OpExpr
    Block OpBlock
    Else *ElseStmt // only one of these is set
    Elif *ElifStmt
}

type ElifStmt IfStmt

type ElseStmt struct {
    ElsePos token.Pos
    Block OpBlock
}

type SwitchStmt struct {
    Pos token.Pos
    Cases []CaseStmt
}

type CaseStmt struct {
    Cond OpExpr         // nil -> default
    ColonPos token.Pos
    Stmts []OpStmt
}

type ThroughStmt struct {
    Pos token.Pos
}

type WhileStmt struct {
    WhilePos token.Pos
    Cond OpExpr
    Dec OpDecVar
    InitVal OpExpr
    Block OpBlock
}

type ForStmt struct {
    ForPos token.Pos
    Dec OpDecVar
    Limit OpExpr
    Start OpExpr
    Step OpExpr
    Block OpBlock
}

type BreakStmt struct {
    Pos token.Pos
}

type ContinueStmt struct {
    Pos token.Pos
}


func (o *BadStmt)      stmt() {}
func (o *IfStmt)       stmt() {}
func (o *ElseStmt)     stmt() {}
func (o *ElifStmt)     stmt() {}
func (o *SwitchStmt)   stmt() {}
func (o *ThroughStmt)  stmt() {}
func (o *CaseStmt)     stmt() {}
func (o *ForStmt)      stmt() {}
func (o *WhileStmt)    stmt() {}
func (o *BreakStmt)    stmt() {}
func (o *ContinueStmt) stmt() {}
func (o *OpBlock)      stmt() {}
func (o *OpDeclStmt)   stmt() {}
func (o *OpExprStmt)   stmt() {}
func (o *OpAssignVar)  stmt() {}


func (o *OpAssignVar) Compile(file *os.File) {
    o.typeCheck()

    size := o.Dest.GetType().Size()

    switch dest := o.Dest.(type) {
    case *UnaryExpr:
        dest.Compile(file)

        switch e := o.Value.(type) {
        case *LitExpr:
            vars.DerefSetVal(file, e.Val, size)

        case *IdentExpr:
            vars.DerefSetVar(file, e.Ident)

        case *UnaryExpr:
            file.WriteString(asm.MovRegReg(asm.RegD, asm.RegA, size))
            o.Value.Compile(file)
            if e.Operator.Type == token.Mul {
                file.WriteString(asm.DerefRax(size))
            }
            file.WriteString(asm.MovDerefReg("rdx", size, asm.RegA))

        default:
            file.WriteString(asm.MovRegReg(asm.RegD, asm.RegA, size))
            o.Value.Compile(file)
            file.WriteString(asm.MovDerefReg("rdx", size, asm.RegA))
        }

    case *IdentExpr:
        v := vars.GetVar(dest.Ident.Str)
        if v == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] variable %s is not declared\n", dest.Ident.Str)
            fmt.Fprintln(os.Stderr, "\t" + dest.Ident.At())
            os.Exit(1)
        }

        switch e := o.Value.(type) {
        case *LitExpr:
            v.SetVal(file, e.Val)

        case *IdentExpr:
            v.SetVar(file, e.Ident)

        case *UnaryExpr:
            o.Value.Compile(file)
            if e.Operator.Type == token.Mul {
                file.WriteString(asm.DerefRax(size))
            }
            v.SetExpr(file)

        default:
            o.Value.Compile(file)
            v.SetExpr(file)
        }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] expected a variable or a derefenced pointer but got \"%t\"\n", dest)
        fmt.Fprintln(os.Stderr, "\t" + o.Pos.At())
        os.Exit(1)
    }
}

func (o *OpBlock) Compile(file *os.File) {
    for _, op := range o.Stmts {
        op.Compile(file)
    }
}

func (o *IfStmt) Compile(file *os.File) {
    o.typeCheck()

    vars.CreateScope()

    hasElse := o.Else != nil || o.Elif != nil

    var count uint = 0
    switch e := o.Cond.(type) {
    case *LitExpr:
        if e.Val.Str == "true" {
            o.Block.Compile(file)
            return
        } else if o.Else != nil {
            o.Else.Block.Compile(file)
            return
        }

    case *IdentExpr:
        count = cond.IfIdent(file, e.Ident, hasElse)
        o.Block.Compile(file)

    default:
        o.Cond.Compile(file)
        count = cond.IfExpr(file, hasElse)
        o.Block.Compile(file)
    }

    vars.RemoveScope()

    if hasElse {
        cond.ElseStart(file, count)

        if o.Else != nil {
            o.Else.Compile(file)
        } else {
            o.Elif.Compile(file)
        }

        cond.ElseEnd(file, count)
    }

    cond.IfEnd(file, count)
}

func (o *ElifStmt) Compile(file *os.File) {
    (*IfStmt)(o).Compile(file)
}

func (o *CaseStmt) Compile(file *os.File, switchCount uint) {
    vars.CreateScope()
    defer vars.RemoveScope()

    block := OpBlock{ Stmts: o.Stmts }

    if o.Cond == nil {
        cond.Default(file)
        block.Compile(file)
        return
    }

    cond.CaseStart(file)

    if i,ok := o.Cond.(*IdentExpr); ok {
        cond.CaseIdent(file, i.Ident)
    } else {
        o.Cond.Compile(file)
        cond.CaseExpr(file)
    }

    cond.CaseBody(file)
    block.Compile(file)
    cond.CaseBodyEnd(file, switchCount)
}

func (o *SwitchStmt) Compile(file *os.File) {
    o.typeCheck()
    count := cond.StartSwitch()

    // TODO: detect unreachable code and throw error
    // * a1 < but case 420 before 86
    // * cases with same cond

    for i := 0; i < len(o.Cases)-1; i++ {
        o.Cases[i].Compile(file, count)
    }
    cond.InLastCase()
    o.Cases[len(o.Cases)-1].Compile(file, count)

    cond.EndSwitch(file)
}

func (o *ThroughStmt) Compile(file *os.File) {
    cond.Through(file, o.Pos)
}

func (o *ElseStmt) Compile(file *os.File) {
    vars.CreateScope()
    o.Block.Compile(file)
    vars.RemoveScope()
}

func (o *WhileStmt) Compile(file *os.File) {
    vars.CreateScope()
    defer vars.RemoveScope()

    if o.InitVal != nil {
        o.Dec.Compile(file)
        def := OpDefVar{ Varname: o.Dec.Varname, Value: o.InitVal }
        def.Compile(file)
    }

    o.typeCheck()

    switch e := o.Cond.(type) {
    case *LitExpr:
        if e.Val.Str == "true" {
            count := loops.WhileStart(file)
            o.Block.Compile(file)
            loops.WhileEnd(file, count)
        }

    case *IdentExpr:
        count := loops.WhileStart(file)
        loops.WhileIdent(file, e.Ident)
        o.Block.Compile(file)
        loops.WhileEnd(file, count)

    default:
        count := loops.WhileStart(file)
        o.Cond.Compile(file)
        loops.WhileExpr(file)
        o.Block.Compile(file)
        loops.WhileEnd(file, count)
    }
}

func (o *ForStmt) Compile(file *os.File) {
    vars.CreateScope()
    defer vars.RemoveScope()

    o.Dec.Compile(file)
    def := OpDefVar{ Varname: o.Dec.Varname, Value: o.Start }
    def.Compile(file)

    o.typeCheck()

    count := loops.ForStart(file)
    if o.Limit != nil {
        cond := BinaryExpr{ Operator: token.Token{ Type: token.Lss }, OperandL: &IdentExpr{ Ident: o.Dec.Varname }, OperandR: o.Limit }
        cond.Compile(file)
        loops.ForExpr(file)
    }

    o.Block.Compile(file)
    loops.ForBlockEnd(file, count)

    step := OpAssignVar{ Dest: &IdentExpr{ Ident: o.Dec.Varname }, Value: o.Step }
    step.Compile(file)
    loops.ForEnd(file, count)
}

func (o *BreakStmt) Compile(file *os.File) {
    loops.Break(file)
}

func (o *ContinueStmt) Compile(file *os.File) {
    loops.Continue(file)
}

func (o *OpExprStmt) Compile(file *os.File) {
    o.Expr.Compile(file)
}

func (o *OpDeclStmt) Compile(file *os.File) {
    o.Decl.Compile(file)
}

func (o *BadStmt) Compile(file *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad statement")
    os.Exit(1)
}
