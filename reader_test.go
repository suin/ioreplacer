package ioreplacer

import (
	"fmt"
	. "github.com/r7kamura/gospel"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func TestDescribe(t *testing.T) {
	Describe(t, "Test for Reader", func() {
		Context("Replace 'abc' with string map", func() {
			inner := strings.NewReader("abc")
			reader := NewReader(inner, map[string]string{"a": "b", "b": "c", "c": "a"})
			bytes, err := ioutil.ReadAll(reader)

			It("should be 'bca'", func() {
				Expect(string(bytes)).To(Equal, "bca")
				Expect(err).To(Equal, nil)
			})
		})

		Context("Replace 'pineapple is a fruit' with string map", func() {
			inner := strings.NewReader("pineapple is a fruit")
			reader := NewReader(inner, map[string]string{"apple": "carrot", "pineapple": "orange", "pine" : "tomato"})
			bytes, err := ioutil.ReadAll(reader)

			It("should be 'orange is a fruit' (long word should be replaced first)", func() {
				Expect(string(bytes)).To(Equal, "orange is a fruit")
				Expect(err).To(Equal, nil)
			})
		})

		Context("Test in poor buffer size", func() {
			inner := strings.NewReader("pineapple is a fruit")
			reader := NewReader(inner, map[string]string{"apple": "carrot", "pineapple": "orange"})
			reader.BufferSize = 1
			bytes, err := ioutil.ReadAll(reader)

			It("should be 'orange is a fruit'", func() {
				Expect(string(bytes)).To(Equal, "orange is a fruit")
				Expect(err).To(Equal, nil)
			})
		})

		Context("Test byte size", func() {
			inner := strings.NewReader("abc")
			reader := NewReader(inner, map[string]string{"a": "b", "b": "c", "c": "a"})

			payload := make([]byte, 10)
			readBytes, err := reader.Read(payload)

			It("should return valid payload", func() {
				Expect(fmt.Sprintf("%v", payload)).To(Equal, "[98 99 97 0 0 0 0 0 0 0]")
				Expect(readBytes).To(Equal, 3)
				Expect(err).To(Equal, nil)
			})
		})

		Context("Test EOF", func() {
			inner := strings.NewReader("abc")
			reader := NewReader(inner, map[string]string{"a": "b"})

			payload := make([]byte, 10)
			readBytes, err := reader.Read(payload)
			readBytes, err = reader.Read(payload)

			It("should return valid payload", func() {
				Expect(readBytes).To(Equal, 0)
				Expect(err).To(Equal, io.EOF)
			})
		})

		Context("Replace long word shorter", func() {
			inner := strings.NewReader("here is a loooooooong word")
			reader := NewReader(inner, map[string]string{"loooooooong": "short"})

			payload := make([]byte, 20)
			readBytes, _ := reader.Read(payload)

			It("should be shorter", func() {
				Expect(readBytes).To(Equal, len(payload))
				Expect(string(payload)).To(Equal, "here is a short word")
			})
		})

		Context("Replace short word longer", func() {
			inner := strings.NewReader("here is a short word")
			reader := NewReader(inner, map[string]string{"short": "looooooong"})

			payload := make([]byte, 25)
			readBytes, _ := reader.Read(payload)

			It("should be longer", func() {
				Expect(readBytes).To(Equal, len(payload))
				Expect(string(payload)).To(Equal, "here is a looooooong word")
			})
		})

		Context("Replace 'strawberry' to blank string on 'apple,orange,strawberry,melon'", func() {
			input := strings.NewReader("apple,orange,strawberry,melon")
			reader := NewReader(input, map[string]string{"strawberry": ""})
			payload := make([]byte, 19)
			readBytes, _ := reader.Read(payload)

			It("should be 'apple,orange,,melon'", func() {
				Expect(len(payload)).To(Equal, readBytes)
				Expect("apple,orange,,melon").To(Equal, string(payload))
			})
		})

		Context("Test empty replace map", func() {
			input := strings.NewReader("apple,orange,strawberry,melon")
			reader := NewReader(input, map[string]string{})
			readBytes, _ := ioutil.ReadAll(reader)

			It("should not be changed", func() {
				Expect("apple,orange,strawberry,melon").To(Equal, string(readBytes))
			})
		})
	})
}
