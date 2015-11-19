package app

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/config"
)

type newIDForTest struct {
	count    uint64
	expected string
}

func testNewIDFor(t *testing.T, a *Application, tests []newIDForTest) {
	for _, test := range tests {
		got := a.newIDFor(test.count)

		if string(got) != test.expected {
			t.Errorf("\n%s == app.newIdFor(%d) not \n%s as expected",
				got, test.count, test.expected)
		}
	}
}

func TestNewIDForWithApplicaitonID(t *testing.T) {
	a := &Application{cfg: &config.Config{BaseConfig: config.BaseConfig{ApplicationID: "ThApp"}}, started: time.Unix(1073741824, 0)}
	var tests = []newIDForTest{
		{0, "ThApp0000004000000000"},
		{1, "ThApp0000004000000001"},
		{16, "ThApp0000004000000010"},
		{10, "ThApp000000400000000a"},
		{1073741824, "ThApp0000004040000000"},
	}
	testNewIDFor(t, a, tests)
}

func TestNewIDForWithoutApplicaitonID(t *testing.T) {
	a := &Application{cfg: new(config.Config), started: time.Unix(1073741824, 0)}
	var tests = []newIDForTest{
		{0, "0000004000000000"},
		{1, "0000004000000001"},
		{16, "0000004000000010"},
		{10, "000000400000000a"},
		{1073741824, "0000004040000000"},
	}
	testNewIDFor(t, a, tests)
}

func BenchmarkNewIDFor(b *testing.B) {
	a := &Application{cfg: &config.Config{BaseConfig: config.BaseConfig{ApplicationID: "ThApp"}}, started: time.Unix(1073741824, 0)}
	var count uint64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.newIDFor(atomic.AddUint64(&count, 1))
		}
	})
}
