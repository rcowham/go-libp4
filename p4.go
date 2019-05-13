/*
Package p4 wraps the Perforce Helix Core command line.

It assumes p4 or p4.exe is in the PATH.
It uses the p4 -G global option which returns Python marshalled dictionary objects.

p4 Python parsing module is based on: https://github.com/hambster/gopymarshal

*/
package p4

import (
	"bytes"
	"io"
	"os/exec"

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

		if codeStop == code {
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
	port string
	user string
}

// NewP4 - create and initialise properly
func NewP4(port string, user string) *P4 {
	var p4 P4
	p4.port = port
	p4.user = user
	return &p4
}

// Run - runs p4 command
func (p4 *P4) Run(args []string) ([]byte, error) {
	cmd := exec.Command("p4", args...)

	if data, err := cmd.CombinedOutput(); err != nil {
		return data, err
	} else {
		return data, nil
	}
}

// RunP - runs p4 command
func (p4 *P4) RunP(args []string) ([]map[interface{}]interface{}, error) {
	nargs := []string{"-G"}
	nargs = append(nargs, args...)
	cmd := exec.Command("p4", nargs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	mainerr := cmd.Run()
	results := make([]map[interface{}]interface{}, 0)
	for {
		r, err := Unmarshal(&stdout)
		if err == io.EOF {
			break
		}
		if err == nil {
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
