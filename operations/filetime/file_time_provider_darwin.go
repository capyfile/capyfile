package filetime

import (
	"os"
	"syscall"
	"time"
)

type PlatformTimeStatProvider struct{}

func (p *PlatformTimeStatProvider) TimeStat(fi os.FileInfo) (*TimeStat, error) {
	s := fi.Sys().(*syscall.Stat_t)

	return &TimeStat{
		Atime: time.Unix(s.Atimespec.Sec, s.Atimespec.Nsec),
		Mtime: time.Unix(s.Mtimespec.Sec, s.Mtimespec.Nsec),
		Ctime: time.Unix(s.Ctimespec.Sec, s.Ctimespec.Nsec),
	}, nil
}
