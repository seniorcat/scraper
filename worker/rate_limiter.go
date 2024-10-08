package worker

import (
	"time"
)

// RateLimiter отвечает за ограничение скорости запросов
type RateLimiter struct {
	Interval time.Duration // Интервал между запросами
	TokenCh  chan struct{}
}

// NewRateLimiter создает новый лимитер с заданным интервалом
func NewRateLimiter(rps int) *RateLimiter {
	rl := &RateLimiter{
		Interval: time.Second / time.Duration(rps), // Интервал между токенами
		TokenCh:  make(chan struct{}, rps),         // Канал для токенов
	}

	// Запускаем горутину для периодической выдачи токенов
	go rl.generateTokens()
	return rl
}

// generateTokens добавляет токены в канал с заданным интервалом
func (rl *RateLimiter) generateTokens() {
	ticker := time.NewTicker(rl.Interval)
	defer ticker.Stop()
	for range ticker.C {
		select {
		case rl.TokenCh <- struct{}{}:
		default:
			// Если канал переполнен, пропускаем токен
		}
	}
}

// TakeToken запрашивает токен из лимитера, блокируя выполнение до его получения
func (rl *RateLimiter) TakeToken() {
	<-rl.TokenCh
}
