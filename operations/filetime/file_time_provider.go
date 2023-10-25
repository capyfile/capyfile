package filetime

import (
	"os"
	"time"
)

type TimeStatProvider interface {
	TimeStat(os.FileInfo) (*TimeStat, error)
}

type TimeStat struct {
	Atime time.Time
	Mtime time.Time
	Ctime time.Time
}
