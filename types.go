package main

import (
    "strings"
)

type gType int
const (
    i32 gType = iota
    str gType = iota
)

func (t gType) readable() string {
    switch t {
    case str:
        return "str"
    case i32:
        return "i32"
    default:
        return ""
    }
}

// -1 if string does not contain a valid type
func toType(s string) gType {
    switch s {
    case "str":
        return str
    case "i32":
        return i32
    default:
        return -1
    }
}

type data struct {
    value string
    size int
}

// later .data in general
var strLits []data

func addStrLit(s string) {
    i := strings.Count(s, "\\\"") * 7

    // replace escape characters
    s = strings.ReplaceAll(s, "\\\"", "\",0x22,\"")     //   \" -> ",0x22," (0x22 = ascii of ")
    s = strings.ReplaceAll(s, "\\\\", "\\")             //   \\ -> \

    size := len(s) - i - 2 + 1 // -2 (don't count ""), -i (don't count ",0x22,"), +1 (for \n)
    s += ",0xa"

    strLits = append(strLits, data{s, size})
}
