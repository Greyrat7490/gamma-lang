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
    return fmt.Sprintf("{%s %v}", v.Name.Str, v.Type)
}
func (v *LocalVar) Get() string {
    return fmt.Sprintf("QWORD [rbp-%d]", v.offset)
}
func (v *LocalVar) Gets() (string, string) {
    return fmt.Sprintf("QWORD [rbp-%d]", v.offset+8), fmt.Sprintf("QWORD [rbp-%d]", v.offset)
}
func (v *LocalVar) GetType() types.Type {
    return v.Type
}

func GetLastOffset() int {
    vars := scopes[len(scopes)-1].vars
    return vars[len(vars)-1].offset
}

func calcOffset(vartype types.Type) (offset int) {
    if !InGlobalScope() {
        if vartype.GetKind() == types.Str {
            offset = localVarOffset + 8
        } else {
            offset = localVarOffset + vartype.Size()
        }

        localVarOffset += vartype.Size()
    }

    return offset
}

func inCurScope(name string) bool {
    for _,v := range scopes[len(scopes)-1].vars {
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

    scopes[len(scopes)-1].vars = append(scopes[len(scopes)-1].vars, LocalVar{
        Name: varname,
        Type: vartype,
        offset: calcOffset(vartype),
    })
}

func defLocalVal(asm *os.File, v *LocalVar, val token.Token) {
    switch v.Type.GetKind() {
    case types.Str:
        strIdx := str.Add(val)
        asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], _str%d\n", v.offset+8, strIdx))
        asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], %d\n", v.offset, str.GetSize(strIdx)))

    case types.I32:
        asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], %s\n", v.offset, val.Str))

    case types.Bool:
        if val.Str == "true" { val.Str = "1" } else { val.Str = "0" }
        asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], %s\n", v.offset, val.Str))

    case types.Ptr:
        fmt.Fprintln(os.Stderr, "TODO defLocalVal PtrType")
        os.Exit(1)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name.Str)
        os.Exit(1)
    }
}

func defLocalExpr(asm *os.File, v *LocalVar, reg string) {
    asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], %s\n", v.offset, reg))
}
