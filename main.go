package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	LeftLimiter      = "{{"
	RightLimiter     = "}}"
	SecretPath       = "/etc/secrets"
	TargetBasePath   = "/etc/rendered"
	TemplateBasePath = "/etc/templates"
)

func main() {

	// definition of cli interface
	continueOnMissingKey := flag.Bool("continue-on-missing-key", false, "enable to not stop when hitting missing keys during templating")
	leftLimiter := flag.String("left-limiter", LeftLimiter, "left limiter for internal go templating")
	rightLimiter := flag.String("right-limiter", RightLimiter, "right limiter for internal go templating")
	secretPath := flag.String("secret-path", SecretPath, "absolute path to directory where secrets are mounted")
	targetBasePath := flag.String("target-base-dir", TargetBasePath, "absolute path to directory containing rendered template files")
	templateBasePath := flag.String("template-base-dir", TemplateBasePath, "absolute path to directory containing template files")
	flag.Parse()

	// retrieve secrets
	secrets, err := getSecretsFromFiles(*secretPath)
	if err != nil {
		log.Panicf("failed to get secrets from files: %s", err)
	}

	// detect templates
	templatePaths, err := getAllTemplateFilePaths(*templateBasePath)
	if err != nil {
		log.Panicf("failed to read paths of template files: %s", err)
	}

	// parse every template file separately
	err = parseTemplates(templatePaths, *leftLimiter, *rightLimiter, *continueOnMissingKey, *targetBasePath, *templateBasePath, secrets)
	if err != nil {
		log.Panicf("failed to parse template: %s", err)
	}
}

func parseTemplates(templatePaths []string, leftLimiter string, rightLimiter string, continueOnMissingKey bool, targetBasePath string, templateBasePath string, secrets map[string]map[string]string) error {
	for _, templatePath := range templatePaths {
		t, err := template.ParseFiles(templatePath)
		if err != nil {
			return fmt.Errorf("failed to parse template files(%q): %w", templatePath, err)
		}
		t.Delims(leftLimiter, rightLimiter)
		if !continueOnMissingKey {
			t.Option("missingkey=error")
		}

		targetPath := path.Join(targetBasePath, strings.TrimPrefix(templatePath, templateBasePath))

		err = mkDirIfNotExists(path.Dir(targetPath))
		if err != nil {
			return fmt.Errorf("failed to create target dir for %q: %w", templatePath, err)
		}
		targetFile, _ := os.Create(targetPath)
		err = t.Execute(targetFile, struct {
			Secrets map[string]map[string]string
		}{
			Secrets: secrets,
		})
		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
	}

	return nil
}

func getAllTemplateFilePaths(templateWalkDir string) (templateFilePaths []string, err error) {
	err = filepath.WalkDir(templateWalkDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			templateFilePaths = append(templateFilePaths, path)
		}
		return nil
	})
	return templateFilePaths, err
}

func getSecretsFromFiles(secretsPath string) (map[string]map[string]string, error) {
	secrets := make(map[string]map[string]string)
	err := filepath.WalkDir(secretsPath, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			secret, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read secret from file %q: %s", filePath, err)
			}
			keyName := path.Base(filePath)
			secretName := path.Base(path.Dir(filePath))
			_, ok := secrets[secretName]
			if !ok {
				secrets[secretName] = make(map[string]string)
			}
			secrets[secretName][keyName] = string(secret)
		}
		return nil
	})
	if err != nil {
		return secrets, fmt.Errorf("failed to get secrets from files: %s", err)
	}
	return secrets, err
}

func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return fileInfo.IsDir()
}

func mkDirIfNotExists(path string) error {
	if !isDirectory(path) {
		err := os.MkdirAll(path, 0775)
		if err != nil {
			return err
		}
	}

	return nil
}
