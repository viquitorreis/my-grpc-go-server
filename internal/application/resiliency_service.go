package application

import (
	"fmt"
	"math/rand/v2"
	"time"
)

type ResiliencyService struct {
}

func (r *ResiliencyService) GenerateResiliency(minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32) (string, uint32) {
	// gerando delay aleatório
	delay := rand.IntN(int(maxDelaySecond-minDelaySecond)) + int(minDelaySecond)
	delaySecond := time.Duration(delay) * time.Second
	time.Sleep(delaySecond)

	// gerando index aleatório a partir do tamanho do array de statusCodes
	idx := rand.IntN(len(statusCodes))
	str := fmt.Sprintf("Tempo agora é %s, delay foi de %d segundos e o status code é %d", time.Now().Format("15:04:05.000"), delay, statusCodes[idx])

	return str, statusCodes[idx]
}
