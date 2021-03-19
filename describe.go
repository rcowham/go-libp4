package p4

import (
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
	/*
		args = append([]string{"fixes"}, args...)
		res, err := p4r.Run(args)
		if err != nil {
			return nil, fmt.Errorf("Failed to run p4 %s\n%v", args, err)
		}
		fs := Fixes{}
		for _, r := range res {
			f := Fix{}
			if v, ok := r["code"]; ok {
				f.Code = v.(string)
				if f.Code == "error" {
					return nil, parseError(r)
				}
			}
			if v, ok := r["Change"]; ok {
				f.Change, err = strconv.Atoi(v.(string))
				if err != nil {
					return nil, fmt.Errorf("Failed to parse change %s, res: %v ", v.(string), r)
				}
			}
			if v, ok := r["Client"]; ok {
				f.Client = v.(string)
			}
			if v, ok := r["Date"]; ok {
				epoch, err := strconv.ParseInt(v.(string), 10, 64)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse Date %s, res: %v ", v.(string), r)
				}
				f.Date = time.Unix(epoch, 0)
			}
			if v, ok := r["Job"]; ok {
				f.Job = v.(string)
			}
			if v, ok := r["Status"]; ok {
				f.Status = v.(string)
			}
			if v, ok := r["User"]; ok {
				f.User = v.(string)
			}

			fs = append(fs, f)
		}
	*/
	return Describe{}, nil
}
