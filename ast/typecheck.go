package ast

import (
	"os"
	"fmt"
	"gorec/token"
	"gorec/types"
	"gorec/identObj/func"
)

func (o *DefVar) typeCheck() {
    t1 := o.Ident.V.GetType()
    t2 := o.Value.GetType()

    if !types.AreCompatible(t1, t2) {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot define \"%s\" (type: %v) with type %v\n", o.Ident.Ident.Str, t1, t2)
        fmt.Fprintln(os.Stderr, "\t" + o.Ident.Ident.At())
        os.Exit(1)
    }
}

func (o *DefConst) typeCheck() {
    t2 := o.Value.GetType()
    if !types.AreCompatible(o.Type, t2) {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot define \"%s\" (type: %v) with type %v\n", o.Ident.Ident.Str, o.Type, t2)
        fmt.Fprintln(os.Stderr, "\t" + o.Ident.Ident.At())
        os.Exit(1)
    }
}


func (o *ExprStmt) typeCheck() { o.Expr.typeCheck() }
func (o *Assign)   typeCheck() {
    t1 := o.Dest.GetType()
    t2 := o.Value.GetType()

    if !types.AreCompatible(t1, t2) {
        fmt.Fprintf(os.Stderr, "[ERROR] cannot assign a type: %v with type: %v\n", t1, t2)
        fmt.Fprintln(os.Stderr, "\t" + o.Pos.At())
        os.Exit(1)
    }
}
func (o *If) typeCheck() {
    if t := o.Cond.GetType(); t.GetKind() != types.Bool {
        fmt.Fprintf(os.Stderr, "[ERROR] expected an bool as if condition but got %v\n", t)
        fmt.Fprintln(os.Stderr, "\t" + o.Pos.At())
        os.Exit(1)
    }
}
func (o *Elif) typeCheck() {
    (*If)(o).typeCheck()
}
func (o *Switch) typeCheck() {
    for _,c := range o.Cases {
        // skip default case
        if c.Cond == nil { continue }

        c.Cond.typeCheck()
        if t := c.Cond.GetType(); t.GetKind() != types.Bool {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a condition of type bool but got \"%v\"\n", t)
            fmt.Fprintln(os.Stderr, "\t" + c.ColonPos.At())
            os.Exit(1)
        }
    }
}

func (o *For) typeCheck() {
    t := o.Def.Type

    if o.Limit != nil {
        if t2 := o.Limit.GetType(); !types.AreCompatible(t, t2) {
            fmt.Fprintf(os.Stderr, "[ERROR] expected %v as for iterator limit type but got %v\n", t, t2)
            fmt.Fprintln(os.Stderr, "\t" + o.ForPos.At())
            os.Exit(1)
        }
    }

    if t2 := o.Step.GetType(); !types.AreCompatible(t, t2) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected %v as for iterator step type but got %v\n", t, t2)
        fmt.Fprintln(os.Stderr, "\t" + o.ForPos.At())
        os.Exit(1)
    }
}

func (o *While) typeCheck() {
    if t := o.Cond.GetType(); t.GetKind() != types.Bool {
        fmt.Fprintf(os.Stderr, "[ERROR] expected an bool as while condition but got %v\n", t)
        fmt.Fprintln(os.Stderr, "\t" + o.WhilePos.At())
        os.Exit(1)
    }
}


func (o *BadExpr) typeCheck() {}
func (o *Lit)     typeCheck() {}
func (o *Paren)   typeCheck() {}
func (o *Ident)   typeCheck() {}
func (o *Unary)   typeCheck() {
    switch o.Operator.Type {
    case token.Mul:
        if _,ok := o.Operand.(*Ident); !ok {
            if _,ok := o.Operand.(*Paren); !ok {
                fmt.Fprintln(os.Stderr, "[ERROR] expected a variable or parentheses after \"*\"")
                fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
                os.Exit(1)
            }
        }

    case token.Amp:
        if _,ok := o.Operand.(*Ident); !ok {
            fmt.Fprintln(os.Stderr, "[ERROR] expected a variable after \"&\"")
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }

    case token.Plus, token.Minus:
        if t := o.Operand.GetType(); t.GetKind() != types.I32 {
            fmt.Fprintf(os.Stderr, "[ERROR] expected i32 after +/- unary op but got %v\n", t)
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected unary op %v\n", o.Operator)
        fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
        os.Exit(1)
    }
}
func (o *Binary) typeCheck() {
    t1 := o.OperandL.GetType()
    t2 := o.OperandR.GetType()

    if o.Operator.Type == token.And || o.Operator.Type == token.Or {
        if t1.GetKind() != types.Bool || t2.GetKind() != types.Bool {
            fmt.Fprintf(os.Stderr, "[ERROR] expected 2 bools for logic op \"%s\" but got %v and %v\n", o.Operator.Str, t1, t2)
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }
    } else {
        if !types.AreCompatible(t1, t2) {
            if (t1.GetKind() == types.Ptr && t2.GetKind() == types.I32) ||
               (t2.GetKind() == types.Ptr && t1.GetKind() == types.I32) {
                if o.Operator.Type == token.Plus || o.Operator.Type == token.Minus {
                    return
                }

                fmt.Fprintf(os.Stderr, "[ERROR] only +/- operators are allowed for binary ops with %v and %v\n", t1, t2)
                fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
                os.Exit(1)
            }

            fmt.Fprintf(os.Stderr, "[ERROR] binary operation has two differente types (left: %v right: %v)\n", t1, t2)
            fmt.Fprintln(os.Stderr, "\t(ptr +/- i32 is allowed)")
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }
    }
}

func (o *XSwitch) typeCheck() {
    ts := make(map[types.Type][]int)

    for i,c := range o.Cases {
        t2 := c.Expr.GetType()
        ts[t2] = append(ts[t2], i)
    }

    if len(ts) > 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] expected every case body to return the same type but got %d differente\n", len(ts))
        for key,val := range ts {
            fmt.Fprintf(os.Stderr, "cases %v\n   type: %v\n", val, key)
        }

        os.Exit(1)
    }
}

func (o *FnCall) typeCheck() {
    f := fn.GetFn(o.Name.Str)
    if f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" is not declared\n", o.Name.Str)
        fmt.Fprintln(os.Stderr, "\t" + o.Name.At())
        os.Exit(1)
    }

    if len(f.Args) != len(o.Values) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected %d args for function \"%s\" but got %d\n", len(f.Args), o.Name.Str, len(o.Values))
        fmt.Fprintf(os.Stderr, "\texpected: %v\n", f.Args)
        fmt.Fprintf(os.Stderr, "\tgot:      %v\n", valuesToTypes(o.Values))
        fmt.Fprintln(os.Stderr, "\t" + o.Name.At())
        os.Exit(1)
    }

    for i, t1 := range f.Args {
        t2 := o.Values[i].GetType()

        if !types.AreCompatible(t1, t2) {
            fmt.Fprintf(os.Stderr, "[ERROR] expected %v as arg %d but got %v for function \"%s\"\n", t1, i, t2, o.Name.Str)
            fmt.Fprintf(os.Stderr, "\texpected: %v\n", f.Args)
            fmt.Fprintf(os.Stderr, "\tgot:      %v\n", valuesToTypes(o.Values))
            fmt.Fprintln(os.Stderr, "\t" + o.Name.At())
            os.Exit(1)
        }
    }
}

func valuesToTypes(values []Expr) (res []types.Type) {
    for _, v := range values {
        res = append(res, v.GetType())
    }

    return res
}

func (o *BadExpr) GetType() types.Type { return nil }
func (o *FnCall)  GetType() types.Type { return nil }
func (o *Lit)     GetType() types.Type { return o.Type }
func (o *Paren)   GetType() types.Type { return o.Expr.GetType() }
func (o *Ident)   GetType() types.Type {
    if o.C != nil {
        return o.C.Type
    }

    return o.V.GetType()
}

func (o *Unary) GetType() types.Type {
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

func (o *Binary) GetType() types.Type {
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

func (o *XSwitch) GetType() types.Type {
    return o.Cases[0].Expr.GetType()
}
