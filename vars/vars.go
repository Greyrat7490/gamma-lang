package vars

import (
    "fmt"
    "os"
    "strconv"
    "gorec/types"
    "gorec/str"
)

type reg struct {
    Name string
    isAddr bool
    value int      // either an actual value or an address(index)
}

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

const maxRegs int = 5
var availReg int = 0

var vars []Var
var globalDefs []string

type Var struct {
    Name string
    Regs []int
    Vartype types.Type
}

func Get(varname string) *Var {
    for _, v := range vars {
        if v.Name == varname {
            return &v
        }
    }

    return nil
}

func Declare(v Var) {
    switch v.Vartype {
    case types.Str:
        if availReg + 1 >= maxRegs {
            fmt.Fprintf(os.Stderr, "[ERROR] not enough registers left for var \"%s\"(string)", v.Name)
            os.Exit(1)
        }

        v.Regs = []int{ availReg, availReg+1 }

        vars = append(vars, v)
        availReg += 2
    case types.I32:
        if availReg >= maxRegs {
            fmt.Fprintf(os.Stderr, "[ERROR] not enough registers left for var \"%s\"(i32)", v.Name)
            os.Exit(1)
        }

        v.Regs = []int{ availReg }
        vars = append(vars, v)
        availReg++
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name)
        os.Exit(1)
    }
}

func Define(isLit bool, v Var, value string) {
    if isLit {
        switch v.Vartype {
        case types.Str:
            if len(v.Regs) != 2 {
                fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) string variable should use 2 registers\n")
                os.Exit(1)
            }

            strIdx := str.Add(value)
            globalDefs = append(globalDefs, fmt.Sprintf("mov %s, str%d\n", Registers[v.Regs[0]].Name, strIdx))
            globalDefs = append(globalDefs, fmt.Sprintf("mov %s, %d\n", Registers[v.Regs[1]].Name, str.GetSize(strIdx)))

        case types.I32:
            i, _ := strconv.Atoi(value)
            globalDefs = append(globalDefs, fmt.Sprintf("mov %s, %d\n", Registers[v.Regs[0]].Name, i))

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name)
            os.Exit(1)
        }
    } else {
        // TODO: check if var is defined
        if otherVar := Get(value); otherVar != nil {
            for ri, r := range otherVar.Regs {
                globalDefs = append(globalDefs, fmt.Sprintf("mov %s, %s\n", Registers[v.Regs[ri]].Name, Registers[r].Name))
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not declared\n", value)
            os.Exit(1)
        }
    }
}

func Add(v Var) {
    vars = append(vars, v)
}

func Remove(varname string) {
    if len(vars) == 1 && vars[0].Name == varname {
        vars = []Var{}
        return
    }

    for i, v := range vars {
        if v.Name == varname {
            vars[i] = vars[len(vars)-1]
            vars = vars[:len(vars)-1]
            return
        }
    }
}

func WriteGlobalVars(asm *os.File) {
    for _, s := range globalDefs {
        asm.WriteString(s)
    }
}
