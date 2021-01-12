package tools

import "fmt"

func BytesToFormalSize(bytes int) string {
	dbSize := float64(bytes)
	measure := "b"
	if dbSize > 1024 {
		dbSize /= 1024
		measure = "kb"
	}
	if dbSize > 1024 {
		dbSize /= 1024
		measure = "mb"
	}
	if dbSize > 1024 {
		dbSize /= 1024
		measure = "gb"
	}
	// Anything more isn't good
	return fmt.Sprintf("%.02f %s", dbSize, measure)
}
