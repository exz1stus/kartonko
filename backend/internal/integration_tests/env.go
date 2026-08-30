package integration_tests

import (
	"os"
	"path/filepath"
	"runtime"
)

func FindEnvTestFile() string {
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(thisFile)
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, ".env.test")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		dir = filepath.Dir(dir)
	}
	return ".env.test"
}
