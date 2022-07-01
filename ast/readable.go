package ast

import (
    "os"
    "fmt"
    "strings"
)

func (o *OpDecVar) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return fmt.Sprintf("%sOP_DEC_VAR:\n%s%s(%s) %v(Typename)\n", s, s2,
        o.Varname.Str, o.Varname.Type.Readable(),
        o.Vartype)
}

func (o *OpDefVar) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return fmt.Sprintf("%sOP_DEF_VAR:\n%s%s(%s)\n", s, s2,
        o.Varname.Str, o.Varname.Type.Readable()) + o.Value.Readable(indent+1)
}

func (o *OpDefFn) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "OP_DEF_FN:\n"

    s := ""
    for _,a := range o.Args {
        s += fmt.Sprintf("%s(Name) %v(Typename), ", a.Varname.Str, a.Vartype)
    }
    if len(s) > 0 { s = s[:len(s)-2] }

    res += fmt.Sprintf("%s%s(%s) [%s]\n", strings.Repeat("   ", indent+1), o.FnName.Str, o.FnName.Type.Readable(), s) +
        o.Block.Readable(indent+2)

    return res
}
func (o *BadDecl) Readable(indent int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] bad declaration")
    os.Exit(1)
    return ""
}


func (o *LitExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(%v)\n", o.Val.Str, o.Type)
}

func (o *IdentExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + o.Ident.Str + "(Name)\n"
}

func (o *OpFnCall) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    res := fmt.Sprintf("%sOP_CALL_FN:\n%s%s\n", s, s2, o.FnName.Str)
    for _, e := range o.Values {
        res += e.Readable(indent+1)
    }

    return res
}

func (o *UnaryExpr) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return fmt.Sprintf("%sOP_UNARY:\n%s%s(%s)\n", s, s2, o.Operator.Str, o.Operator.Type.Readable()) +
        o.Operand.Readable(indent+1)
}

func (o *BinaryExpr) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return s + "OP_BINARY:\n" +
        o.OperandL.Readable(indent+1) +
        s2 + fmt.Sprintf("%s(%s)\n", o.Operator.Str, o.Operator.Type.Readable()) +
        o.OperandR.Readable(indent+1)
}

func (o *ParenExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "PAREN:\n" + o.Expr.Readable(indent+1)
}

func (o *CaseExpr) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    if o.Cond == nil {
        s += "XDEFAULT:\n" + o.Expr.Readable(indent+1)
    } else {
        s += "XCASE:\n" + o.Cond.Readable(indent+1) + o.Expr.Readable(indent+1)
    }

    return s
}

func (o *SwitchExpr) Readable(indent int) string {
    s := strings.Repeat("   ", indent) + "XSWITCH:\n"

    for _, c := range o.Cases {
        s += c.Readable(indent+1)
    }

    return s
}

func (o *BadExpr) Readable(indent int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
    return ""
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

func (o *ElifStmt) Readable(indent int) string {
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

func (o *ElseStmt) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "ELSE:\n" +
        o.Block.Readable(indent+1)
}

func (o *CaseStmt) Readable(indent int) string {
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

func (o *SwitchStmt) Readable(indent int) string {
    s := strings.Repeat("   ", indent) + "COND-SWITCH:\n"

    for _, c := range o.Cases {
        s += c.Readable(indent+1)
    }

    return s
}

func (o *ThroughStmt) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "THROUGH\n"
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
