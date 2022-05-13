package vars

import (
    "fmt"
    "os"
    "gorec/str"
    "gorec/types"
    "gorec/token"
)

const maxRegs int = 6           // 4
var availReg int = 2            // rax and rbx are reserved for expr (only tmp maybe)

var IsGlobalScope bool = true

var vars []Var
var globalScope []string

// TODO: register allocator for variables
var Registers []reg = []reg {
    {Name: "rax"},
    {Name: "rbx"},
    {Name: "rcx"},
    {Name: "rdx"},
    {Name: "r8"},
    {Name: "r9"},
    {Name: "r10"},
    {Name: "r11"},
}


type reg struct {
    Name string
    isAddr bool
    value int      // either an actual value or an address(index)
}

type Var struct {
    Name string
    Regs []int
    Vartype types.Type
}

func ShowVars() {
    for _, v := range vars {
        fmt.Printf("%s { type:%s regs:%v }\n", v.Name, v.Vartype.Readable(), v.Regs)
    }
}

func GetVar(varname string) *Var {
    for _, v := range vars {
        if v.Name == varname {
            return &v
        }
    }

    return nil
}

func Declare(varname token.Token, vartype types.Type) {
    // maybe implement shadowing later (TODO)
    if GetVar(varname.Str) != nil {
        fmt.Fprintf(os.Stderr, "[ERROR] a variable with the name \"%s\" is already declared\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    v := Var{ Name: varname.Str, Vartype: vartype }

    const _ uint = 3 - types.TypesCount
    switch vartype {
    case types.Str:
        if availReg + 1 >= maxRegs {
            fmt.Fprintf(os.Stderr, "[ERROR] not enough registers left for var \"%s\"(string)", v.Name)
            fmt.Fprintln(os.Stderr, "\t" + varname.At())
            os.Exit(1)
        }

        v.Regs = []int{ availReg, availReg+1 }

        vars = append(vars, v)
        availReg += 2
    case types.I32:
        if availReg >= maxRegs {
            fmt.Fprintf(os.Stderr, "[ERROR] not enough registers left for var \"%s\"(i32)", v.Name)
            fmt.Fprintln(os.Stderr, "\t" + varname.At())
            os.Exit(1)
        }

        v.Regs = []int{ availReg }

        vars = append(vars, v)
        availReg++
    case types.Bool:
        if availReg >= maxRegs {
            fmt.Fprintf(os.Stderr, "[ERROR] not enough registers left for var \"%s\"(bool)", v.Name)
            fmt.Fprintln(os.Stderr, "\t" + varname.At())
            os.Exit(1)
        }

        v.Regs = []int{ availReg }

        vars = append(vars, v)
        availReg++
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }
}

// TODO: merge Define and Assign
func Define(asm *os.File, varname token.Token, value token.Token) {
    if v := GetVar(value.Str); v != nil {
        DefineByVar(asm, varname, value)
    } else {
        DefineByValue(asm, varname, value)
    }
}

func DefineByValue(asm *os.File, varname token.Token, value token.Token) {
    v := GetVar(varname.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) Op(OP_DEF_VAR) var \"%s\" is not declared\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    if v.Name == value.Str {
        fmt.Fprintln(os.Stderr, "[ERROR] cannot define a variable with itself")
        fmt.Fprintln(os.Stderr, "\t" + value.At())
        os.Exit(1)
    }

    if value.Type == token.Boolean || value.Type == token.Number || value.Type == token.Str {
        const _ uint = 3 - types.TypesCount
        switch v.Vartype {
        case types.Str:
            if len(v.Regs) != 2 {
                fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) string variable should use 2 registers\n")
                os.Exit(1)
            }

            strIdx := str.Add(value.Str)
            WriteVar(asm, fmt.Sprintf("mov %s, str%d\n", Registers[v.Regs[0]].Name, strIdx))
            WriteVar(asm, fmt.Sprintf("mov %s, %d\n", Registers[v.Regs[1]].Name, str.GetSize(strIdx)))

        case types.I32:
            WriteVar(asm, fmt.Sprintf("mov %s, %s\n", Registers[v.Regs[0]].Name, value.Str))

        case types.Bool:
            if value.Str == "true" {
                WriteVar(asm, fmt.Sprintf("mov %s, %d\n", Registers[v.Regs[0]].Name, 1))
            } else {
                WriteVar(asm, fmt.Sprintf("mov %s, %d\n", Registers[v.Regs[0]].Name, 0))
            }

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name)
        }
    }
}

func DefineByVar(asm *os.File, destVar token.Token, srcVar token.Token) {
    if v := GetVar(destVar.Str); v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", destVar.Str)
        fmt.Fprintln(os.Stderr, "\t" + destVar.At())
        os.Exit(1)
    } else {
        // TODO: check if var is defined
        if otherVar := GetVar(srcVar.Str); otherVar == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not declared\n", srcVar.Str)
            fmt.Fprintln(os.Stderr, "\t" + srcVar.At())
            os.Exit(1)
        } else {
            for ri, r := range otherVar.Regs {
                WriteVar(asm, fmt.Sprintf("mov %s, %s\n", Registers[v.Regs[ri]].Name, Registers[r].Name))
            }
        }
    }
}

func AssignByReg(asm *os.File, destVar token.Token, reg string) {
    if v := GetVar(destVar.Str); v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", destVar.Str)
        fmt.Fprintln(os.Stderr, "\t" + destVar.At())
        os.Exit(1)
    } else {
        WriteVar(asm, fmt.Sprintf("mov %s, %s\n", Registers[v.Regs[0]].Name, reg))
    }
}

func Assign(asm *os.File, varname token.Token, value token.Token) {
    v := GetVar(varname.Str)

    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) Op(OP_ASSIGN_VAR) var \"%s\" is not declared\n", varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + varname.At())
        os.Exit(1)
    }

    // skip assigning a variable to itself (redundant)
    if v.Name == value.Str {
        return
    }

    if value.Type == token.Number || value.Type == token.Str {
        const _ uint = 3 - types.TypesCount
        switch v.Vartype {
        case types.Str:
            if len(v.Regs) != 2 {
                fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) string variable should use 2 registers\n")
                os.Exit(1)
            }

            strIdx := str.Add(value.Str)
            WriteVar(asm, fmt.Sprintf("mov %s, str%d\n", Registers[v.Regs[0]].Name, strIdx))
            WriteVar(asm, fmt.Sprintf("mov %s, %d\n", Registers[v.Regs[1]].Name, str.GetSize(strIdx)))

        case types.I32, types.Bool:
            WriteVar(asm, fmt.Sprintf("mov %s, %s\n", Registers[v.Regs[0]].Name, value.Str))

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name)
            os.Exit(1)
        }
    } else {
        // TODO: check if var is defined
        if otherVar := GetVar(value.Str); otherVar != nil {
            for ri, r := range otherVar.Regs {
                WriteVar(asm, fmt.Sprintf("mov %s, %s\n", Registers[v.Regs[ri]].Name, Registers[r].Name))
            }
        }
    }
}

func Add(v Var) {
    vars = append(vars, v)
}

func Remove(varname string) {
    if len(vars) == 1 && vars[0].Name == varname {
        vars = []Var{}
        availReg = maxRegs
        return
    }

    for _, v := range vars {
        if v.Name == varname {
            v = vars[len(vars)-1]
            vars = vars[:len(vars)-1]

            // "free" registers
            if v.Vartype == types.Str {
                availReg-=2
            } else {
                availReg--
            }
            return
        }
    }

}

func WriteVar(asm *os.File, s string) {
    if IsGlobalScope {
        globalScope = append(globalScope, s)
    } else {
        asm.WriteString(s)
    }
}

func WriteGlobalScope(asm *os.File) {
    for _, s := range globalScope {
        asm.WriteString(s)
    }
}
