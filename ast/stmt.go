package ast

import (
    "os"
    "fmt"
    "strings"
    "gorec/vars"
    "gorec/token"
    "gorec/loops"
    "gorec/conditions"
)

type OpStmt interface {
    Op
    Compile(asm *os.File)
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
    Varname token.Token
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


func (o *OpAssignVar) Compile(asm *os.File) {
    if l, ok := o.Value.(*LitExpr); ok {
        vars.DefineByValue(asm, o.Varname, l.Val)
    } else if ident, ok := o.Value.(*IdentExpr); ok {
        vars.DefineByVar(asm, o.Varname, ident.Ident)
    } else {
        o.Value.Compile(asm)
        vars.AssignByReg(asm, o.Varname, "rax")
    }
}

func (o *OpBlock) Compile(asm *os.File) {
    for _, op := range o.Stmts {
        op.Compile(asm)
    }
}

func (o *IfStmt) Compile(asm *os.File) {
    if l, ok := o.Cond.(*LitExpr); ok {
        if l.Val.Str == "true" {
            o.Block.Compile(asm)
        }
    } else if ident, ok := o.Cond.(*IdentExpr); ok {
        count := cond.IfIdent(asm, ident.Ident)
        o.Block.Compile(asm)
        cond.IfEnd(asm, count)
    } else {
        o.Cond.Compile(asm)
        count := cond.IfReg(asm, "rax")

        o.Block.Compile(asm)
        cond.IfEnd(asm, count)
    }
}

func (o *IfElseStmt) Compile(asm *os.File) {
    if l, ok := o.If.Cond.(*LitExpr); ok {
        if l.Val.Str == "true" {
            o.If.Block.Compile(asm)
        } else {
            o.Block.Compile(asm)
        }
    } else if ident, ok := o.If.Cond.(*IdentExpr); ok {
        count := cond.IfIdent(asm, ident.Ident)

        o.If.Block.Compile(asm)

        cond.ElseStart(asm, count)
        o.Block.Compile(asm)

        cond.ElseEnd(asm, count)
    } else {
        o.If.Cond.Compile(asm)
        count := cond.IfReg(asm, "rax")

        o.If.Block.Compile(asm)

        cond.ElseStart(asm, count)
        o.Block.Compile(asm)

        cond.ElseEnd(asm, count)
    }
}

func (o *WhileStmt) Compile(asm *os.File) {
    if o.InitVal != nil {
        o.Dec.Compile(asm)
        def := OpDefVar{ Varname: o.Dec.Varname, Value: o.InitVal }
        def.Compile(asm)
    }

    if l, ok := o.Cond.(*LitExpr); ok {
        if l.Val.Str == "true" {
            count := loops.WhileStart(asm)
            o.Block.Compile(asm)
            loops.WhileEnd(asm, count)
        }
    } else if ident, ok := o.Cond.(*IdentExpr); ok {
        count := loops.WhileStart(asm)
        loops.WhileIdent(asm, ident.Ident)
        o.Block.Compile(asm)
        loops.WhileEnd(asm, count)
    } else {
        count := loops.WhileStart(asm)
        o.Cond.Compile(asm)
        loops.WhileReg(asm, "rax")

        o.Block.Compile(asm)
        loops.WhileEnd(asm, count)
    }

    if o.InitVal != nil {
        vars.Remove(o.Dec.Varname.Str)
    }
}

func (o *ForStmt) Compile(asm *os.File) {
    o.Dec.Compile(asm)
    def := OpDefVar{ Varname: o.Dec.Varname, Value: o.Start }
    def.Compile(asm)

    count := loops.ForStart(asm)
    if o.Limit != nil {
        cond := BinaryExpr{ Operator: token.Token{ Type: token.Lss }, OperandL: &IdentExpr{ Ident: o.Dec.Varname }, OperandR: o.Limit }
        cond.Compile(asm)
        loops.ForReg(asm, "rax")
    }

    o.Block.Compile(asm)
    loops.ForBlockEnd(asm, count)

    step := OpAssignVar{ Varname: o.Dec.Varname, Value: o.Step }
    step.Compile(asm)
    loops.ForEnd(asm, count)

    vars.Remove(o.Dec.Varname.Str)
}

func (o *BreakStmt) Compile(asm *os.File) {
    loops.Break(asm)
}

func (o *ContinueStmt) Compile(asm *os.File) {
    loops.Continue(asm)
}

func (o *OpExprStmt) Compile(asm *os.File) {
    o.Expr.Compile(asm)
}

func (o *OpDeclStmt) Compile(asm *os.File) {
    o.Decl.Compile(asm)
}

func (o *BadStmt) Compile(asm *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad statement")
    os.Exit(1)
}


func (o *OpAssignVar) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("OP_ASSIGN: %s\n", o.Varname.Str) + o.Value.Readable(indent+1)
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
