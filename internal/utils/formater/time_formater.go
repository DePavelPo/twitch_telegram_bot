package formater

import (
	"fmt"
	"time"
)

func CreateStreamDuration(streamTime time.Time) string {

	location := time.FixedZone("MSK", 3*60*60)
	streamStartTime := streamTime.In(location)

	streamDuration := time.Now().Sub(streamStartTime)
	hours := streamDuration / time.Hour
	streamDuration = streamDuration % time.Hour
	minutes := streamDuration / time.Minute
	streamDuration = streamDuration % time.Minute
	seconds := streamDuration / time.Second
	streamDurationStr := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)

	return streamDurationStr
}
