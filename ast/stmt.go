package ast

type OpStmt interface {
    Op
    stmt()  // to differenciate OpStmt from OpDecl and OpExpr
}
