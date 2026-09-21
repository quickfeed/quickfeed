package interceptor

import (
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestLogNamesRefresh(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		name, calls := "old-name", 0
		cache := logNames{lookup: func(uint64) (string, error) {
			calls++
			return name, nil
		}}
		check := func(want string, wantCalls int) {
			t.Helper()
			got, err := cache.get(1)
			if err != nil || got != want || calls != wantCalls {
				t.Errorf("get(1) = %q, %v; calls = %d, want %q, nil; calls = %d", got, err, calls, want, wantCalls)
			}
		}
		check("old-name", 1)
		name = "new-name"
		time.Sleep(logNameTTL - time.Second)
		check("old-name", 1)
		time.Sleep(time.Second)
		check("new-name", 2)
		check("new-name", 2)
	})
}

func TestLogNamesRetryMissingNames(t *testing.T) {
	for _, lookupErr := range []error{nil, errors.New("database unavailable")} {
		t.Run(strconv.FormatBool(lookupErr != nil), func(t *testing.T) {
			calls := 0
			cache := logNames{lookup: func(uint64) (string, error) {
				calls++
				if calls == 1 {
					return "", lookupErr
				}
				return "recovered", nil
			}}
			if got, err := cache.get(0); got != "" || err != nil || calls != 0 {
				t.Fatalf("zero ID: get = %q, %v; calls = %d", got, err, calls)
			}
			if got, err := cache.get(1); got != "" || !errors.Is(err, lookupErr) {
				t.Fatalf("first get = %q, %v; want empty name and %v", got, err, lookupErr)
			}
			for range 2 {
				if got, err := cache.get(1); got != "recovered" || err != nil {
					t.Errorf("get after recovery = %q, %v", got, err)
				}
			}
			if calls != 2 {
				t.Errorf("lookup calls = %d, want 2", calls)
			}
		})
	}
}

func TestLogNamesConcurrentAndBounded(t *testing.T) {
	var calls atomic.Int64
	cache := logNames{lookup: func(id uint64) (string, error) {
		calls.Add(1)
		return strconv.FormatUint(id, 10), nil
	}}
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			for id := uint64(1); id <= 20; id++ {
				if got, err := cache.get(id); got != strconv.FormatUint(id, 10) || err != nil {
					t.Errorf("get(%d) = %q, %v", id, got, err)
				}
			}
		})
	}
	wg.Wait()
	before := calls.Load()
	for id := uint64(1); id <= 20; id++ {
		if _, err := cache.get(id); err != nil {
			t.Fatal(err)
		}
	}
	if got := calls.Load(); got != before {
		t.Errorf("cached lookups hit database: calls = %d, want %d", got, before)
	}
	for id := uint64(21); id <= maxLogNameKeys+20; id++ {
		if _, err := cache.get(id); err != nil {
			t.Fatal(err)
		}
	}
	if got := len(cache.values); got > maxLogNameKeys {
		t.Errorf("cache contains %d entries, want at most %d", got, maxLogNameKeys)
	}
}

// Refreshing an expired entry in a full cache must not evict another entry.
func TestLogNamesRefreshAtCapacity(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cache := logNames{lookup: func(id uint64) (string, error) {
			return strconv.FormatUint(id, 10), nil
		}}
		for id := uint64(1); id <= maxLogNameKeys; id++ {
			if _, err := cache.get(id); err != nil {
				t.Fatal(err)
			}
		}
		time.Sleep(logNameTTL)
		if _, err := cache.get(1); err != nil {
			t.Fatal(err)
		}
		if got := len(cache.values); got != maxLogNameKeys {
			t.Errorf("cache contains %d entries after refresh, want %d", got, maxLogNameKeys)
		}
	})
}
