package freeze

import (
    "crypto/sha1"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    "sync"
)

type hashResult struct {
    path string
    hash string
}

func calcHash(wg *sync.WaitGroup, paths <-chan string, hashes chan<- *hashResult) {
    defer wg.Done()
    for path := range paths {
        func() {
            f, err := os.Open(path)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Failed to open %s\n", path)
                return
            }
            defer f.Close()

            h := sha1.New()
            _, err = io.Copy(h, f)
            if err != nil {
                panic(err)
            }
            hashes <- &hashResult{path, fmt.Sprintf("%x", h.Sum(nil))}
        }()
    }
}

func collectResults(hashes <-chan *hashResult, results chan<- map[string]string) {
    resultsMap := make(map[string]string)
    for result := range hashes {
        resultsMap[result.path] = result.hash
    }
    results <- resultsMap
}

func Generate() map[string]string {
    const COUNT = 64
    cwd, err := os.Getwd()
    if err != nil {
        panic(err)
    }
    var wg sync.WaitGroup
    paths := make(chan string, COUNT)
    hashes := make(chan *hashResult, COUNT)
    results := make(chan map[string]string)
    for i := 0; i < COUNT; i++ {
        wg.Add(1)
        go calcHash(&wg, paths, hashes)
    }
    go collectResults(hashes, results)

    filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.Mode().IsRegular() && !strings.HasPrefix(info.Name(), ".") {
            path, _ := filepath.Rel(cwd, path)
            paths <- path
        }
        return nil
    })
    close(paths)
    wg.Wait()
    close(hashes)

    resultsMap := <-results
    return resultsMap
}

func Verify(expected map[string]string) {
    results := Generate()
    for path, hash := range expected {
        if expectedHash, ok := results[path]; ok {
            if hash != expectedHash {
                fmt.Printf("%s modified\n", path)
            }
            delete(results, path)
        } else {
            fmt.Printf("%s missing\n", path)
        }
    }
    for path, _ := range results {
        fmt.Printf("%s new\n", path)
    }
}
