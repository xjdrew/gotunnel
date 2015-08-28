package tunnel

import "testing"

var input string = "hello, world"

func produce(buffer *LinkBuffer) {
	for i := 0; i < len(input); i++ {
		buffer.Put([]byte(input[i : i+1]))
	}
}

func consume(buffer *LinkBuffer) bool {
	var output string
	for {
		data, ok := buffer.Pop()
		if !ok {
			break
		}
		output += string(data)
		if len(output) == len(input) {
			break
		}
	}
	if input != output {
		return false
	}
	return true
}

func TestBuffer(t *testing.T) {
	buffer := NewLinkBuffer(1)

	produce(buffer)
	if !consume(buffer) {
		t.FailNow()
	}
	produce(buffer)
	if !consume(buffer) {
		t.FailNow()
	}
}
