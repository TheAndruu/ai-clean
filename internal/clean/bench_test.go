package clean

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkClean(b *testing.B) {
	fixtures := []string{
		"claude_summary_1_sample.txt",
		"full_border_padded_sample.txt",
		"wrapped_padded_indented_sample.txt",
	}
	for _, name := range fixtures {
		data, err := os.ReadFile(filepath.Join("../../testdata/examples", name))
		if err != nil {
			b.Skipf("fixture %s not present: %v", name, err)
			return
		}
		text := string(data)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for i := 0; i < b.N; i++ {
				_, _ = Clean(text, Opts{})
			}
		})
	}
}
