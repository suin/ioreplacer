package ioreplacer

import (
	"bytes"
	"io"
	"sort"
)

type Reader struct {
	BufferSize                 int
	source                     io.Reader
	replaceMap                 []pair
	unreadBytes, replacedBytes []byte
	longestNeedleLength        int
	err                        error
}

func NewReader(source io.Reader, replaceMap map[string]string) *Reader {
	newReplaceMap := [][][]byte{}

	for from, to := range replaceMap {
		newReplaceMap = append(newReplaceMap, [][]byte{[]byte(from), []byte(to)})
	}

	return NewBytesReader(source, newReplaceMap)
}

func NewBytesReader(source io.Reader, replaceMap [][][]byte) (this *Reader) {
	this = new(Reader)
	this.source = source
	this.BufferSize = 1024 * 32 // 32KB

	for _, fromTo := range replaceMap {
		from := fromTo[0]
		to := fromTo[1]
		this.replaceMap = append(this.replaceMap, pair{from, to})
	}

	sort.Sort(byNeedleLength(this.replaceMap))

	if len(this.replaceMap) > 0 {
		this.longestNeedleLength = len(this.replaceMap[0].from)
	} else {
		this.longestNeedleLength = 1 // this must be at least one, because it is used for range
	}

	return
}

func (this *Reader) Read(payload []byte) (readBytes int, err error) {
	for i := 0; i < len(payload); i ++ {
		aByte, end := this.readByte()

		if end == true {
			break
		}

		if len(aByte) > 0 {
			payload[i] = aByte[0]
			readBytes += 1
		} else {
			i -= 1 // TODO: refactor this
		}
	}

	if readBytes > 0 && io.EOF == this.err {
		err = nil
	} else {
		err = this.err
	}

	return
}

// self encapsulation
func (this *Reader) setUnreadBytes(unreadBytes []byte) {
	this.unreadBytes = unreadBytes
}

// self encapsulation
func (this *Reader) setReplacedBytes(replacedBytes []byte) {
	this.replacedBytes = replacedBytes
}

func (this *Reader) readByte() (b []byte, end bool) {
	// when there are replaced bytes are stacked, pop it
	if len(this.replacedBytes) > 0 {
		b = []byte{this.replacedBytes[0]}
		this.setReplacedBytes(this.replacedBytes[1:])
		return
	}

	// load data when unread bytes pool is empty or has been consumed all
	if len(this.unreadBytes) == 0 {
		readBytes := this.fillUnreadBytes()
		if readBytes == 0 {
			// no more data
			return b, true
		}
	}

	thisBytes := make([]byte, this.longestNeedleLength)
	copy(thisBytes, this.unreadBytes)

	for _, pair := range this.replaceMap {
		if bytes.Index(thisBytes, pair.from) == 0 {
			// a pair matched!
			if len(pair.to) > 0 {
				b = []byte{pair.to[0]}
				this.setReplacedBytes(append(this.replacedBytes, pair.to[1:]...)) // enqueue bytes
			}

			this.setUnreadBytes(this.unreadBytes[len(pair.from):]) // consume one byte or some bytes
			return
		}
	}

	// no pair not matched
	b = []byte{thisBytes[0]}
	this.setUnreadBytes(this.unreadBytes[1:]) // consume one byte

	return
}

func (this *Reader) fillUnreadBytes() (readBytes int) {
	var bufferSize int

	if this.longestNeedleLength > this.BufferSize {
		bufferSize = this.longestNeedleLength
	} else {
		bufferSize = this.BufferSize
	}

	payload := make([]byte, bufferSize)

	readBytes, err := this.source.Read(payload)

	if err != nil {
		this.err = err
	}

	for i := 0; i < readBytes; i++ {
		this.setUnreadBytes(append(this.unreadBytes, payload[i]))
	}

	return
}

// for sort replace maps
type pair struct {
	from []byte
	to   []byte
}

type byNeedleLength []pair

func (a byNeedleLength) Len() int           { return len(a) }
func (a byNeedleLength) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byNeedleLength) Less(i, j int) bool { return len(a[i].from) > len(a[j].from) }
