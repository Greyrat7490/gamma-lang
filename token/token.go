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
var idx int = -1

type TokenType uint
const (
    Unknown TokenType = iota

    EOF             // end of file

    Name            // var/func name
    Typename        // i32, str, bool
    Str             // "string"
    Number          // 1234
    Boolean         // true/false

    Plus            // +
    Minus           // -
    Mul             // *
    Div             // /
    Mod             // %

    And             // &&
    Or              // ||

    Amp             // &

    Eql             // ==
    Neq             // !=
    Geq             // >=
    Leq             // <=
    Lss             // <
    Grt             // >
    Not             // !

    ParenL          // (
    ParenR          // )
    BrackL          // [
    BrackR          // ]
    BraceL          // {
    BraceR          // }
    UndScr          // _

    Comma           // ,
    Colon           // :
    SemiCol         // ;

    Comment         // // ..., /* ... */

    Dec_var         // var
    Def_var         // :=
    Assign          // =
    Def_fn          // fn
    If              // if
    Elif            // elif
    Else            // else
    While           // while
    For             // for
    Break           // break
    Continue        // continue
    Through         // through
    XSwitch         // $

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
    case "%":
        return Mod

    case "&&":
        return And
    case "||":
        return Or

    case "&":
        return Amp

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
    case "!":
        return Not

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
    case "_":
        return UndScr

    case ",":
        return Comma
    case ";":
        return SemiCol
    case ":":
        return Colon

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
    case "elif":
        return Elif
    case "else":
        return Else
    case "while":
        return While
    case "for":
        return For
    case "break":
        return Break
    case "continue":
        return Continue
    case "through":
        return Through
    case "$":
        return XSwitch

    default:
        if types.ToType(s) != nil {
            return Typename
        } else if s[0] == '"' && s[len(s) - 1] == '"' {
            return Str
        } else if _, err := strconv.Atoi(s); err == nil {
            return Number
        }

        return Name
    }
}

// TODO: to string()
func (t TokenType) Readable() string {
    switch t {
    case EOF:
        return "EOF"

    case Plus:
        return "Plus"
    case Minus:
        return "Minus"
    case Mul:
        return "Mul"
    case Div:
        return "Div"
    case Mod:
        return "Mod"

    case And:
        return "And"
    case Or:
        return "Or"

    case Amp:
        return "Amp"

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
    case Not:
        return "Not"

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
    case UndScr:
        return "UnderS"

    case Comma:
        return "Comma"
    case SemiCol:
        return "SemiCol"
    case Colon:
        return "Colon"

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
    case Elif:
        return "Elif"
    case Else:
        return "Else"
    case While:
        return "While"
    case For:
        return "For"
    case Break:
        return "Break"
    case Continue:
        return "Continue"
    case Through:
        return "Through"
    case XSwitch:
        return "XSwitch"

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

func Tokenize(file []byte) {
    keySigns := "(){}+-*/%=,:;&$"
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

                    t := TokenTypeOfStr(s)
                    if t == Typename && tokens[len(tokens)-1].Type == Mul {     // *typename
                        tokens[len(tokens)-1].Str += s
                        tokens[len(tokens)-1].Type = Typename
                    } else {
                        tokens = append(tokens, Token{t, s, Pos{line, col + start - i} })
                    }
                }
                start = i + 1

                if strings.Contains(keySigns, string(r)) {
                    t := TokenTypeOfStr(string(r))

                    if t == Amp && tokens[len(tokens)-1].Type == Amp {
                        tokens[len(tokens)-1].Str  = "&&"
                        tokens[len(tokens)-1].Type = And
                    } else if t == Assign {
                        switch tokens[len(tokens)-1].Type {
                        case Colon:
                            tokens[len(tokens)-1].Str  = ":="
                            tokens[len(tokens)-1].Type = Def_var
                        case Not:
                            tokens[len(tokens)-1].Str  = "!="
                            tokens[len(tokens)-1].Type = Neq
                        case Assign:
                            tokens[len(tokens)-1].Str  = "=="
                            tokens[len(tokens)-1].Type = Eql
                        case Lss:
                            tokens[len(tokens)-1].Str  = "<="
                            tokens[len(tokens)-1].Type = Leq
                        case Grt:
                            tokens[len(tokens)-1].Str  = ">="
                            tokens[len(tokens)-1].Type = Geq
                        default:
                            tokens = append(tokens, Token{t, string(r), Pos{line, col}})
                        }
                    } else {
                        tokens = append(tokens, Token{t, string(r), Pos{line, col}})
                    }
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

    tokens = append(tokens, Token{EOF, "EOF", Pos{line, col}})

    if mlSkip {
        fmt.Fprintln(os.Stderr, "you have not terminated your comment (missing \"*/\")")
        os.Exit(1)
    }
}

func Cur() Token {
    return tokens[idx]
}

func Next() Token {
    idx++

    if idx >= len(tokens) {
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected end of file")
        os.Exit(1)
    }

    return tokens[idx]
}

func Peek() Token {
    if idx+1 >= len(tokens) {
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected end of file")
        os.Exit(1)
    }

    return tokens[idx+1]
}

func Peek2() Token {
    if idx+2 >= len(tokens) {
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected end of file")
        os.Exit(1)
    }

    return tokens[idx+2]
}

func Last() Token {
    if idx < 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected beginning of file (expected 1 word more at the start of the file)")
        os.Exit(1)
    }

    return tokens[idx-1]
}

func Last2() Token {
    if idx < 2 {
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected beginning of file (expected %d words more at the start of the file)\n", 2-idx)
        os.Exit(1)
    }

    return tokens[idx-2]
}

// returns Pos{ -1, -1 } if not found
func FindNext (t TokenType) Pos {
    for i := idx+1; i < len(tokens); i++ {
        if tokens[i].Type == t {
            return tokens[i].Pos
        }
    }

    return Pos{ -1, -1 }
}
