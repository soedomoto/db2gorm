package tools

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func tidyFile(filePath string) {
	// Parse the Go source file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Failed to parse file:", err)
		return
	}

	// Create a map to track imported packages
	imports := make(map[string]bool)

	// Traverse the AST and remove duplicate import declarations
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}

		// Iterate through import specs and remove duplicates
		var specs []ast.Spec
		for _, spec := range genDecl.Specs {
			importSpec := spec.(*ast.ImportSpec)
			importPath := strings.Trim(importSpec.Path.Value, "\"")

			if !imports[importPath] {
				specs = append(specs, spec)
				imports[importPath] = true
			}
		}

		// Update the import specs with the deduplicated ones
		genDecl.Specs = specs
	}

	// Generate the modified Go code
	var output strings.Builder
	if err := printer.Fprint(&output, fset, file); err != nil {
		fmt.Println("Failed to generate modified code:", err)
		return
	}

	// Format the code using gofmt
	formattedCode, err := format.Source([]byte(output.String()))
	if err != nil {
		fmt.Println("Failed to format code:", err)
		return
	}

	// Write the formatted code to a file
	err = ioutil.WriteFile(filePath, formattedCode, 0644)
	if err != nil {
		fmt.Println("Failed to write code to file:", err)
		return
	}

	fmt.Println("Code successfully written to file.")
}

func Tidy() {
	root := "./" // Specify the root directory to search in

	err := filepath.Walk(root, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", filePath, err)
			return nil
		}

		// Check if the file has a .go extension
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			tidyFile(filePath)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %s: %v\n", root, err)
	}
}
