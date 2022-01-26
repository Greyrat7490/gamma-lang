package main

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
