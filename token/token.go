package token

import (
    "os"
    "fmt"
    "strconv"
    "unicode"
    "strings"
    "gorec/types"
)

var tokens []Token

type TokenType uint
const (
    Unknown TokenType = iota

    Name            // var/func name
    Typename        // i32, str
    Str             // "string"
    Number          // 1234
    Boolean         // true/false

    Plus            // +
    Minus           // -
    Mul             // *
    Div             // /

    Eql             // ==
    Neq             // !=
    Geq             // >=
    Leq             // <=
    Lss             // <
    Grt             // >

    ParenL          // (
    ParenR          // )
    BrackL          // [
    BrackR          // ]
    BraceL          // {
    BraceR          // }

    Comma           // ,
    SemiCol         // ;

    Comment         // // ..., /* ... */

    Dec_var         // var
    Def_var         // :=
    Assign          // =
    Def_fn          // fn
    If              // if
    Else            // else
    While           // while
    For             // for
    Break           // break

    TokenTypeCount uint = iota
)

func TokenTypeOfStr(s string) TokenType {
    switch s {
    case "true", "false":
        return Boolean

    case "+":
        return Plus
    case "-":
        return Minus
    case "*":
        return Mul
    case "/":
        return Div

    case "==":
        return Eql
    case "!=":
        return Neq
    case ">=":
        return Geq
    case "<=":
        return Leq
    case ">":
        return Grt
    case "<":
        return Lss

    case "(":
        return ParenL
    case ")":
        return ParenR
    case "[":
        return BrackL
    case "]":
        return BrackR
    case "{":
        return BraceL
    case "}":
        return BraceR

    case ",":
        return Comma
    case ";":
        return SemiCol

    case "//", "/*", "*/":
        return Comment

    case "var":
        return Dec_var
    case ":=":
        return Def_var
    case "=":
        return Assign
    case "fn":
        return Def_fn
    case "if":
        return If
    case "else":
        return Else
    case "while":
        return While
    case "for":
        return For
    case "break":
        return Break

    default:
        if types.ToType(s) != -1 {
            return Typename
        } else if s[0] == '"' && s[len(s) - 1] == '"' {
            return Str
        } else if _, err := strconv.Atoi(s); err == nil {
            return Number
        }

        return Name
    }
}

func (t TokenType) Readable() string {
    switch t {
    case Plus:
        return "Plus"
    case Minus:
        return "Minus"
    case Mul:
        return "Mul"
    case Div:
        return "Div"

    case Eql:
        return "Eql"
    case Neq:
        return "Neq"
    case Geq:
        return "Geq"
    case Leq:
        return "Leq"
    case Grt:
        return "Grt"
    case Lss:
        return "Lss"

    case ParenL:
        return "ParenL"
    case ParenR:
        return "ParenR"
    case BrackL:
        return "BrackL"
    case BrackR:
        return "BrackR"
    case BraceL:
        return "BraceL"
    case BraceR:
        return "BraceR"

    case Comma:
        return "Comma"
    case SemiCol:
        return "SemiCol"

    case Comment:
        return "Comment"

    case Dec_var:
        return "Dec_var"
    case Def_var:
        return "Def_var"
    case Assign:
        return "Assign"
    case Def_fn:
        return "Def_fn"
    case If:
        return "If"
    case Else:
        return "Else"
    case While:
        return "While"
    case For:
        return "For"
    case Break:
        return "Break"

    case Typename:
        return "Typename"
    case Str:
        return "Str"
    case Number:
        return "Number"
    case Boolean:
        return "Boolean"
    case Name:
        return "Name"

    default:
        return "Unknown"
    }
}

type Pos struct {
    Line int
    Col int
    // later filename
}

func (p Pos) At() string {
    return fmt.Sprintf("at line: %d, col: %d", p.Line, p.Col)
}

type Token struct {
    Type TokenType
    Str string
    Pos Pos
}

func (t Token) At() string {
    return t.Pos.At()
}

// escape chars (TODO: \n, \t, \r, ...) (done: \\, \")
func Tokenize(file []byte) {
    keySigns := "(){}+-*/,;"
    f := string(file)

    start := 0

    line := 1
    col := 1

    skip := false
    mlSkip := false
    strLit := false
    escape := false

    for i, r := range(f) {
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
                    s := f[start:i]

                    tokens = append(tokens, Token{TokenTypeOfStr(s), s, Pos{line, col + start - i} })
                }
                start = i + 1

                if strings.Contains(keySigns, string(r)) {
                    tokens = append(tokens, Token{TokenTypeOfStr(string(r)), string(r), Pos{line, col}})
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

func GetTokens() []Token {
    return tokens
}
