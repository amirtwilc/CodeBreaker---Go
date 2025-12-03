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
	CodeLength                  = 4
	MinCode                     = 1000
	MaxCode                     = 9999
	CodeRange                   = 9000 // MaxCode - MinCode + 1
	TimePrefixLabel             = "TIME"
	TimePrefixFormat            = TimePrefixLabel + ": %s - "
	TimeLayout                  = "15:04:05" // ✅ HH:mm:ss in Go format
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
	HintFirstHalf               = "One of your correct digits is in the first half of the number"
	HintSecondHalf              = "One of your correct digits is in the second half of the number"
	HintNone                    = "No correct digits found"
)

type Feedback struct {
	CorrectPlace int
	WrongPlace   int
	Hint         string
}

func ValidateGuess(input string) (int, error) {
	trimmed := strings.TrimSpace(input)

	// Must be exactly 4 digits
	if len(trimmed) != CodeLength {
		return 0, fmt.Errorf("guess must contain exactly %d digits", CodeLength)
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

func GenerateSecretCode() int {
	return GenerateSecretCodeWithDifficulty(DifficultyMedium)
}

func GenerateSecretCodeWithDifficulty(d Difficulty) int {
	for {
		base := rand.Intn(CodeRange) + MinCode

		digits := splitToDigits(base)
		sum := digitSum(digits)

		var modified []int

		// MEDIUM RULE (existing logic)
		if sum%2 == 0 {
			modified = reverseDigits(digits)
		} else {
			modified = incrementDigits(digits)
		}

		final := digitsToNumber(modified)

		// Palindrome override rule (applies to all difficulties)
		if isPalindrome(final) {
			final = 7777
		}

		//in case final is not in correct size
		if final < MinCode || final > MaxCode {
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
			if hasRepeatingDigit(final) && isPrime(digitSum(splitToDigits(final))) {
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
	out := make([]int, CodeLength)

	for i := CodeLength - 1; i >= 0; i-- {
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
	out := make([]int, CodeLength)
	for i := 0; i < CodeLength; i++ {
		out[i] = d[CodeLength-1-i]
	}
	return out
}

func incrementDigits(d []int) []int {
	out := make([]int, CodeLength)
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
	return d[0]*1000 + d[1]*100 + d[2]*10 + d[3]
}

func isPalindrome(n int) bool {
	d := splitToDigits(n)
	return d[0] == d[3] && d[1] == d[2]
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

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func GenerateFeedback(secret, guess int, rng *rand.Rand) Feedback {
	secretDigits := splitToDigits(secret)
	guessDigits := splitToDigits(guess)

	usedSecret := make([]bool, CodeLength)
	usedGuess := make([]bool, CodeLength)

	correctPlace := 0
	wrongPlace := 0

	// Pass 1: correct place
	for i := 0; i < CodeLength; i++ {
		if secretDigits[i] == guessDigits[i] {
			correctPlace++
			usedSecret[i] = true
			usedGuess[i] = true
		}
	}

	// Pass 2: misplaced digits
	for i := 0; i < CodeLength; i++ {
		if usedGuess[i] {
			continue
		}
		for j := 0; j < CodeLength; j++ {
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

	// Hint logic
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

	for i := 0; i < CodeLength; i++ {
		if guessDigits[i] == secretDigits[i] {
			if i < CodeLength/2 {
				firstHalfMatches++
			} else {
				secondHalfMatches++
			}
		}
	}

	if firstHalfMatches > 0 {
		hints = append(hints,
			fmt.Sprintf("At least 1 of the correctly placed digit(s) are in the FIRST half"),
		)
	}

	if secondHalfMatches > 0 {
		hints = append(hints,
			fmt.Sprintf("At least 1 of the correctly placed digit(s) are in the SECOND half"),
		)
	}

	// 3️⃣ Even / Odd distribution
	evenCount := 0
	for _, d := range secretDigits {
		if d%2 == 0 {
			evenCount++
		}
	}

	if evenCount >= 3 {
		hints = append(hints, "The secret contains mostly EVEN digits")
	}
	if evenCount <= 1 {
		hints = append(hints, "The secret contains mostly ODD digits")
	}

	// 4️⃣ Repetition analysis
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
				"Guess repeated a digit that does NOT exist in the secret",
			)
		}
	}

	for d, sCount := range secretMap {
		if sCount > 1 && guessMap[d] == 1 {
			hints = append(hints,
				"The secret contains a repeating digit",
			)
		}
	}

	// 5️⃣ High vs Low majority
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
		hints = append(hints, "Most digits in the secret are HIGH (5–9)")
	}
	if low >= 3 {
		hints = append(hints, "Most digits in the secret are LOW (0–4)")
	}

	// ✅ ✅ ✅ NEW ADVANCED HINT IDEAS

	// 6️⃣ Sum of secret digits (range, not exact)
	sum := 0
	for _, d := range secretDigits {
		sum += d
	}

	switch {
	case sum < 10:
		hints = append(hints, "The sum of the secret digits is lower than 10")
	case sum <= 20:
		hints = append(hints, "The sum of the secret digits is between 10 and 20")
	case sum <= 30:
		hints = append(hints, "The sum of the secret digits is between 20 and 30")
	default:
		hints = append(hints, "The sum of the secret digits is greater than 30")
	}

	// 7️⃣ Increasing / decreasing pattern
	if isStrictlyIncreasing(secretDigits) {
		hints = append(hints, "The secret digits are in a strictly INCREASING order")
	}
	if isStrictlyDecreasing(secretDigits) {
		hints = append(hints, "The secret digits are in a strictly DECREASING order")
	}

	// 9️⃣ Distance hint
	totalDistance := 0
	for i := 0; i < CodeLength; i++ {
		diff := secretDigits[i] - guessDigits[i]
		if diff < 0 {
			diff = -diff
		}
		totalDistance += diff
	}

	if totalDistance < 10 {
		hints = append(hints, "Guess is VERY close overall")
	} else if totalDistance < 20 {
		hints = append(hints, "Guess is getting closer")
	} else {
		hints = append(hints, "Guess is still FAR from the secret")
	}

	// RANDOMIZE HINT SELECTION
	rng.Shuffle(len(hints), func(i, j int) {
		hints[i], hints[j] = hints[j], hints[i]
	})

	// Always return ONE randomized hint
	if len(hints) > 0 {
		return hints[0]
	}

	// Absolute fallback (should never happen)
	return "You are the best!"
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
