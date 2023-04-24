package char

func Escape(r rune) (uint8, bool) {
    switch r {
    case '0':
        return 0, true
    case 't':
        return uint8('\t'), true
    case 'r':
        return uint8('\r'), true
    case 'n':
        return uint8('\n'), true
    case '"':
        return uint8('"'), true
    case '\\':
        return uint8('\\'), true
    case '\'':
        return uint8('\''), true
    case '{':
        return uint8('{'), true
    case '}':
        return uint8('}'), true
    default:
        return 0, false
    }
}

func EscapeByte(c byte) (uint8, bool) {
    switch c {
    case '0':
        return 0, true
    case 't':
        return uint8('\t'), true
    case 'r':
        return uint8('\r'), true
    case 'n':
        return uint8('\n'), true
    case '"':
        return uint8('"'), true
    case '\\':
        return uint8('\\'), true
    case '\'':
        return uint8('\''), true
    case '{':
        return uint8('{'), true
    case '}':
        return uint8('}'), true
    default:
        return 0, false
    }
}
