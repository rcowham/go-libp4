package p4

import (
	"fmt"
	"strconv"
	"time"
)

type Revision struct {
	Action    string
	Rev       int
	DepotFile string
	Type      string
	Digest    string
	FileSize  int
}

type JobDescription struct {
	Job    string
	Status string
}

// Fix is a single fix from p4 fixes result
type Describe struct {
	Code       string
	Change     int
	OldChange  int
	ChangeType string
	Client     string
	Desc       string
	Path       string
	Time       time.Time
	Status     string
	User       string
	Jobs       []JobDescription
	Revisions  []Revision
}

// RunFixes runs p4 fixes args...
func RunDescribe(p4r Runner, args []string) (Describe, error) {
	args = append([]string{"describe"}, args...)
	res, err := p4r.Run(args)
	if err != nil {
		return Describe{}, fmt.Errorf("Failed to run p4 %s\n%v", args, err)
	}
	d := Describe{}
	if len(res) == 0 {
		// No response, should we error?
		return Describe{}, nil
	}
	r := res[0]
	if v, ok := r["code"]; ok {
		d.Code = v.(string)
		if d.Code == "error" {
			return Describe{}, parseError(r)
		}
	}
	if v, ok := r["change"]; ok {
		d.Change, err = strconv.Atoi(v.(string))
		if err != nil {
			return Describe{}, fmt.Errorf("Failed to parse change %s, res: %v ", v.(string), r)
		}
	}
	if v, ok := r["oldChange"]; ok {
		d.OldChange, err = strconv.Atoi(v.(string))
		if err != nil {
			return Describe{}, fmt.Errorf("Failed to parse old change %s, res: %v ", v.(string), r)
		}
	}
	if v, ok := r["changeType"]; ok {
		d.ChangeType = v.(string)
	}
	if v, ok := r["client"]; ok {
		d.Client = v.(string)
	}
	if v, ok := r["desc"]; ok {
		d.Desc = v.(string)
	}
	if v, ok := r["path"]; ok {
		d.Path = v.(string)
	}
	if v, ok := r["time"]; ok {
		epoch, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			return Describe{}, fmt.Errorf("Failed to parse Date %s, res: %v ", v.(string), r)
		}
		d.Time = time.Unix(epoch, 0)
	}
	if v, ok := r["status"]; ok {
		d.User = v.(string)
	}
	if v, ok := r["user"]; ok {
		d.User = v.(string)
	}

	return d, nil
}
