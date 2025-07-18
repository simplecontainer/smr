package tail

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

func File(ctx context.Context, path string, follow bool) (io.ReadCloser, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("error accessing file: %w", err)
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("%s is a directory, not a file", path)
	}

	pipeReader, pipeWriter := io.Pipe()

	tailCtx, cancelFunc := context.WithCancel(ctx)

	tailer := &FileTailer{
		path:       path,
		follow:     follow,
		reader:     pipeReader,
		writer:     pipeWriter,
		done:       make(chan struct{}),
		ctx:        tailCtx,
		cancelFunc: cancelFunc,
	}

	go tailer.tailLoop()

	return tailer, nil
}

func (t *FileTailer) tailLoop() {
	defer close(t.done)
	defer t.writer.Close()

	file, err := os.Open(t.path)
	if err != nil {
		t.writer.CloseWithError(fmt.Errorf("error opening file: %w", err))
		return
	}
	defer file.Close()

	buf := make([]byte, 4096)
	var offset int64 = 0
	keepAliveTicker := time.NewTicker(10 * time.Second)
	defer keepAliveTicker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-keepAliveTicker.C:
			if t.follow {
				_, writeErr := t.writer.Write([]byte{})
				if writeErr != nil {
					t.writer.CloseWithError(fmt.Errorf("error writing keepalive: %w", writeErr))
					return
				}
			}
		default:
			// Continue processing
		}

		_, err := file.Seek(offset, io.SeekStart)
		if err != nil {
			t.writer.CloseWithError(fmt.Errorf("error seeking in file: %w", err))
			return
		}

		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			t.writer.CloseWithError(fmt.Errorf("error reading file: %w", err))
			return
		}

		if n > 0 {
			written, err := t.writer.Write(buf[:n])
			if err != nil {
				t.writer.CloseWithError(fmt.Errorf("error writing to pipe: %w", err))
				return
			}
			offset += int64(written)
		}

		if err == io.EOF {
			if !t.follow {
				return
			} else {
				stat, statErr := os.Stat(t.path)
				if statErr != nil {
					if os.IsNotExist(statErr) {
						time.Sleep(100 * time.Millisecond)
						continue
					}

					t.writer.CloseWithError(fmt.Errorf("error checking file stats: %w", statErr))
					return
				}

				if stat.Size() < offset {
					offset = 0
					file.Close()

					file, err = os.Open(t.path)
					if err != nil {
						t.writer.CloseWithError(fmt.Errorf("error reopening file: %w", err))
						return
					}
				} else {
					time.Sleep(100 * time.Millisecond)
				}
				continue
			}
		} else if err != nil {
			t.writer.CloseWithError(fmt.Errorf("error reading file: %w", err))
			return
		}
	}
}

func (t *FileTailer) Read(p []byte) (n int, err error) {
	return t.reader.Read(p)
}

func (t *FileTailer) Close() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	t.cancelFunc()

	<-t.done

	return t.reader.Close()
}
