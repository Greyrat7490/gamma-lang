package prs

import (
    "fmt"
    "gorec/types"
    "os"
    "unicode"
    "strconv"
)

var Ops []interface{ Op }
var tokens []Token
var isMainDefined bool = false

type TokenType uint
const (
    unknown TokenType = iota

    name            // var/func name
    typename        // i32, str
    str             // "string"
    number          // 1234

    plus            // +
    minus           // -
    mul             // *
    div             // /

    parenL          // (
    parenR          // )
    brackL          // [
    brackR          // ]
    braceL          // {
    braceR          // }

    comment         // //,/*,*/

    dec_var         // var
    def_var         // :=
    assign          // =
    def_fn          // fn

    tokenTypeCount uint = iota
)

func TokenTypeOfStr(s string) TokenType {
    switch s {
    case "+":
        return plus
    case "-":
        return minus
    case "*":
        return mul
    case "/":
        return div

    case "(":
        return parenL
    case ")":
        return parenR
    case "[":
        return brackL
    case "]":
        return brackR
    case "{":
        return braceL
    case "}":
        return braceR

    case "//", "/*", "*/":
        return comment

    case "var":
        return dec_var
    case ":=":
        return def_var
    case "=":
        return assign
    case "fn":
        return def_fn

    default:
        if types.ToType(s) != -1 {
            return typename
        } else if s[0] == '"' && s[len(s) - 1] == '"' {
            return str
        } else if _, err := strconv.Atoi(s); err == nil {
            return number
        }

        return name
    }
}

func IsLit(s string) bool {
    if s[0] == '"' && s[len(s) - 1] == '"' {
        return true
    }

    if _, err := strconv.Atoi(s); err == nil {
        return true
    }

    return false
}

type Token struct {
    Type TokenType
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
    OP_DEF_VAR
    OP_DEF_FN
    OP_END_FN
    OP_CALL_FN
    OP_DEC_ARGS
    OP_DEF_ARGS
    OP_COUNT      uint = iota
)

func (o OpType) Readable() string {
    // compile time reminder to add cases when Operants are added
    const _ uint = 7 - OP_COUNT

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

type Op interface {
    Readable() string
    Compile(asm *os.File)
}

func ShowOps() {
    for i, o := range Ops {
        fmt.Printf("%d: %s\n", i, o.Readable())
    }
}

func Parse() {
    for i := 0; i < len(tokens); i++ {
        switch tokens[i].Type {
        case dec_var:
            i = prsDecVar(tokens, i)
        case def_var:
            i = prsDefVar(tokens, i)
        case def_fn:
            i = prsDefFn(tokens, i)
        case name:
            if tokens[i+1].Type == parenL {
                fmt.Fprintln(os.Stderr, "[ERROR] function calls are not allowed in global scope")
                fmt.Fprintln(os.Stderr, "\t" + tokens[i].At())
                os.Exit(1)
            }
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
func Tokenize(file string) {
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
                    s := file[start:i]

                    tokens = append(tokens, Token{TokenTypeOfStr(s), s, line, col + start - i})
                }
                start = i + 1

                if r == '(' || r == ')' || r == '{' || r == '}' {
                    tokens = append(tokens, Token{TokenTypeOfStr(string(r)), string(r), line, col - 1})
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
