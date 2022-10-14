package structLit

import (
	"gamma/cmpTime/constVal"
)

var structLits []structLit

type structLit struct {
    name string
    fields []constVal.ConstVal
}

func GetValues(structLitIdx uint64) []constVal.ConstVal {
    return structLits[structLitIdx].fields
}

func Add(name string, fields []constVal.ConstVal) int {
    structLits = append(structLits, structLit{ name: name, fields: fields })
    return len(structLits)-1
}
