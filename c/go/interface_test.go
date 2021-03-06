package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/dropbox/rust-brotli/c/go/brotli"
)

var options = brotli.CompressionOptions{
	NumThreads: 1,
	Quality:    7,
	Catable:    true,
	Appendable: true,
	Magic:      true,
}

func TestCompressWriter(*testing.T) {
	data := testData()
	outBuffer := bytes.NewBuffer(nil)
	var options = brotli.CompressionOptions{
		NumThreads: 1,
		Quality:    4,
		Catable:    true,
		Appendable: true,
		Magic:      true,
	}
	writer := brotli.NewMultiCompressionWriter(
		outBuffer,
		options,
	)
	_, err := writer.Write(data[:])
	if err != nil {
		panic(err)
	}
	err = writer.Close()
	if err != nil {
		panic(err)
	}
	if len(outBuffer.Bytes()) == 0 {
		panic("Zero output buffer")
	}
	if len(outBuffer.Bytes()) > 800000 {
		panic(fmt.Sprintf("Buffer too large: %d", len(outBuffer.Bytes())))
	}
	version, size, err := brotli.BrotliParseHeader(outBuffer.Bytes())
	if err != nil {
		panic(err)
	}
	if version != byte(brotli.BrotliEncoderVersion()&0xff) {
		panic(version)
	}
	if size != uint64(len(data)) {
		panic(size)
	}
}

func TestCompressRoundtrip(*testing.T) {
	tmp := testData()
	data := tmp[:len(tmp)-17]
	outBuffer := bytes.NewBuffer(nil)
	var options = brotli.CompressionOptions{
		NumThreads: 1,
		Quality:    9,
		Catable:    true,
		Appendable: true,
		Magic:      true,
	}
	writer := brotli.NewMultiCompressionWriter(
		brotli.NewDecompressionWriter(
			outBuffer,
		),
		options,
	)
	_, err := writer.Write(data[:])
	if err != nil {
		panic(err)
	}
	err = writer.Close()
	if err != nil {
		panic(err)
	}
	if len(outBuffer.Bytes()) == 0 {
		panic("Zero output buffer")
	}
	if !bytes.Equal(outBuffer.Bytes(), data[:]) {
		panic(fmt.Sprintf("Bytes not equal %d, %d", len(outBuffer.Bytes()), len(data)))
	}
}

func TestCompressRoundtripMulti(*testing.T) {
	tmp := testData()
	data := tmp[:len(tmp)-17]
	outBuffer := bytes.NewBuffer(nil)
	var options = brotli.CompressionOptions{
		NumThreads: 16,
		Quality:    9,
		Catable:    true,
		Appendable: true,
		Magic:      true,
	}
	writer := brotli.NewMultiCompressionWriter(
		brotli.NewDecompressionWriter(
			outBuffer,
		),
		options,
	)
	_, err := writer.Write(data[:])
	if err != nil {
		panic(err)
	}
	err = writer.Close()
	if err != nil {
		panic(err)
	}
	if len(outBuffer.Bytes()) == 0 {
		panic("Zero output buffer")
	}
	if !bytes.Equal(outBuffer.Bytes(), data[:]) {
		panic(fmt.Sprintf("Bytes not equal %d, %d", len(outBuffer.Bytes()), len(data)))
	}
}

func TestRejectCorruptBuffers(*testing.T) {
	tmp := testData()
	data := tmp[:len(tmp)-17]
	outBuffer := bytes.NewBuffer(nil)
	compressedBuffer := bytes.NewBuffer(nil)
	var options = brotli.CompressionOptions{
		NumThreads: 1,
		Quality:    4,
		Catable:    true,
		Appendable: true,
		Magic:      true,
	}
	writer := brotli.NewMultiCompressionWriter(
		compressedBuffer,
		options,
	)
	_, err := writer.Write(data[:])
	if err != nil {
		panic(err)
	}
	err = writer.Close()
	if err != nil {
		panic(err)
	}
	decompressorWriter := brotli.NewDecompressionWriter(
		outBuffer,
	)
	// early EOF
	_, err = decompressorWriter.Write(compressedBuffer.Bytes()[:len(compressedBuffer.Bytes())-1])
	if err != nil {
		panic(err)
	}
	err = decompressorWriter.Close()
	if err == nil {
		panic("Expected error")
	}
	decompressorWriter = brotli.NewDecompressionWriter(
		outBuffer,
	)
	_, err = decompressorWriter.Write(compressedBuffer.Bytes()[:len(compressedBuffer.Bytes())/2])
	if err != nil {
		panic(err)
	}
	// missed a byte
	_, err = decompressorWriter.Write(compressedBuffer.Bytes()[len(compressedBuffer.Bytes())/2+1:])
	if err == nil {
		panic("ExpectedError")
	}
	_ = decompressorWriter.Close()
	corruptBuffer := bytes.NewBuffer(compressedBuffer.Bytes()[:len(compressedBuffer.Bytes())-1])
	decompressorReader := brotli.NewDecompressionReader(corruptBuffer)
	_, err = ioutil.ReadAll(decompressorReader)
	if err == nil {
		panic("ExpectedError")
	}
	decompressorReader = brotli.NewDecompressionReader(compressedBuffer)
	_, err = ioutil.ReadAll(decompressorReader)
	if err != nil {
		panic(err)
	}
}
func TestCompressRoundtripZero(*testing.T) {
	var data []byte
	outBuffer := bytes.NewBuffer(nil)
	var options = brotli.CompressionOptions{
		NumThreads: 1,
		Quality:    9,
		Catable:    true,
		Appendable: true,
		Magic:      true,
	}
	compressedForm := bytes.NewBuffer(nil)
	writer := brotli.NewMultiCompressionWriter(
		io.MultiWriter(compressedForm, brotli.NewDecompressionWriter(
			outBuffer,
		),
		),
		options,
	)
	err := writer.Close()
	if err != nil {
		panic(err)
	}
	if len(compressedForm.Bytes()) == 0 {
		panic("Zero output buffer")
	}
	if !bytes.Equal(outBuffer.Bytes(), data[:]) {
		panic(fmt.Sprintf("Bytes not equal %d, %d", len(outBuffer.Bytes()), len(data)))
	}
}

func TestCompressReader(*testing.T) {
	data := testData()
	inBuffer := bytes.NewBuffer(data[:])
	outBuffer := bytes.NewBuffer(nil)
	var options = brotli.CompressionOptions{
		NumThreads: 1,
		Quality:    4,
		Appendable: true,
		Magic:      true,
	}
	reader := brotli.NewMultiCompressionReader(
		inBuffer,
		options,
	)
	_, err := io.Copy(outBuffer, reader)
	if err != nil {
		panic(err)
	}
	if len(outBuffer.Bytes()) == 0 {
		panic("Zero output buffer")
	}
	if len(outBuffer.Bytes()) > 800000 {
		panic(fmt.Sprintf("Buffer too large: %d", len(outBuffer.Bytes())))
	}
	version, size, err := brotli.BrotliParseHeader(outBuffer.Bytes())
	if err != nil {
		panic(err)
	}
	if version != byte(brotli.BrotliEncoderVersion()&0xff) {
		panic(version)
	}
	if size != uint64(len(data)) {
		panic(size)
	}
}
func TestCompressReaderClose(*testing.T) {
	data := testData()
	inBuffer := bytes.NewBuffer(data[:])
	outBuffer := bytes.NewBuffer(nil)
	var options = brotli.CompressionOptions{
		NumThreads: 1,
		Quality:    2,
		Catable:    true,
		Appendable: true,
		Magic:      true,
	}
	reader := brotli.NewMultiCompressionReader(
		inBuffer,
		options,
	)
	_, err := io.Copy(outBuffer, reader)
	if err != nil {
		panic(err)
	}
	if len(outBuffer.Bytes()) == 0 {
		panic("Zero output buffer")
	}
	if len(outBuffer.Bytes()) > 1850280 {
		panic(fmt.Sprintf("Buffer too large: %d", len(outBuffer.Bytes())))
	}
	err = reader.Close()
	if err != nil {
		panic(err)
	}
	version, size, err := brotli.BrotliParseHeader(outBuffer.Bytes())
	if err != nil {
		panic(err)
	}
	if version != byte(brotli.BrotliEncoderVersion()&0xff) {
		panic(version)
	}
	if size != uint64(len(data)) {
		panic(size)
	}
}

func TestCompressReaderEarlyClose(*testing.T) {
	data := testData()
	inBuffer := bytes.NewBuffer(data[:])
	var options = brotli.CompressionOptions{
		NumThreads: 1,
		Quality:    2,
		Catable:    true,
		Appendable: true,
		Magic:      true,
	}
	reader := brotli.NewMultiCompressionReader(
		inBuffer,
		options,
	)
	var smallBuf [1024]byte
	count, err := reader.Read(smallBuf[:])
	if err != nil {
		panic(err)
	}
	if count != len(smallBuf) {
		panic("Underflow for test data: too few bytes of test data")
	}
	err = reader.Close()
	if err != nil {
		panic(err)
	}
}

func TestCompressReaderRoundtrip(*testing.T) {
	data := testData()
	inBuffer := bytes.NewBuffer(data[:])
	outBuffer := bytes.NewBuffer(nil)
	var options = brotli.CompressionOptions{
		NumThreads: 1,
		Quality:    4,
		Catable:    true,
		Appendable: true,
		Magic:      true,
	}
	reader := brotli.NewDecompressionReader(
		brotli.NewMultiCompressionReader(
			inBuffer,
			options,
		),
	)
	_, err := io.Copy(outBuffer, reader)
	if err != nil {
		panic(err)
	}
	if len(outBuffer.Bytes()) == 0 {
		panic("Zero output buffer")
	}
	if !bytes.Equal(outBuffer.Bytes(), data[:]) {
		panic(fmt.Sprintf("Bytes not equal %d, %d", len(outBuffer.Bytes()), len(data)))
	}
}

func TestDecompressReaderEarlyClose(*testing.T) {
	data := testData()
	inBuffer := bytes.NewBuffer(data[:])
	var options = brotli.CompressionOptions{
		NumThreads: 1,
		Quality:    4,
		Catable:    true,
		Appendable: true,
		Magic:      true,
	}
	reader := brotli.NewDecompressionReader(
		brotli.NewMultiCompressionReader(
			inBuffer,
			options,
		),
	)
	var smallBuffer [1027]byte
	count, err := reader.Read(smallBuffer[:])
	if err != nil {
		panic(err)
	}
	if count < 1024 {
		panic("Too small a test buffer")
	}
	err = reader.Close()
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(smallBuffer[:], data[:len(smallBuffer)]) {
		panic(fmt.Sprintf("Bytes not equal %x, %x", smallBuffer[:], data[:len(smallBuffer)]))
	}
}

func TestCompressReaderRoundtripZero(*testing.T) {
	var data []byte
	inBuffer := bytes.NewBuffer(data[:])
	outBuffer := bytes.NewBuffer(nil)
	var options = brotli.CompressionOptions{
		NumThreads: 1,
		Quality:    11,
		Catable:    true,
		Appendable: true,
		Magic:      true,
	}
	compressedForm := bytes.NewBuffer(nil)
	reader := brotli.NewDecompressionReader(
		io.TeeReader(
			brotli.NewMultiCompressionReader(
				inBuffer,
				options,
			),
			compressedForm),
	)
	_, err := io.Copy(outBuffer, reader)
	if err != nil {
		panic(err)
	}
	if len(compressedForm.Bytes()) == 0 {
		panic("Zero output buffer")
	}
	if !bytes.Equal(outBuffer.Bytes(), data[:]) {
		panic(fmt.Sprintf("Bytes not equal %d, %d", len(outBuffer.Bytes()), len(data)))
	}
}

func TestConcatFlatFunction(*testing.T) {
	data := testData()
	inBufferAa := bytes.NewBuffer(data[:len(data)/5])
	inBufferBa := bytes.NewBuffer(data[len(data)/5 : 2*(len(data)/5)])
	inBufferCa := bytes.NewBuffer(data[2*(len(data)/5) : 3*(len(data)/5)])
	inBufferDa := bytes.NewBuffer(data[3*(len(data)/5):])
	midBufferA := bytes.NewBuffer(nil)
	var err error
	_, err = io.Copy(midBufferA, brotli.NewMultiCompressionReader(
		inBufferAa,
		options,
	))
	if err != nil {
		panic(err)
	}
	midBufferB := bytes.NewBuffer(nil)
	_, err = io.Copy(midBufferB, brotli.NewMultiCompressionReader(
		inBufferBa,
		options,
	))
	if err != nil {
		panic(err)
	}
	midBufferC := bytes.NewBuffer(nil)
	_, err = io.Copy(midBufferC, brotli.NewMultiCompressionReader(
		inBufferCa,
		options,
	))
	if err != nil {
		panic(err)
	}
	midBufferD := bytes.NewBuffer(nil)
	_, err = io.Copy(midBufferD, brotli.NewMultiCompressionReader(
		inBufferDa,
		options,
	))
	if err != nil {
		panic(err)
	}
	final, err := brotli.BroccoliConcat([][]byte{midBufferA.Bytes(), midBufferB.Bytes(), midBufferC.Bytes(), midBufferD.Bytes()}...)
	if err != nil {
		panic(err)
	}
	finalBuffer := bytes.NewBuffer(final)
	rtBuffer := bytes.NewBuffer(nil)
	_, err = io.Copy(rtBuffer, brotli.NewDecompressionReader(finalBuffer))
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(rtBuffer.Bytes(), data[:]) {
		panic(fmt.Sprintf("Bytes not equal %d, %d", len(rtBuffer.Bytes()), len(data)))
	}
}

func TestConcatReaderRoundtrip(*testing.T) {
	data := testData()
	inBufferA := bytes.NewBuffer(data[:len(data)/5-1])
	inBufferB := bytes.NewBuffer(data[len(data)/5-1 : 2+2*(len(data)/5)])
	inBufferC := bytes.NewBuffer(data[2+2*(len(data)/5) : 3*(len(data)/5)])
	inBufferD := bytes.NewBuffer(data[3*(len(data)/5):])
	outBuffer := bytes.NewBuffer(nil)
	var options = brotli.CompressionOptions{
		NumThreads: 1,
		Quality:    4,
		Catable:    true,
		Appendable: true,
		Magic:      true,
	}

	reader := brotli.NewDecompressionReader(
		brotli.NewBroccoliConcatReader(
			brotli.NewMultiCompressionReader(
				inBufferA,
				options,
			),
			brotli.NewMultiCompressionReader(
				inBufferB,
				options,
			),
			brotli.NewMultiCompressionReader(
				inBufferC,
				options,
			),
			brotli.NewMultiCompressionReader(
				inBufferD,
				options,
			),
		))
	_, err := io.Copy(outBuffer, reader)
	if err != nil {
		panic(err)
	}
	if len(outBuffer.Bytes()) == 0 {
		panic("Zero output buffer")
	}
	if !bytes.Equal(outBuffer.Bytes(), data[:]) {
		panic(fmt.Sprintf("Bytes not equal %d, %d", len(outBuffer.Bytes()), len(data)))
	}
}

func TestVersions(*testing.T) {
	if brotli.BrotliEncoderVersion() == 0 {
		panic(fmt.Sprintf("Bad version %d\n", brotli.BrotliEncoderVersion()))
	}
	if brotli.BrotliDecoderVersion() == 0 {
		panic(fmt.Sprintf("Bad version %d\n", brotli.BrotliDecoderVersion()))
	}
}
