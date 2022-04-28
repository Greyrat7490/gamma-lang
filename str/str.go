package str

import (
    "strings"
    "os"
    "fmt"
)

type data struct {
    value string
    size int
}

// later .data in general
var strLits []data

func GetSize(idx int) int {
    return strLits[idx].size
}

func Add(s string) (idx int) {
    i := strings.Count(s, "\\\"") * 7

    // replace escape characters
    s = strings.ReplaceAll(s, "\\\"", "\",0x22,\"")     //   \" -> ",0x22," (0x22 = ascii of ")
    s = strings.ReplaceAll(s, "\\\\", "\\")             //   \\ -> \

    size := len(s) - i - 2 + 1 // -2 (don't count ""), -i (don't count ",0x22,"), +1 (for \n)
    s += ",0xa"

    idx = len(strLits)
    strLits = append(strLits, data{s, size})

    return idx
}

func WriteStrLits(asm *os.File) {
    for i, str := range strLits {
        asm.WriteString(fmt.Sprintf("str%d: db %s\n", i, str.value))
    }
}
