package clickhouse

import (
	"testing"

	driver "github.com/ClickHouse/clickhouse-go/v2"
)

func TestIngestionActivityHardLimitsCannotBeOverridden(t *testing.T) {
	source := driver.Settings{"max_execution_time": 30, "max_threads": 100, "max_rows_to_read": 0, "max_bytes_to_read": 0, "read_overflow_mode": "break"}
	got := ingestionActivitySettings(source)
	want := driver.Settings{"max_execution_time": 3, "max_threads": 2, "max_rows_to_read": 10000000, "max_bytes_to_read": 2147483648, "read_overflow_mode": "throw"}
	for key, v := range want {
		if got[key] != v {
			t.Fatalf("%s=%v, want %v", key, got[key], v)
		}
	}
	if source["max_execution_time"] != 30 || source["max_rows_to_read"] != 0 {
		t.Fatal("source settings mutated")
	}
}
