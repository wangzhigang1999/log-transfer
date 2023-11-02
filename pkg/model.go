package pkg

import (
	"errors"
	"log/slog"
	"regexp"
	"sync"
)

type ReaderMSG struct {
	Namespace string `json:"namespace"`
	Workload  string `json:"workload"`
	TailLines int64  `json:"tailLines"`
	Mode      string `json:"mode"`
}

var allowNamespace = sync.Map{}
var disallowNamespace = sync.Map{}

var allowNamespaceRegList = []string{
	"wanz",
	"schedule",
	"train-job",
}

var MaxTailLines = int64(100)

var MaxErrorCount = 10

// enum mode
const (
	ModePod = "pod"
	ModeJob = "job"
)

func (r *ReaderMSG) Valid() (bool, error) {
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
	// 允许的 namespace
	if _, ok := allowNamespace.Load(ns); ok {
		return true
	}

	// 不允许的 namespace
	if _, ok := disallowNamespace.Load(ns); ok {
		return false
	}

	//  不知道，需要判断
	for _, reg := range allowNamespaceRegList {
		if ok, _ := regexp.MatchString(reg, ns); ok {
			// 判断通过，添加到允许列表
			allowNamespace.Store(ns, true)
			return true
		}
	}

	// 否则添加到不允许列表
	disallowNamespace.Store(ns, true)
	return false
}
