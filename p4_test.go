package p4

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	debug    bool = true
	logger   *logrus.Logger
	testRoot string
)

func init() {
	flag.BoolVar(&debug, "debug", true, "Set to have debug logging for tests.")
}

func createLogger() *logrus.Logger {
	if logger != nil {
		return logger
	}
	logger = logrus.New()
	logger.Level = logrus.InfoLevel
	if debug {
		logger.Level = logrus.DebugLevel
	}
	return logger
}

type P4Test struct {
	startDir   string
	p4d        string
	port       string
	testRoot   string
	serverRoot string
	clientRoot string
}

func MakeP4Test(startDir string) *P4Test {
	var err error
	p4t := &P4Test{}
	p4t.startDir = startDir
	if err != nil {
		panic(err)
	}
	p4t.testRoot = filepath.Join(p4t.startDir, "testroot")
	p4t.serverRoot = filepath.Join(p4t.testRoot, "server")
	p4t.clientRoot = filepath.Join(p4t.testRoot, "client")
	p4t.ensureDirectories()
	p4t.p4d = "p4d"
	p4t.port = fmt.Sprintf("rsh:%s -r \"%s\" -L log -vserver=3 -i", p4t.p4d, p4t.serverRoot)
	os.Chdir(p4t.clientRoot)
	p4config := filepath.Join(p4t.startDir, os.Getenv("P4CONFIG"))
	writeToFile(p4config, fmt.Sprintf("P4PORT=%s", p4t.port))
	os.Chdir(p4t.serverRoot)
	cmd := exec.Command("p4d", "-xu")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Error %v %q %q", err, out.String(), stderr.String())
	}
	// log.Debugf("Upgraded", out.String(), stderr.String())
	return p4t
}

// writeToFile - write contents to file
func writeToFile(fname, contents string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(f, contents)
	if err != nil {
		_ = f.Close()
		return err
	}
	err = f.Close()
	return err
}

func (p4t *P4Test) ensureDirectories() {
	for _, d := range []string{p4t.serverRoot, p4t.clientRoot} {
		err := os.MkdirAll(d, 0777)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create %s: %v", d, err)
		}
	}
}

// TestMain - required as a wrapper for 'go test' to set directory appropriatly so that
// p4 will pick up .p4config etc.
func TestMain(m *testing.M) {
	logger := createLogger()
	logger.Debugf("TestMain: %s", os.Environ())
	dir, err := os.Getwd()
	if err != nil {
		logger.Debugf("Failed to Getwd: %v\n", err)
	}
	logger.Debugf("TM Cwd: %s\n", dir)
	os.Setenv("PWD", dir)
	code := m.Run()
	os.Exit(code)
}

// Call at start of all tests - we recreate a blank DVCS repo
func init() {
	_, filename, _, _ := runtime.Caller(0)
	testRoot = path.Join(path.Dir(filename), "_testdata")
	p4t := MakeP4Test(testRoot)
	logger := createLogger()
	logger.Debugf("Created server: %s\n", p4t.testRoot)
}

func TestInfo(t *testing.T) {
	logger := createLogger()
	logger.Debugf("======== Test: %s", t.Name())
	dir, err := os.Getwd()
	if err != nil {
		logger.Debugf("Failed to Getwd: %v\n", err)
	}
	logger.Debugf("Test Cwd: %s\n", dir)

	cmd := exec.Command("p4", "info")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Error %v %q %q", err, out.String(), stderr.String())
	}
	logger.Debugf("P4 info: %s\nerr: %s", out.String(), stderr.String())

	p4 := NewP4()
	result, err := p4.Run([]string{"info"})
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(result))
	logger.Debugf("P4 info result: %v\n", result[0])
	assertMapContains(t, result[0], "serverAddress", "unknown")
	assertMapContains(t, result[0], "clientName", "*unknown*")
	for _, k := range []string{"caseHandling", "clientName", "serverRoot", "serverUptime"} {
		assertMapKey(t, result[0], k)
	}
}

func check(t *testing.T, e error) {
	if e != nil {
		t.Error(e)
	}
}

func TestAdd(t *testing.T) {
	logger := createLogger()
	logger.Debugf("======== Test: %s", t.Name())
	file1 := "file1"

	err := os.WriteFile(file1, []byte("Some text"), 0644)
	check(t, err)

	p4 := NewP4()
	p4.client = "test_ws"
	client, err := p4.Fetch("client")
	assert.Equal(t, nil, err)
	logger.Debugf("client: %+v", client)
	client["View"] = "//depot/... //test_ws/..."
	sresult, err := p4.SaveTxt("client", client)
	logger.Debugf("save: %v, %v", sresult, err)

	result, err := p4.Run([]string{"add", file1})
	logger.Debugf("add err: %v", err)
	assert.Equal(t, 1, len(result))
	logger.Debugf("P4 add result: %v\n", result[0])
	assertMapContains(t, result[0], "depotFile", "//stream/main/file1")

	result, err = p4.Run([]string{"opened", file1})
	logger.Debugf("opened err: %v", err)
	assert.Equal(t, 1, len(result))
	logger.Debugf("P4 opened result: %v\n", result[0])
	assertMapContains(t, result[0], "depotFile", "//stream/main/file1")
	assertMapContains(t, result[0], "clientFile", fmt.Sprintf("%s/%s", testRoot, file1))
	assertMapContains(t, result[0], "change", "default")

}

func runUnmarshall(t *testing.T, testFile string) ([]map[interface{}]interface{}, []error) {
	results := make([]map[interface{}]interface{}, 0)
	errors := []error{}
	fname := path.Join(testRoot, "..", "testdata", testFile)
	buf, err := os.ReadFile(fname)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("Can't read file: %s", fname))
	}
	mbuf := bytes.NewBuffer(buf)
	for {
		r, err := Unmarshal(mbuf)
		if err == io.EOF {
			break
		}
		if err == nil {
			if r == nil {
				// Empty result for the end of the object
				break
			}
			results = append(results, r.(map[interface{}]interface{}))
		} else {
			errors = append(errors, err)
			break
		}
	}
	return results, errors
}

func assertMapContains(t *testing.T, result map[interface{}]interface{}, key string, expected string) {
	val, ok := result[key]
	assert.True(t, ok, fmt.Sprintf("key not found: %s", key))
	assert.Equal(t, expected, val)
}

func assertMapKey(t *testing.T, result map[interface{}]interface{}, key string) {
	_, ok := result[key]
	assert.True(t, ok, fmt.Sprintf("key not found: %s", key))
}

func TestUnmarshallInfo(t *testing.T) {
	logger := createLogger()
	logger.Debugf("======== Test: %s", t.Name())
	dir, err := os.Getwd()
	if err != nil {
		logger.Debugf("Failed to Getwd: %v\n", err)
	}
	logger.Debugf("Test Cwd: %s\n", dir)
	results, errors := runUnmarshall(t, "info.bin")
	assert.Equal(t, 1, len(results))
	if !assert.Equal(t, 0, len(errors)) {
		t.Fatalf("Unexpected number of errors: %v", errors)
	}

	assertMapContains(t, results[0], "serverAddress", "unknown")
}

func TestUnmarshallChanges(t *testing.T) {
	logger := createLogger()
	logger.Debugf("======== Test: %s", t.Name())
	results, errors := runUnmarshall(t, "changes.bin")
	assert.Equal(t, 3, len(results))
	if !assert.Equal(t, 0, len(errors)) {
		t.Fatalf("Unexpected number of errors: %v", errors)
	}
	assertMapContains(t, results[0], "change", "3")
	assertMapContains(t, results[1], "change", "2")
	assertMapContains(t, results[2], "change", "1")

	assertMapContains(t, results[1], "time", "1557746038")
	assertMapContains(t, results[1], "user", "rcowham")
	assertMapContains(t, results[1], "client", "rcowham-dvcs-1557689468")
	assertMapContains(t, results[1], "status", "submitted")
	assertMapContains(t, results[1], "changeType", "public")
	assertMapContains(t, results[1], "path", "//stream/main/p4cmdf/*")
	assertMapContains(t, results[1], "desc", "second")

	assertMapContains(t, results[0], "desc", "Multi line change description\nS")
}

func TestUnmarshallChangesLongDesc(t *testing.T) {
	logger := createLogger()
	logger.Debugf("======== Test: %s", t.Name())
	results, errors := runUnmarshall(t, "changes-l.bin")
	assert.Equal(t, 3, len(results))
	if !assert.Equal(t, 0, len(errors)) {
		t.Fatalf("Unexpected number of errors: %v", errors)
	}
	assertMapContains(t, results[0], "change", "3")
	assertMapContains(t, results[1], "change", "2")
	assertMapContains(t, results[2], "change", "1")

	assertMapContains(t, results[0], "desc", "Multi line change description\nSecond line\nThird line\n")
}

func TestUnmarshallFetchChange(t *testing.T) {
	logger := createLogger()
	logger.Debugf("======== Test: %s", t.Name())
	results, errors := runUnmarshall(t, "change-o.bin")
	assert.Equal(t, 1, len(results))
	if !assert.Equal(t, 0, len(errors)) {
		t.Fatalf("Unexpected number of errors: %v", errors)
	}
	assertMapContains(t, results[0], "Change", "new")
	assertMapContains(t, results[0], "Status", "new")
	assertMapContains(t, results[0], "Description", "<enter description here>\n")
	assertMapContains(t, results[0], "Client", "rcowham-dvcs-1557689468")
	assertMapContains(t, results[0], "User", "rcowham")
}

func TestUnmarshallFetchProtects(t *testing.T) {
	logger := createLogger()
	logger.Debugf("======== Test: %s", t.Name())
	results, errors := runUnmarshall(t, "protects.bin")
	assert.Equal(t, 4, len(results))
	if !assert.Equal(t, 0, len(errors)) {
		t.Fatalf("Unexpected number of errors: %v", errors)
	}
	assertMapContains(t, results[0], "code", "stat")
	assertMapContains(t, results[0], "perm", "write")
	assertMapContains(t, results[0], "host", "*")
	assertMapContains(t, results[0], "user", "*")
	assertMapContains(t, results[0], "line", "1")
	assertMapContains(t, results[0], "depotFile", "//...")
	assertMapContains(t, results[3], "code", "stat")
	assertMapContains(t, results[3], "perm", "super")
	assertMapContains(t, results[3], "host", "*")
	assertMapContains(t, results[3], "user", "*")
	assertMapContains(t, results[3], "line", "4")
	assertMapContains(t, results[3], "depotFile", "//...")
}

func TestFormatSpec(t *testing.T) {
	logger := createLogger()
	logger.Debugf("======== Test: %s", t.Name())
	spec := map[string]string{"Change": "new",
		"Description": "My line\nSecond line\nThird line\n",
	}
	// Order of lines isn't deterministic, maps don't retain order
	res := formatSpec(spec)
	assert.Regexp(t, regexp.MustCompile("Change: new\n\n"), res)
	assert.Regexp(t, regexp.MustCompile("Description:\n My line\n Second line\n Third line\n\n"), res)

}

type parseErrorTest struct {
	input map[interface{}]interface{}
	want  error
}

var parseErrorTests = []parseErrorTest{
	{
		input: map[interface{}]interface{}{
			"code":     "error",
			"data":     "//fake/depot/... - must refer to client 'HOSTNAME'.",
			"generic":  "2",
			"severity": "3",
		},
		want: errors.New("P4Error -> No such area '//fake/depot/...', please check your path"),
	},
	{
		input: map[interface{}]interface{}{
			"code":     "error",
			"data":     "some unknown error",
			"generic":  "2",
			"severity": "3",
		},
		want: errors.New("P4Error -> some unknown error"),
	},
}

func TestParseError(t *testing.T) {
	logger := createLogger()
	logger.Debugf("======== Test: %s", t.Name())
	for _, tst := range parseErrorTests {
		err := parseError(tst.input)
		assert.Equal(t, tst.want, err)
	}
}

// func TestSave(t *testing.T) {
// 	logger := createLogger()
// 	logger.Debugf("======== Test: %s", t.Name())
// 	ds := map[string]string{
// 		"Job":         "DEV-123",
// 		"Title":       "A failing update",
// 		"Status":      "UNKNOWN",
// 		"Assignee":    "a.person@email.com",
// 		"Description": "Desc2",
// 	}
// 	// p4 := NewP4Params("p4training.hh.imgtec.org:1666", "brett.bates", "p4go_test_ws")
// 	p4 := NewP4()
// 	res, err := p4.SaveTxt("job", ds, []string{})
// 	assert.Nil(t, err)
// 	fmt.Println(res)
// }
