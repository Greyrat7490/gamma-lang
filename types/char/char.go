package char

// 0 -> unexpected escape sequence
func Escape(r rune) int {
    // TODO uint8 instead of int
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

// 0 -> unexpected escape sequence
func EscapeByte(c byte) uint8 {
    switch c {
    case 't':
        return uint8('\t')
    case 'r':
        return uint8('\r')
    case 'n':
        return uint8('\n')
    case '"':
        return uint8('"')
    case '\\':
        return uint8('\\')
    default:
        return 0
    }
}
