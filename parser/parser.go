package prs

import (
    "fmt"
    "unicode"
    "os"
)

var Tokens []Token
var Ops []Op


type Token struct {
    Str string
    Line int
    Col int
    // later filename
}

func (w Token) At() string {
    return fmt.Sprintf("at line: %d, col: %d", w.Line, w.Col)
}

type OpType uint
const (
    OP_DEC_VAR  OpType = iota
    OP_DEF_VAR  OpType = iota
    OP_DEF_FN   OpType = iota
    OP_END_FN   OpType = iota
    OP_CALL_FN  OpType = iota
    OP_DEC_ARGS OpType = iota
    OP_DEF_ARGS OpType = iota
)

func (o OpType) Readable() string {
    switch o {
    case OP_DEC_VAR:
        return "OP_DEC_VAR"
    case OP_DEF_VAR:
        return "OP_DEF_VAR"
    case OP_DEF_FN:
        return "OP_DEF_FN"
    case OP_END_FN:
        return "OP_END_FN"
    case OP_CALL_FN:
        return "OP_CALL_FN"
    case OP_DEC_ARGS:
        return "OP_DEC_ARGS"
    case OP_DEF_ARGS:
        return "OP_DEF_ARGS"
    default:
        return ""
    }
}

type Op struct {
    Type OpType
    Token Token
    Operants []string
}

func (o Op) Readable() string {
    return fmt.Sprintf("%s %v", o.Type.Readable(), o.Operants)
}

func ShowOps() {
    for i, o := range Ops {
        fmt.Printf("%d: %s\n", i, o.Readable())
    }
}

// escape chars (TODO: \n, \t, \r, ...) (done: \\, \")
func Split(file string) {
    start := 0

    line := 1
    col := 1

    skip := false
    mlSkip := false
    strLit := false
    escape := false

    for i, r := range(file) {
        // comments
        if skip {
            if mlSkip {
                if r == '*' && file[i+1] == '/' {
                    skip = false
                    mlSkip = false
                    start = i + 2
                }
            } else {
                if r == '\n' {
                    skip = false
                    start = i + 1
                }
            }

        // string literales
        } else if strLit {
            if !escape {
                if r == '"' {
                    strLit = false
                } else if r == '\\' {
                    escape = true
                }
            } else {
                escape = false
            }

        } else {
            if r == '"' {       // start string literal
                strLit = true
            }

            if r == '/' {       // start comment
                if file[i+1] == '/' {
                    skip = true
                } else if file[i+1] == '*' {
                    skip = true
                    mlSkip = true
                }

            // split
            } else if unicode.IsSpace(r) || r == '(' || r == ')' || r == '{' || r == '}' {
                if start != i {
                    Tokens = append(Tokens, Token{file[start:i], line, col + start - i})
                }
                start = i + 1

                if r == '(' || r == ')' || r == '{' || r == '}' {
                    Tokens = append(Tokens, Token{string(r), line, col - 1})
                }
            }
        }

        // set word position
        if r == '\n' {
            line++
            col = 0
        }
        col++
    }

    if mlSkip {
        fmt.Fprintln(os.Stderr, "you have not terminated your comment (missing \"*/\")")
        os.Exit(1)
    }
}
