package types

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
