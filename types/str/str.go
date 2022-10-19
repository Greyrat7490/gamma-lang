package str

import (
    "os"
    "fmt"
    "strings"
    "gamma/token"
    "gamma/types/char"
    "gamma/gen/asm/x86_64/nasm"
)

type strLit struct {
    value string
    size int
}

var strLits []strLit

func GetSize(idx int) int {
    return strLits[idx].size
}

func Add(s token.Token) uint64 {
    str, size := escape(s)

    if idx,ok := find(str); ok {
        return idx
    }

    idx := uint64(len(strLits))
    strLits = append(strLits, strLit{str, size})
    nasm.AddRodata(GetDefineStr(idx))

    return idx
}

func GetDefineStr(idx uint64) string {
    return fmt.Sprintf("_str%d: db %s", idx, strLits[idx].value)
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
            if i := char.Escape(r); i != 0 {
                s.Str = strings.Replace(s.Str, fmt.Sprintf("\\%c", r), fmt.Sprintf("\",%d,\"", i), 1)
                escape = false
            } else {
                s.Pos.Col += size

                fmt.Fprintf(os.Stderr, "[ERROR] unknown escape sequence \"\\%c\"\n", r)
                fmt.Fprintln(os.Stderr, "\t" + s.At())
                os.Exit(1)
            }
        }

        size++
    }

    size -= 2   // " at the beginning and the end

    s.Str = strings.ReplaceAll(s.Str, "\"\",", "")
    s.Str = strings.ReplaceAll(s.Str, ",\"\"", "")

    return s.Str, size
}

func find(s string) (uint64, bool) {
    for i, v := range strLits {
        if v.value == s {
            return uint64(i), true
        }
    }

    return 0, false
}
