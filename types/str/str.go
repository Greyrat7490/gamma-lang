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
    repr string
    size uint
}

var strLits []strLit
var strs map[string]uint64 = make(map[string]uint64)


func GetSize(idx uint64) uint {
    return strLits[idx].size
}

func Add(s token.Token) uint64 {
    if idx,ok := strs[s.Str]; ok {
        return idx
    }

    idx := uint64(len(strLits))
    strLits = append(strLits, escape(s))
    strs[s.Str] = idx

    return idx
}

func Gen() {
    for idx,lit := range strLits {
        nasm.AddRodata(fmt.Sprintf("_str%d: db %s", idx, lit.repr))
    }
}

func escape(s token.Token) strLit {
    size := uint(0)
    escape := false
    for _,r := range s.Str {
        if !escape {
            if r == '\\' {
                escape = true
                continue
            }
        } else {
            if i,ok := char.Escape(r); ok {
                s.Str = strings.Replace(s.Str, fmt.Sprintf("\\%c", r), fmt.Sprintf("\",%d,\"", i), 1)
                escape = false
            } else {
                s.Pos.Col += int(size)

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

    return strLit{ s.Str, size }
}

func CmpStrLits(idx1 uint64, idx2 uint64) bool {
    if idx1 == idx2 {
        return true
    }

    return strLits[idx1].repr == strLits[idx2].repr
}
