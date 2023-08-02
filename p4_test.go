package p4

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	verboseLog = false
	testRoot   string
)

// TestMain - required as a wrapper for 'go test' to set directory appropriatly so that
// p4 will pick up .p4config etc.
func TestMain(m *testing.M) {
	if verboseLog {
		log.Printf("TestMain: %s", os.Environ())
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to Getwd: %v\n", err)
	}
	if verboseLog {
		log.Printf("TM Cwd: %s\n", dir)
	}
	os.Setenv("PWD", dir)
	code := m.Run()
	os.Exit(code)
}

// Call at start of all tests - we recreate a blank DVCS repo
func init() {
	_, filename, _, _ := runtime.Caller(0)
	testRoot = path.Join(path.Dir(filename), "_testdata")
	setupServer()
}

// Creates blank DVCS repo
func setupServer() {
	os.RemoveAll(testRoot + "/")
	err := os.Mkdir(testRoot, 0755)
	if err != nil {
		log.Printf("Failed to mkdir: %v\n", err)
	}
	err = os.Chdir(testRoot)
	if err != nil {
		log.Printf("Failed to chdir: %v\n", err)
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to Getwd: %v\n", err)
	}
	if verboseLog {
		log.Printf("Init Cwd: %s\n", dir)
	}
	os.Setenv("PWD", dir)
	cmd := exec.Command("p4", "init", "-n", "-C1")

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Printf("Failed to start server: %s\n", stderr.String())
		log.Fatal(err)
	}
	if verboseLog {
		log.Printf("Started server: %q\n", out.String())
	}
}

func TestInfo(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Logf("Failed to Getwd: %v\n", err)
	}
	t.Logf("Test Cwd: %s\n", dir)

	cmd := exec.Command("p4", "info")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("P4 info: %s\nerr: %s", out.String(), stderr.String())

	p4 := NewP4()
	result, err := p4.Run([]string{"info"})
	assert.Equal(t, 1, len(result))
	t.Logf("P4 info result: %v\n", result[0])
	assertMapContains(t, result[0], "serverAddress", "unknown")
	assertMapContains(t, result[0], "clientStream", "//stream/main")
	for _, k := range []string{"caseHandling", "clientName", "serverName", "serverUptime"} {
		assertMapKey(t, result[0], k)
	}
}

func check(t *testing.T, e error) {
	if e != nil {
		t.Error(e)
	}
}

func TestAdd(t *testing.T) {
	file1 := "file1"

	err := ioutil.WriteFile(file1, []byte("Some text"), 0644)
	check(t, err)

	p4 := NewP4()
	result, err := p4.Run([]string{"add", file1})
	t.Logf("add err: %v", err)
	assert.Equal(t, 1, len(result))
	t.Logf("P4 add result: %v\n", result[0])
	assertMapContains(t, result[0], "depotFile", "//stream/main/file1")

	result, err = p4.Run([]string{"opened", file1})
	t.Logf("opened err: %v", err)
	assert.Equal(t, 1, len(result))
	t.Logf("P4 opened result: %v\n", result[0])
	assertMapContains(t, result[0], "depotFile", "//stream/main/file1")
	assertMapContains(t, result[0], "clientFile", fmt.Sprintf("%s/%s", testRoot, file1))
	assertMapContains(t, result[0], "change", "default")

}

func runUnmarshall(t *testing.T, testFile string) ([]map[interface{}]interface{}, []error) {
	results := make([]map[interface{}]interface{}, 0)
	errors := []error{}
	fname := path.Join("testdata", testFile)
	buf, err := ioutil.ReadFile(fname)
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
	results, errors := runUnmarshall(t, "info.bin")
	assert.Equal(t, 1, len(results))
	if !assert.Equal(t, 0, len(errors)) {
		t.Fatalf("Unexpected number of errors: %v", errors)
	}

	assertMapContains(t, results[0], "serverAddress", "unknown")
}

func TestUnmarshallChanges(t *testing.T) {
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
	for _, tst := range parseErrorTests {
		err := parseError(tst.input)
		assert.Equal(t, tst.want, err)
	}
}

func TestSave(t *testing.T) {
	ds := map[string]string{
		"Job":         "DEV-123",
		"Title":       "A failing update",
		"Status":      "UNKNOWN",
		"Assignee":    "a.person@email.com",
		"Description": "Desc2",
	}
	p4 := NewP4Params("p4training.hh.imgtec.org:1666", "brett.bates", "p4go_test_ws")
	res, err := p4.SaveTxt("job", ds, []string{})
	assert.Nil(t, err)
	fmt.Println(res)
}
