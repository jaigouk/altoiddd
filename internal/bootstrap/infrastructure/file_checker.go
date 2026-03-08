// Package infrastructure provides adapters for the Bootstrap bounded context.
package infrastructure

import "os"

// OSFileChecker implements the application.FileChecker port using os.Stat.
type OSFileChecker struct{}

// Exists returns true if the file at path exists.
func (c *OSFileChecker) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
