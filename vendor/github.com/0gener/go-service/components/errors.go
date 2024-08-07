package components

import "fmt"

// ErrWrongComponent is a custom error type for type assertion failures
type ErrWrongComponent struct {
	ExpectedType string
}

func (e *ErrWrongComponent) Error() string {
	return fmt.Sprintf("component is not of the expected type: %s", e.ExpectedType)
}
