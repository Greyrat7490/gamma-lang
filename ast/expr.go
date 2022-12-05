package ast

import (
    "os"
    "fmt"
    "strings"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
    "gamma/ast/identObj"
    "gamma/ast/identObj/func"
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

type IntLit struct {
    Repr int64
    Val token.Token
    Type types.IntType
}

type UintLit struct {
    Repr uint64
    Val token.Token
    Type types.UintType
}

type BoolLit struct {
    Repr bool
    Val token.Token
}

type CharLit struct {
    Repr uint8
    Val token.Token
}

type PtrLit struct {
    Addr addr.Addr
    Local bool
    Val token.Token
    Type types.PtrType
}

type StrLit struct {
    Idx uint64
    Val token.Token
}

type ArrayLit struct {
    Idx uint64
    Pos token.Pos
    Type types.ArrType
    BraceLPos token.Pos
    Values []Expr
    BraceRPos token.Pos
}

type VectorLit struct {
    Pos token.Pos
    Type types.VecType
    BraceLPos token.Pos
    Cap Expr
    Len Expr
    BraceRPos token.Pos
}

type StructLit struct {
    Pos token.Pos
    StructType types.StructType
    BraceLPos token.Pos
    Fields []FieldLit
    BraceRPos token.Pos
}

type FieldLit struct {
    Pos token.Pos
    Name token.Token
    Value Expr
}

type Indexed struct {
    ArrType types.Type
    Type types.Type
    ArrExpr Expr
    BrackLPos token.Pos
    Index Expr
    BrackRPos token.Pos
}

type Field struct {
    StructType types.StructType
    Type types.Type
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
    Type types.Type
    Operator token.Token
    Operand Expr
}

type Binary struct {
    Pos token.Pos
    Type types.Type
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
    Type types.Type
    BraceLPos token.Pos
    Cases []XCase
    BraceRPos token.Pos
}

type XCase struct {
    Cond Expr
    ColonPos token.Pos
    Expr Expr
}

type Cast struct {
    Expr Expr
    AsPos token.Pos
    DestType types.Type
}


func (e *IntLit) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(%v)\n", e.Val.Str, e.Type)
}

func (e *UintLit) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(%v)\n", e.Val.Str, e.Type)
}

func (e *BoolLit) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(bool)\n", e.Val.Str)
}

func (e *CharLit) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(char)\n", e.Val.Str)
}

func (e *PtrLit) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(ptr)\n", e.Val.Str)
}

func (o *StrLit) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(str)\n", o.Val.Str)
}

func (o *FieldLit) Readable(indent int) string {
    if o.Name.Type == token.Unknown {
        return o.Value.Readable(indent)
    } else {
        return strings.Repeat("   ", indent) + fmt.Sprintf("%s: \n%s", o.Name, o.Value.Readable(indent+1))
    }
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

func (e *VectorLit) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "VECTOR_LIT:\n"
    if e.Cap != nil {
        res += strings.Repeat("   ", indent+1) + "cap: " + e.Cap.Readable(0)
    }
    if e.Len != nil {
        res += strings.Repeat("   ", indent+1) + "len: " + e.Len.Readable(0)   
    }

    return res
}

func (o *Indexed) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "INDEXED:\n" +
        o.ArrExpr.Readable(indent+1) + o.Index.Readable(indent+1)

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

func (o *Cast) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "AS:\n" +
        strings.Repeat("   ", indent+1) + o.DestType.String() + "\n" +
        o.Expr.Readable(indent+1)
}

func (o *BadExpr) Readable(indent int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
    return ""
}

func (e *Indexed) Flatten() Expr {
    if t,ok := e.ArrType.(types.ArrType); ok {
        idxType := types.CreateUint(types.Ptr_Size)

        if i,ok := e.ArrExpr.(*Indexed); ok {
            return &Binary{
                Type: idxType,
                Operator: token.Token{ Str: "+", Type: token.Plus },
                OperandR: e.Index,
                OperandL: &Binary{
                    Type: idxType,
                    Operator: token.Token{ Str: "*", Type: token.Mul },
                    OperandL: i.Flatten(),
                    OperandR: &UintLit{ Repr: t.Lens[0], Type: idxType },
                },
            }
        }
    }

    return e.Index
}


func (e *BadExpr)   GetType() types.Type { return nil }
func (e *IntLit)    GetType() types.Type { return e.Type }
func (e *UintLit)   GetType() types.Type { return e.Type }
func (e *BoolLit)   GetType() types.Type { return types.BoolType{} }
func (e *CharLit)   GetType() types.Type { return types.CharType{} }
func (e *PtrLit)    GetType() types.Type { return e.Type }
func (e *StrLit)    GetType() types.Type { return types.StrType{} }
func (e *FieldLit)  GetType() types.Type { return e.Value.GetType() }
func (e *StructLit) GetType() types.Type { return e.StructType }
func (e *ArrayLit)  GetType() types.Type { return e.Type }
func (e *VectorLit) GetType() types.Type { return e.Type }
func (e *FnCall)    GetType() types.Type { return e.F.GetRetType() }
func (e *Indexed)   GetType() types.Type { return e.Type }
func (e *Field)     GetType() types.Type { return e.Type }
func (e *Ident)     GetType() types.Type { return e.Obj.GetType() }
func (e *Unary)     GetType() types.Type { return e.Type }
func (e *Binary)    GetType() types.Type { return e.Type }
func (e *Paren)     GetType() types.Type { return e.Expr.GetType() }
func (e *XSwitch)   GetType() types.Type { return e.Type }
func (e *XCase)     GetType() types.Type { return e.Expr.GetType() }
func (e *Cast)      GetType() types.Type { return e.DestType }


func (e *BadExpr)   expr() {}
func (e *IntLit)    expr() {}
func (e *UintLit)   expr() {}
func (e *BoolLit)   expr() {}
func (e *CharLit)   expr() {}
func (e *PtrLit)    expr() {}
func (e *StrLit)    expr() {}
func (e *FieldLit)  expr() {}
func (e *StructLit) expr() {}
func (e *ArrayLit)  expr() {}
func (e *VectorLit) expr() {}
func (e *FnCall)    expr() {}
func (e *Indexed)   expr() {}
func (e *Field)     expr() {}
func (e *Ident)     expr() {}
func (e *Unary)     expr() {}
func (e *Binary)    expr() {}
func (e *Paren)     expr() {}
func (e *XSwitch)   expr() {}
func (e *XCase)     expr() {}
func (e *Cast)      expr() {}

func (e *BadExpr)   At() string { return "" }
func (e *IntLit)    At() string { return e.Val.At() }
func (e *UintLit)   At() string { return e.Val.At() }
func (e *BoolLit)   At() string { return e.Val.At() }
func (e *CharLit)   At() string { return e.Val.At() }
func (e *PtrLit)    At() string { return e.Val.At() }
func (e *StrLit)    At() string { return e.Val.At() }
func (e *FieldLit)  At() string { return e.Pos.At() }
func (e *StructLit) At() string { return e.Pos.At() }
func (e *ArrayLit)  At() string { return e.Pos.At() }
func (e *VectorLit) At() string { return e.Pos.At() }
func (e *FnCall)    At() string { return e.Ident.At() }
func (e *Indexed)   At() string { return e.ArrExpr.At() }
func (e *Field)     At() string { return e.Obj.At() }
func (e *Ident)     At() string { return e.Pos.At() }
func (e *Unary)     At() string { return e.Operator.At() }
func (e *Binary)    At() string { return e.OperandL.At() }    // TODO: At() of Operand with higher precedence
func (e *Paren)     At() string { return e.ParenLPos.At() }
func (e *XSwitch)   At() string { return e.Pos.At() }
func (e *XCase)     At() string { return e.ColonPos.At() }
func (e *Cast)      At() string { return e.Expr.At() }

func (e *BadExpr)   End() string { return "" }
func (e *IntLit)    End() string { return e.Val.At() }
func (e *UintLit)   End() string { return e.Val.At() }
func (e *BoolLit)   End() string { return e.Val.At() }
func (e *CharLit)   End() string { return e.Val.At() }
func (e *PtrLit)    End() string { return e.Val.At() }
func (e *StrLit)    End() string { return e.Val.At() }
func (e *FieldLit)  End() string { return e.Value.End() }
func (e *StructLit) End() string { return e.BraceRPos.At() }
func (e *ArrayLit)  End() string { return e.BraceRPos.At() }
func (e *VectorLit) End() string { return e.BraceRPos.At() }
func (e *FnCall)    End() string { return e.ParenRPos.At() }
func (e *Indexed)   End() string { return e.BrackRPos.At() }
func (e *Field)     End() string { return e.FieldName.At() }
func (e *Ident)     End() string { return e.Pos.At() }
func (e *Unary)     End() string { return e.Operand.At() }
func (e *Binary)    End() string { return e.OperandR.At() }
func (e *Paren)     End() string { return e.ParenRPos.At() }
func (e *XSwitch)   End() string { return e.BraceRPos.At() }
func (e *XCase)     End() string { return e.Expr.At() }
func (e *Cast)      End() string { return e.AsPos.At() }
