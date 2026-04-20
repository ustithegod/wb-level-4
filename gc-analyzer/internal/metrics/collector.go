package metrics

import (
	"runtime"
	"time"
)

type Metric struct {
	Name  string
	Help  string
	Type  string
	Value float64
}

type Collector struct{}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) Snapshot(gcPercent int) []Metric {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	lastGCTimestamp := float64(0)
	if mem.LastGC != 0 {
		lastGCTimestamp = float64(mem.LastGC) / float64(time.Second)
	}

	return []Metric{
		{
			Name:  "go_gc_alloc_bytes_total",
			Help:  "Total bytes allocated during the lifetime of the process.",
			Type:  "counter",
			Value: float64(mem.TotalAlloc),
		},
		{
			Name:  "go_gc_mallocs_total",
			Help:  "Total number of heap objects allocated.",
			Type:  "counter",
			Value: float64(mem.Mallocs),
		},
		{
			Name:  "go_gc_frees_total",
			Help:  "Total number of heap objects freed.",
			Type:  "counter",
			Value: float64(mem.Frees),
		},
		{
			Name:  "go_gc_cycles_total",
			Help:  "Total number of completed GC cycles.",
			Type:  "counter",
			Value: float64(mem.NumGC),
		},
		{
			Name:  "go_gc_pause_total_ns",
			Help:  "Total GC pause time in nanoseconds.",
			Type:  "counter",
			Value: float64(mem.PauseTotalNs),
		},
		{
			Name:  "go_gc_last_run_time_seconds",
			Help:  "Unix timestamp of the last completed GC cycle.",
			Type:  "gauge",
			Value: lastGCTimestamp,
		},
		{
			Name:  "go_memory_alloc_bytes",
			Help:  "Bytes of allocated heap objects.",
			Type:  "gauge",
			Value: float64(mem.Alloc),
		},
		{
			Name:  "go_memory_heap_inuse_bytes",
			Help:  "Bytes in in-use heap spans.",
			Type:  "gauge",
			Value: float64(mem.HeapInuse),
		},
		{
			Name:  "go_memory_heap_idle_bytes",
			Help:  "Bytes in idle heap spans.",
			Type:  "gauge",
			Value: float64(mem.HeapIdle),
		},
		{
			Name:  "go_memory_heap_objects",
			Help:  "Number of allocated heap objects.",
			Type:  "gauge",
			Value: float64(mem.HeapObjects),
		},
		{
			Name:  "go_memory_stack_inuse_bytes",
			Help:  "Bytes in stack spans that are in use.",
			Type:  "gauge",
			Value: float64(mem.StackInuse),
		},
		{
			Name:  "go_runtime_goroutines",
			Help:  "Current number of goroutines.",
			Type:  "gauge",
			Value: float64(runtime.NumGoroutine()),
		},
		{
			Name:  "go_gc_target_percent",
			Help:  "Configured GC percent value passed to debug.SetGCPercent.",
			Type:  "gauge",
			Value: float64(gcPercent),
		},
	}
}
