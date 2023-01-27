package utilstrings

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNameFormat(t *testing.T) {
	name := GetRandomName()

	if !strings.Contains(name, "-") {
		t.Fatalf("Generated name does not contain a '-'")
	}
}

func TestEnsureUnique(t *testing.T) {
	name := GetRandomName()
	second := GetRandomName()
	require.NotEqual(t, name, second)
}

func BenchmarkGetRandomName(b *testing.B) {
	b.ReportAllocs()
	var out string
	for n := 0; n < b.N; n++ {
		out = GetRandomName()
	}
	b.Log("Last result:", out)
}
