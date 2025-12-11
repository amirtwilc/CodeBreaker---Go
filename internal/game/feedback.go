package game

import (
	"math/rand"
)

// Feedback holds result of evaluating a guess.
type Feedback struct {
	CorrectPlace int
	WrongPlace   int
	Hint         string
}

// GenerateFeedback compares secret vs guess and returns counts and a hint.
// RNG is injected to allow deterministic tests.
func GenerateFeedback(secret, guess int, rng *rand.Rand) Feedback {
	secretDigits := splitToDigits(secret)
	guessDigits := splitToDigits(guess)

	usedSecret := make([]bool, codeDigits)
	usedGuess := make([]bool, codeDigits)

	correctPlace := 0
	wrongPlace := 0

	// exact matches
	for i := 0; i < codeDigits; i++ {
		if secretDigits[i] == guessDigits[i] {
			correctPlace++
			usedSecret[i] = true
			usedGuess[i] = true
		}
	}

	// misplaced matches
	for i := 0; i < codeDigits; i++ {
		if usedGuess[i] {
			continue
		}
		for j := 0; j < codeDigits; j++ {
			if usedSecret[j] {
				continue
			}
			if guessDigits[i] == secretDigits[j] {
				wrongPlace++
				usedGuess[i] = true
				usedSecret[j] = true
				break
			}
		}
	}

	hint := GenerateSmartHint(secretDigits, guessDigits, rng)

	return Feedback{
		CorrectPlace: correctPlace,
		WrongPlace:   wrongPlace,
		Hint:         hint,
	}
}

// GenerateSmartHint builds possible hints and returns ONE randomized hint.
// It mirrors the original logic.
func GenerateSmartHint(secretDigits, guessDigits []int, rng *rand.Rand) string {
	var hints []string

	// First / second half placement
	firstHalfMatches := 0
	secondHalfMatches := 0
	for i := 0; i < codeDigits; i++ {
		if guessDigits[i] == secretDigits[i] {
			if i < codeDigits/2 {
				firstHalfMatches++
			} else {
				secondHalfMatches++
			}
		}
	}
	if firstHalfMatches > 0 {
		hints = append(hints, HintFirstHalfPlacement)
	}
	if secondHalfMatches > 0 {
		hints = append(hints, HintSecondHalfPlacement)
	}

	// Even / odd majority
	evenCount := 0
	for _, d := range secretDigits {
		if d%2 == 0 {
			evenCount++
		}
	}
	if evenCount >= codeDigits/2+1 {
		hints = append(hints, HintMostlyEvenDigits)
	}
	if evenCount <= codeDigits/2-1 {
		hints = append(hints, HintMostlyOddDigits)
	}

	// Repetition maps
	secretMap := make(map[int]int)
	guessMap := make(map[int]int)
	for _, d := range secretDigits {
		secretMap[d]++
	}
	for _, d := range guessDigits {
		guessMap[d]++
	}
	for d, gCount := range guessMap {
		if gCount > 1 && secretMap[d] == 0 {
			hints = append(hints, HintGuessRepeatedWrong)
		}
	}
	for d, sCount := range secretMap {
		if sCount > 1 && guessMap[d] == 1 {
			hints = append(hints, HintSecretRepeatingDigit)
		}
	}

	// High / low majority
	high := 0
	low := 0
	for _, d := range secretDigits {
		if d >= 5 {
			high++
		} else {
			low++
		}
	}
	if high >= 3 {
		hints = append(hints, HintMostlyHighDigits)
	}
	if low >= 3 {
		hints = append(hints, HintMostlyLowDigits)
	}

	// Sum range
	sum := 0
	for _, d := range secretDigits {
		sum += d
	}
	switch {
	case sum < 10:
		hints = append(hints, HintSumLow)
	case sum <= 20:
		hints = append(hints, HintSumMidLow)
	case sum <= 30:
		hints = append(hints, HintSumMidHigh)
	default:
		hints = append(hints, HintSumHigh)
	}

	// Increasing / decreasing
	if isStrictlyIncreasing(secretDigits) {
		hints = append(hints, HintIncreasingOrder)
	}
	if isStrictlyDecreasing(secretDigits) {
		hints = append(hints, HintDecreasingOrder)
	}

	// Shuffle hints deterministically via injected RNG
	if len(hints) > 1 {
		rng.Shuffle(len(hints), func(i, j int) {
			hints[i], hints[j] = hints[j], hints[i]
		})
	}

	if len(hints) > 0 {
		return hints[0]
	}
	return HintDefault // Should not reach here
}

func isStrictlyIncreasing(digits []int) bool {
	for i := 1; i < len(digits); i++ {
		if digits[i] <= digits[i-1] {
			return false
		}
	}
	return true
}

func isStrictlyDecreasing(digits []int) bool {
	for i := 1; i < len(digits); i++ {
		if digits[i] >= digits[i-1] {
			return false
		}
	}
	return true
}
