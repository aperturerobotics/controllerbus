package hot_loader

import (
	"os"
	"time"
)

// PluginStat contains plugin stats.
type PluginStat struct {
	binarySize int64
	mTime      time.Time
}

// NewPluginStat stats a plugin file.
func NewPluginStat(filePath string) (*PluginStat, error) {
	fileSt, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	return &PluginStat{
		binarySize: fileSt.Size(),
		mTime:      fileSt.ModTime(),
	}, nil
}

// GetBinarySize returns the binary size.
func (s *PluginStat) GetBinarySize() int64 {
	return s.binarySize
}

// GetModificationTime returns the modification time.
func (s *PluginStat) GetModificationTime() time.Time {
	return s.mTime
}

// Equal compares two plugin stats.
func (s *PluginStat) Equal(other *PluginStat) bool {
	if s == other {
		return true
	}
	if s == nil || other == nil {
		return false
	}
	return s.mTime.Equal(other.mTime) && s.binarySize == other.binarySize
}
