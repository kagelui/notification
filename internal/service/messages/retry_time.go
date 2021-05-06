package messages

import "time"

var retryIntervalInMinute0 = 15
var retryIntervalInMinute1 = 45
var retryIntervalInMinute2 = 120
var retryIntervalInMinute3 = 180
var retryIntervalInMinute4 = 360
var retryIntervalInMinute5 = 720

func getRetryTime(curr time.Time, retryCount int) time.Time {
	var increment time.Duration
	switch retryCount {
	case 0:
		increment = time.Minute * time.Duration(retryIntervalInMinute0)
	case 1:
		increment = time.Minute * time.Duration(retryIntervalInMinute1)
	case 2:
		increment = time.Minute * time.Duration(retryIntervalInMinute2)
	case 3:
		increment = time.Minute * time.Duration(retryIntervalInMinute3)
	case 4:
		increment = time.Minute * time.Duration(retryIntervalInMinute4)
	case 5:
		increment = time.Minute * time.Duration(retryIntervalInMinute5)
	default:
		increment = 0
	}
	return curr.Add(increment)
}
