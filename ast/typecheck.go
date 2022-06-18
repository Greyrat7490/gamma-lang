package ast

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/vars"
)

func (o *BadDecl)  typeCheck() {}
func (o *OpDecVar) typeCheck() {
    vars.Declare(o.Varname, o.Vartype)
}

func (o *OpDefFn) typeCheck() {
    vars.CreateScope()

    for _,a := range o.Args {
        a.typeCheck()
    }

    o.Block.typeCheck()
    vars.RemoveScope()
}

func (o *OpDefVar) typeCheck() {
    v := vars.GetVar(o.Varname.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", o.Varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + o.Varname.At())
        os.Exit(1)
    }

    t1 := v.GetType()
    t2 := o.Value.GetType()

    if t1 != t2 {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot define \"%s\" (type: %v) with type %v\n", o.Varname.Str, t1, t2)
        fmt.Fprintln(os.Stderr, "\t" + o.Varname.At())
        os.Exit(1)
    }
}


func (o *BadStmt)      typeCheck() {}
func (o *BreakStmt)    typeCheck() {}
func (o *ContinueStmt) typeCheck() {}
func (o *OpDeclStmt)   typeCheck() { o.Decl.typeCheck() }
func (o *OpExprStmt)   typeCheck() { o.Expr.typeCheck() }
func (o *OpAssignVar)  typeCheck() {
    t1 := o.Dest.GetType()
    t2 := o.Value.GetType()

    if t1 != t2 {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot assign a type: %v with type: %v\n",  t1, t2)
        fmt.Fprintln(os.Stderr, "\t" + o.Pos.At())
        os.Exit(1)
    }
}
func (o *OpBlock) typeCheck() {
    for _,o := range o.Stmts {
        o.typeCheck()
    }
}
func (o *IfStmt) typeCheck() {
    if t := o.Cond.GetType(); t.GetKind() != types.Bool {
        fmt.Fprintf(os.Stderr, "[ERROR] expected an bool as if condition but got %v\n", t)
        fmt.Fprintln(os.Stderr, "\t" + o.IfPos.At())
        os.Exit(1)
    }

    vars.CreateScope()
    o.Block.typeCheck()
    vars.RemoveScope()
}
func (o *IfElseStmt) typeCheck() {
    if t := o.If.Cond.GetType(); t.GetKind() != types.Bool {
        fmt.Fprintf(os.Stderr, "[ERROR] expected an bool as if condition but got %v\n", t)
        fmt.Fprintln(os.Stderr, "\t" + o.If.IfPos.At())
        os.Exit(1)
    }

    vars.CreateScope()
    o.If.Block.typeCheck()
    vars.RemoveScope()

    vars.CreateScope()
    o.Block.typeCheck()
    vars.RemoveScope()
}
func (o *ForStmt) typeCheck() {
    vars.CreateScope()
    vars.Declare(o.Dec.Varname, o.Dec.Vartype)

    t := o.Dec.Vartype

    if t2 := o.Start.GetType(); t != t2 {
        fmt.Fprintf(os.Stderr, "[ERROR] expected %v as for iterator start type but got %v\n", t, t2)
        fmt.Fprintln(os.Stderr, "\t" + o.ForPos.At())
        os.Exit(1)
    }

    if o.Limit != nil {
        if t2 := o.Limit.GetType(); t != t2 {
            fmt.Fprintf(os.Stderr, "[ERROR] expected %v as for iterator limit type but got %v\n", t, t2)
            fmt.Fprintln(os.Stderr, "\t" + o.ForPos.At())
            os.Exit(1)
        }
    }

    if t2 := o.Step.GetType(); t != t2 {
        fmt.Fprintf(os.Stderr, "[ERROR] expected %v as for iterator step type but got %v\n", t, t2)
        fmt.Fprintln(os.Stderr, "\t" + o.ForPos.At())
        os.Exit(1)
    }

    vars.RemoveScope()
}
func (o *WhileStmt) typeCheck() {
    if o.InitVal != nil {
        vars.CreateScope()
        vars.Declare(o.Dec.Varname, o.Dec.Vartype)
        t1 := o.Dec.Vartype
        t2 := o.InitVal.GetType()

        if t1 != t2 {
            fmt.Fprintf(os.Stderr, "[ERROR] expected %v as while iterator init type but got %v\n", t1, t2)
            fmt.Fprintln(os.Stderr, "\t" + o.WhilePos.At())
            os.Exit(1)
        }
    }

    if t := o.Cond.GetType(); t.GetKind() != types.Bool {
        fmt.Fprintf(os.Stderr, "[ERROR] expected an bool as while condition but got %v\n", t)
        fmt.Fprintln(os.Stderr, "\t" + o.WhilePos.At())
        os.Exit(1)
    }

    if o.InitVal != nil {
        vars.RemoveScope()
    }
}


func (o *BadExpr)    typeCheck() {}
func (o *LitExpr)    typeCheck() {}
func (o *ParenExpr)  typeCheck() {}
func (o *IdentExpr)  typeCheck() {}
func (o *UnaryExpr)  typeCheck() {
    if o.Operator.Type == token.Plus || o.Operator.Type == token.Minus {
        if t := o.Operand.GetType(); t.GetKind() != types.I32 {
            fmt.Fprintf(os.Stderr, "[ERROR] expected i32 after +/- unary op but got %v\n", t)
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }
    }
}
func (o *BinaryExpr) typeCheck() {
    t1 := o.OperandL.GetType()
    t2 := o.OperandR.GetType()

    if t1 != t2 {
        if (t1.GetKind() == types.Ptr && t2.GetKind() == types.I32) ||
           (t2.GetKind() == types.Ptr && t1.GetKind() == types.I32) {
            if o.Operator.Type != token.Plus && o.Operator.Type != token.Minus {
                fmt.Fprintf(os.Stderr, "[ERROR] only +/- operator are allowed for binary ops with %v and %v\n", t1, t2)
                fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
                os.Exit(1)
            }
        }

        fmt.Fprintf(os.Stderr, "[ERROR] binary operation has two diffrente types (left: %v right: %v)\n", t1, t2)
        fmt.Fprintln(os.Stderr, "\t(ptr +/- i32 is allowed)")
        fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
        os.Exit(1)
    }
}
func (o *OpFnCall) typeCheck() {
    // TODO
}


func (o *BadExpr)   GetType() types.Type { return nil }
func (o *OpFnCall)  GetType() types.Type { return nil }
func (o *LitExpr)   GetType() types.Type { return o.Type }
func (o *ParenExpr) GetType() types.Type { return o.Expr.GetType() }
func (o *IdentExpr) GetType() types.Type {
    v := vars.GetVar(o.Ident.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared)\n", o.Ident.Str)
        fmt.Fprintln(os.Stderr, "\t" + o.Ident.At())
        os.Exit(1)
    }

    return v.GetType()
}

func (o *UnaryExpr) GetType() types.Type {
    if o.Operator.Type == token.Amp {
        return types.PtrType{ BaseType: o.Operand.GetType() }
    }

    if o.Operator.Type == token.Mul {
        if ptr, ok := o.Operand.GetType().(types.PtrType); ok {
            return ptr.BaseType
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] you cannot deref this expre (expected a pointer/address)")
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }
    }

    return o.Operand.GetType()
}

func (o *BinaryExpr) GetType() types.Type {
    if  o.Operator.Type == token.Eql || o.Operator.Type == token.Neq ||
        o.Operator.Type == token.Grt || o.Operator.Type == token.Lss ||
        o.Operator.Type == token.Geq || o.Operator.Type == token.Leq {
        return types.BoolType{}
    }

    t := o.OperandL.GetType()
    if t == nil {
        return o.OperandR.GetType()
    }

    if other := o.OperandR.GetType(); other.GetKind() == types.Ptr {
        // check for cases like 420 + &v1
        if t.GetKind() == types.I32 {
            return other
        }

        // check for cases like ptr1 - ptr2
        if t.GetKind() == types.Ptr {
            return types.I32Type{}
        }
    }

    return t
}

