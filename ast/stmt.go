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
    stmt()  // to differenciate OpStmt from OpDecl and OpExpr
}

type OpDeclStmt struct {
    Decl OpDecl
}

type OpExprStmt struct {
    Expr OpExpr
}

type OpAssignVar struct {
    Varname token.Token
    Value token.Token
}

type OpBlock struct {
    Stmts []OpStmt
}


func (o OpBlock)    stmt() {}
func (o OpDeclStmt) stmt() {}
func (o OpExprStmt) stmt() {}
func (o OpAssignVar) stmt() {}

func (o OpAssignVar) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("OP_ASSIGN: %s %s\n", o.Varname.Str, o.Value.Str)
}
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

func (o OpAssignVar) Compile(asm *os.File) {
    vars.Assign(asm, o.Varname, o.Value)
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
