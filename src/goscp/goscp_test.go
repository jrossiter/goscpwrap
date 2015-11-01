package goscp

import (
	"os"
	"reflect"
	"regexp"
	"testing"
)

/*
todo
	Chdir
*/

func TestMain(m *testing.M) {
	setUp()

	code := m.Run()

	tearDown()

	os.Exit(code)
}

func setUp() {

}

func tearDown() {

}

func TestUpDirectory(t *testing.T) {
	/*
		c.DestinationPath = c.DestinationPath[:len(c.DestinationPath)-1]
	*/
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
			expectedError(t, "TestUpDirectory", c.DestinationPath, v.Expected)
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
				expectedError(t, "TestParseMessage", err, v.ExpectedError)
			}
			continue
		}

		if !reflect.DeepEqual(output, v.Expected) {
			expectedError(t, "TestParseMessage", output, v.Expected)
		}
	}
}

func expectedError(t *testing.T, fn string, received, expected interface{}) {
	t.Errorf("%s - received: %v, expected: %v", fn, received, expected)
}
