package profiler

import (
	"fmt"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cornelk/hashmap"
	"github.com/sirupsen/logrus"
)

type stats struct {
	elapsed, user, system int64
	calls                 int64
}

var durations hashmap.HashMap
var prev = &hashmap.HashMap{}

func toTime(t syscall.Timeval) time.Time {
	return time.Unix(t.Sec, t.Usec*1000)
}

func before() (time.Time, syscall.Rusage) {
	var rstart syscall.Rusage
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &rstart)
	start := time.Now()
	return start, rstart
}

func after(label string, start time.Time, rstart syscall.Rusage) {
	var rend syscall.Rusage
	end := time.Now()
	duration := end.Sub(start)
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &rend)
	ptr, _ := durations.GetOrInsert(label, &stats{})
	current := ptr.(*stats)
	atomic.AddInt64(&current.elapsed, duration.Nanoseconds())
	atomic.AddInt64(&current.calls, 1)
	atomic.AddInt64(&current.user, (toTime(rend.Utime).Sub(toTime(rstart.Utime)).Nanoseconds()))
	atomic.AddInt64(&current.system, (toTime(rend.Stime).Sub(toTime(rstart.Stime)).Nanoseconds()))
}

func TimeIt(f func(), label string) {
	start, rstart := before()
	f()
	after(label, start, rstart)
}

func formatter() []string {
	ret := make([]string, 0)
	newPrev := &hashmap.HashMap{}
	for kv := range durations.Iter() {
		v := kv.Value.(*stats)
		newValue := *v
		newPrev.Set(kv.Key, &newValue)
		prevValueIf, ok := prev.Get(kv.Key)
		if !ok {
			ret = append(ret, fmt.Sprintf("%s  elapsed %+v ms user %+v ms sys %+v ms calls %d", kv.Key, v.elapsed/1000000, v.user/1000000, v.system/1000000, v.calls))
		} else {
			prevValue := prevValueIf.(*stats)
			ret = append(ret, fmt.Sprintf("%s  elapsed %+v/%+v ms user %+v/%+v ms sys %+v/%+v ms calls %d/%d", kv.Key,
				(v.elapsed-prevValue.elapsed)/1000000, v.elapsed/1000000,
				(v.user-prevValue.user)/1000000, v.user/1000000,
				(v.system-prevValue.system)/1000000, v.system/1000000,
				v.calls-prevValue.calls, v.calls))
		}
	}
	prev = newPrev
	return ret
}

func PrintTimes() {
	for _, s := range formatter() {
		fmt.Println(s)
	}
}

func Measure(label string) func() {
	start, rstart := before()
	return func() {
		after(label, start, rstart)
	}
}

func PeriodicPrinter(log logrus.FieldLogger) {
	for {
		time.Sleep(time.Minute)
		for _, s := range formatter() {
			log.Infof("stats: %s", s)
		}
	}
}
