/*
Package p4 wraps the Perforce Helix Core command line.

It assumes p4 or p4.exe is in the PATH.
It uses the p4 -G global option which returns Python marshalled dictionary objects.

p4 Python parsing module is based on: https://github.com/hambster/gopymarshal
*/
package p4

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"encoding/binary"
	"errors"
	"math"
)

// Parsing constants
const (
	codeNone     = 'N' //None
	codeInt      = 'i' //integer
	codeInt2     = 'c' //integer2
	codeFloat    = 'g' //float
	codeString   = 's' //string
	codeUnicode  = 'u' //unicode string
	codeTString  = 't' //tstring?
	codeTuple    = '(' //tuple
	codeList     = '[' //list
	codeDict     = '{' //dict
	codeStop     = '0'
	codeEnd      = 0 //end of the object
	dictInitSize = 64
)

// Parse error
var (
	ErrParse       = errors.New("invalid data")
	ErrUnknownCode = errors.New("unknown code")
)

// Unmarshal data serialized by python
func Unmarshal(buffer *bytes.Buffer) (ret interface{}, retErr error) {
	ret, _, retErr = Unmarshal2(buffer)
	return
}

// Unmarshal2 data serialized by python, returning the unused portion.
func Unmarshal2(buffer *bytes.Buffer) (ret interface{}, remainder []byte, retErr error) {
	code, err := buffer.ReadByte()
	if nil != err {
		retErr = err
	}
	ret, retErr = unmarshal(code, buffer)
	remainder = buffer.Bytes()
	return
}

func unmarshal(code byte, buffer *bytes.Buffer) (ret interface{}, retErr error) {
	switch code {
	case codeNone:
		ret = nil
	case codeInt:
		fallthrough
	case codeInt2:
		ret, retErr = readInt32(buffer)
	case codeFloat:
		ret, retErr = readFloat64(buffer)
	case codeString:
		fallthrough
	case codeUnicode:
		fallthrough
	case codeTString:
		ret, retErr = readString(buffer)
	case codeTuple:
		fallthrough
	case codeList:
		ret, retErr = readList(buffer)
	case codeDict:
		ret, retErr = readDict(buffer)
	case codeEnd:
		ret, retErr = nil, nil
	default:
		retErr = ErrUnknownCode
	}

	return
}

func readInt32(buffer *bytes.Buffer) (ret int32, retErr error) {
	var tmp int32
	retErr = ErrParse
	if retErr = binary.Read(buffer, binary.LittleEndian, &tmp); nil == retErr {
		ret = tmp
	}

	return
}

func readFloat64(buffer *bytes.Buffer) (ret float64, retErr error) {
	retErr = ErrParse
	tmp := make([]byte, 8)
	if num, err := buffer.Read(tmp); nil == err && 8 == num {
		bits := binary.LittleEndian.Uint64(tmp)
		ret = math.Float64frombits(bits)
		retErr = nil
	}

	return
}

func readString(buffer *bytes.Buffer) (ret string, retErr error) {
	var strLen int32
	strLen = 0
	retErr = ErrParse
	if err := binary.Read(buffer, binary.LittleEndian, &strLen); nil != err {
		retErr = err
		return
	}

	retErr = nil
	buf := make([]byte, strLen)
	buffer.Read(buf)
	ret = string(buf)
	return
}

func readList(buffer *bytes.Buffer) (ret []interface{}, retErr error) {
	var listSize int32
	if retErr = binary.Read(buffer, binary.LittleEndian, &listSize); nil != retErr {
		return
	}

	var code byte
	var err error
	var val interface{}
	ret = make([]interface{}, int(listSize))
	for idx := 0; idx < int(listSize); idx++ {
		code, err = buffer.ReadByte()
		if nil != err {
			break
		}

		val, err = unmarshal(code, buffer)
		if nil != err {
			retErr = err
			break
		}
		ret = append(ret, val)
	} //end of read loop

	return
}

func readDict(buffer *bytes.Buffer) (ret map[interface{}]interface{}, retErr error) {
	var code byte
	var err error
	var key interface{}
	var val interface{}
	ret = make(map[interface{}]interface{})
	for {
		code, err = buffer.ReadByte()
		if nil != err {
			break
		}

		if code == codeStop {
			break
		}

		key, err = unmarshal(code, buffer)
		if nil != err {
			retErr = err
			break
		}

		code, err = buffer.ReadByte()
		if nil != err {
			break
		}

		val, err = unmarshal(code, buffer)
		if nil != err {
			retErr = err
			break
		}
		ret[key] = val
	} //end of read loop

	return
}

// P4 - environment for P4
type P4 struct {
	port   string
	user   string
	client string
}

// NewP4 - create and initialise properly
func NewP4() *P4 {
	var p4 P4
	return &p4
}

// NewP4Params - create and initialise with params
func NewP4Params(port string, user string, client string) *P4 {
	var p4 P4
	p4.port = port
	p4.user = user
	p4.client = client
	return &p4
}

// RunBytes - runs p4 command and returns []byte output
func (p4 *P4) RunBytes(args []string) ([]byte, error) {
	cmd := exec.Command("p4", args...)

	data, err := cmd.CombinedOutput()
	if err != nil {
		return data, err
	}
	return data, nil
}

// Get options that go before the p4 command
func (p4 *P4) getOptions() []string {
	opts := []string{"-G"}

	if p4.port != "" {
		opts = append(opts, "-p", p4.port)
	}
	if p4.user != "" {
		opts = append(opts, "-u", p4.user)
	}
	if p4.client != "" {
		opts = append(opts, "-c", p4.client)
	}
	return opts
}

// Get options that go before the p4 command
func (p4 *P4) getOptionsNonMarshal() []string {
	opts := []string{}

	if p4.port != "" {
		opts = append(opts, "-p", p4.port)
	}
	if p4.user != "" {
		opts = append(opts, "-u", p4.user)
	}
	if p4.client != "" {
		opts = append(opts, "-c", p4.client)
	}
	return opts
}

// Runner is an interface to make testing p4 commands more easily
type Runner interface {
	Run([]string) ([]map[interface{}]interface{}, error)
}

// Run - runs p4 command and returns map
func (p4 *P4) Run(args []string) ([]map[interface{}]interface{}, error) {
	opts := p4.getOptions()
	args = append(opts, args...)
	cmd := exec.Command("p4", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	mainerr := cmd.Run()
	// May not be the correct place to do this
	// But we are ignoring the actual error otherwise
	if stderr.Len() > 0 {
		return nil, errors.New(stderr.String())
	}
	results := make([]map[interface{}]interface{}, 0)
	for {
		r, err := Unmarshal(&stdout)
		if err == io.EOF {
			break
		}
		if err == nil {
			if r == nil {
				// End of object
				break
			}
			results = append(results, r.(map[interface{}]interface{}))
		} else {
			if mainerr == nil {
				mainerr = err
			}
			break
		}
	}
	return results, mainerr
}

// parseError turns perforce error messages into go error's
func parseError(res map[interface{}]interface{}) error {
	var err error
	var e string
	if v, ok := res["data"]; ok {
		e = v.(string)
	} else {
		// I don't know if we can get in this situation
		e = fmt.Sprintf("Failed to parse error %v", err)
		return errors.New(e)
	}
	// Search for non-existent depot error
	nodepot, err := regexp.Match(`must refer to client`, []byte(e))
	if err != nil {
		return err // Do we need to return (error, error) for real error and parsed one?
	}
	if nodepot {
		path := strings.Split(e, " - must")[0]
		return errors.New("P4Error -> No such area '" + path + "', please check your path")
	}
	err = fmt.Errorf("P4Error -> %s", e)
	return err
}

// Assume multiline entries should be on seperate lines
func formatSpec(specContents map[string]string) string {
	var output bytes.Buffer
	for k, v := range specContents {
		if strings.Index(v, "\n") > -1 {
			output.WriteString(fmt.Sprintf("%s:", k))
			lines := strings.Split(v, "\n")
			for i := range lines {
				if len(strings.TrimSpace(lines[i])) > 0 {
					output.WriteString(fmt.Sprintf("\n %s", lines[i]))
				}
			}
			output.WriteString("\n\n")
		} else {
			output.WriteString(fmt.Sprintf("%s: %s\n\n", k, v))
		}
	}
	return output.String()
}

// Save - runs p4 -i for specified spec returns result
func (p4 *P4) Save(specName string, specContents map[string]string, args ...string) ([]map[interface{}]interface{}, error) {
	opts := p4.getOptions()
	nargs := []string{specName, "-i"}
	nargs = append(nargs, args...)
	args = append(opts, nargs...)

	log.Println(args)
	cmd := exec.Command("p4", args...)
	var stdout, stderr bytes.Buffer
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("An error occured: ", err)
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	mainerr := cmd.Start()
	if mainerr != nil {
		fmt.Println("An error occured: ", mainerr)
	}
	spec := formatSpec(specContents)
	log.Println(spec)
	io.WriteString(stdin, spec)
	stdin.Close()
	cmd.Wait()

	results := make([]map[interface{}]interface{}, 0)
	for {
		r, err := Unmarshal(&stdout)
		if err == io.EOF || r == nil {
			break
		}
		if err == nil {
			results = append(results, r.(map[interface{}]interface{}))
			fmt.Println(r)
		} else {
			if mainerr == nil {
				mainerr = err
			}
			break
		}
	}
	return results, mainerr
}

// Fetch - runs p4 <cmd> -o for specified spec and returns result
func (p4 *P4) Fetch(specName string, args ...string) (map[string]string, error) {
	opts := p4.getOptions()
	nargs := []string{specName, "-o"}
	nargs = append(nargs, args...)
	args = append(opts, nargs...)

	cresult, err := p4.Run(args)
	result := make(map[string]string, 0)
	if len(cresult) == 0 {
		return result, err
	}
	for i, v := range cresult[0] {
		log.Printf("%v: %v", i, v)
		// result[k.(string)] = v.(string)
	}
	return result, err
}

// The Save() func doesn't work as it needs the data marshalled instead of
// map[string]string
// This is a quick fix, the real fix is writing a marshal() function or try
// using gopymarshal
func (p4 *P4) SaveTxt(specName string, specContents map[string]string, args ...string) (string, error) {
	opts := p4.getOptionsNonMarshal()
	nargs := []string{specName, "-i"}
	nargs = append(nargs, args...)
	args = append(opts, nargs...)

	log.Println(args)
	cmd := exec.Command("p4", args...)
	var stdout, stderr bytes.Buffer
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("An error occured: ", err)
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	mainerr := cmd.Start()
	if mainerr != nil {
		fmt.Println("An error occured: ", mainerr)
	}
	spec := formatSpec(specContents)
	log.Println(spec)
	io.WriteString(stdin, spec)
	// Need to explicitly call this for the command to fire
	stdin.Close()
	cmd.Wait()

	e, err := io.ReadAll(&stderr)
	log.Println(e)
	if len(e) > 0 {
		return "", errors.New(string(e))
	}
	x, err := io.ReadAll(&stdout)
	s := string(x)
	log.Println(s)
	return s, mainerr
}
