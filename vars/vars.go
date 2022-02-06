package vars

import (
    "fmt"
    "os"
    "strconv"
    "gorec/parser"
    "gorec/types"
)

type reg struct {
    Name string
    IsAddr bool
    Value int      // either an actual value or an address(index)
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

var vars []Variable
var globalDefs []string

type Variable struct {
    Name string
    Regs []int
    Vartype types.Type
}

func GetVar(varname string) *Variable {
    for _, v := range vars {
        if v.Name == varname {
            return &v
        }
    }

    return nil
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

func DeclareVar(words []prs.Word, i int) int {
    if len(words) < i + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] neither name nor type provided for the variable declaration")
        fmt.Fprintln(os.Stderr, "\t" + words[i].At())
        os.Exit(1)
    }
    if len(words) < i + 2 {
        fmt.Fprintln(os.Stderr, "[ERROR] no name or type provided for the variable")
        fmt.Fprintln(os.Stderr, "\t" + words[i+1].At())
        os.Exit(1)
    }
    if len(vars) >= len(Registers) {
        fmt.Fprintf(os.Stderr, "[ERROR] a maximum of only %d variables is allowed yet\n", len(Registers))
        fmt.Fprintln(os.Stderr, "\t" + words[i+1].At())
        os.Exit(1)
    }
    // maybe implement shadowing later (TODO)
    if GetVar(words[i+1].Str) != nil {
        fmt.Fprintf(os.Stderr, "[ERROR] a variable with the name \"%s\" is already declared\n", words[i+1].Str)
        fmt.Fprintln(os.Stderr, "\t" + words[i+1].At())
        os.Exit(1)
    }

    switch types.ToType(words[i+2].Str) {
    case types.Str:
        if availReg + 1 >= maxRegs {
            fmt.Fprintf(os.Stderr, "[ERROR] not enough registers left for var \"%s\"(string)", words[i+1].Str)
            fmt.Fprintln(os.Stderr, "\t" + words[i+1].At())
            os.Exit(1)
        }

        vars = append(vars, Variable{words[i+1].Str, []int{ availReg, availReg+1 }, types.Str})
        availReg += 2
    case types.I32:
        if availReg >= maxRegs {
            fmt.Fprintf(os.Stderr, "[ERROR] not enough registers left for var \"%s\"(i32)", words[i+1].Str)
            fmt.Fprintln(os.Stderr, "\t" + words[i+1].At())
            os.Exit(1)
        }

        vars = append(vars, Variable{words[i+1].Str, []int{ availReg }, types.I32})
        availReg++
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", words[i+2].Str)
        fmt.Fprintln(os.Stderr, "\t" + words[i+2].At())
        os.Exit(1)
    }

    return i + 2
}

func DefineVar(words []prs.Word, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] no value provided to define the variable\n")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    if IsLit(words[idx+1].Str) {
        if v := GetVar(words[idx-2].Str); v != nil {
            switch v.Vartype {
            case types.Str:
                if len(v.Regs) != 2 {
                    fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) string variable should use 2 registers\n")
                    fmt.Fprintln(os.Stderr, "\t" + words[idx-2].At())
                    os.Exit(1)
                }

                globalDefs = append(globalDefs, fmt.Sprintf("mov %s, str%d\n", Registers[v.Regs[0]].Name, len(types.StrLits)))
                types.AddStrLit(words[idx+1].Str)
                globalDefs = append(globalDefs, fmt.Sprintf("mov %s, %d\n", Registers[v.Regs[1]].Name, types.StrLits[len(types.StrLits)-1].Size))

            case types.I32:
                i, _ := strconv.Atoi(words[idx+1].Str)
                globalDefs = append(globalDefs, fmt.Sprintf("mov %s, %d\n", Registers[v.Regs[0]].Name, i))

            default:
                fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.Name)
                fmt.Fprintln(os.Stderr, "\t" + words[idx-2].At())
                os.Exit(1)
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" not declared\n", words[idx-2].Str)
            fmt.Fprintln(os.Stderr, "\t" + words[idx-2].At())
            os.Exit(1)
        }
    } else {
        // TODO: check if var is defined
        if otherVar := GetVar(words[idx+1].Str); otherVar != nil {
            if v := GetVar(words[idx-2].Str); v != nil {
                for ri, r := range otherVar.Regs {
                    globalDefs = append(globalDefs, fmt.Sprintf("mov %s, %s\n", Registers[v.Regs[ri]].Name, Registers[r].Name))
                }
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" not declared\n", words[idx-2].Str)
                fmt.Fprintln(os.Stderr, "\t" + words[idx-2].At())
                os.Exit(1)
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not declared\n", words[idx+1].Str)
            fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
            os.Exit(1)
        }
    }

    return idx + 1
}

func AddVar(v Variable) {
    vars = append(vars, v)
}

func RmVar(varname string) {
    if len(vars) == 1 && vars[0].Name == varname {
        vars = []Variable{}
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

func WriteStrLits(asm *os.File) {
    if len(types.StrLits) > 0 {
        asm.WriteString("\nsection .data\n")

        for i, str := range types.StrLits {
            asm.WriteString(fmt.Sprintf("str%d: db %s\n", i, str.Value))
        }
    }
}
