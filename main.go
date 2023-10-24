package main

import (
	"flag"
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
	SecretPrefix     = "SECRET_"
	TargetBasePath   = "/etc/rendered"
	TemplateBasePath = "/etc/templates"
)

func main() {

	// definition of cli interface
	continueOnMissingKey := flag.Bool("continue-on-missing-key", false, "enable to not stop when hitting missing keys during templating")
	leftLimiter := flag.String("left-limiter", LeftLimiter, "left limiter for internal go templating")
	rightLimiter := flag.String("right-limiter", RightLimiter, "right limiter for internal go templating")
	secretEnvPrefix := flag.String("secret-env-prefix", SecretPrefix, "prefix for the environment variables containing secrets")
	targetBasePath := flag.String("target-base-dir", TargetBasePath, "absolute path to directory containing rendered template files")
	templateBasePath := flag.String("template-base-dir", TemplateBasePath, "absolute path to directory containing template files")
	flag.Parse()

	// retrieve secrets
	secrets := getSecretsFromEnv(*secretEnvPrefix)

	// detect templates
	templatePaths, err := getAllTemplateFilePaths(*templateBasePath)
	if err != nil {
		log.Panicf("failed to read paths of template files: %s", err)
	}

	// parse every template file separately
	for _, templatePath := range templatePaths {
		t := template.Must(template.ParseFiles(templatePath)).Delims(*leftLimiter, *rightLimiter)
		if !*continueOnMissingKey {
			t.Option("missingkey=error")
		}

		targetPath := path.Join(*targetBasePath, strings.TrimPrefix(templatePath, *templateBasePath))

		err = mkDirIfNotExists(path.Dir(targetPath))
		if err != nil {
			log.Panicf("failed to create target dir for %q: %s", templatePath, err)
		}
		targetFile, _ := os.Create(targetPath)
		err := t.Execute(targetFile, secrets)
		if err != nil {
			log.Panic(err)
		}
	}
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

func getSecretsFromEnv(prefix string) map[string]string {
	var secrets = make(map[string]string)
	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, prefix) {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				secrets[strings.TrimPrefix(parts[0], prefix)] = parts[1]
			}
		}
	}
	return secrets
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
