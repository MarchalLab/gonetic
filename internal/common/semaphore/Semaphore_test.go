package semaphore

import (
	"encoding/json"
	"testing"
)

// TestNewSemaphore tests the creation of a new Semaphore
func TestNewSemaphore(t *testing.T) {
	sem := NewSemaphore(3)
	if cap(sem) != 3 {
		t.Errorf("Expected semaphore capacity to be 3, got %d", cap(sem))
	}
}

// TestAcquireRelease tests acquiring and releasing the Semaphore
func TestAcquireRelease(t *testing.T) {
	sem := NewSemaphore(1)
	sem.Acquire()
	select {
	case sem <- struct{}{}:
		t.Error("Expected semaphore to be acquired, but it was not")
	default:
	}

	sem.Release()
	select {
	case sem <- struct{}{}:
	default:
		t.Error("Expected semaphore to be released, but it was not")
	}
}

// TestWait tests the Wait method of the Semaphore
func TestWait(t *testing.T) {
	sem := NewSemaphore(2)
	sem.Acquire()
	sem.Acquire()

	go func() {
		sem.Release()
		sem.Release()
	}()

	sem.Wait()
	select {
	case sem <- struct{}{}:
	default:
		t.Error("Expected semaphore to be fully released, but it was not")
	}
}

// TestMarshalJSON tests the JSON marshaling of the Semaphore
func TestMarshalJSON(t *testing.T) {
	sem := NewSemaphore(5)
	data, err := json.Marshal(sem)
	if err != nil {
		t.Errorf("Error marshaling semaphore: %v", err)
	}

	expected := `5`
	if string(data) != expected {
		t.Errorf("Expected JSON %s, got %s", expected, string(data))
	}
}

// TestUnmarshalJSON tests the JSON unmarshaling of the Semaphore
func TestUnmarshalJSON(t *testing.T) {
	data := `5`
	var sem Semaphore
	if err := json.Unmarshal([]byte(data), &sem); err != nil {
		t.Errorf("Error unmarshaling semaphore: %v", err)
	}

	if cap(sem) != 5 {
		t.Errorf("Expected semaphore capacity to be 5, got %d", cap(sem))
	}
}

func TestSemaphore_WaitAndClose(t *testing.T) {
	sem := NewSemaphore(2)
	// Acquire both resources
	sem.Acquire()
	sem.Acquire()
	// Release both resources asynchronously
	go func() { sem.Release() }()
	go func() { sem.Release() }()
	// Wait for both resources to be released and then close the semaphore
	sem.WaitAndClose()
	// Try to acquire a resource
	select {
	case _, ok := <-sem:
		if ok {
			t.Error("Expected semaphore to be closed, but it was not")
		}
	default:
	}
}
