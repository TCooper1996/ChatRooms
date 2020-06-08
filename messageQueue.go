package main

import (
	"errors"
	"fmt"
	"strings"
)

//We must keep a limit of how many bytes are being stored as messages in order to avoid unbounded memory issues.
//When the limit is reached, the oldest messages are dropped.
//The queue implemented as a linked list, though a circular linked list may or may not be a better option.
const maxByteSize = 1024 * 20

//Set a max size for each message
const maxMessageSize = 1024

//Suppose that when the queue exceed its maxByteSize, it removes just enough of the older messages to stay under
//The maximum. If this were the case, then nearly every new message may cause it to exceed and require another truncation.
//As a result, the following constant indicates by what factor of the maxByteSize the queue will be reduced.
const truncationFactor = 0.5

type node struct {
	message  string
	next     *node
	byteSize int
}

type messageQueue struct {
	head     *node
	tail     *node
	byteSize int
	count    int
}

func (q messageQueue) Push(m string) error {
	m = strings.TrimRight(m, "\n") + string('\n')
	byteCount := len([]byte(m))
	if byteCount > maxMessageSize {
		return errors.New(fmt.Sprintf("Message size (%d) exceeded,", maxMessageSize))
	}

	q.byteSize += byteCount

	//If memory is exceeded, pop off older messages
	if q.byteSize > maxByteSize {
		bytesToRemove := int(maxByteSize * truncationFactor)
		//Iterate over each node keeping count of how many bytes have been written.
		//When the loop condition ends, the value for n is the new head.
		n := q.head
		for ; bytesToRemove < 0; n = n.next {
			bytesToRemove -= n.byteSize
		}
		q.head = n
	}

	n := &node{message: m, byteSize: byteCount}
	if q.head == nil {
		q.head = n
		q.tail = n
	} else {
		q.tail.next = n
		q.tail = n
	}
	q.count++

	return nil
}

//Range returns the contents of the queue
func (q messageQueue) Range() []string {
	arr := make([]string, q.count)
	n := q.head
	for i := 0; i < q.count; i++ {
		arr[i] = n.message
		n = n.next
	}
	return arr
}
