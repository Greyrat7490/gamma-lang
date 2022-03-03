package prs

import (
    "fmt"
    "gorec/types"
    "gorec/vars"
    "os"
)

type OpDecVar struct {
    Varname string
    Vartype types.Type
}

func (o OpDecVar) Readable() string {
    return fmt.Sprintf("%s: %s %s", OP_DEC_VAR.Readable(), o.Varname, o.Vartype.Readable())
}

func (o OpDecVar) Compile(asm *os.File) {
    vars.Declare(o.Varname, o.Vartype)
}


type OpDefVar struct {
    Varname string
    Value string
    ValueType types.Type
}

func (o OpDefVar) Readable() string {
    return fmt.Sprintf("%s: %s %s %s", OP_DEF_VAR.Readable(), o.Varname, o.Value, o.ValueType.Readable())
}

func (o OpDefVar) Compile(asm *os.File) {
    vars.Define(o.Varname, o.Value)
}


func prsDecVar(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] neither name nor type provided for the variable declaration")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    if len(words) < idx + 2 {
        if words[idx+1].Type == name {
            fmt.Fprintln(os.Stderr, "[ERROR] no type provided for the variable")
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] no name provided for the variable")
        }
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    t := types.ToType(words[idx+2].Str)
    if t == -1 {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", words[idx+2].Str)
        fmt.Fprintln(os.Stderr, "\t" + words[idx+2].At())
        os.Exit(1)
    }

    op := OpDecVar{ Varname: words[idx+1].Str, Vartype: t }
    Ops = append(Ops, op)

    return idx + 2
}

func prsDefVar(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] no value provided to define the variable\n")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    value := words[idx+1].Str
    t := types.TypeOfVal(value)
    v := words[idx-2].Str

    op := OpDefVar{ Varname: v, Value: value, ValueType: t }
    Ops = append(Ops, op)

    return idx + 1
}
