package temperature

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// FileSource implements the Source interface for reading from a local file.
type FileSource struct {
	path string
}

// NewFileSource creates a new file-based temperature source.
func NewFileSource(path string) *FileSource {
	return &FileSource{path: path}
}

// GetTemperature reads the temperature from the file.
// It expects the file to contain an integer value in millidegrees Celsius.
func (f *FileSource) GetTemperature() (float64, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		return 0, fmt.Errorf("failed to read thermal zone: %w", err)
	}

	tempStr := strings.TrimSpace(string(data))
	tempMilli, err := strconv.ParseInt(tempStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse temperature: %w", err)
	}

	return float64(tempMilli) / 1000.0, nil
}

// Close implements the Source interface.
func (f *FileSource) Close() error {
	// No cleanup needed for file source
	return nil
}
