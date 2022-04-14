package ast

import (
    "os"
    "strings"
)

type OpStmt interface {
    Op
    stmt()  // to differenciate OpStmt from OpDecl and OpExpr
}

type OpDeclStmt struct {
    Decl OpDecl
}

type OpExprStmt struct {
    Expr OpExpr
}

type OpBlock struct {
    Stmts []OpStmt
}


func (o OpBlock)    stmt() {}
func (o OpDeclStmt) stmt() {}
func (o OpExprStmt) stmt() {}


func (o OpBlock) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "OP_BLOCK:\n"
    for _, op := range o.Stmts {
        res += op.Readable(indent+1)
    }

    return res
}
func (o OpExprStmt) Readable(indent int) string {
    return o.Expr.Readable(indent)
}
func (o OpDeclStmt) Readable(indent int) string {
    return o.Decl.Readable(indent)
}


func (o OpBlock) Compile(asm *os.File) {
    for _, op := range o.Stmts {
        op.Compile(asm);
    }
}
func (o OpExprStmt) Compile(asm *os.File) {
    o.Expr.Compile(asm)
}
func (o OpDeclStmt) Compile(asm *os.File) {
    o.Decl.Compile(asm)
}
