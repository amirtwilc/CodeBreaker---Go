package game

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Difficulty string

const (
	TimePrefixLabel             = "TIME"
	TimePrefixFormat            = TimePrefixLabel + ": %s - "
	TimeLayout                  = "15:04:05" // HH:mm:ss in Go format
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

const (
	HintFirstHalfPlacement  = "At least 1 of the correctly placed digit(s) are in the FIRST half"
	HintSecondHalfPlacement = "At least 1 of the correctly placed digit(s) are in the SECOND half"

	HintMostlyEvenDigits = "The secret contains mostly EVEN digits"
	HintMostlyOddDigits  = "The secret contains mostly ODD digits"

	HintGuessRepeatedWrong   = "Guess repeated a digit that does NOT exist in the secret"
	HintSecretRepeatingDigit = "The secret contains a repeating digit"

	HintMostlyHighDigits = "Most digits in the secret are HIGH (5–9)"
	HintMostlyLowDigits  = "Most digits in the secret are LOW (0–4)"

	HintSumLow     = "The sum of the secret digits is lower than 10"
	HintSumMidLow  = "The sum of the secret digits is between 10 and 20"
	HintSumMidHigh = "The sum of the secret digits is between 20 and 30"
	HintSumHigh    = "The sum of the secret digits is greater than 30"

	HintIncreasingOrder = "The secret digits are in a strictly INCREASING order"
	HintDecreasingOrder = "The secret digits are in a strictly DECREASING order"

	HintDefault = "You are the best!"
)

// default values
var codeDigits = 4
var minCode = 1000
var maxCode = 9999
var codeRange = 9000 // MaxCode - MinCode + 1

type Feedback struct {
	CorrectPlace int
	WrongPlace   int
	Hint         string
}

func SetCodeDigits(digits int) {
	if digits < 2 || digits > 8 {
		panic("code digits must be between 2 and 8")
	}

	codeDigits = digits
	minCode, maxCode = computeCodeBounds(codeDigits)
	codeRange = maxCode - minCode + 1
}

func computeCodeBounds(digits int) (min int, max int) {
	min = 1
	for i := 1; i < digits; i++ {
		min *= 10
	}
	max = min*10 - 1
	return
}

func ValidateGuess(input string) (int, error) {
	trimmed := strings.TrimSpace(input)

	// Must be exactly 4 digits
	if len(trimmed) != codeDigits {
		return 0, fmt.Errorf("guess must contain exactly %d digits", codeDigits)
	}

	// Ensure all characters are digits
	for _, ch := range trimmed {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("guess must contain only digits")
		}
	}

	guess, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("invalid number format")
	}

	return guess, nil
}

func GenerateSecretCodeWithDifficulty(codeLength int, d Difficulty) int {
	for {
		if d == DifficultyHard && codeLength < 3 {
			panic("Hard difficulty must have at least 3 digits")
		}

		SetCodeDigits(codeLength)

		base := rand.Intn(codeRange) + minCode

		digits := splitToDigits(base)
		sum := digitSum(digits)

		var modified []int

		// Business requirement - based on even or odd number
		if sum%2 == 0 {
			modified = reverseDigits(digits)
		} else {
			modified = incrementDigits(digits)
		}

		final := digitsToNumber(modified)

		// Business requirement - Palindrome override rule (applies to all difficulties)
		if isPalindrome(final) {
			final = 7777
		}

		//in case final is not in correct size - try again
		if final < minCode || final > maxCode {
			continue
		}

		// Difficulty filters
		switch d {
		case DifficultyEasy:
			if !hasRepeatingDigit(final) {
				return final
			}

		case DifficultyMedium:
			return final

		case DifficultyHard:
			if hasRepeatingDigit(final) {
				return final
			}
		}
	}
}

// GenerateTimestampPrefix generates a textual prefix containing the current time
func GenerateTimestampPrefix() string {
	now := time.Now().Format(TimeLayout)
	return fmt.Sprintf(TimePrefixFormat, now)
}

func splitToDigits(n int) []int {
	out := make([]int, codeDigits)

	for i := codeDigits - 1; i >= 0; i-- {
		out[i] = n % 10
		n /= 10
	}
	return out
}

func digitSum(d []int) int {
	sum := 0
	for _, v := range d {
		sum += v
	}
	return sum
}

func reverseDigits(d []int) []int {
	out := make([]int, codeDigits)
	for i := 0; i < codeDigits; i++ {
		out[i] = d[codeDigits-1-i]
	}
	return out
}

func incrementDigits(d []int) []int {
	out := make([]int, codeDigits)
	for i, v := range d {
		if v == 9 {
			out[i] = 0
		} else {
			out[i] = v + 1
		}
	}
	return out
}

func digitsToNumber(d []int) int {
	result := 0
	for _, digit := range d {
		result = result*10 + digit
	}
	return result
}

func isPalindrome(n int) bool {
	d := splitToDigits(n)

	i := 0
	j := len(d) - 1

	for i < j {
		if d[i] != d[j] {
			return false
		}
		i++
		j--
	}
	return true
}

func hasRepeatingDigit(n int) bool {
	seen := make(map[int]bool)
	for _, d := range splitToDigits(n) {
		if seen[d] {
			return true
		}
		seen[d] = true
	}
	return false
}

// GenerateFeedback provides a feedback regarding the guess
func GenerateFeedback(secret, guess int, rng *rand.Rand) Feedback {
	secretDigits := splitToDigits(secret)
	guessDigits := splitToDigits(guess)

	usedSecret := make([]bool, codeDigits)
	usedGuess := make([]bool, codeDigits)

	correctPlace := 0
	wrongPlace := 0

	// Check guess place correctness
	for i := 0; i < codeDigits; i++ {
		if secretDigits[i] == guessDigits[i] {
			correctPlace++
			usedSecret[i] = true
			usedGuess[i] = true
		}
	}

	// Check guess misplaced digits
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

func GenerateSmartHint(secretDigits, guessDigits []int, rng *rand.Rand) string {
	var hints []string

	// Region-based correct placement hints
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
		hints = append(hints,
			fmt.Sprintf(HintFirstHalfPlacement),
		)
	}

	if secondHalfMatches > 0 {
		hints = append(hints,
			fmt.Sprintf(HintSecondHalfPlacement),
		)
	}

	// Even / Odd distribution
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

	// Repetition analysis
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
			hints = append(hints,
				HintGuessRepeatedWrong,
			)
		}
	}

	for d, sCount := range secretMap {
		if sCount > 1 && guessMap[d] == 1 {
			hints = append(hints,
				HintSecretRepeatingDigit,
			)
		}
	}

	// High vs Low majority
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

	// Sum of secret digits (range, not exact)
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

	// Increasing / decreasing pattern
	if isStrictlyIncreasing(secretDigits) {
		hints = append(hints, HintIncreasingOrder)
	}
	if isStrictlyDecreasing(secretDigits) {
		hints = append(hints, HintDecreasingOrder)
	}

	// Randomize hints - decrease chance of receiving same hint every time
	rng.Shuffle(len(hints), func(i, j int) {
		hints[i], hints[j] = hints[j], hints[i]
	})

	// Always return ONE randomized hint
	if len(hints) > 0 {
		return hints[0]
	}

	// Absolute fallback (should never happen)
	return HintDefault
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
