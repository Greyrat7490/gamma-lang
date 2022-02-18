package vars

import (
    "fmt"
    "os"
    "strconv"
    "gorec/types"
    "gorec/str"
    "gorec/parser"
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

func IsLit(w string) bool {
    if w[0] == '"' && w[len(w) - 1] == '"' {
        return true
    }

    if _, err := strconv.Atoi(w); err == nil {
        return true
    }

    return false
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

func Declare(op *prs.Op) {
    if op.Type != prs.OP_DEC_VAR {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) OpType should be OP_DEC_VAR\n")
        os.Exit(1)
    }

    if len(op.Operants) != 2 {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) Op(OP_DEC_VAR) should have 2 operants\n")
        os.Exit(1)
    }

    varname := op.Operants[0]
    vartype := types.ToType(op.Operants[1])

    // maybe implement shadowing later (TODO)
    if GetVar(varname) != nil {
        fmt.Fprintf(os.Stderr, "[ERROR] a variable with the name \"%s\" is already declared\n", op.Token.Str)
        fmt.Fprintln(os.Stderr, "\t" + op.Token.At())
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

func Define(op *prs.Op) {
    if op.Type != prs.OP_DEF_VAR {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) OpType should be OP_DEF_VAR\n")
        os.Exit(1)
    }

    if len(op.Operants) != 2 {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) Op(OP_DEF_VAR) should have 2 operants\n")
        os.Exit(1)
    }

    v := GetVar(op.Operants[0])
    value := op.Operants[1]

    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) Op(OP_DEF_VAR) var \"%s\" is not declared\n", op.Operants[0])
        os.Exit(1)
    }

    if IsLit(value) {
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
    } else {
        // TODO: check if var is defined
        if otherVar := GetVar(value); otherVar != nil {
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
