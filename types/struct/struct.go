package structLit

import "gamma/token"

var structLits []structLit

type structLit struct {
    name string
    fields []token.Token
}

func GetValues(structLitIdx uint64) (res []token.Token) {
    return structLits[structLitIdx].fields
}

func Add(name string, fields []token.Token) int {
    structLits = append(structLits, structLit{ name: name, fields: fields })
    return len(structLits)-1
}
