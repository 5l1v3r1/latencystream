package latencystream

import (
	"bytes"
	"io"
	"math/rand"
	"testing"
	"time"
)

type testReader struct {
	buffer   []byte
	incoming chan []byte
}

func newTestReader() *testReader {
	return &testReader{incoming: make(chan []byte, 10)}
}

func (t *testReader) Read(output []byte) (count int, err error) {
	if len(t.buffer) > 0 {
		copy(output, t.buffer)
		usedLen := len(t.buffer)
		if usedLen > len(output) {
			copy(t.buffer, t.buffer[:len(output)])
			t.buffer = t.buffer[:len(t.buffer)-len(output)]
			return len(output), nil
		} else {
			l := len(t.buffer)
			t.buffer = t.buffer[:0]
			return l, nil
		}
	} else {
		incoming := <-t.incoming
		if incoming == nil {
			return 0, io.EOF
		}
		t.buffer = append(t.buffer, incoming...)
		return t.Read(output)
	}
}

func TestNewChunkStreamReader(t *testing.T) {
	reader := newTestReader()
	chunkStream := NewChunkStreamReader(reader)
	testChunk := generateTestData(ReaderMaxChunkSize)
	reader.incoming <- testChunk
	readChunk := <-chunkStream.Chunks
	if !bytes.Equal(readChunk, testChunk) {
		t.Error("invalid result chunk for input")
	}
	reader.incoming <- testChunk[:10]
	reader.incoming <- testChunk[10:20]

	readChunk = <-chunkStream.Chunks
	if !bytes.Equal(readChunk, testChunk[:10]) {
		t.Error("invalid result chunk for input")
	}

	readChunk = <-chunkStream.Chunks
	if !bytes.Equal(readChunk, testChunk[10:20]) {
		t.Error("invalid result chunk for input")
	}

	close(chunkStream.Close)
	reader.incoming <- testChunk[:10]
	reader.incoming <- testChunk[:10]
	time.Sleep(time.Millisecond * 100)
	select {
	case res := <-chunkStream.Chunks:
		if res != nil {
			t.Error("should not have gotten anything")
		}
	default:
		t.Error("chunk stream not closed")
	}

	reader = newTestReader()
	chunkStream = NewChunkStreamReader(reader)
	reader.incoming <- testChunk
	close(reader.incoming)
	readChunk = <-chunkStream.Chunks
	if !bytes.Equal(readChunk, testChunk) {
		t.Error("invalid result chunk for input")
	}
	if <-chunkStream.Chunks != nil {
		t.Error("close not captured correctly")
	}
}

func generateTestData(length int) []byte {
	res := make([]byte, length)
	for i := range res {
		res[i] = byte(rand.Intn(256))
	}
	return res
}

