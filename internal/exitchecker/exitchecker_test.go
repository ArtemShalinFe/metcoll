package exitchecker

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAll(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		diagCount int
	}{
		{
			name: "1_case.go",
			src: `package main 

			import "os"
	
			func main() { 
				os.Exit(0) // want "called os.Exit in the main function"
			}`,
			diagCount: 1,
		},
		{
			name: "2_case.go",
			src: `package main

			func main() {
				print("hello, world!")
			}`,
			diagCount: 0,
		},
	}
	for _, tt := range tests {
		filemap := make(map[string]string)
		filemap[tt.name] = tt.src

		testdata, cleanup, err := analysistest.WriteFiles(filemap)
		if err != nil {
			t.Errorf("cannot create go files err: %v", err)
		}
		defer cleanup()

		analysistest.Run(t, filepath.Join(testdata, "src"), NewAnalyzer())
	}
}
