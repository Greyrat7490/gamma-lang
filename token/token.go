package token

import (
	"bufio"
	"fmt"
	"gamma/types"
	"os"
	"strconv"
)

type TokenType uint
const (
    Unknown TokenType = iota

    EOF             // end of file

    Name            // var/func name
    Typename        // i32, str, bool
    Str             // "string"
    Char            // 'a'
    Number          // 1234, 0xffff
    Boolean         // true/false

    Plus            // +
    Minus           // -
    Mul             // *
    Div             // /
    Mod             // %

    Shl             // <<
    Shr             // >>
    Xor             // ^
    BitOr           // |
    BitNot          // ~

    PlusEq          // +=
    MinusEq         // -=
    MulEq           // *=
    DivEq           // /=
    ModEq           // %=

    ShlEq           // <<=
    ShrEq           // >>=
    BitAndEq        // &=
    BitOrEq         // |=
    XorEq           // ^=

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

    Arrow           // ->

    Dot             // .
    Comma           // ,
    Colon           // :
    SemiCol         // ;

    Comment         // // ..., /* ... */

    DefVar          // :=
    DefConst        // ::
    Assign          // =
    Fn              // fn
    ConstFn         // cfn
    Ret             // ret
    If              // if
    Elif            // elif
    Else            // else
    While           // while
    For             // for
    Break           // break
    Continue        // continue
    Through         // through
    Struct          // struct
    Interface       // interface
    Enum            // Enum
    Impl            // impl
    Self            // self
    SelfType        // Self
    Import          // import
    XSwitch         // $
    As              // as

    TokenTypeCount uint = iota
)

func ToTokenType(s string, p Pos) TokenType {
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

    case "<<":
        return Shl
    case ">>":
        return Shr
    case "|":
        return BitOr
    case "^":
        return Xor
    case "~":
        return BitNot

    case "+=":
        return PlusEq
    case "-=":
        return MinusEq
    case "*=":
        return MulEq
    case "/=":
        return DivEq
    case "%=":
        return ModEq
    case "<<=":
        return ShlEq
    case ">>=":
        return ShrEq
    case "&=":
        return BitAndEq
    case "|=":
        return BitOrEq
    case "^=":
       return XorEq

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

    case "->":
        return Arrow

    case ".":
        return Dot
    case ",":
        return Comma
    case ";":
        return SemiCol
    case ":":
        return Colon

    case "//", "/*", "*/":
        return Comment

    case ":=":
        return DefVar
    case "::":
        return DefConst
    case "=":
        return Assign
    case "fn":
        return Fn
    case "cfn":
        return ConstFn
    case "ret":
        return Ret
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
    case "struct":
        return Struct
    case "interface":
        return Interface
    case "enum":
        return Enum
    case "impl":
        return Impl
    case "self":
        return Self
    case "Self":
        return SelfType
    case "import":
        return Import
    case "$":
        return XSwitch
    case "as":
        return As

    default:
        if types.ToBaseType(s) != nil {
            return Typename
        } else {
            switch {
            case s[0] == '"' && s[len(s) - 1] == '"':
                return Str
            case s[0] == '\'' && s[len(s) - 1] == '\'':
                return Char
            default:
                if _, err := strconv.ParseUint(s, 0, 64); err == nil {
                    return Number
                } else {
                    if e,ok := err.(*strconv.NumError); ok && e.Err == strconv.ErrRange {
                        fmt.Fprintf(os.Stderr, "[ERROR] %s is too big (out of u64 range)\n", s)
                        fmt.Fprintln(os.Stderr, "\t" + p.At())
                        os.Exit(1);
                    }
                }
            }
        }

        return Name
    }
}

func (t TokenType) String() string {
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

    case Shl:
        return "Shl"
    case Shr:
        return "Shr"
    case BitOr:
        return "BitOr"
    case Xor:
        return "Xor"
    case BitNot:
        return "BitNot"

    case PlusEq:
        return "PlusEq"
    case MinusEq:
        return "MinusEq"
    case MulEq:
        return "MulEq"
    case DivEq:
        return "DivEq"
    case ModEq:
        return "ModEq"
    case ShlEq:
        return "ShlEq"
    case ShrEq:
        return "ShrEq"
    case BitAndEq:
        return "BitAndEq"
    case BitOrEq:
        return "BitOrEq"
    case XorEq:
       return "XorEq"

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

    case Arrow:
        return "Arrow"

    case Dot:
        return "Dot"
    case Comma:
        return "Comma"
    case SemiCol:
        return "SemiCol"
    case Colon:
        return "Colon"

    case Comment:
        return "Comment"

    case DefVar:
        return "DefVar"
    case DefConst:
        return "DefConst"
    case Assign:
        return "Assign"
    case Fn:
        return "Fn"
    case ConstFn:
        return "ConstFn"
    case Ret:
        return "Ret"
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
    case Struct:
        return "Struct"
    case Interface:
        return "Interface"
    case Enum:
        return "Enum"
    case Impl:
        return "Impl"
    case Self:
        return "Self"
    case SelfType:
        return "SelfType"
    case Import:
        return "Import"
    case XSwitch:
        return "XSwitch"
    case As:
        return "As"

    case Typename:
        return "Typename"
    case Str:
        return "Str"
    case Char:
        return "Char"
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
    File string
}

func (p Pos) At() string {
    return fmt.Sprintf("at: %s:%d:%d", p.File, p.Line, p.Col)
}

type Token struct {
    Type TokenType
    Str string
    Pos Pos
}

func (t Token) String() string {
    return fmt.Sprintf("%s (%v)", t.Str, t.Type)
}

func (t Token) At() string {
    return t.Pos.At()
}

type Tokens struct {
    tokens []Token
    idx int
    path string
    lastImport bool
}

func (t *Tokens) split(s string, start int, end int, line int, file string) {
    if start != end {
        s := s[start:end]
        typ := ToTokenType(s, Pos{line, start+1, file})

        t.tokens = append(t.tokens, Token{typ, s, Pos{line, start+1, file} })
    }
}

func (t *Tokens) GetPath() string {
    return t.path
}

func (t *Tokens) IsFileStart() bool {
    return t.lastImport
}

func (t *Tokens) SetLastImport() {
    if t.Peek().Type != Import {
        t.lastImport = false
    }
}

func Tokenize(path string, src *os.File) (tokens Tokens) {
    tokens.idx = -1
    tokens.path = path
    tokens.lastImport = true

    scanner := bufio.NewScanner(src)

    comment := false
    mlComment := false
    strLit := false
    escape := false

    line := ""
    lineNum := 0
    for lineNum = 1; scanner.Scan(); lineNum++ {
        line = scanner.Text()

        start := 0
        comment = false
        for i := 0; i < len(line); i++ {
            // in single line comment
            if comment {
                break
            }

            // in multiline comment
            if mlComment {
                if i+2 <= len(line) && line[i:i+2] == "*/" {
                    mlComment = false
                    start = i+2
                    i++
                }

                continue
            }

            // in string literal
            if strLit {
                if escape {
                    escape = false
                } else {
                    if line[i] == '"' {
                        strLit = false
                    } else if line[i] == '\\' {
                        escape = true
                    }
                }

                continue
            }

            switch line[i] {
            // start string literal
            case '"':
                strLit = true

            // char literal
            case '\'':
                if line[start+1] == '\\' {
                    i++
                }
                tokens.split(line, start, i+3, lineNum, path)
                start = i+3
                i += 2

            // split at space
            case ' ', '\t':
                tokens.split(line, start, i, lineNum, path)
                start = i+1

            // split at <<=, >>=
            case '<', '>':
                if i+3 <= len(line) {
                    s := line[i:i+3]
                    switch s {
                    case "<<=", ">>=":
                        tokens.split(line, start, i, lineNum, path)
                        tokens.tokens = append(tokens.tokens, Token{ ToTokenType(s, Pos{lineNum, i+1, path}), s, Pos{lineNum, i+1, path} })
                        start = i+3
                        i+=2
                        continue
                    }
                }

                fallthrough

            // split at //, /*, :=, ::, <=, >=, ==, !=, &&, ->, <<, >>, +=, -=, *=, /=, &=, |=, %=, ^=
            case '/', ':', '=', '-', '!', '&', '|', '+', '*', '%', '^':
                if i+2 <= len(line) {
                    s := line[i:i+2]
                    switch s {
                    // start single line comment
                    case "//":
                        tokens.split(line, start, i, lineNum, path)
                        comment = true
                        i++
                        continue
                    // start multiline comment
                    case "/*":
                        tokens.split(line, start, i, lineNum, path)
                        mlComment = true
                        i++
                        continue

                    case "&&", "||", ":=", "::", "!=", "==", "<=", ">=", "->", "<<", ">>", "+=", "-=", "*=", "/=", "&=", "|=", "%=", "^=":
                        tokens.split(line, start, i, lineNum, path)
                        tokens.tokens = append(tokens.tokens, Token{ ToTokenType(s, Pos{lineNum, i+1, path}), s, Pos{lineNum, i+1, path} })
                        start = i+2
                        i++
                        continue
                    }
                }

                fallthrough

            // split at non space char (and keep char)
            case '(', ')', '{', '}', '[', ']', '.', ',', ';', '$', '~':
                tokens.split(line, start, i, lineNum, path)
                tokens.tokens = append(tokens.tokens, Token{ ToTokenType(string(line[i]), Pos{lineNum, i+1, path}), string(line[i]), Pos{lineNum, i+1, path} })
                start = i+1
            }
        }

        if !comment && !mlComment && len(line) > start {
            tokens.split(line, start, len(line), lineNum, path)
        }
    }

    tokens.tokens = append(tokens.tokens, Token{EOF, "EOF", Pos{ Line: lineNum, Col: len(line), File: path }})

    if strLit {
        fmt.Fprintln(os.Stderr, "string literal not terminated (missing '\"')")
        os.Exit(1)
    }
    if mlComment {
        fmt.Fprintln(os.Stderr, "comment not terminated (missing \"*/\")")
        os.Exit(1)
    }

    return
}

func (t *Tokens) Cur() Token {
    return t.tokens[t.idx]
}

func (t *Tokens) Next() Token {
    t.idx++

    if t.idx >= len(t.tokens) {
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected end of file")
        os.Exit(1)
    }

    return t.tokens[t.idx]
}

func (t *Tokens) Peek() Token {
    if t.idx+1 >= len(t.tokens) {
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected end of file")
        os.Exit(1)
    }

    return t.tokens[t.idx+1]
}

func (t *Tokens) Peek2() Token {
    if t.idx+2 >= len(t.tokens) {
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected end of file")
        os.Exit(1)
    }

    return t.tokens[t.idx+2]
}

func (t *Tokens) Last() Token {
    if t.idx < 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected beginning of file (expected 1 word more at the start of the file)")
        os.Exit(1)
    }

    return t.tokens[t.idx-1]
}

func (t *Tokens) Last2() Token {
    if t.idx < 2 {
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected beginning of file (expected %d words more at the start of the file)\n", 2-t.idx)
        os.Exit(1)
    }

    return t.tokens[t.idx-2]
}

// returns Pos{ -1, -1 } if not found
func (t *Tokens) FindNext (typ TokenType) Pos {
    for i := t.idx+1; i < len(t.tokens); i++ {
        if t.tokens[i].Type == typ {
            return t.tokens[i].Pos
        }
    }

    return Pos{ -1, -1, "" }
}

func (t *Tokens) SaveIdx() int {
    return t.idx
}

func (t *Tokens) ResetIdx(idx int) {
    t.idx = idx
}
