package main

import (
	"bytes"
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
		secrets              map[string]map[string]string
		tempPaths            []string
		continueOnMissingKey bool
		wantError            bool
		expectedResult       string
	}{
		{
			name: "working example",
			secrets: map[string]map[string]string{
				"example": {
					"TEST1": "value1",
					"TEST2": "value2",
				},
			},
			tempPaths:            []string{"examples/simple/templates/etc/config"},
			continueOnMissingKey: false,
			wantError:            false,
			expectedResult:       "test1=value1\ntest2=value2",
		},
		{
			name: "missing secret",
			secrets: map[string]map[string]string{
				"example": {
					"TEST1": "value1",
				},
			},
			tempPaths:            []string{"examples/simple/templates/etc/config"},
			continueOnMissingKey: false,
			wantError:            true,
		},
		{
			name: "missing secret with continue",
			secrets: map[string]map[string]string{
				"example": {
					"TEST1": "value1",
				},
			},
			tempPaths:            []string{"examples/simple/templates/etc/config"},
			continueOnMissingKey: true,
			wantError:            false,
			expectedResult:       "test1=value1\ntest2=<no value>",
		},
		{
			name: "wrong template path",
			secrets: map[string]map[string]string{
				"example": {
					"TEST1": "value1",
					"TEST2": "value2",
				},
			},
			tempPaths:            []string{"examples/simple/templates/etc/config12345"},
			continueOnMissingKey: false,
			wantError:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderSecretsIntoTemplates(tt.tempPaths, LeftDelimiter, RightDelimiter, tt.continueOnMissingKey, testdir, tempBasePath, tt.secrets)
			if ((got == nil) && tt.wantError) || ((got != nil) && !tt.wantError) {
				t.Errorf("unexpected result: renderSecretsIntoTemplates() = %v, want %v", got != nil, tt.wantError)
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

func Test_getSecretsFromFiles(t *testing.T) {
	secrets, err := getSecretsFromFiles("tests/secrets")
	if err != nil {
		t.Errorf("failed to get secrets from files: %s", err)
	}
	if secrets["sec1"]["key1"] != "thisisavalue" {
		t.Errorf("failed to map sec1[key1]: thisisavalue != %s", secrets["sec1"]["key1"])
	}
	if secrets["sec1"]["key2"] != "thisisanothervalue" {
		t.Errorf("failed to map sec1[key2]: thisisanothervalue != %s", secrets["sec1"]["key2"])
	}
	if secrets["sec2"]["key1"] != "thisisjustavalue" {
		t.Errorf("failed to map sec2[key1]: thisisjustavalue != %s", secrets["sec2"]["key1"])
	}
	if secrets["sec2"]["key2"] != "thisisjustanothervalue" {
		t.Errorf("failed to map sec2[key2]: thisisjustanothervalue != %s", secrets["sec2"]["key2"])
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

func Test_getValueByFirstMatchingKey(t *testing.T) {
	tests := []struct {
		name      string
		stringMap map[string]string
		keys      []string
		want      string
		wantErr   bool
	}{
		{
			name:      "one-match",
			stringMap: map[string]string{"key1": "val1"},
			keys:      []string{"key1"},
			want:      "val1",
			wantErr:   false,
		},
		{
			name:      "match-first",
			stringMap: map[string]string{"key1": "val1", "key2": "val2"},
			keys:      []string{"key1"},
			want:      "val1",
			wantErr:   false,
		},
		{
			name:      "match-second",
			stringMap: map[string]string{"key1": "val1", "key2": "val2"},
			keys:      []string{"key2"},
			want:      "val2",
			wantErr:   false,
		},
		{
			name:      "skip-first-match-second",
			stringMap: map[string]string{"key1": "val1", "key2": "val2"},
			keys:      []string{"key3", "key2"},
			want:      "val2",
			wantErr:   false,
		},
		{
			name:      "key-not-found",
			stringMap: map[string]string{"key1": "val1", "key2": "val2"},
			keys:      []string{"key3"},
			want:      "",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getValueByFirstMatchingKey(tt.stringMap, tt.keys...)
			if (err != nil) != tt.wantErr {
				t.Errorf("getValueByFirstMatchingKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getValueByFirstMatchingKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}
