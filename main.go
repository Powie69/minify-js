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
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "n, name",
				Usage: "Specify the output file name",
			},
		},
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
	res, err := http.PostForm(
		"https://www.toptal.com/developers/javascript-minifier/api/raw",
		url.Values{"input": {jsCode}},
	)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return "", fmt.Errorf("bad status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read failed: %w", err)
	}
	return string(body), nil
}

func outputFileName(flagValue string, file string) string {
	if flagValue == "" {
		fileName := filepath.Base(file)
		fileExt := filepath.Ext(fileName)
		baseName := fileName[:len(fileName)-len(fileExt)]
		flagValue = fmt.Sprintf("%s.min%s", baseName, fileExt)
	}

	return filepath.Join(filepath.Dir(file), flagValue)
}

func isOutputAlreadyExist(minifiedFileName string) (bool, error) {
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
				return false, err
			}
		case 2:
			os.Exit(0)
		}
	}
	return false, nil
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

	minifiedFileName := outputFileName(c.String("name"), file) // string

	for {
		if result, err := isOutputAlreadyExist(minifiedFileName); err != nil {
			return err
		} else if !result {
			break
		}

		// Prompt user to enter a new name
		prompt := promptui.Prompt{
			Label: "File already exists. Please enter a new name",
		}
		newName, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("error getting new name: %v", err)
		}
		minifiedFileName = outputFileName(newName, file)
	}

	fmt.Printf("Processing file: %s\n", file)
	contents, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	minifiedOutput, err := minifyJavaScript(string(contents))
	if err != nil {
		return fmt.Errorf("error minifying JavaScript: %v", err)
	}

	if err := os.WriteFile(minifiedFileName, []byte(minifiedOutput), 0644); err != nil {
		return fmt.Errorf("error writing minifiedOutput file: %v", err)
	}

	fmt.Printf("Minified output:\n%s\n", minifiedOutput)
	fmt.Printf("Minified file created: %s\n", minifiedFileName)

	return nil
}
