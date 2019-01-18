package helper

import (
	"fmt"
	"log"
	"time"
)

func MeasureTitle(s string, args ...interface{}) (string, time.Time) {
	title := fmt.Sprintf(s, args...)

	log.Println("Start:	", title)
	return title, time.Now()
}

func MeasureTime(s string, startTime time.Time) {
	endTime := time.Now()
	log.Println("End:	", s, "took", endTime.Sub(startTime).Nanoseconds())
}

//useage
//////////////

//func a() {
//	defer MeasureTime(MeasureTitle("testing"))
//	doSomething()
//}
