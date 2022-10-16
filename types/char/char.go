package char

// 0 -> unexpected escape sequence
func Escape(r rune) uint8 {
    switch r {
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
