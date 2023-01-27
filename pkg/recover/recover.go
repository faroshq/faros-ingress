package recover

import (
	"fmt"
	"runtime/debug"

	"k8s.io/klog"
)

// Panic recovers a panic
func Panic() {
	if e := recover(); e != nil {
		klog.Error(fmt.Sprint("%w", e))
		klog.Error(string(debug.Stack()))
	}
}
