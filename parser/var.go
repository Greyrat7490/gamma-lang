package prs

import (
    "os"
    "fmt"
    "strings"
    "gorec/token"
    "gorec/types"
    "gorec/vars"
)

type OpDecVar struct {
    Varname token.Token
    Vartype types.Type
}

func (o OpDecVar) Readable(indent int) string {
    return strings.Repeat("   ", indent) +
        fmt.Sprintf("OP_DEC_VAR: %s(%s) %s(Typename)\n",
        o.Varname.Str, o.Varname.Type.Readable(),
        o.Vartype.Readable())
}

func (o OpDecVar) Compile(asm *os.File) {
    vars.Declare(o.Varname, o.Vartype)
}


type OpDefVar struct {
    Varname token.Token
    Value token.Token
    ValueType types.Type
}

func (o OpDefVar) Readable(indent int) string {
    return strings.Repeat("   ", indent) +
        fmt.Sprintf("OP_DEF_VAR: %s(%s) %s(%s) %s(Typename)\n",
        o.Varname.Str, o.Varname.Type.Readable(),
        o.Value.Str, o.Value.Type.Readable(),
        o.ValueType.Readable())
}

func (o OpDefVar) Compile(asm *os.File) {
    vars.Define(o.Varname, o.Value)
}


func prsDecVar(idx int) (OpDecVar, int) {
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

    return op, idx + 2
}

func prsDefVar(idx int) (OpDefVar, int) {
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

    return op, idx + 1
}
