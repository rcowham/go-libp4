package p4

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testRoot string

func TestMain(m *testing.M) {
	fmt.Printf("TestMain: %s", os.Environ())
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to Getwd: %v\n", err)
	}
	fmt.Printf("Cwd: %s\n", dir)
	os.Setenv("PWD", dir)
	code := m.Run()
	os.Exit(code)
}

func init() {
	_, filename, _, _ := runtime.Caller(0)
	testRoot = path.Join(path.Dir(filename), "_testdir")
	err := os.Mkdir(testRoot, 0755)
	if err != nil {
		fmt.Printf("Failed to mkdir: %v\n", err)
	}
	err = os.Chdir(testRoot)
	if err != nil {
		fmt.Printf("Failed to chdir: %v\n", err)
	}
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to Getwd: %v\n", err)
	}
	fmt.Printf("Cwd: %s\n", dir)
}

func setupServer() {
	// testRoot, err := ioutil.TempDir("", "p4_test")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Temp dir: %s\n", testRoot)
	// defer os.RemoveAll(testRoot) // clean up

	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to Getwd: %v\n", err)
	}
	fmt.Printf("Cwd: %s\n", dir)
	cmd := exec.Command("pwd")
	// cmd.Dir = testRoot

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to run pwd: %s\n", stderr.String())
		log.Fatal(err)
	}
	fmt.Printf("Pwd: %q\n", out.String())

	// cmd := exec.Command("p4", "init", "-n", "-C1")
	// cmd.Dir = testRoot

	// var out, stderr bytes.Buffer
	// cmd.Stdout = &out
	// cmd.Stderr = &stderr
	// err = cmd.Run()
	// if err != nil {
	// 	fmt.Printf("Failed to start server: %s\n", stderr.String())
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Started server: %q\n", out.String())
}

func TestInfo(t *testing.T) {
	// setupServer()

	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to Getwd: %v\n", err)
	}
	fmt.Printf("Cwd: %s\n", dir)

	// out1, err := exec.Command("p4", "set").Output()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("p4 set is %s\n", out1)

	cmd := exec.Command("p4", "info")
	// cmd.Dir = testRoot
	// cmd.Env = os.Environ()
	// cmd.Env = append(cmd.Env, "P4CONFIG=.p4config")

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("P4 info: %s\nerr: %s", out.String(), stderr.String())

	// p4 := NewP4()
	// result, err := p4.Run([]string{"info"})
	// assert.Equal(t, 1, len(result))
	// fmt.Printf("P4 info result: %v\n", result[0])
	// assertMapContains(t, result[0], "serverAddress", "unknown")
}

func runUnmarshall(t *testing.T, testFile string) []map[interface{}]interface{} {
	results := make([]map[interface{}]interface{}, 0)
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
			results = append(results, r.(map[interface{}]interface{}))
		} else {
			break
		}
	}
	return results
}

func assertMapContains(t *testing.T, result map[interface{}]interface{}, key string, expected string) {
	val, ok := result[key]
	assert.True(t, ok, fmt.Sprintf("key not found: %s", key))
	assert.Equal(t, expected, val)
}

// func TestUnmarshallInfo(t *testing.T) {
// 	results := runUnmarshall(t, "info.bin")
// 	assert.Equal(t, 1, len(results))
// 	assertMapContains(t, results[0], "serverAddress", "unknown")
// }

// func TestUnmarshallChanges(t *testing.T) {
// 	results := runUnmarshall(t, "changes.bin")
// 	assert.Equal(t, 3, len(results))
// 	assertMapContains(t, results[0], "change", "3")
// 	assertMapContains(t, results[1], "change", "2")
// 	assertMapContains(t, results[2], "change", "1")

// 	assertMapContains(t, results[1], "time", "1557746038")
// 	assertMapContains(t, results[1], "user", "rcowham")
// 	assertMapContains(t, results[1], "client", "rcowham-dvcs-1557689468")
// 	assertMapContains(t, results[1], "status", "submitted")
// 	assertMapContains(t, results[1], "changeType", "public")
// 	assertMapContains(t, results[1], "path", "//stream/main/p4cmdf/*")
// 	assertMapContains(t, results[1], "desc", "second")

// 	assertMapContains(t, results[0], "desc", "Multi line change description\nS")
// }

// func TestUnmarshallChangesLongDesc(t *testing.T) {
// 	results := runUnmarshall(t, "changes-l.bin")
// 	assert.Equal(t, 3, len(results))
// 	assertMapContains(t, results[0], "change", "3")
// 	assertMapContains(t, results[1], "change", "2")
// 	assertMapContains(t, results[2], "change", "1")

// 	assertMapContains(t, results[0], "desc", "Multi line change description\nSecond line\nThird line\n")
// }

// func TestUnmarshallFetchChange(t *testing.T) {
// 	results := runUnmarshall(t, "change-o.bin")
// 	assert.Equal(t, 1, len(results))
// 	assertMapContains(t, results[0], "Change", "new")
// 	assertMapContains(t, results[0], "Status", "new")
// 	assertMapContains(t, results[0], "Description", "<enter description here>\n")
// 	assertMapContains(t, results[0], "Client", "rcowham-dvcs-1557689468")
// 	assertMapContains(t, results[0], "User", "rcowham")
// }

// func TestFormatSpec(t *testing.T) {
// 	spec := map[string]string{"Change": "new",
// 		"Description": "My line\nSecond line\nThird line\n",
// 	}
// 	assert.Equal(t, "Change: new\n\nDescription:\n My line\n Second line\n Third line\n\n", formatSpec(spec))
// }
