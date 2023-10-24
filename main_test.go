package main

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func Test_parseTemplates(t *testing.T) {
	testdir := ".testresults"
	tempBasePath := "examples/simple/templates"
	err := mkDirIfNotExists(testdir)
	if err != nil {
		t.Errorf("failed to init filesystem for tests: %s", err)
	}
	t.Cleanup(func() {
		err = os.RemoveAll(testdir)
		if err != nil {
			t.Errorf("failed to cleanup testdir: %s", err)
		}
	})
	tests := []struct {
		name                 string
		secrets              map[string]string
		tempPaths            []string
		continueOnMissingKey bool
		wantError            bool
		expectedResult       string
	}{
		{
			name: "working example",
			secrets: map[string]string{
				"TEST1": "value1",
				"TEST2": "value2",
			},
			tempPaths:            []string{"examples/simple/templates/etc/config"},
			continueOnMissingKey: false,
			wantError:            false,
			expectedResult:       fmt.Sprintf("test1=value1\ntest2=value2"),
		},
		{
			name: "missing secret",
			secrets: map[string]string{
				"TEST1": "value1",
			},
			tempPaths:            []string{"examples/simple/templates/etc/config"},
			continueOnMissingKey: false,
			wantError:            true,
		},
		{
			name: "missing secret with continue",
			secrets: map[string]string{
				"TEST1": "value1",
			},
			tempPaths:            []string{"examples/simple/templates/etc/config"},
			continueOnMissingKey: true,
			wantError:            false,
			expectedResult:       fmt.Sprintf("test1=value1\ntest2=<no value>"),
		},
		{
			name: "wrong template path",
			secrets: map[string]string{
				"TEST1": "value1",
				"TEST2": "value2",
			},
			tempPaths:            []string{"examples/simple/templates/etc/config12345"},
			continueOnMissingKey: false,
			wantError:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTemplates(tt.tempPaths, LeftLimiter, RightLimiter, tt.continueOnMissingKey, testdir, tempBasePath, tt.secrets)
			if ((got == nil) && tt.wantError) || ((got != nil) && !tt.wantError) {
				t.Errorf("unexpected result: parseTemplates() = %v, want %v", got != nil, tt.wantError)
			}
			if !tt.wantError {
				content, err := os.ReadFile(testdir + "/etc/config")
				if err != nil {
					t.Errorf("failed to read rendered file (%q): %s", testdir+"/etc/config", err)
				}
				expectedContent := []byte(tt.expectedResult)
				if !bytes.Equal(content, expectedContent) {
					t.Errorf("content differes from expected content, \ncontent:\n%s\n\nexpected content:\n%s\n", content, expectedContent)
				}
			}
		})
	}
}

func Test_getSecretsFromEnv(t *testing.T) {
	// prepare envs
	_ = os.Setenv("TEST1", "value1")
	_ = os.Setenv("SECRET_TEST1", "secretValue1")

	wantedResult := map[string]string{
		"TEST1": "secretValue1",
	}
	if got := getSecretsFromEnv(SecretPrefix); !reflect.DeepEqual(got, wantedResult) {
		t.Errorf("getSecretsFromEnv() = %v, want %v", got, wantedResult)
	}
}

func Test_getAllTemplateFilePaths(t *testing.T) {
	tempBasePath := "examples/simple/templates"
	tempPaths, err := getAllTemplateFilePaths(tempBasePath)
	if err != nil {
		t.Errorf("failed to get template paths(%q): %s", tempBasePath, err)
	}
	wantedResult := []string{tempBasePath + "/etc/config"}
	if !reflect.DeepEqual(tempPaths, wantedResult) {
		t.Errorf("getAllTemplateFilePaths() = %v, want %v", tempPaths, wantedResult)
	}
}
