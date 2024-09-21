package utils

import (
	"fmt"
	"github.com/pkg/errors"
)

func RecoverGO(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("%+v", errors.Errorf("%+v", r))
			}
		}()
		f()
	}()
}
