package benchmarks

import (
	"math/rand"
	"testing"
	"time"

	"code_breaker/internal/game"
)

func BenchmarkGenerateSmartHint(b *testing.B) {
	game.SetCodeDigits(4)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	secret := []int{1, 3, 5, 7}
	guess := []int{2, 4, 6, 8}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = game.GenerateSmartHint(secret, guess, rng)
	}
}

func BenchmarkGenerateFeedback(b *testing.B) {
	game.SetCodeDigits(4)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	secret := 1234
	guess := 1342
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = game.GenerateFeedback(secret, guess, rng)
	}
}

func BenchmarkGenerateSmartHint_8Digits(b *testing.B) {
	// 8-digit secret and guess
	secret := []int{1, 2, 3, 4, 5, 6, 7, 8}
	guess := []int{8, 7, 6, 5, 4, 3, 2, 1}

	// Deterministic RNG for benchmarking
	rng := rand.New(rand.NewSource(42))

	b.ResetTimer() // do not count setup time

	for i := 0; i < b.N; i++ {
		_ = game.GenerateSmartHint(secret, guess, rng)
	}
}

func BenchmarkGenerateSecret_8Digits(b *testing.B) {

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = game.GenerateSecretCodeWithDifficulty(8, game.DifficultyHard)
	}
}
