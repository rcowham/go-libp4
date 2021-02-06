package p4

import (
	"fmt"
	"strconv"
	"time"
)

// Fix is a single fix from p4 fixes result
type Fix struct {
	Code   string
	Change int
	Client string
	Date   time.Time
	Job    string
	Status string
	User   string
}

// Fixes is all of the results from p4 fixes
type Fixes []Fix

// RunFixes runs p4 fixes args...
func RunFixes(p4r Runner, args []string) ([]Fix, error) {
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
	return fs, nil
}
