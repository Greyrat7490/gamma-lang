package char

func Escape(r rune) int {
    switch r {
    case 't':
        return int('\t')
    case 'r':
        return int('\r')
    case 'n':
        return int('\n')
    case '"':
        return int('"')
    case '\\':
        return int('\\')
    default:
        return -1
    }
}

func EscapeByte(c byte) int {
    switch c {
    case 't':
        return int('\t')
    case 'r':
        return int('\r')
    case 'n':
        return int('\n')
    case '"':
        return int('"')
    case '\\':
        return int('\\')
    default:
        return -1
    }
}
