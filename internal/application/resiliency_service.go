package application

import (
	"fmt"
	"math/rand"
	"time"
)

type Resiliencyservice struct{}

func (r *Resiliencyservice) GenerateResiliency(minDelaySec int32, maxDelaySec int32, statusCodes []uint32) (string, uint32) {
	delay := rand.Intn(int(maxDelaySec-minDelaySec)) + int(minDelaySec)
	delaySecond := time.Duration(delay) * time.Second
	time.Sleep(delaySecond)

	idx := rand.Intn(len(statusCodes))
	str := fmt.Sprintf("The time now is %v, Delay: %d seconds, Status code: %d", time.Now().Format("15:04:05.0000"), delay, statusCodes[idx])

	return str, statusCodes[idx]
}
