package ast

import (
    "os"
    "fmt"
    "strings"
    "gorec/vars"
    "gorec/token"
    "gorec/loops"
    "gorec/conditions"
    "gorec/asm/x86_64"
)

type OpStmt interface {
    Op
    Compile(file *os.File)
    stmt()  // to differenciate OpStmt from OpDecl and OpExpr
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
    IfPos token.Pos
    Cond OpExpr
    Block OpBlock
}

type IfElseStmt struct {
    If IfStmt
    ElsePos token.Pos
    Block OpBlock
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
func (o *IfElseStmt)   stmt() {}
func (o *ForStmt)      stmt() {}
func (o *WhileStmt)    stmt() {}
func (o *BreakStmt)    stmt() {}
func (o *ContinueStmt) stmt() {}
func (o *OpBlock)      stmt() {}
func (o *OpDeclStmt)   stmt() {}
func (o *OpExprStmt)   stmt() {}
func (o *OpAssignVar)  stmt() {}


func (o *OpAssignVar) Compile(file *os.File) {
    t1 := o.Dest.GetType()
    t2 := o.Value.GetType()
    if t1 != t2 {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot assign a type: %v with type: %v\n",  t1, t2)
        fmt.Fprintln(os.Stderr, "\t" + o.Pos.At())
        os.Exit(1)
    }

    size := t1.Size()

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
    vars.CreateScope()

    switch e := o.Cond.(type) {
    case *LitExpr:
        if e.Val.Str == "true" {
            o.Block.Compile(file)
        }

    case *IdentExpr:
        count := cond.IfIdent(file, e.Ident)
        o.Block.Compile(file)
        cond.IfEnd(file, count)

    default:
        o.Cond.Compile(file)
        count := cond.IfExpr(file)
        o.Block.Compile(file)
        cond.IfEnd(file, count)
    }

    vars.RemoveScope()
}

func (o *IfElseStmt) Compile(file *os.File) {
    switch e := o.If.Cond.(type) {
    case *LitExpr:
        vars.CreateScope()

        if e.Val.Str == "true" {
            o.If.Block.Compile(file)
        } else {
            o.Block.Compile(file)
        }

        vars.RemoveScope()

    case *IdentExpr:
        count := cond.IfElseIdent(file, e.Ident)

        vars.CreateScope()
        o.If.Block.Compile(file)
        vars.RemoveScope()

        cond.ElseStart(file, count)

        vars.CreateScope()
        o.Block.Compile(file)
        vars.RemoveScope()

        cond.IfElseEnd(file, count)

    default:
        o.If.Cond.Compile(file)
        count := cond.IfElseExpr(file)

        vars.CreateScope()
        o.If.Block.Compile(file)
        vars.RemoveScope()

        cond.ElseStart(file, count)

        vars.CreateScope()
        o.Block.Compile(file)
        vars.RemoveScope()

        cond.IfElseEnd(file, count)
    }
}

func (o *WhileStmt) Compile(file *os.File) {
    vars.CreateScope()

    if o.InitVal != nil {
        o.Dec.Compile(file)
        def := OpDefVar{ Varname: o.Dec.Varname, Value: o.InitVal }
        def.Compile(file)
    }

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

    vars.RemoveScope()
}

func (o *ForStmt) Compile(file *os.File) {
    vars.CreateScope()

    o.Dec.Compile(file)
    def := OpDefVar{ Varname: o.Dec.Varname, Value: o.Start }
    def.Compile(file)

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

    vars.RemoveScope()
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


func (o *OpAssignVar) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "OP_ASSIGN:\n" +
        o.Dest.Readable(indent+1) +
        o.Value.Readable(indent+1)
}

func (o *OpBlock) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "OP_BLOCK:\n"
    for _, op := range o.Stmts {
        res += op.Readable(indent+1)
    }

    return res
}

func (o *IfStmt) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "IF:\n" +
        o.Cond.Readable(indent+1) +
        o.Block.Readable(indent+1)
}

func (o *IfElseStmt) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "IF_ELSE:\n" +
        o.If.Readable(indent+1) +
        strings.Repeat("   ", indent+1) + "ELSE:\n" +
        o.Block.Readable(indent+2)
}

func (o *WhileStmt) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "WHILE:\n" +
        o.Cond.Readable(indent+1)
    if o.InitVal != nil {
        res += o.Dec.Readable(indent+1) +
        o.InitVal.Readable(indent+1)
    }
    res += o.Block.Readable(indent+1)

    return res
}

func (o *ForStmt) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "FOR:\n" +
        o.Dec.Readable(indent+1)
    if o.Limit != nil {
        res += o.Limit.Readable(indent+1)
    }

    res += o.Start.Readable(indent+1) +
    o.Step.Readable(indent+1) +
    o.Block.Readable(indent+1)

    return res
}

func (o *BreakStmt) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "BREAK\n"
}

func (o *ContinueStmt) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "CONTINUE\n"
}

func (o *OpExprStmt) Readable(indent int) string {
    return o.Expr.Readable(indent)
}

func (o *OpDeclStmt) Readable(indent int) string {
    return o.Decl.Readable(indent)
}

func (o *BadStmt) Readable(indent int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] bad statement")
    os.Exit(1)
    return ""
}
