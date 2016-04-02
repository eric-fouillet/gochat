/*
Simple util methods, such as error checking
*/

package gochatutil

import (
	"log"
	"os"
)

type ErrCode int

const (
	_            ErrCode = iota
	ERR_CONTINUE ErrCode = iota
	ERR_STOP     ErrCode = iota
)

func CheckError(err error) bool {
	return CheckErrorAndAct(err, ERR_CONTINUE)
}

// CheckErrorAndAct checks if there has been an error
// and perform the action
func CheckErrorAndAct(err error, action ErrCode) bool {
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
