package ast

import (
    "os"
    "fmt"
    "strings"
    "gamma/token"
    "gamma/types"
    "gamma/ast/identObj"
    "gamma/ast/identObj/func"
    "gamma/ast/identObj/vars"
    "gamma/ast/identObj/consts"
    "gamma/ast/identObj/struct"
)

type Expr interface {
    Node
    GetType() types.Type
    expr()  // to distinguish Expr from Stmt and Decl
}

type BadExpr struct{}

type FnCall struct {
    F *fn.Func
    Ident Ident
    ParenLPos token.Pos
    Values []Expr
    ParenRPos token.Pos
}

type Lit struct {
    Val token.Token
    Type types.Type
}

type ArrayLit struct {
    Idx int
    Pos token.Pos
    Type types.ArrType
    BraceLPos token.Pos
    Values []Expr
    BraceRPos token.Pos
}

type StructLit struct {
    Idx int
    Pos token.Pos
    StructType types.StructType
    BraceLPos token.Pos
    Fields []FieldLit
    BraceRPos token.Pos
}

type FieldLit struct {
    Name token.Token
    Pos token.Pos
    Value Expr
}

type Indexed struct {
    ArrExpr Expr
    BrackLPos token.Pos
    Indices []Expr
    BrackRPos token.Pos
}

type Field struct {
    Pos token.Pos
    Obj Expr
    DotPos token.Pos
    FieldName token.Token
}

type Ident struct {
    Name string
    Pos token.Pos
    Obj identObj.IdentObj
}

type Unary struct {
    Operator token.Token
    Operand Expr
}

type Binary struct {
    Pos token.Pos
    OperandL Expr
    Operator token.Token
    OperandR Expr
}

type Paren struct {
    ParenLPos token.Pos
    Expr Expr
    ParenRPos token.Pos
}

type XSwitch struct {
    Pos token.Pos
    BraceLPos token.Pos
    Cases []XCase
    BraceRPos token.Pos
}

type XCase struct {
    Cond Expr
    ColonPos token.Pos
    Expr Expr
}


func (o *Lit) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(%v)\n", o.Val.Str, o.Type)
}
func (o *FieldLit) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s: \n%s", o.Name, o.Value.Readable(indent+1))
}
func (o *StructLit) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "STRUCT_LIT:\n"

    for _,f := range o.Fields {
        res += f.Readable(indent+1)
    }

    return res
}
func (o *ArrayLit) Readable(indent int) string {
    s := strings.Repeat("   ", indent+1)
    res := strings.Repeat("   ", indent) + "ARRAY_LIT:\n" +
           fmt.Sprintf("%sType: %v\n", s, o.Type)

    for _,v := range o.Values {
        res += v.Readable(indent+1)
    }

    return res
}

func (o *Indexed) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "INDEXED:\n" +
        o.ArrExpr.Readable(indent+1)

    for _, idx := range o.Indices {
        res += idx.Readable(indent+1)
    }

    return res
}

func (o *Field) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "FIELD:\n" +
        o.Obj.Readable(indent+1) +
        strings.Repeat("   ", indent+1) + o.FieldName.String() + "\n"
}

func (o *Ident) Readable(indent int) string {
    return strings.Repeat("   ", indent) + o.Name + "(Name)\n"
}

func (o *FnCall) Readable(indent int) string {
    s  := strings.Repeat("   ", indent)
    s2 := strings.Repeat("   ", indent+1)

    res := s + "CALL_FN:\n" +
          s2 + o.Ident.Name + "(Name)\n"

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



func (e *FnCall) GetType() types.Type {
    return e.F.GetRetType()
}

func (e *Lit) GetType() types.Type {
    return e.Type
}

func (e *FieldLit) GetType() types.Type {
    return e.Value.GetType()
}

func (e *StructLit) GetType() types.Type {
    return e.StructType
}

func (e *ArrayLit) GetType() types.Type {
    return e.Type
}

func (e *Indexed) GetType() types.Type {
    if t,ok := e.ArrExpr.GetType().(types.ArrType); ok {
        return t.Ptr.BaseType
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] you can only index an array but got %v\n", t)
        os.Exit(1)
        return nil
    }
}

func (e *Field) GetType() types.Type {
    t := e.Obj.GetType()

    if sType,ok := t.(types.StructType); ok {
        obj := identObj.Get(sType.Name)
        if s,ok := obj.(*structDec.Struct); ok {
            return s.GetTypeOfField(e.FieldName.Str)
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] expected struct but got %v\n", t)
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    return nil
}

func (e *Paren) GetType() types.Type {
    return e.Expr.GetType()
}

func (e *Ident) GetType() types.Type {
    if c,ok := e.Obj.(*consts.Const); ok {
        return c.GetType()
    }

    if v,ok := e.Obj.(vars.Var); ok {
        return v.GetType()
    }

    if s,ok := e.Obj.(*structDec.Struct); ok {
        return s.GetType()
    }

    // TODO: function

    fmt.Fprintf(os.Stderr, "[ERROR] could not get type of %s\n", e.Name)
    os.Exit(1)
    return nil
}

func (e *Unary) GetType() types.Type {
    if e.Operator.Type == token.Amp {
        return types.PtrType{ BaseType: e.Operand.GetType() }
    }

    if e.Operator.Type == token.Mul {
        if ptr, ok := e.Operand.GetType().(types.PtrType); ok {
            return ptr.BaseType
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] you cannot deref this expre (expected a pointer/address)")
            fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
            os.Exit(1)
        }
    }

    return e.Operand.GetType()
}

func (e *Binary) GetType() types.Type {
    if  e.Operator.Type == token.Eql || e.Operator.Type == token.Neq ||
        e.Operator.Type == token.Grt || e.Operator.Type == token.Lss ||
        e.Operator.Type == token.Geq || e.Operator.Type == token.Leq {
        return types.BoolType{}
    }

    t := e.OperandL.GetType()
    if t == nil {
        return e.OperandR.GetType()
    }

    if other := e.OperandR.GetType(); other.GetKind() == types.Ptr {
        // check for cases like 420 + &v1
        if t.GetKind() == types.Int {
            return other
        }

        // check for cases like ptr1 - ptr2
        if t.GetKind() == types.Ptr {
            return types.CreateInt(types.Ptr_Size)
        }
    }

    return t
}

func (e *XSwitch) GetType() types.Type {
    return e.Cases[0].Expr.GetType()
}

func (e *XCase) GetType() types.Type {
    return e.Expr.GetType()
}

func (e *BadExpr) GetType() types.Type {
    return nil
}


func (e *Indexed) Flatten() Expr {
    // for dim = 1 (no need to flatten)
    if len(e.Indices) == 1 {
        return e.Indices[0]
    }

    res := Binary{ Operator: token.Token{ Str: "+", Type: token.Plus }, OperandR: e.Indices[len(e.Indices)-1] }
    expr := &res.OperandL

    // for dim > 3
    arrType,_ := e.ArrExpr.GetType().(types.ArrType)
    var innerLen uint64 = arrType.Lens[0]
    for i := 1; i < len(e.Indices)-1; i++ {
        b := Binary{
            Operator: token.Token{ Str: "+", Type: token.Plus }, OperandR: &Binary{
                Operator: token.Token{ Str: "*", Type: token.Mul },
                OperandL: e.Indices[len(e.Indices)-1-i],
                OperandR: &Lit{ Val: token.Token{ Str: fmt.Sprint(innerLen), Type: token.Number }, Type: types.CreateInt(types.I32_Size) },
        } }

        *expr = &b
        expr = &b.OperandL

        innerLen *= arrType.Lens[i]
    }

    *expr = &Binary{
        Operator: token.Token{ Str: "*", Type: token.Mul },
        OperandL: e.Indices[0],
        OperandR: &Lit{ Val: token.Token{ Str: fmt.Sprint(innerLen), Type: token.Number }, Type: types.CreateInt(types.I32_Size) },
    }

    return &res
}


func (e *BadExpr)   expr() {}
func (e *FnCall)    expr() {}
func (e *Lit)       expr() {}
func (e *FieldLit)  expr() {}
func (e *StructLit) expr() {}
func (e *ArrayLit)  expr() {}
func (e *Indexed)   expr() {}
func (e *Field)     expr() {}
func (e *Ident)     expr() {}
func (e *Unary)     expr() {}
func (e *Binary)    expr() {}
func (e *Paren)     expr() {}
func (e *XSwitch)   expr() {}
func (e *XCase)     expr() {}

func (e *BadExpr)   At() string { return "" }
func (e *FnCall)    At() string { return e.Ident.At() }
func (e *Lit)       At() string { return e.Val.At() }
func (e *FieldLit)  At() string { return e.Name.At() }
func (e *StructLit) At() string { return e.Pos.At() }
func (e *ArrayLit)  At() string { return e.Pos.At() }
func (e *Indexed)   At() string { return e.ArrExpr.At() }
func (e *Field)     At() string { return e.Pos.At() }
func (e *Ident)     At() string { return e.Pos.At() }
func (e *Unary)     At() string { return e.Operator.At() }
func (e *Binary)    At() string { return e.OperandL.At() }    // TODO: At() of Operand with higher precedence
func (e *Paren)     At() string { return e.ParenLPos.At() }
func (e *XSwitch)   At() string { return e.Pos.At() }
func (e *XCase)     At() string { return e.ColonPos.At() }

func (e *BadExpr)   End() string { return "" }
func (e *FnCall)    End() string { return e.ParenRPos.At() }
func (e *Lit)       End() string { return e.Val.At() }
func (e *FieldLit)  End() string { return e.Value.End() }
func (e *StructLit) End() string { return e.BraceRPos.At() }
func (e *ArrayLit)  End() string { return e.BraceRPos.At() }
func (e *Indexed)   End() string { return e.BrackRPos.At() }
func (e *Field)     End() string { return e.FieldName.At() }
func (e *Ident)     End() string { return e.Pos.At() }
func (e *Unary)     End() string { return e.Operand.At() }
func (e *Binary)    End() string { return e.OperandR.At() }
func (e *Paren)     End() string { return e.ParenRPos.At() }
func (e *XSwitch)   End() string { return e.BraceRPos.At() }
func (e *XCase)     End() string { return e.Expr.At() }
