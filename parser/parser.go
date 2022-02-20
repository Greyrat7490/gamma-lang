package prs

import (
    "fmt"
    "os"
    "strings"
    "unicode"
    "strconv"
)

var Ops []Op
var tokens []Token
var isMainDefined bool = false

type Token struct {
    Str string
    line int
    col int
    // later filename
}

func (w Token) At() string {
    return fmt.Sprintf("at line: %d, col: %d", w.line, w.col)
}

type OpType uint
const (
    OP_DEC_VAR  OpType = iota
    OP_DEF_VAR
    OP_ASSIGN_VAR
    OP_DEF_FN
    OP_END_FN
    OP_CALL_FN
    OP_DEC_ARGS
    OP_DEF_ARGS
    OP_ADD
    OP_SUB
    OP_MUL
    OP_DIV
    OP_COUNT      uint = iota
)

func (o OpType) Readable() string {
    // compile time reminder to add cases when Operants are added
    const _ uint = 12 - OP_COUNT

    switch o {
    case OP_DEC_VAR:
        return "OP_DEC_VAR"
    case OP_DEF_VAR:
        return "OP_DEF_VAR"
    case OP_ASSIGN_VAR:
        return "OP_ASSIGN_VAR"
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
    case OP_ADD:
        return "OP_ADD"
    case OP_SUB:
        return "OP_SUB"
    case OP_MUL:
        return "OP_MUL"
    case OP_DIV:
        return "OP_DIV"
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

func IsLit(w string) bool {
    if w[0] == '"' && w[len(w) - 1] == '"' {
        return true
    }

    if _, err := strconv.Atoi(w); err == nil {
        return true
    }

    return false
}

func Tokenize(src []byte) {
    split(string(src))

    for i := 0; i < len(tokens); i++ {
        switch tokens[i].Str {
        case "var":
            i = prsDecVar(tokens, i)
        case ":=":
            i = prsDefVar(tokens, i)
        case "fn":
            i = prsDefFn(tokens, i)
        case "+":
            i = prsAdd(tokens, i)
        case "-":
            i = prsSub(tokens, i)
        case "*":
            i = prsMul(tokens, i)
        case "/":
            i = prsDiv(tokens, i)
        case "printInt", "printStr", "exit":
            fmt.Fprintln(os.Stderr, "[ERROR] function calls outside of main are not allowed")
            fmt.Fprintln(os.Stderr, "\t" + tokens[i].At())
            os.Exit(1)
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] unknown word \"%s\"\n", tokens[i].Str)
            fmt.Fprintln(os.Stderr, "\t" + tokens[i].At())
            os.Exit(1)
        }
    }

    if !isMainDefined {
        fmt.Fprintln(os.Stderr, "[ERROR] no \"main\" function was defined")
        os.Exit(1)
    }
}

// escape chars (TODO: \n, \t, \r, ...) (done: \\, \")
func split(file string) {
    keySigns := "(){}+-*/"

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
            } else if unicode.IsSpace(r) || strings.Contains(keySigns, string(r)) {
                if start != i {
                    tokens = append(tokens, Token{file[start:i], line, col + start - i})
                }
                start = i + 1

                if strings.Contains(keySigns, string(r)) {
                    tokens = append(tokens, Token{string(r), line, col - 1})
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
