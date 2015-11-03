package goscp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
		os.Remove(v)
	}
}

func expectedError(t *testing.T, received, expected interface{}) {
	t.Errorf("received: %q, expected: %q", received, expected)
}

func TestUpDirectory(t *testing.T) {
	tests := []struct {
		Input    []string
		Expected []string
	}{
		{
			// Up one directory
			Input:    []string{"one", "two", "three"},
			Expected: []string{"one", "two"},
		},
		{
			// Up one directory
			Input:    []string{"one"},
			Expected: []string{},
		},
		{
			// Current directory
			Input:    []string{},
			Expected: []string{},
		},
	}

	c := Client{}
	for _, v := range tests {
		c.DestinationPath = v.Input
		c.upDirectory()

		// Check paths match
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
			// Create file message
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
			// Create directory message
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
			// Timestamp message
			Input: "T1234567890 0 9876543210 0",
			Regex: timestampRx,
			Expected: map[string]string{
				"":      "T1234567890 0 9876543210 0",
				"mtime": "1234567890",
				"atime": "9876543210",
			},
		},
		{
			// Invalid message
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

		// Check parts match
		if !reflect.DeepEqual(output, v.Expected) {
			expectedError(t, output, v.Expected)
		}
	}
}

func TestDirectory(t *testing.T) {
	uts := time.Now().Unix()
	dirName := fmt.Sprintf("%s-%v", "goscp-mydir", uts)

	tests := []struct {
		StartPath               string
		InputPath               string
		ExpectedPath            string
		ExpectedDestinationPath []string
	}{
		{
			// Directory message
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

		// Check dir was created
		path := filepath.Join(c.DestinationPath...)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			expectedError(t, err, path)
			continue
		}
		created = append(created, path)

		// Check destination paths match
		if !reflect.DeepEqual(c.DestinationPath, v.ExpectedDestinationPath) {
			expectedError(t, c.DestinationPath, v.ExpectedDestinationPath)
		}
	}
}

func TestFile(t *testing.T) {
	uts := time.Now().Unix()
	fileName := fmt.Sprintf("%s-%v", "goscp-test-file", uts)
	fileContent := "hello world"

	tests := []struct {
		StartPath    string
		InputPath    string
		FileContent  string
		ExpectedPath string
	}{
		{
			// File message
			StartPath:    ".",
			InputPath:    fmt.Sprintf("C0755 %d %s", len(fileContent), fileName),
			FileContent:  fileContent,
			ExpectedPath: fileName,
		},
		{
			// Empty file message
			StartPath:    ".",
			InputPath:    fmt.Sprintf("C0755 %d %s", 0, fileName),
			FileContent:  "",
			ExpectedPath: fileName,
		},
	}

	for _, v := range tests {
		c := Client{}
		c.SetDestinationPath(v.StartPath)

		dummy := bytes.NewBuffer([]byte(v.FileContent))
		rdr := &readCanceller{Reader: bufio.NewReader(dummy)}
		c.scpStdoutPipe = rdr

		c.file(v.InputPath)

		// Check file was created
		if _, err := os.Stat(v.ExpectedPath); os.IsNotExist(err) {
			expectedError(t, err, v.ExpectedPath)
			continue
		}

		// Check file content
		bytes, _ := ioutil.ReadFile(v.ExpectedPath)
		if string(bytes) != v.FileContent {
			expectedError(t, string(bytes), v.FileContent)
		}

		os.Remove(v.ExpectedPath)
	}
}

func TestHandleItem(t *testing.T) {
	tests := []struct {
		Type                    string
		Name                    string
		Content                 []byte
		ExpectedMessages        []string
		DestinationPath         []string
		ExpectedDestinationPath []string
	}{
		{
			// File creation
			Type:    "file",
			Name:    "goscp-test-content.txt",
			Content: []byte("hello-world test text\n"),
			ExpectedMessages: []string{
				"C0644 22 goscp-test-content.txt\n",
				"hello-world test text\n",
				"\x00\n",
			},
		},
		{
			// Empty file creation
			Type:    "file",
			Name:    "goscp-test-content.txt",
			Content: []byte(""),
			ExpectedMessages: []string{
				"C0644 0 goscp-test-content.txt\n",
				"\x00\n",
			},
		},

		{
			// Directory creation and traversing upward
			Type: "directory",
			Name: "goscp-test-dir/two",
			ExpectedMessages: []string{
				"E\n",
				"E\n",
				"D0644 0 two\n",
			},
			DestinationPath:         []string{"goscp-test-dir", "hello", "one"},
			ExpectedDestinationPath: []string{"goscp-test-dir/two"},
		},
		{
			// Directory creation in same level
			Type: "directory",
			Name: "goscp-test-dir/one",
			ExpectedMessages: []string{
				"E\n",
				"D0644 0 one\n",
			},
			DestinationPath:         []string{"goscp-test-dir", "two"},
			ExpectedDestinationPath: []string{"goscp-test-dir/one"},
		},
		{
			// Directory creation in traversing downward
			Type: "directory",
			Name: "goscp-test-dir/one/two",
			ExpectedMessages: []string{
				"D0644 0 two\n",
			},
			DestinationPath:         []string{"goscp-test-dir", "one"},
			ExpectedDestinationPath: []string{"goscp-test-dir/one/two"},
		},
		{
			// Directory creation
			Type: "directory",
			Name: "goscp-test-dir",
			ExpectedMessages: []string{
				"E\n",
				"D0644 0 goscp-test-dir\n",
			},
			DestinationPath:         []string{"."},
			ExpectedDestinationPath: []string{"goscp-test-dir"},
		},
	}

	for _, v := range tests {
		r, w := io.Pipe()
		c := Client{
			scpStdinPipe:    w,
			ShowProgressBar: false,
		}

		filePath := v.Name
		var stats os.FileInfo
		if v.Type == "file" {
			f, err := os.Create(filePath)
			if err != nil {
				t.Error("Unexpected error:", err)
			}

			f.Write(v.Content)
			f.Close()
		} else if v.Type == "directory" {
			err := os.MkdirAll(filePath, 0755)
			if err != nil {
				t.Error("Unexpected error:", err)
			}

			c.DestinationPath = v.DestinationPath
		}

		created = append(created, filePath)
		stats, _ = os.Stat(filePath)

		go func() {
			br := bufio.NewReader(r)

			msgCounter := 0
			for {
				msg, err := br.ReadString('\n')
				if err != nil {
					t.Error("Unexpected error:", err)
					break
				}

				if msg != v.ExpectedMessages[msgCounter] {
					expectedError(t, msg, v.ExpectedMessages[msgCounter])
				}

				msgCounter++
			}
		}()

		err := c.handleItem(filePath, stats, nil)
		if err != nil {
			t.Error("Unexpected error:", err)
		}

		if v.Type == "file" {
			// Output one more newline for convenience in reading from the pipe
			fmt.Fprintf(c.scpStdinPipe, "\n")
		} else if v.Type == "directory" {
			if !reflect.DeepEqual(c.DestinationPath, v.ExpectedDestinationPath) {
				expectedError(t, c.DestinationPath, v.ExpectedDestinationPath)
			}
		}

		os.Remove(filePath)
	}
}

func TestCancel(t *testing.T) {
	// Send creation message
	// Cancel
	// Send another creation message
	testsMessages := []string{
		"C0644 15 goscp-cancel.txt",
		"Cancel incoming\x00",
		"C0644 15 goscp-cancel.txt",
		"Transfer cancelled",
		"io: read/write on closed pipe",
	}

	r, w := io.Pipe()
	c := Client{
		scpStdinPipe:    w,
		ShowProgressBar: false,
	}

	filePath := "goscp-cancel.txt"
	f, err := os.Create(filePath)
	if err != nil {
		t.Error("Unexpected error:", err)
	}

	f.Write([]byte("Cancel incoming"))
	f.Close()

	created = append(created, filePath)
	stats, _ := os.Stat(filePath)
	msgCounter := 0

	go func() {
		c.scpStdoutPipe = &readCanceller{
			Reader: bufio.NewReader(r),
			cancel: make(chan struct{}, 1),
		}

		scanner := bufio.NewScanner(c.scpStdoutPipe)

		for scanner.Scan() {
			txt := scanner.Text()

			if txt != testsMessages[msgCounter] {
				expectedError(t, txt, testsMessages[msgCounter])
			}
			msgCounter++
		}

		err := scanner.Err()
		if err != nil {
			if err.Error() != testsMessages[msgCounter] {
				expectedError(t, err.Error(), testsMessages[msgCounter])
			}
			msgCounter++
		}
		c.scpStdinPipe.Close()
	}()

	err = c.handleItem(filePath, stats, nil)
	if err != nil {
		t.Error("Unexpected error:", err)
	}

	// Output one more newline for convenience in reading from the pipe
	fmt.Fprintf(c.scpStdinPipe, "\n")

	go c.Cancel()

	time.Sleep(time.Millisecond * 100)

	err = c.handleItem(filePath, stats, nil)
	if err != nil {
		if err.Error() != testsMessages[msgCounter] {
			expectedError(t, err.Error(), testsMessages[msgCounter])
		}
		msgCounter++
	}

	// Output one more newline for convenience in reading from the pipe
	fmt.Fprintf(c.scpStdinPipe, "\n")
}
