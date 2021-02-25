package librsync

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"
)

func signatureFromRand(b *testing.B, src io.Reader) *SignatureType {
	var (
		magic            = BLAKE2_SIG_MAGIC
		blockLen  uint32 = 512
		strongLen uint32 = 32
		bufSize          = 65536
	)

	s, err := Signature(
		bufio.NewReaderSize(src, bufSize),
		ioutil.Discard,
		blockLen, strongLen, magic)
	if err != nil {
		b.Error(err)
	}

	return s
}

func BenchmarkSignature(b *testing.B) {
	var totalBytes int64 = 1000000000 // 1 GB
	src := io.LimitReader(rand.New(rand.NewSource(time.Now().UnixNano())), totalBytes)

	b.SetBytes(totalBytes)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		signatureFromRand(b, src)
	}
}

func BenchmarkDelta(b *testing.B) {
	var totalBytes int64 = 1000000000 // 1 GB

	var srcBuf bytes.Buffer
	src := io.TeeReader(
		io.LimitReader(rand.New(rand.NewSource(time.Now().UnixNano())), totalBytes),
		&srcBuf)
	s := signatureFromRand(b, src)

	b.SetBytes(totalBytes)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer

		newBytes := totalBytes / 10
		srcBuf.Truncate(int(totalBytes - newBytes))
		_, err := io.CopyN(&srcBuf, rand.New(rand.NewSource(time.Now().UnixNano())), newBytes)
		if err != nil {
			b.Error(err)
		}

		if err := Delta(s, &srcBuf, &buf); err != nil {
			b.Error(err)
		}

		b.Logf("target size:    %v bytes", totalBytes)
		b.Logf("delta size:     %v bytes", len(buf.Bytes()))
		b.Logf("==> raw diff:   %.2f %%", (float64(newBytes)/float64(totalBytes))*100)
		b.Logf("==> delta diff: %.2f %%", (float64(len(buf.Bytes()))/float64(totalBytes))*100)
	}
}
