package vars

import (
    "fmt"
    "os"
    "gorec/types"
    "gorec/str"
)

const maxRegs int = 5
var availReg int = 0

var vars []Var
var globalDefs []string

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

func Declare(varname string, vartype types.Type) {
    // maybe implement shadowing later (TODO)
    if GetVar(varname) != nil {
        fmt.Fprintf(os.Stderr, "[ERROR] a variable with the name \"%s\" is already declared\n", varname)
        os.Exit(1)
    }

    v := Var{ Name: varname, Vartype: vartype }

    const _ uint = 2 - types.TypesCount
    switch vartype {
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

func Define(varname string, value string) {
    if v := GetVar(value); v != nil {
        DefineByVar(varname, value)
    } else {
        DefineByValue(varname, value)
    }
}

func DefineByValue(varname string, value string) {
    v := GetVar(varname)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) Op(OP_DEF_VAR) var \"%s\" is not declared\n", varname)
        os.Exit(1)
    }

    const _ uint = 2 - types.TypesCount
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
        globalDefs = append(globalDefs, fmt.Sprintf("mov %s, %s\n", Registers[v.Regs[0]].Name, value))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name)
        os.Exit(1)
    }
}

func DefineByVar(destVarName string, srcVarName string) {
    if v := GetVar(destVarName); v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", destVarName)
        os.Exit(1)
    } else {
        // TODO: check if var is defined
        if otherVar := GetVar(srcVarName); otherVar == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not declared\n", srcVarName)
            os.Exit(1)
        } else {
            for ri, r := range otherVar.Regs {
                globalDefs = append(globalDefs, fmt.Sprintf("mov %s, %s\n", Registers[v.Regs[ri]].Name, Registers[r].Name))
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
