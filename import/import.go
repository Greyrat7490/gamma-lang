package imprt

import (
    "os"
    "fmt"
    "gamma/token"
    "path/filepath"
)

var basePath string

var imported map[string]bool = make(map[string]bool)
// true: fully imported
// false: not fully import -> import cycle if imported again


func ImportMain(path string) token.Tokens {
    file, err := os.Open(path)
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR]", err)
        os.Exit(1)
    }

    basePath = filepath.Dir(path)

    addImport(path)

    return token.Tokenize(path, file)
}

func Import(importPath token.Token) (*token.Tokens, bool) {
    path := preparePath(importPath)

    if addImport(path) {
        file, err := os.Open(path)
        if err != nil {
            fmt.Fprintln(os.Stderr, "[ERROR]", err)
            os.Exit(1)
        }

        tokens := token.Tokenize(path, file)
        return &tokens, true
    }

    return nil, false
}

func EndImport(path string) {
    imported[path] = true
}

func addImport(path string) (newImport bool) {
    if importable, notNew := imported[path]; !notNew {
        imported[path] = false
        newImport = true
    } else {
        if !importable {
            fmt.Fprintln(os.Stderr, "[ERROR] import cycle detected:")
            for path, importable := range imported {
                if !importable {
                    fmt.Fprintln(os.Stderr, "\t", path)
                }
            }
            fmt.Fprintln(os.Stderr, "\t", path)
            os.Exit(1)
        }
    }

    return
}

func preparePath(path token.Token) string {
    path.Str = path.Str[1:len(path.Str)-1]

    if !filepath.IsAbs(path.Str) {
        // relative path to main file (file passed as arg to compiler)
        // std path (std/<path>)
        // relative path
        basePaths := []string{ basePath, "std", "./" }

        for _,basePath := range basePaths {
            path := filepath.Join(basePath, path.Str)
            if _, err := os.Stat(path); !os.IsNotExist(err) {
                return path
            }
        }
    }
    // absolute path

    return path.Str
}
