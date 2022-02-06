package prs

import (
    "fmt"
    "unicode"
    "os"
)

type Word struct {
    Str string
    Line int
    Col int
    // later filename
}

func (w Word) At() string {
    return fmt.Sprintf("at line: %d, col: %d", w.Line, w.Col)
}

var Words []Word

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
                    Words = append(Words, Word{file[start:i], line, col + start - i})
                }
                start = i + 1

                if r == '(' || r == ')' || r == '{' || r == '}' {
                    Words = append(Words, Word{string(r), line, col - 1})
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
