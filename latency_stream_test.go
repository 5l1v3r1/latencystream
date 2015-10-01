package latencystream

import (
	"bytes"
	"testing"
	"time"
)

const ShortTimeout = time.Millisecond * 10
const LongTimeout = time.Millisecond * 50

func TestNewLatencyStreamInfLatency(t *testing.T) {
	unbuffered := make(chan []byte)
	closeChan := make(chan struct{})
	chunkStream := ChunkStream{unbuffered, closeChan}
	latencyStream := NewLatencyStream(chunkStream, 10, time.Hour)

	unbuffered <- []byte("hey")
	unbuffered <- []byte("testing")
	valueOut := <-latencyStream.Chunks
	if !bytes.Equal(valueOut, []byte("heytesting")) {
		t.Error("bad chunk")
	}

	unbuffered <- []byte("hey")
	unbuffered <- []byte("testing123")
	valueOut = <-latencyStream.Chunks
	if !bytes.Equal(valueOut, []byte("heytesting")) {
		t.Error("bad chunk")
	}

	unbuffered <- []byte("testing")
	valueOut = <-latencyStream.Chunks
	if !bytes.Equal(valueOut, []byte("123testing")) {
		t.Error("bad chunk")
	}

	close(latencyStream.Close)
	time.Sleep(time.Millisecond * 100)
	select {
	case <-closeChan:
	default:
		t.Error("closing did not result in more closes")
	}
}

func TestNewLatencyStream(t *testing.T) {
	unbuffered := make(chan []byte, 10)
	closeChan := make(chan struct{})
	chunkStream := ChunkStream{unbuffered, closeChan}
	latencyStream := NewLatencyStream(chunkStream, 10, LongTimeout)

	unbuffered <- []byte("testing")
	time.Sleep(ShortTimeout)
	select {
	case <-latencyStream.Chunks:
		t.Error("did not buffer")
	default:
	}
	time.Sleep(LongTimeout * 2)
	select {
	case valueOut := <-latencyStream.Chunks:
		if !bytes.Equal(valueOut, []byte("testing")) {
			t.Error("bad chunk")
		}
	default:
		t.Error("did not flush")
	}

	unbuffered <- []byte("testing123456")
	time.Sleep(ShortTimeout)
	select {
	case buff := <-latencyStream.Chunks:
		if !bytes.Equal(buff, []byte("testing123")) {
			t.Error("invalid buffer")
		}
	default:
		t.Error("did not flush")
	}

	select {
	case <-latencyStream.Chunks:
		t.Error("did not buffer")
	default:
	}

	time.Sleep(LongTimeout * 2)
	select {
	case valueOut := <-latencyStream.Chunks:
		if !bytes.Equal(valueOut, []byte("456")) {
			t.Error("bad chunk")
		}
	default:
		t.Error("did not flush")
	}

	unbuffered <- []byte("hey")
	time.Sleep(LongTimeout * 2)
	unbuffered <- []byte("there")
	time.Sleep(LongTimeout * 2)
	select {
	case valueOut := <-latencyStream.Chunks:
		if !bytes.Equal(valueOut, []byte("heythere")) {
			t.Error("bad chunk")
		}
	default:
		t.Error("did not flush")
	}
}
