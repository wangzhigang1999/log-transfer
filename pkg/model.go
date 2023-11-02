package pkg

import (
	"errors"
	"log/slog"
	"regexp"
	"sync"
)

type TargetWorkloadMSG struct {
	Namespace string `json:"namespace"`
	Workload  string `json:"workload"`
	TailLines int64  `json:"tailLines"`
	Mode      string `json:"mode"`
}

var allowNamespace = sync.Map{}
var disallowNamespace = sync.Map{}

var allowNamespaceRegList = []string{
	".*",
}

var MaxTailLines = int64(100)

var MaxErrorCount = 10

// enum mode
const (
	ModePod = "pod"
	ModeJob = "job"
)

func (r *TargetWorkloadMSG) Valid() (bool, error) {
	if r.Namespace == "" || r.Workload == "" {
		return false, errors.New("namespace and workload must not be empty")
	}

	if r.Mode == "" {
		slog.Info("mode is empty,use default mode:pod")
		r.Mode = ModePod
	}

	if allowAccess(r.Namespace) {
		return true, nil
	}

	return false, errors.New("namespace is not allowed to access")
}

func allowAccess(ns string) bool {
	// namespace allow cache
	if _, ok := allowNamespace.Load(ns); ok {
		return true
	}

	// namespace  disallow cache
	if _, ok := disallowNamespace.Load(ns); ok {
		return false
	}

	//  check whether the namespace is allowed
	for _, reg := range allowNamespaceRegList {
		if ok, _ := regexp.MatchString(reg, ns); ok {
			// the namespace is allowed,add to allow cache
			allowNamespace.Store(ns, true)
			return true
		}
	}

	// the namespace is not allowed,add to disallow cache
	disallowNamespace.Store(ns, true)
	return false
}
