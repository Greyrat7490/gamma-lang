package ast

import (
    "os"
    "fmt"
    "strings"
)

func (o *DecVar) Readable(indent int) string {
    s  := strings.Repeat("   ", indent)
    s2 := strings.Repeat("   ", indent+1)

    return s + "DEC_VAR:\n" +
          s2 + fmt.Sprintf("%v(Name)\n", o.Name.Str) +
          s2 + fmt.Sprintf("%v(Typename)\n", o.Type)
}

func (o *DefVar) Readable(indent int) string {
    s  := strings.Repeat("   ", indent)
    s2 := strings.Repeat("   ", indent+1)

    res := s + "DEF_VAR:\n" +
        s2 + fmt.Sprintf("%v(Name)\n", o.Name.Str)

    if o.Type == nil {
        res += s2 + "infer type\n"
    } else {
        res += s2 + fmt.Sprintf("%v(Typename)\n", o.Type)
    }

    return res + o.Value.Readable(indent+1)
}

func (o *DefFn) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "DEF_FN:\n"

    s := ""
    for _,a := range o.Args {
        s += fmt.Sprintf("%s(Name) %v(Typename), ", a.Name.Str, a.Type)
    }
    if len(s) > 0 { s = s[:len(s)-2] }

    res += fmt.Sprintf("%s%s(%v) [%s]\n", strings.Repeat("   ", indent+1), o.FnName.Str, o.FnName.Type, s) +
        o.Block.Readable(indent+2)

    return res
}
func (o *BadDecl) Readable(indent int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] bad declaration")
    os.Exit(1)
    return ""
}


func (o *Lit) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(%v)\n", o.Val.Str, o.Type)
}

func (o *Ident) Readable(indent int) string {
    return strings.Repeat("   ", indent) + o.Ident.Str + "(Name)\n"
}

func (o *FnCall) Readable(indent int) string {
    s  := strings.Repeat("   ", indent)
    s2 := strings.Repeat("   ", indent+1)

    res := s + "CALL_FN:\n" +
          s2 + o.Name.Str + "(Name)\n"

    for _,e := range o.Values {
        res += e.Readable(indent+1)
    }

    return res
}

func (o *Unary) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return fmt.Sprintf("%sUNARY:\n%s%s(%v)\n", s, s2, o.Operator.Str, o.Operator.Type) +
        o.Operand.Readable(indent+1)
}

func (o *Binary) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return s + "BINARY:\n" +
        o.OperandL.Readable(indent+1) +
        s2 + fmt.Sprintf("%s(%v)\n", o.Operator.Str, o.Operator.Type) +
        o.OperandR.Readable(indent+1)
}

func (o *Paren) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "PAREN:\n" + o.Expr.Readable(indent+1)
}

func (o *XCase) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    if o.Cond == nil {
        s += "XDEFAULT:\n" + o.Expr.Readable(indent+1)
    } else {
        s += "XCASE:\n" + o.Cond.Readable(indent+1) + o.Expr.Readable(indent+1)
    }

    return s
}

func (o *XSwitch) Readable(indent int) string {
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
