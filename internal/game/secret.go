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
	TimeLayout                  = "15:04:05" // HH:mm:ss
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

// Hint constants kept from original business logic.
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

// default values (configurable via SetCodeDigits)
var codeDigits = 4
var minCode = 1000
var maxCode = 9999
var codeRange = 9000 // max-min+1

// SetCodeDigits configures the code length (2..8); panics otherwise (same as original project).
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

func GenerateTimestampPrefix() string {
	now := time.Now().Format(TimeLayout)
	return fmt.Sprintf(TimePrefixFormat, now)
}

// ValidateGuess validates an input guess string according to codeDigits.
func ValidateGuess(input string) (int, error) {
	trimmed := strings.TrimSpace(input)

	if len(trimmed) != codeDigits {
		return 0, fmt.Errorf("guess must contain exactly %d digits", codeDigits)
	}
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

// GenerateSecretCodeWithDifficulty implements the original business rules for generating the secret.
// It loops until a value matching difficulty constraints is produced.
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
		if sum%2 == 0 {
			modified = reverseDigits(digits)
		} else {
			modified = incrementDigits(digits)
		}

		final := digitsToNumber(modified)

		// Palindrome override: convert palindromes to 7777 (same behavior as original)
		if isPalindrome(final) {
			final = 7777
		}

		// ensure within bounds
		if final < minCode || final > maxCode {
			continue
		}

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

// Helpers
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
