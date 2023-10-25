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
		Atime: time.Unix(s.Atim.Sec, s.Atim.Nsec),
		Mtime: time.Unix(s.Mtim.Sec, s.Mtim.Nsec),
		Ctime: time.Unix(s.Ctim.Sec, s.Ctim.Nsec),
	}, nil
}
