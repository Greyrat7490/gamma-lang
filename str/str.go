package str

import (
	"fmt"
	"os"
	"strings"
	"gorec/token"
)

type strLit struct {
    value string
    size int
}

var strLits []strLit

func GetSize(idx int) int {
    return strLits[idx].size
}

func Add(s token.Token) (idx int) {
    str, size := escape(s)

    if idx = find(str); idx != -1 {
        return idx
    }

    idx = len(strLits)
    strLits = append(strLits, strLit{str, size})

    return idx
}

func WriteStrLits(asm *os.File) {
    for i, str := range strLits {
        asm.WriteString(fmt.Sprintf("str%d: db %s\n", i, str.value))
    }
}

func escape(s token.Token) (string, int) {
    size := 0
    escape := false
    for _,r := range s.Str {
        if !escape {
            if r == '\\' {
                escape = true
                continue
            }
        } else {
            var i int
            switch r {
            case 't':
                i = int('\t')
            case 'r':
                i = int('\r')
            case 'n':
                i = int('\n')
            case '"':
                i = int('"')
            case '\\':
                i = int('\\')
            default:
                s.Pos.Col += size

                fmt.Fprintf(os.Stderr, "[ERROR] unknown escape sequence \"\\%c\"\n", r)
                fmt.Fprintln(os.Stderr, "\t" + s.At())
                os.Exit(1)
            }

            s.Str = strings.Replace(s.Str, fmt.Sprintf("\\%c", r), fmt.Sprintf("\",%d,\"", i), 1)
            escape = false
        }

        size++
    }

    size -= 2   // " at the beginning and the end

    s.Str = strings.ReplaceAll(s.Str, "\"\",", "")
    s.Str = strings.ReplaceAll(s.Str, ",\"\"", "")

    return s.Str, size
}

func find(s string) int {
    for i, v := range strLits {
        if v.value == s {
            return i
        }
    }

    return -1
}
