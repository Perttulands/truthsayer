package testdata

import "fmt"

func goodHandler() error {
	err := fmt.Errorf("something failed")
	if err != nil {
		return fmt.Errorf("handler failed: %w", err)
	}
	return nil
}
