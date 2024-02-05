package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var logr = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logr.Printf("$ %s\n%s", cmd.String(), output)
		return err
	}
	logr.Printf("$ %s\n%s", cmd.String(), output)
	return nil
}

func runTW(inPath, outPath string, withWatch bool, wg *sync.WaitGroup) {
	defer wg.Done()
	flags := []string{"-i", inPath, "-o", outPath, "--minify"}
	if withWatch {
		flags = append(flags, "--watch")
	}
	if err := runCmd("tailwindcss", flags...); err != nil {
		logr.Fatalf("Error running TW: %v", err)
	}
}

func runAir(wg *sync.WaitGroup) {
	defer wg.Done()
	if err := runCmd("air"); err != nil {
		logr.Fatalf("Error running Air: %v", err)
	}
}

func main() {
	parallel := true

	rootDir, err := filepath.Abs(filepath.Join(filepath.Dir(os.Args[0]), ".."))
	if err != nil {
		log.Fatal(err)
	}

	staticDir := filepath.Join(rootDir, "static")
	templatesDir := filepath.Join(rootDir, "templates")
	templatesCSSDir := filepath.Join(templatesDir, "css")
	inCSSPath := filepath.Join(templatesCSSDir, "globals.css")
	outCSSPath := filepath.Join(staticDir, "css", "style.css")

	var wg sync.WaitGroup

	if parallel {
		wg.Add(2)
		go runTW(inCSSPath, outCSSPath, true, &wg)
		go runAir(&wg)
		wg.Wait()
	} else {
		runTW(inCSSPath, outCSSPath, false, nil)
		runAir(nil)
	}
}
