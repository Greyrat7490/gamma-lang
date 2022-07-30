package test

import (
    "os"
    "fmt"
    "flag"
    "os/exec"
    "strings"
    "testing"
    "io/ioutil"
    "path/filepath"
)

var rec bool
var keepAsm bool

var failed bool = false

func record(t *testing.T, path string, name string, stdout string, stderr string) {
    os.MkdirAll(path, 0644)

    output := []byte(stdout + "\n" + stderr)
    file := path + "/" + name + ".rec"

    err := ioutil.WriteFile(file, output, 0644)
    if err != nil {
        t.Fatalf("[ERROR] could not record results\n\t%v\n", err)
    }

    fmt.Println("[RECORDED]")
}

func check(path string, name string, stdout string, stderr string) {
    result := stdout + "\n" + stderr
    file := path + "/" + name + ".rec"

    expected, err := ioutil.ReadFile(file)
    if err != nil {
        fmt.Printf("[ERROR] could not compare with recorded results\n\t%v\n", err)
        return
    }

    if string(expected) == result {
        fmt.Println("[PASSED]")
    } else {
        fmt.Println("[FAILED]")
        fmt.Println(diff(string(expected), result))
        failed = true
    }
}

// https://en.wikipedia.org/wiki/Longest_common_subsequence_problem
func diff(expected string, res string) string {
    expLines := strings.Split(expected, "\n")
    resLines := strings.Split(res, "\n")

    LCS_Table := make([][]int, len(expLines)+1)
    for i := range LCS_Table {
        LCS_Table[i] = make([]int, len(resLines)+1)
    }

    // init LCS_table
    for ie := 1; ie <= len(expLines); ie++ {
        for ir := 1; ir <= len(resLines); ir++ {
            if expLines[ie-1] == resLines[ir-1] {
                LCS_Table[ie][ir] = LCS_Table[ie-1][ir-1] + 1

            } else if LCS_Table[ie-1][ir] >= LCS_Table[ie][ir-1] {
                LCS_Table[ie][ir] = LCS_Table[ie-1][ir]

            } else {
                LCS_Table[ie][ir] = LCS_Table[ie][ir-1]
            }
        }
    }

    return getDiff(LCS_Table, expLines, resLines, len(expLines), len(resLines))
}

func getDiff(LCS_Table [][]int, a []string, b []string, i int, j int) string {
    if i == 0 || j == 0 {
        return ""
    }

    if a[i-1] == b[j-1] {
        return getDiff(LCS_Table, a, b, i-1, j-1)
    }

    if LCS_Table[i][j-1] >= LCS_Table[i-1][j] {
        return getDiff(LCS_Table, a, b, i, j-1) + fmt.Sprintf("@%d + %s\n", j, b[j-1])
    } else {
        return getDiff(LCS_Table, a, b, i-1, j) + fmt.Sprintf("@%d - %s\n", i, a[i-1])
    }
}

// removes executable, object and assembly files
// if 'srcname' != "" -> the assembly file will be renamed instead
func clearBuilds(t *testing.T, path string, srcname string) {
    err := os.Remove(path + "output")
    if err != nil && !os.IsNotExist(err) {
        t.Fatalf("[ERROR] could not remove output\n%v", err)
    }

    err = os.Remove(path + "output.o")
    if err != nil && !os.IsNotExist(err) {
        t.Fatalf("[ERROR] could not remove output.o\n%v", err)
    }

    if len(srcname) == 0 {
        err = os.Remove(path + "output.asm")
        if err != nil && !os.IsNotExist(err)  {
            t.Fatalf("[ERROR] could not remove output.asm\n%v", err)
        }
    } else {
        err = os.Rename(path + "output.asm", srcname + ".asm")
        if err != nil && !os.IsNotExist(err) {
            t.Fatalf("[ERROR] could not rename output.asm\n%v", err)
        }
    }
}

func init() {
    flag.BoolVar(&rec, "rec", false, "record the stdout and stderr results")
    flag.BoolVar(&keepAsm, "asm", false, "keep the assembly files generated")
}

func test(t *testing.T, flagStr string, recDir string, fileDir string) {
    flag.Parse()

    path, err := os.Getwd()
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }

    path += "/" + fileDir + "/"

    files, err := ioutil.ReadDir(path)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }

    for _, f := range files {
        if filepath.Ext(f.Name()) == ".gma" {
            fmt.Print(f.Name())
            cmd := exec.Command("go", "run", "gamma", flagStr, f.Name())

            var stdout, stderr strings.Builder
            cmd.Stdout = &stdout
            cmd.Stderr = &stderr

            cmd.Run()

            stdoutStr := stdout.String()
            stderrStr := stderr.String()

            if rec {
                record(t, recDir, f.Name(), stdoutStr, stderrStr)
            } else {
                check(recDir, f.Name(), stdoutStr, stderrStr)
            }

            if keepAsm {
                clearBuilds(t, path, f.Name())
            } else {
                clearBuilds(t, path, "")
            }
        }
    }
}

func TestError(t *testing.T) {
    test(t, "", "recs/err", "errTests")

    if failed { os.Exit(1) }
}

func TestAst(t *testing.T) {
    test(t, "-ast", "recs/ast", "")
}

func TestRun(t *testing.T) {
    test(t, "-r", "recs/run", "")

    if failed { os.Exit(1) }
}
