package p4

import (
	"bytes"
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
func TestUnmarshallInfo(t *testing.T) {
	results := make([]map[interface{}]interface{}, 0)
	buf, err := ioutil.ReadFile(path.Join("testdata", "info.bin"))
	if err != nil {
		assert.Fail(t, "Can't read file")
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
	assert.Equal(t, 1, len(results))
	r := results[0]
	val, ok := r["serverAddress"]
	assert.True(t, ok)
	assert.Equal(t, "unknown", val)
	// for _, r := range results {
	// 	for k, v := range r {
	// 		fmt.Printf("%v: %v\n", k, v)
	// 	}
	// 	fmt.Printf("\n")
	// }

}
