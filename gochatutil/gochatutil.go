// Simple util methods, such as error checking

package gochatutil

import (
	"log"
	"os"
)

type Err_code int

const (
	_            Err_code = iota
	ERR_CONTINUE Err_code = iota
	ERR_STOP     Err_code = iota
)

func CheckError(err error) bool {
	return CheckErrorAndAct(err, ERR_CONTINUE)
}

// CheckErrorAndAct checks if there has been an error
// and perform the action
func CheckErrorAndAct(err error, action Err_code) bool {
	if err != nil {
		log.Printf("Error %v\n", err)
		switch action {
		case ERR_STOP:
			os.Exit(-1)
		case ERR_CONTINUE:
			return true
		default:
			return true
		}
	}
	return false
}
