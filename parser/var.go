package prs

import (
    "os"
    "fmt"
    "gorec/vars"
    "gorec/types"
    "gorec/token"
    "gorec/ast"
)

type OpDecVar struct {
    Varname token.Token
    Vartype types.Type
}

func (o OpDecVar) Readable() string {
    return fmt.Sprintf("%s: %s %s", ast.OP_DEC_VAR.Readable(), o.Varname.Str, o.Vartype.Readable())
}

func (o OpDecVar) Compile(asm *os.File) {
    vars.Declare(o.Varname, o.Vartype)
}


type OpDefVar struct {
    Varname token.Token
    Value token.Token
    ValueType types.Type
}

func (o OpDefVar) Readable() string {
    return fmt.Sprintf("%s: %s %s %s", ast.OP_DEF_VAR.Readable(), o.Varname.Str, o.Value.Str, o.ValueType.Readable())
}

func (o OpDefVar) Compile(asm *os.File) {
    vars.Define(o.Varname, o.Value)
}


func prsDecVar(idx int) int {
    tokens := token.GetTokens()

    if len(tokens) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] neither name nor type provided for the variable declaration")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }
    if len(tokens) < idx + 2 {
        if tokens[idx+1].Type == token.Name {
            fmt.Fprintln(os.Stderr, "[ERROR] no type provided for the variable")
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] no name provided for the variable")
        }
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+1].At())
        os.Exit(1)
    }

    if (tokens[idx+1].Type != token.Name) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %s(\"%s\")\n", tokens[idx+1].Type.Readable(), tokens[idx+1].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+1].At())
        os.Exit(1)
    }
    if (tokens[idx+2].Type != token.Typename) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Typename but got %s(\"%s\")\n", tokens[idx+2].Type.Readable(), tokens[idx+2].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+2].At())
        os.Exit(1)
    }


    t := types.ToType(tokens[idx+2].Str)
    if t == -1 {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", tokens[idx+2].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+2].At())
        os.Exit(1)
    }

    op := OpDecVar{ Varname: tokens[idx+1], Vartype: t }
    ast.Ast = append(ast.Ast, op)

    return idx + 2
}

func prsDefVar(idx int) int {
    tokens := token.GetTokens()

    if len(tokens) < idx + 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] no value provided to define the variable\n")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    if (tokens[idx-2].Type != token.Name) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %s(\"%s\")\n", tokens[idx-2].Type.Readable(), tokens[idx-2].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx-2].At())
        os.Exit(1)
    }
    if (!(tokens[idx+1].Type == token.Name || tokens[idx+1].Type == token.Number || tokens[idx+1].Type == token.Str)) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name or a literal but got %s(\"%s\")\n", tokens[idx+1].Type.Readable(), tokens[idx+1].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+1].At())
        os.Exit(1)
    }

    value := tokens[idx+1]
    t := types.TypeOfVal(value.Str)
    v := tokens[idx-2]

    op := OpDefVar{ Varname: v, Value: value, ValueType: t }
    ast.Ast = append(ast.Ast, op)

    return idx + 1
}
