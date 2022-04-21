package ast

import (
    "os"
    "fmt"
    "strings"
    "gorec/arithmetic"
    "gorec/func"
    "gorec/token"
    "gorec/types"
    "gorec/vars"
)

type OpExpr interface {
    Op
    Compile(asm *os.File)
    expr()  // to differenciate OpExpr from OpDecl and OpStmt
}

type BadExpr struct{}

type OpFnCall struct {
    FnName token.Token
    Values []string
}

type LitExpr struct {
    Val token.Token
    Type types.Type
}

type IdentExpr struct {
    Ident token.Token
}

type UnaryExpr struct {
    Operator token.Token
    Operand OpExpr
}

type BinaryExpr struct {
    OperandL OpExpr
    Operator token.Token
    OperandR OpExpr
}

func (o *BadExpr)    expr() {}
func (o *LitExpr)    expr() {}
func (o *IdentExpr)  expr() {}
func (o *OpFnCall)   expr() {}
func (o *UnaryExpr)  expr() {}
func (o *BinaryExpr) expr() {}


func (o *LitExpr)   Compile(asm *os.File) {}
func (o *IdentExpr) Compile(asm *os.File) {}
func (o *UnaryExpr) Compile(asm *os.File) { 
    if l, ok := o.Operand.(*LitExpr); ok {
        vars.WriteVar(asm, fmt.Sprintf("mov rax, %s\n", l.Val.Str))
    } else if ident, ok := o.Operand.(*IdentExpr); ok {
        if v := vars.GetVar(ident.Ident.Str); v != nil {
            // TODO is i32
            reg := vars.Registers[v.Regs[0]].Name
            vars.WriteVar(asm, fmt.Sprintf("mov rax, %s\n", reg))
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] variable %s is not declared\n", ident.Ident.Str) 
            fmt.Fprintln(os.Stderr, "\t" + ident.Ident.At())
            os.Exit(1)
        }
    }

    o.Operand.Compile(asm)

    if o.Operator.Type == token.Minus {
        vars.WriteVar(asm, "neg rax\n")
    }
}
func (o *BinaryExpr) Compile(asm *os.File) {
    if l, ok := o.OperandL.(*LitExpr); ok {
        vars.WriteVar(asm, fmt.Sprintf("mov rax, %s\n", l.Val.Str))
    } else if ident, ok := o.OperandL.(*IdentExpr); ok {
        if v := vars.GetVar(ident.Ident.Str); v != nil {
            // TODO is i32
            reg := vars.Registers[v.Regs[0]].Name
            vars.WriteVar(asm, fmt.Sprintf("mov rax, %s\n", reg))
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] variable %s is not declared\n", ident.Ident.Str) 
            fmt.Fprintln(os.Stderr, "\t" + ident.Ident.At())
            os.Exit(1)
        }
    }

    o.OperandL.Compile(asm)

    if l, ok := o.OperandR.(*LitExpr); ok {
        arith.BinaryOp(asm, o.Operator.Type, l.Val.Str)
    } else if ident, ok := o.OperandR.(*IdentExpr); ok {
        if v := vars.GetVar(ident.Ident.Str); v != nil {
            // TODO is i32
            reg := vars.Registers[v.Regs[0]].Name
            arith.BinaryOp(asm, o.Operator.Type, reg)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] variable %s is not declared\n", ident.Ident.Str) 
            fmt.Fprintln(os.Stderr, "\t" + ident.Ident.At())
            os.Exit(1)
        }
    } else {
        vars.WriteVar(asm, "push rbx\n")
        vars.WriteVar(asm, "mov rbx, rax\n")
        o.OperandR.Compile(asm)

        arith.BinaryOp(asm, o.Operator.Type, "rbx")
        vars.WriteVar(asm, "pop rbx\n")
    }
}

func (o *OpFnCall) Compile(asm *os.File) {
    fn.DefineArgs(asm, o.FnName, o.Values)
    fn.CallFunc(asm, o.FnName)
}

func (o *BadExpr) Compile(asm *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
}


func (o *LitExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(%s)\n", o.Val.Str, o.Type.Readable())
}

func (o *IdentExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + o.Ident.Str + "(Name)\n"
}

func (o *OpFnCall) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "
    return fmt.Sprintf("%sOP_CALL_FN:\n%s%s %v\n", s, s2, o.FnName.Str, o.Values)
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
        s2 + o.Operator.Str + o.Operator.Type.Readable() + "\n" +
        o.OperandR.Readable(indent+1)
}

func (o *BadExpr) Readable(indent int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
    return ""
}
