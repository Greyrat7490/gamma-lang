package types

import (
    "strings"
)

type Type int
const (
    I32 Type = iota
    Str Type = iota
)

func (t Type) Readable() string {
    switch t {
    case Str:
        return "str"
    case I32:
        return "i32"
    default:
        return ""
    }
}

// -1 if string does not contain a valid type
func ToType(s string) Type {
    switch s {
    case "str":
        return Str
    case "i32":
        return I32
    default:
        return -1
    }
}

type data struct {
    Value string
    Size int
}

// later .data in general
var StrLits []data

func AddStrLit(s string) {
    i := strings.Count(s, "\\\"") * 7

    // replace escape characters
    s = strings.ReplaceAll(s, "\\\"", "\",0x22,\"")     //   \" -> ",0x22," (0x22 = ascii of ")
    s = strings.ReplaceAll(s, "\\\\", "\\")             //   \\ -> \

    size := len(s) - i - 2 + 1 // -2 (don't count ""), -i (don't count ",0x22,"), +1 (for \n)
    s += ",0xa"

    StrLits = append(StrLits, data{s, size})
}
