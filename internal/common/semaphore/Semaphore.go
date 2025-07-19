package semaphore

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Semaphore chan struct{}

// NewSemaphore creates a new Semaphore with the given maximum number of concurrent access.
func NewSemaphore(n int) Semaphore {
	return make(Semaphore, n)
}

// Acquire acquires the Semaphore, blocking until resources are available.
func (s Semaphore) Acquire() {
	s <- struct{}{}
}

// Release releases the Semaphore.
func (s Semaphore) Release() {
	<-s
}

// Wait waits until the Semaphore is fully released.
func (s Semaphore) Wait() {
	for i := 0; i < cap(s); i++ {
		s.Acquire()
	}
	for i := 0; i < cap(s); i++ {
		s.Release()
	}
}

// Close closes the Semaphore.
func (s Semaphore) Close() {
	close(s)
}

// WaitAndClose waits until the Semaphore is fully released and then closes it.
func (s Semaphore) WaitAndClose() {
	s.Wait()
	s.Close()
}

// String returns the string representation of the Semaphore.
func (s Semaphore) String() string {
	return fmt.Sprintf("Semaphore with cap %d", cap(s))
}

// MarshalJSON implements the json.Marshaler interface.
func (s Semaphore) MarshalJSON() ([]byte, error) {
	return json.Marshal(cap(s))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *Semaphore) UnmarshalJSON(data []byte) error {
	var n int
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	if n <= 0 {
		return errors.New("invalid capacity for Semaphore")
	}
	*s = NewSemaphore(n)
	return nil
}
