package downloads

import (
	"os"
	"sync"
)

type SynchronizedFileWriter struct {
	mutex *sync.Mutex
	file  *os.File
}

func NewSynchronizedFileWriter(filePath string) SynchronizedFileWriter {
	var file, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic("failed to open file") // TODO
	}
	return SynchronizedFileWriter{
		mutex: &sync.Mutex{},
		file:  file,
	}
}

func (writer *SynchronizedFileWriter) Write(buffer []byte, at int64, length int64) {
	writer.mutex.Lock()
	writer.file.WriteAt(buffer[:length], at)
	writer.mutex.Unlock()
}

func (writer *SynchronizedFileWriter) Close() {
	writer.file.Close()
}
