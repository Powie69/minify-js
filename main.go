package main

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func main() {
	app := &cli.App{
		Name:  "miniJs",
		Usage: "Process JavaScript files",
		Action: func(c *cli.Context) error {
			return readFile(c)
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func minifyJavaScript(jsCode string) (string, error) {
	resp, err := http.PostForm(
		"https://www.toptal.com/developers/javascript-minifier/api/raw",
		url.Values{"input": {jsCode}},
	)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 status
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Read and return body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read failed: %w", err)
	}
	return string(body), nil
}

func readFile(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("must specify file")
	}

	filePath := c.Args().First()
	file, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid file: %v", err)
	}

	file = filepath.Clean(file)
	fileInfo, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", file)
		}
		return fmt.Errorf("error accessing file: %v", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("%s is a directory, not a file", file)
	}

	fileName := filepath.Base(file)
	fileExt := filepath.Ext(fileName)
	baseName := fileName[:len(fileName)-len(fileExt)]
	minifiedFileName := filepath.Join(filepath.Dir(file), fmt.Sprintf("%s.min%s", baseName, fileExt))

	if _, err := os.Stat(minifiedFileName); err == nil {
		prompt := promptui.Select{
			Label: "Minified file already exist",
			Items: []string{
				"Override output file",
				"Rename output File",
				"Cancel operation",
			},
		}

		i, _, err := prompt.Run()
		if err != nil {
			log.Fatalf("Prompt failed %v\n", err)
		}

		switch i {
		case 1:
			fmt.Println("Output file shall be named:")
			if _, err := fmt.Scanln(&minifiedFileName); err != nil {
				return err
			}
		case 2:
			os.Exit(0)
		}
	}

	fmt.Printf("Processing file: %s\n", file)
	contents, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Minify the JavaScript
	minified, err := minifyJavaScript(string(contents))
	if err != nil {
		return fmt.Errorf("error minifying JavaScript: %v", err)
	}

	// Write minified content to new file
	if err := os.WriteFile(minifiedFileName, []byte(minified), 0644); err != nil {
		return fmt.Errorf("error writing minified file: %v", err)
	}

	fmt.Printf("Minified output:\n%s\n", minified)
	fmt.Printf("Minified file created: %s\n", minifiedFileName)

	return nil
}
