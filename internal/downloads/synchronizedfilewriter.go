package downloads

import (
	"os"
	"sync"
)

type SynchronizedFileWriter struct {
	mutex *sync.Mutex
	file  *os.File
}

func NewSynchronizedFileWriter(filePath string) *SynchronizedFileWriter {
	var file, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic("failed to open file") // TODO
	}
	return &SynchronizedFileWriter{
		mutex: &sync.Mutex{},
		file:  file,
	}
}

func (writer *SynchronizedFileWriter) WriteAt(buffer []byte, at int64) (int, error) {
	writer.mutex.Lock()
	n, err := writer.file.WriteAt(buffer, at)
	writer.mutex.Unlock()
	return n, err
}

func (writer *SynchronizedFileWriter) Close() {
	writer.file.Close()
}
