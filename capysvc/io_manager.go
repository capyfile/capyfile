package capysvc

import (
	"capyfile/files"
	"sync"
)

// Provides the basic methods to manage the operation's input and output.
type ioManager struct {
	inOutLock *sync.RWMutex
	in        []files.ProcessableFile
	out       []files.ProcessableFile

	procCnt int64
}

func (m *ioManager) isQueueEmpty() bool {
	m.inOutLock.RLock()
	defer m.inOutLock.RUnlock()

	return len(m.in) == 0 && len(m.out) == 0
}

func (m *ioManager) isInputQueueEmpty() bool {
	m.inOutLock.RLock()
	defer m.inOutLock.RUnlock()

	return len(m.in) == 0
}

func (m *ioManager) isOutputQueueEmpty() bool {
	m.inOutLock.RLock()
	defer m.inOutLock.RUnlock()

	return len(m.out) == 0
}

func (m *ioManager) enqueueInput(pf ...files.ProcessableFile) {
	m.inOutLock.Lock()
	defer m.inOutLock.Unlock()

	m.in = append(m.in, pf...)
}

func (m *ioManager) process(
	inSize int,
	f func(in []files.ProcessableFile) (out []files.ProcessableFile),
) {
	m.inOutLock.Lock()
	defer m.inOutLock.Unlock()

	var in []files.ProcessableFile
	if inSize == 0 {
		in = m.in
		m.in = nil
	} else {
		if len(m.in) < inSize {
			inSize = len(m.in)
		}

		in = m.in[:inSize]
		m.in = m.in[inSize:]
	}

	out := f(in)

	m.out = append(m.out, out...)

	m.procCnt += 1
}

func (m *ioManager) dequeueOutput(
	f func(out []files.ProcessableFile),
) {
	m.inOutLock.Lock()
	defer m.inOutLock.Unlock()

	if len(m.out) == 0 {
		return
	}

	f(m.out)

	m.out = nil
}
