package p4

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

type p4server struct {
	test_root   string
	server_root string
	client_root string
	port        string
	user        string
	p4d         string
	p4c         P4
}

func (p4s *p4server) setupServer() {
	// string p4d

	// dir, err := ioutil.TempDir("", "p4_test")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// p4s.test_root = dir
	// fmt.Println("Temp dir: %s", p4s.test_root)
	// defer os.RemoveAll(p4s.test_root) // clean up

	// p4s.server_root = path.Join(p4s.test_root, "server")
	// p4s.client_root = path.Join(p4s.test_root, "client")
	// os.MkdirAll(p4s.server_root, os.ModeDir)
	// os.MkdirAll(p4s.client_root, os.ModeDir)

	// p4s.p4d = "p4d"
	// p4s.port = fmt.Sprintf("rsh:%s -r \"%s\" -L log -i", p4s.p4d, p4s.server_root)
	// p4s.p4c = NewP4(p4s.port, p4s.user)
	// p4s.p4c.client = p4client

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

/*
TODO this returns in wrong order
func TestFormatSpec(t *testing.T) {
	spec := map[string]string{"Change": "new",
		"Description": "My line\nSecond line\nThird line\n",
	}
	assert.Equal(t, "Change: new\n\nDescription:\n My line\n Second line\n Third line\n\n", formatSpec(spec))
}
*/
