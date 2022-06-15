package util

import "time"

//Utility functions

var startTime time.Time

func init() {
	startTime = time.Now()
}

//Primitive to pointer functions
//Golang does not allow constant pointers, which makes structs and functions that only takes pointers look really stupid.

//Create int64 pointer from constant
func Int64Ptr(x int64) *int64 {
	return &x
}

//Create int32 pointer from constant
func Int32Ptr(x int32) *int32 {
	return &x
}

//Create bool pointer from constant
func BoolPtr(b bool) *bool {
	return &b
}

//Controller uptime
func Uptime() time.Duration {
	return time.Now().Sub(startTime)
}
