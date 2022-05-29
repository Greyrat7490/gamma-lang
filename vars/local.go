package vars

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/str"
)

var localVarOffset int = 0

type LocalVar struct {
    Name token.Token
    Type types.Type
    offset int
}

func (v *LocalVar) String() string {
    return fmt.Sprintf("{%s %s}", v.Name.Str, v.Type.Readable())
}
func (v *LocalVar) Get() string {
    return fmt.Sprintf("QWORD [rbp-%d]", v.offset)
}
func (v *LocalVar) Gets() (string, string) {
    return fmt.Sprintf("QWORD [rbp-%d]", v.offset), fmt.Sprintf("QWORD [rbp-%d]", v.offset+8)
}
func (v *LocalVar) GetType() types.Type {
    return v.Type
}

func GetLastOffset() int {
    return scopes[curScope].vars[len(scopes[curScope].vars)-1].offset
}

func calcOffset(vartype types.Type) (offset int) {
    if !InGlobalScope() {
        if vartype == types.Str {
            offset = localVarOffset + 8
        } else {
            offset = localVarOffset + vartype.Size()
        }

        localVarOffset += vartype.Size()
    }

    return offset
}

func inCurScope(name string) bool {
    for _,v := range scopes[curScope].vars {
        if v.Name.Str == name {
            return true
        }
    }

    return false
}

func declareLocal(varname token.Token, vartype types.Type) {
    if inCurScope(varname.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] local var \"%s\" is already declared in this scope\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    scopes[curScope].vars = append(scopes[curScope].vars, LocalVar{
        Name: varname,
        Type: vartype,
        offset: calcOffset(vartype),
    })
}

func defLocalVal(asm *os.File, v *LocalVar, val string) {
    const _ uint = 3 - types.TypesCount
    switch v.Type {
    case types.Str:
        strIdx := str.Add(val)
        asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], str%d\n", v.offset, strIdx))
        asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], %d\n", v.offset+8, str.GetSize(strIdx)))

    case types.I32:
        asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], %s\n", v.offset, val))

    case types.Bool:
        if val == "true" { val = "1" } else { val = "0" }
        asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], %s\n", v.offset, val))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name.Str)
    }
}

func defLocalExpr(asm *os.File, v *LocalVar, reg string) {
    asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], %s\n", v.offset, reg))
}
