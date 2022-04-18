package ast

import (
    "fmt"
    "gorec/token"
    "gorec/vars"
    "os"
    "strings"
)

type OpStmt interface {
    Op
    Compile(asm *os.File)
    stmt()  // to differenciate OpStmt from OpDecl
}

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
    Stmts []OpStmt
}


func (o *OpBlock)     stmt() {}
func (o *OpDeclStmt)  stmt() {}
func (o *OpExprStmt)  stmt() {}
func (o *OpAssignVar) stmt() {}

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
func (o *OpExprStmt) Readable(indent int) string {
    return o.Expr.Readable(indent)
}
func (o *OpDeclStmt) Readable(indent int) string {
    return o.Decl.Readable(indent)
}

func (o *OpAssignVar) Compile(asm *os.File) {
    fmt.Printf("assign: %s = %s\n", o.Varname.Str, o.Value.GetValue().Str)
    vars.Assign(asm, o.Varname, o.Value.GetValue())
    o.Value.Compile(asm, o.Varname)
}
func (o *OpBlock) Compile(asm *os.File) {
    for _, op := range o.Stmts {
        op.Compile(asm);
    }
}
func (o *OpExprStmt) Compile(asm *os.File) {
    o.Expr.Compile(asm, token.Token{})
}
func (o *OpDeclStmt) Compile(asm *os.File) {
    o.Decl.Compile(asm)
}
