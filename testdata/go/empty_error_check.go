package testdata

import "fmt"

func badHandler() error {
	err := fmt.Errorf("something failed")
	if err != nil {
		return nil
	}
	return nil
}
