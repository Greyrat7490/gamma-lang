package prs

/*
import (
	"gorec/token"
	"os"
	"fmt"
)

func sortBinOps(destIdx int) {
    // set assign value to first operant of mul or div
    if Ops[destIdx+1].Type != OP_MUL && Ops[destIdx+1].Type != OP_DIV {
        srcIdx := len(Ops)-2

        tmp := Ops[destIdx].Operants[1]

        if Ops[srcIdx].Type == OP_SUB {
            Ops[destIdx].Operants[1] = "-" + Ops[srcIdx].Operants[1]
            Ops[srcIdx].Type = OP_ADD
        } else {
            Ops[destIdx].Operants[1] = Ops[srcIdx].Operants[1]
        }

        Ops[srcIdx].Operants[1] = tmp
    }

    // put OP_MUL and OP_DIV before OP_ADD and OP_SUB
    for i := len(Ops)-1; Ops[i-1].Type == OP_ADD || Ops[i-1].Type == OP_SUB; i-- {
        tmp := Ops[i]
        Ops[i] = Ops[i-1]
        Ops[i-1] = tmp
    }
}

func prsAdd(idx int) int {
    tokens := token.GetTokens()

    if len(tokens) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '+' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    destOpIdx := len(Ops)-1
    for isBinaryOp(Ops[destOpIdx].Type) { destOpIdx-- }

    if Ops[destOpIdx].Type == OP_DEF_VAR || Ops[destOpIdx].Type == OP_ASSIGN_VAR {
        op := Op{ Type: OP_ADD, Token: tokens[idx], Operants: []string{ Ops[destOpIdx].Operants[0], tokens[idx+1].Str } }
        Ops = append(Ops, op)
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] not using result (assigning or defining a var)")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    return idx + 1
}

func prsSub(idx int) int {
    tokens := token.GetTokens()

    if len(tokens) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '-' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    destOpIdx := len(Ops)-1
    for isBinaryOp(Ops[destOpIdx].Type) { destOpIdx-- }

    if Ops[destOpIdx].Type == OP_DEF_VAR || Ops[destOpIdx].Type == OP_ASSIGN_VAR {
        op := Op{ Type: OP_SUB, Token: tokens[idx], Operants: []string{ Ops[destOpIdx].Operants[0], tokens[idx+1].Str } }
        Ops = append(Ops, op)
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] not using result (assigning or defining a var)")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    return idx + 1
}

func prsMul(idx int) int {
    tokens := token.GetTokens()

    if len(tokens) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '*' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    destOpIdx := len(Ops)-1
    for isBinaryOp(Ops[destOpIdx].Type) { destOpIdx-- }

    if Ops[destOpIdx].Type == OP_DEF_VAR || Ops[destOpIdx].Type == OP_ASSIGN_VAR {
        op := Op{ Type: OP_MUL, Token: tokens[idx], Operants: []string{ Ops[destOpIdx].Operants[0], tokens[idx+1].Str } }
        Ops = append(Ops, op)
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] not using result (assigning or defining a var)")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    // sortBinOps(destOpIdx)

    return idx + 1
}

func prsDiv(idx int) int {
    tokens := token.GetTokens()

    if len(tokens) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '/' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    destOpIdx := len(Ops)-1
    for isBinaryOp(Ops[destOpIdx].Type) { destOpIdx-- }

    if Ops[destOpIdx].Type == OP_DEF_VAR || Ops[destOpIdx].Type == OP_ASSIGN_VAR {
        op := Op{ Type: OP_DIV, Token: tokens[idx], Operants: []string{ Ops[destOpIdx].Operants[0], tokens[idx+1].Str } }
        Ops = append(Ops, op)
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] not using result (assigning or defining a var)")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    // sortBinOps(destOpIdx)

    return idx + 1
}
*/
