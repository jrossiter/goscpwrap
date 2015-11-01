package goscp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
	"time"
)

var (
	// Items created during testing
	created []string
)

func TestMain(m *testing.M) {
	setUp()

	code := m.Run()

	tearDown()

	os.Exit(code)
}

func setUp() {
	err := os.Chdir(os.TempDir())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Running tests in:", os.TempDir())
}

func tearDown() {
	for _, v := range created {
		log.Println("Removing", v)
		os.Remove(v)
	}
}

func expectedError(t *testing.T, received, expected interface{}) {
	t.Errorf("received: %v, expected: %v", received, expected)
}

func TestUpDirectory(t *testing.T) {
	tests := []struct {
		Input    []string
		Expected []string
	}{
		{
			Input:    []string{"one", "two", "three"},
			Expected: []string{"one", "two"},
		},
		{
			Input:    []string{"one"},
			Expected: []string{},
		},
		{
			Input:    []string{},
			Expected: []string{},
		},
	}

	c := Client{}
	for _, v := range tests {
		c.DestinationPath = v.Input
		c.upDirectory()
		if !reflect.DeepEqual(c.DestinationPath, v.Expected) {
			expectedError(t, c.DestinationPath, v.Expected)
		}
	}
}

func TestParseMessage(t *testing.T) {
	tests := []struct {
		Input         string
		Regex         *regexp.Regexp
		Expected      map[string]string
		ExpectedError string
	}{
		{
			Input: "C0644 25 helloworld.txt",
			Regex: fileCopyRx,
			Expected: map[string]string{
				"":         "C0644 25 helloworld.txt",
				"mode":     "0644",
				"length":   "25",
				"filename": "helloworld.txt",
			},
		},
		{
			Input: "D0755 0 mydir",
			Regex: dirCopyRx,
			Expected: map[string]string{
				"":        "D0755 0 mydir",
				"mode":    "0755",
				"length":  "0",
				"dirname": "mydir",
			},
		},
		{
			Input: "T1234567890 0 9876543210 0",
			Regex: timestampRx,
			Expected: map[string]string{
				"":      "T1234567890 0 9876543210 0",
				"mtime": "1234567890",
				"atime": "9876543210",
			},
		},
		{
			Input:         "Invalid msg",
			Regex:         fileCopyRx,
			ExpectedError: "Could not parse protocol message: Invalid msg",
		},
	}

	c := Client{}
	for _, v := range tests {
		output, err := c.parseMessage(v.Input, v.Regex)
		if err != nil {
			if err.Error() != v.ExpectedError {
				expectedError(t, err, v.ExpectedError)
			}
			continue
		}

		if !reflect.DeepEqual(output, v.Expected) {
			expectedError(t, output, v.Expected)
		}
	}
}

func TestDirectory(t *testing.T) {
	uts := time.Now().Unix()
	dirName := fmt.Sprintf("%s-%v", "mydir", uts)

	tests := []struct {
		StartPath               string
		InputPath               string
		ExpectedPath            string
		ExpectedDestinationPath []string
	}{
		{
			StartPath:               ".",
			InputPath:               fmt.Sprintf("D0755 0 %s", dirName),
			ExpectedPath:            dirName,
			ExpectedDestinationPath: []string{".", dirName},
		},
	}

	for _, v := range tests {
		c := Client{}
		c.SetDestinationPath(v.StartPath)
		c.directory(v.InputPath)

		path := filepath.Join(c.DestinationPath...)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			expectedError(t, err, path)
			continue
		}
		created = append(created, path)

		if !reflect.DeepEqual(c.DestinationPath, v.ExpectedDestinationPath) {
			expectedError(t, c.DestinationPath, v.ExpectedDestinationPath)
		}
	}

}

/*
todo
file
handleItem
cancel


*/
