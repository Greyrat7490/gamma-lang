package imprt

import (
    "os"
    "fmt"
    "gamma/token"
    "path/filepath"
)

// TODO: no .gma file extension
// TODO: <file> for importDir / "file" for projectDir

const buildinDir string = "../buildin/buildin.gma"
var projectDir string
var importDir string

var imported map[string]bool = make(map[string]bool)
// true: fully imported
// false: not fully import -> import cycle if imported again

func ImportMain(path string) token.Tokens {
    return ImportFile(path)
}

func ImportBuildin() token.Tokens {
    return ImportFile(preparePath(buildinDir))
}

func ImportFile(path string) token.Tokens {
    file, err := os.Open(path)
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR]", err)
        os.Exit(1)
    }

    addImport(path)

    return token.Tokenize(path, file)
}

func Import(importPath token.Token) (*token.Tokens, bool) {
    path := preparePath(extractPath(importPath))

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

func SetImportDirs(filePath string, path string) {
    importDir = path
    projectDir = filepath.Dir(filePath)
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

func extractPath(path token.Token) string {
    return path.Str[1:len(path.Str)-1]
}

func preparePath(path string) string {
    // relative path
    if !filepath.IsAbs(path) {
        // project path (main file dir (file passed as arg to compiler))
        // import path (default ./std)
        std := filepath.Join(filepath.Join(importDir, path))

        if _, err := os.Stat(std); os.IsNotExist(err) {
            return filepath.Join(projectDir, path)
        } else {
            return std
        }
    }
    // absolute path

    return path
}
