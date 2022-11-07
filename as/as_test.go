package as_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/as"
)

func TestBasic(t *testing.T) {
	assertOk(t, "Equal", func(t testing.TB) {
		as.Equal(t, 1, 1)
	})
	assertOk(t, "NotEqual", func(t testing.TB) {
		as.NotEqual(t, 1, 2)
	})
	assertOk(t, "Error", func(t testing.TB) {
		as.Error(t, errors.New("sth wrong"))
	})
	assertOk(t, "NoError", func(t testing.TB) {
		as.NoError(t, nil)
	})
	assertOk(t, "Panics", func(t testing.TB) {
		as.Panics(t, func() {
			panic(0)
		})
	})
	assertOk(t, "NotPanics", func(t testing.TB) {
		as.NotPanics(t, func() {})
	})
	assertOk(t, "Regexp", func(t testing.TB) {
		as.Regexp(t, "^start", "start of the line")
	})
	assertOk(t, "NotRegexp", func(t testing.TB) {
		as.NotRegexp(t, "^asdfastart", "Not the start of the line")
	})
	assertOk(t, "True", func(t testing.TB) {
		as.True(t, true)
	})
	assertOk(t, "False", func(t testing.TB) {
		as.False(t, false)
	})
	assertOk(t, "Eventually", func(t testing.TB) {
		as.Eventually(t, func(ctx context.Context) bool {
			time.Sleep(time.Millisecond * 500)
			return true
		}, time.Second, time.Millisecond*100)
	})
	assertOk(t, "Never", func(t testing.TB) {
		as.Never(t, func(ctx context.Context) bool {
			time.Sleep(time.Millisecond * 500)
			return false
		}, time.Second, time.Millisecond*100)
	})
}

func TestFail(t *testing.T) {
	assertFail(t, "Equal", func(t testing.TB) {
		as.Equal(t, 1, 2)
	})
	assertFail(t, "NotEqual", func(t testing.TB) {
		as.NotEqual(t, 1, 1)
	})
	assertFail(t, "Error", func(t testing.TB) {
		as.Error(t, nil)
	})
	assertFail(t, "NoError", func(t testing.TB) {
		as.NoError(t, errors.New("sth wrong"))
	})
	assertFail(t, "Panics", func(t testing.TB) {
		as.Panics(t, func() {})
	})
	assertFail(t, "NotPanics", func(t testing.TB) {
		as.NotPanics(t, func() {
			panic(0)
		})
	})
	assertFail(t, "Regexp", func(t testing.TB) {
		as.Regexp(t, "^asdfastart", "Not the start of the line")
	})
	assertFail(t, "NotRegexp", func(t testing.TB) {
		as.NotRegexp(t, "^start", "start of the line")
	})
	assertFail(t, "True", func(t testing.TB) {
		as.True(t, false)
	})
	assertFail(t, "False", func(t testing.TB) {
		as.False(t, true)
	})
	assertFail(t, "Eventually", func(t testing.TB) {
		as.Eventually(t, func(ctx context.Context) bool {
			time.Sleep(time.Millisecond * 500)
			return false
		}, time.Second, time.Millisecond*100)
	})
	assertFail(t, "Never", func(t testing.TB) {
		as.Never(t, func(ctx context.Context) bool {
			time.Sleep(time.Millisecond * 500)
			return true
		}, time.Second, time.Millisecond*100)
	})
}

func TestFailDirectly(t *testing.T) {
	assertFail(t, "True", func(t testing.TB) {
		tf := as.New(t)
		tf.FailDirectly(true).True(false)
		t.(*testTester).checkFailNow(true)

		tf.FailDirectly(false).True(false)
		t.(*testTester).checkFailNow(false)
	})
}

type testTester struct {
	*testing.T
	errorMsg   string
	hasFailNow bool
}

func (t *testTester) Errorf(message string, args ...interface{}) {
	t.errorMsg = fmt.Sprintf(message, args...)
}

func (t *testTester) Error(args ...interface{}) {
	t.errorMsg = fmt.Sprint(args...)
}

func (t *testTester) FailNow() { t.hasFailNow = true }

func (t *testTester) Fail() { t.hasFailNow = false }

func (t *testTester) checkFailNow(failNow bool) {
	if t.hasFailNow != failNow {
		t.Fatalf("Fail now not match, expect: %v", failNow)
	}
}

func assertOk(t *testing.T, name string, fn func(t testing.TB)) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		t.Helper()
		tester := &testTester{T: t}
		fn(tester)
		if tester.errorMsg != "" {
			t.Fatal("Should not have failed with:\n", tester.errorMsg)
		}
	})
}

func assertFail(t *testing.T, name string, fn func(t testing.TB)) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		t.Helper()
		tester := &testTester{T: t}
		fn(tester)
		if tester.errorMsg == "" {
			t.Fatal("Should have failed")
		} else {
			t.Log(tester.errorMsg)
		}
	})
}
