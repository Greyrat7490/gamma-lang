package ast

import (
	"os"
	"fmt"
	"gorec/vars"
	"gorec/func"
	"gorec/token"
	"gorec/types"
)

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
func (o *IfStmt) typeCheck() {
    if t := o.Cond.GetType(); t.GetKind() != types.Bool {
        fmt.Fprintf(os.Stderr, "[ERROR] expected an bool as if condition but got %v\n", t)
        fmt.Fprintln(os.Stderr, "\t" + o.Pos.At())
        os.Exit(1)
    }
}

func (o *ForStmt) typeCheck() {
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
}

func (o *WhileStmt) typeCheck() {
    if o.InitVal != nil {
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

    if o.Operator.Type == token.And || o.Operator.Type == token.Or {
        if t1.GetKind() != types.Bool || t2.GetKind() != types.Bool {
            fmt.Fprintf(os.Stderr, "[ERROR] expected 2 bools for logic op \"%s\" but got %v and %v\n", o.Operator.Str, t1, t2)
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }
    } else {
        if t1 != t2 {
            if (t1.GetKind() == types.Ptr && t2.GetKind() == types.I32) ||
               (t2.GetKind() == types.Ptr && t1.GetKind() == types.I32) {
                if o.Operator.Type == token.Plus || o.Operator.Type == token.Minus {
                    return
                }

                fmt.Fprintf(os.Stderr, "[ERROR] only +/- operators are allowed for binary ops with %v and %v\n", t1, t2)
                fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
                os.Exit(1)
            }

            fmt.Fprintf(os.Stderr, "[ERROR] binary operation has two diffrente types (left: %v right: %v)\n", t1, t2)
            fmt.Fprintln(os.Stderr, "\t(ptr +/- i32 is allowed)")
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }
    }
}
func (o *OpFnCall) typeCheck() {
    f := fn.GetFn(o.FnName.Str)
    if f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" is not declared\n", o.FnName.Str)
        fmt.Fprintln(os.Stderr, "\t" + o.FnName.At())
        os.Exit(1)
    }

    if len(f.Args) != len(o.Values) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected %d args for function \"%s\" but got %d\n", len(f.Args), o.FnName.Str, len(o.Values))
        fmt.Fprintf(os.Stderr, "\texpected: %v\n", f.Args)
        fmt.Fprintf(os.Stderr, "\tgot:      %v\n", valuesToTypes(o.Values))
        fmt.Fprintln(os.Stderr, "\t" + o.FnName.At())
        os.Exit(1)
    }

    for i, t1 := range f.Args {
        t2 := o.Values[i].GetType()

        if t1 != t2 {
            fmt.Fprintf(os.Stderr, "[ERROR] expected %v as arg %d but got %v for function \"%s\"\n", t1, i, t2, o.FnName.Str)
            fmt.Fprintf(os.Stderr, "\texpected: %v\n", f.Args)
            fmt.Fprintf(os.Stderr, "\tgot:      %v\n", valuesToTypes(o.Values))
            fmt.Fprintln(os.Stderr, "\t" + o.FnName.At())
            os.Exit(1)
        }
    }
}

func valuesToTypes(values []OpExpr) (res []types.Type) {
    for _, v := range values {
        res = append(res, v.GetType())
    }

    return res
}

func (o *BadExpr)   GetType() types.Type { return nil }
func (o *OpFnCall)  GetType() types.Type { return nil }
func (o *LitExpr)   GetType() types.Type { return o.Type }
func (o *ParenExpr) GetType() types.Type { return o.Expr.GetType() }
func (o *IdentExpr) GetType() types.Type {
    v := vars.GetVar(o.Ident.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", o.Ident.Str)
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

