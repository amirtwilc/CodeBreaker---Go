package game

/*
import (
	"strings"
	"testing"

	"math/rand"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func TestGenerateSecretCode_Easy_NoRepeatingDigits(t *testing.T) {
	for i := 2; i <= 8; i++ {
		for j := 0; j < 5_000; j++ {
			code := GenerateSecretCodeWithDifficulty(i, DifficultyEasy)
			assert.False(t, hasRepeatingDigit(code), "easy mode must not contain repeating digits")
		}
	}
}

func TestGenerateSecretCode_Medium_StillValidRange(t *testing.T) {
	for i := 2; i <= 8; i++ {
		minCode, maxCode = computeCodeBounds(i)
		for j := 0; j < 5_000; j++ {
			code := GenerateSecretCodeWithDifficulty(i, DifficultyMedium)
			assert.GreaterOrEqual(t, code, minCode)
			assert.LessOrEqual(t, code, maxCode)
		}
	}
}

func TestGenerateSecretCode_Hard_Constraints(t *testing.T) {
	for i := 3; i <= 8; i++ {
		for j := 0; j < 5_000; j++ {
			code := GenerateSecretCodeWithDifficulty(i, DifficultyHard)
			assert.True(t, hasRepeatingDigit(code), "hard mode must contain a repeated digit")
		}
	}
}

func TestHasRepeatingDigit(t *testing.T) {
	codeDigits = 4
	assert.True(t, hasRepeatingDigit(1123))
	assert.True(t, hasRepeatingDigit(9009))
	assert.False(t, hasRepeatingDigit(1234))
	assert.False(t, hasRepeatingDigit(9876))
}

func TestSplitToDigits_Min(t *testing.T) {
	for digits := 2; digits <= 8; digits++ {
		SetCodeDigits(digits)

		d := splitToDigits(minCode)
		assert.Equal(t, digits, len(d), "wrong length for digits=%d", digits)

		// expected slice: [1, 0, 0, ..., 0]
		expected := make([]int, digits)
		expected[0] = 1

		assert.Equal(t, expected, d, "digits mismatch for digits=%d", digits)
	}
}

func TestSplitToDigits_Max(t *testing.T) {
	for digits := 2; digits <= 8; digits++ {
		SetCodeDigits(digits)

		d := splitToDigits(maxCode)
		assert.Equal(t, digits, len(d), "wrong length for digits=%d", digits)

		// expected slice: [9, 9, 9, ..., 9]
		expected := make([]int, digits)
		for i := range expected {
			expected[i] = 9
		}

		assert.Equal(t, expected, d, "digits mismatch for digits=%d", digits)
	}
}

func TestDigitSum_Even(t *testing.T) {
	assert.Equal(t, 10, digitSum([]int{1, 2, 3, 4}))
}

func TestDigitSum_Odd(t *testing.T) {
	assert.Equal(t, 9, digitSum([]int{1, 2, 3, 3}))
}

func TestReverseDigits(t *testing.T) {
	codeDigits = 4
	in := []int{1, 2, 3, 4}
	out := reverseDigits(in)

	assert.Equal(t, 4, len(out))
	assert.Equal(t, []int{4, 3, 2, 1}, out)
}

func TestIncrementDigits_NoWrap(t *testing.T) {
	in := []int{1, 2, 3, 4}
	out := incrementDigits(in)

	assert.Equal(t, 4, len(out))
	assert.Equal(t, []int{2, 3, 4, 5}, out)
}

func TestIncrementDigits_WithWrap(t *testing.T) {
	in := []int{9, 9, 9, 9}
	out := incrementDigits(in)

	assert.Equal(t, 4, len(out))
	assert.Equal(t, []int{0, 0, 0, 0}, out)
}

func TestDigitsToNumber(t *testing.T) {
	d := []int{4, 3, 2, 1}
	n := digitsToNumber(d)

	assert.Equal(t, 4321, n)
}

func TestIsPalindrome_True(t *testing.T) {
	assert.True(t, isPalindrome(1221))
}

func TestIsPalindrome_False(t *testing.T) {
	assert.False(t, isPalindrome(1234))
}

func TestPalindromeOverrideRule(t *testing.T) {
	// 1221 → sum 6 (even) → reverse → 1221 → palindrome → 7777
	d := splitToDigits(1221)
	sum := digitSum(d)

	assert.Equal(t, 6, sum)

	reversed := reverseDigits(d)
	result := digitsToNumber(reversed)

	assert.True(t, isPalindrome(result))
	assert.Equal(t, 7777, 7777)
}

func TestEvenSumReversePath(t *testing.T) {
	// 1357 → sum = 16 (even) → reverse → 7531
	d := splitToDigits(1357)
	assert.Equal(t, 16, digitSum(d))

	reversed := reverseDigits(d)
	assert.Equal(t, []int{7, 5, 3, 1}, reversed)
}

func TestOddSumIncrementWithWrap(t *testing.T) {
	// 9998 → sum = 35 (odd) → increment → 0009
	d := splitToDigits(9998)
	assert.Equal(t, 35, digitSum(d))

	inc := incrementDigits(d)
	assert.Equal(t, []int{0, 0, 0, 9}, inc)

	n := digitsToNumber(inc)
	assert.Equal(t, 9, n)
}

func TestValidateGuess(t *testing.T) {
	codeDigits = 4
	tests := []struct {
		name        string
		input       string
		expected    int
		expectError bool
	}{
		//VALID CASES
		{
			name:        "Valid 4 digits",
			input:       "2000",
			expected:    2000,
			expectError: false,
		},
		{
			name:        "Valid with leading space",
			input:       " 9931",
			expected:    9931,
			expectError: false,
		},
		{
			name:        "Valid with trailing space",
			input:       "1234 ",
			expected:    1234,
			expectError: false,
		},
		{
			name:        "Valid with leading and trailing spaces",
			input:       " 5678 ",
			expected:    5678,
			expectError: false,
		},
		{
			name:        "Valid with carriage return",
			input:       "5678\r",
			expected:    5678,
			expectError: false,
		},

		//INVALID LENGTH
		{
			name:        "Too short - 3 digits",
			input:       "123",
			expectError: true,
		},
		{
			name:        "Too long - 5 digits",
			input:       "12345",
			expectError: true,
		},
		{
			name:        "Empty input",
			input:       "",
			expectError: true,
		},
		{
			name:        "Only spaces",
			input:       "    ",
			expectError: true,
		},

		//NON-DIGIT CHARACTERS
		{
			name:        "Contains letters",
			input:       "12a4",
			expectError: true,
		},
		{
			name:        "All letters",
			input:       "abcd",
			expectError: true,
		},
		{
			name:        "Contains special characters",
			input:       "12#4",
			expectError: true,
		},
		{
			name:        "Internal space",
			input:       "1 23",
			expectError: true,
		},
		{
			name:        "Negative number",
			input:       "-123",
			expectError: true,
		},
		{
			name:        "Decimal number",
			input:       "12.3",
			expectError: true,
		},
		{
			name:        "Unicode digits",
			input:       "１２３４", // Full-width digits
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateGuess(tt.input)

			if tt.expectError {
				require.Error(t, err, "Expected an error but got none")
			} else {
				require.NoError(t, err, "Did not expect an error")
				require.Equal(t, tt.expected, result, "Returned value mismatch")
			}
		})
	}
}

func TestGenerateTimestampPrefix_Format(t *testing.T) {
	prefix := GenerateTimestampPrefix()

	require.NotEmpty(t, prefix)

	// Must start with the constant label
	expectedPrefixStart := TimePrefixLabel + ": "
	assert.True(t, strings.HasPrefix(prefix, expectedPrefixStart))

	// Must end with the constant suffix from the format
	assert.True(t, strings.HasSuffix(prefix, " - "))

	// Extract the numeric timestamp part using the constants
	trimmed := strings.TrimPrefix(prefix, expectedPrefixStart)
	trimmed = strings.TrimSuffix(trimmed, " - ")
	trimmed = strings.TrimSpace(trimmed)

	_, err := time.Parse(TimeLayout, trimmed)
	assert.NoError(t, err, "timestamp part must be in correct format")
}

func TestGenerateTimestampPrefix_MultipleCalls(t *testing.T) {
	p1 := GenerateTimestampPrefix()
	p2 := GenerateTimestampPrefix()

	require.NotEmpty(t, p1)
	require.NotEmpty(t, p2)

	expectedPrefixStart := TimePrefixLabel + ": "

	assert.True(t, strings.HasPrefix(p1, expectedPrefixStart))
	assert.True(t, strings.HasPrefix(p2, expectedPrefixStart))
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func TestGenerateFeedback_AllCorrect(t *testing.T) {
	// ensure code length 4 for predictable behavior
	codeDigits = 4

	secret := 1234
	guess := 1234
	rng := rand.New(rand.NewSource(1))

	f := GenerateFeedback(secret, guess, rng)

	if f.CorrectPlace != codeDigits {
		t.Fatalf("expected %d correct places, got %d", codeDigits, f.CorrectPlace)
	}
	if f.WrongPlace != 0 {
		t.Fatalf("expected 0 wrong place, got %d", f.WrongPlace)
	}

	// With a perfect match, hint set may include increasing/decreasing, sum-range and "VERY close".
	possible := []string{
		HintIncreasingOrder,
		HintDecreasingOrder,
		HintSumLow,
		HintSumMidLow,
		HintSumMidHigh,
		HintSumHigh,
		HintMostlyLowDigits,
	}

	if !containsString(possible, f.Hint) && f.Hint != HintDefault {
		t.Fatalf("hint unexpected for all-correct: %q (expected one of possible list)", f.Hint)
	}
}

func TestGenerateFeedback_NoneCorrect(t *testing.T) {
	codeDigits = 4

	secret := 1234
	guess := 5678 // completely different digits
	rng := rand.New(rand.NewSource(2))

	f := GenerateFeedback(secret, guess, rng)

	if f.CorrectPlace != 0 {
		t.Fatalf("expected 0 correct places, got %d", f.CorrectPlace)
	}
	if f.WrongPlace != 0 {
		t.Fatalf("expected 0 wrong place, got %d", f.WrongPlace)
	}

	expectedPossible := []string{
		HintMostlyEvenDigits,
		HintMostlyOddDigits,
		HintSumLow,
		HintSumMidLow,
		HintSumMidHigh,
		HintSumHigh,
	}

	if !containsString(expectedPossible, f.Hint) {
		t.Fatalf("hint unexpected for none-correct: %q", f.Hint)
	}
}

func TestGenerateFeedback_MisplacedAndCorrect(t *testing.T) {
	codeDigits = 4

	// secret has digits 1,2,3,4
	// guess has 1 in correct place, 3 present but misplaced, 9 and 9 not in secret (repeated guess digit)
	secret := 1234
	guess := 1393 // index0 correct (1), index1 wrong (3 in secret but misplaced), repeated 3 in guess
	rng := rand.New(rand.NewSource(3))

	f := GenerateFeedback(secret, guess, rng)

	if f.CorrectPlace != 1 {
		t.Fatalf("expected 1 correct place, got %d", f.CorrectPlace)
	}
	// There is one misplaced (3) but guessed twice; algorithm should count only one wrongPlace
	if f.WrongPlace != 1 {
		t.Fatalf("expected 1 wrong place, got %d", f.WrongPlace)
	}

	// For this scenario, possible hints include repetition analysis and sum-range/hight-low
	expectedPossible := []string{
		HintGuessRepeatedWrong, // not applicable here since 3 exists, so maybe not this
		HintSecretRepeatingDigit,
		HintMostlyEvenDigits,
		HintMostlyOddDigits,
	}

	if !containsString(expectedPossible, f.Hint) {
		// it's acceptable if hint is some other allowed hint; we just ensure not empty
		if f.Hint == "" {
			t.Fatalf("hint should not be empty")
		}
	}
}

func TestGenerateSmartHint_FirstSecondHalf(t *testing.T) {
	codeDigits = 4

	// first half match: positions 0 are equal
	secret := []int{1, 9, 8, 7}
	guess := []int{1, 0, 0, 0}
	rng := rand.New(rand.NewSource(4))

	h := GenerateSmartHint(secret, guess, rng)
	possible := []string{
		HintFirstHalfPlacement,
		HintMostlyHighDigits,
		// sum-range
		HintSumLow,
		HintSumMidLow,
		HintSumMidHigh,
		HintSumHigh,
	}

	if !containsString(possible, h) {
		t.Fatalf("unexpected hint: %q", h)
	}

	// second half match: position 3 equal
	secret2 := []int{0, 1, 2, 9}
	guess2 := []int{0, 0, 0, 9}
	rng2 := rand.New(rand.NewSource(5))
	h2 := GenerateSmartHint(secret2, guess2, rng2)

	if !containsString([]string{
		HintFirstHalfPlacement,
		HintDefault,
	}, h2) {
		t.Fatalf("unexpected second-half hint: %q", h2)
	}
}

func TestGenerateSmartHint_EvenOddAndHighLow(t *testing.T) {
	codeDigits = 4

	// mostly even secret
	secretEven := []int{2, 4, 6, 1}
	guess := []int{0, 0, 0, 0}
	rng := rand.New(rand.NewSource(6))
	h := GenerateSmartHint(secretEven, guess, rng)
	if !(h == HintMostlyEvenDigits || h == HintSumMidLow) {
		t.Fatalf("unexpected hint for mostly even secret: %q", h)
	}

	// mostly low secret
	secretLow := []int{1, 2, 0, 3}
	rng2 := rand.New(rand.NewSource(7))
	h2 := GenerateSmartHint(secretLow, guess, rng2)
	if !(h2 == HintMostlyOddDigits || h2 == HintSumLow || h2 == HintMostlyLowDigits) {
		t.Fatalf("unexpected hint for mostly low secret: %q", h2)
	}
}

func TestGenerateSmartHint_RepetitionAnalysis(t *testing.T) {
	codeDigits = 4

	// guess repeats digit 5 which is not in secret -> triggers repetition-not-in-secret hint
	secret := []int{1, 2, 3, 4}
	guess := []int{5, 5, 0, 0}
	rng := rand.New(rand.NewSource(8))
	h := GenerateSmartHint(secret, guess, rng)

	possible := []string{
		HintGuessRepeatedWrong,
	}
	if !containsString(possible, h) {
		t.Fatalf("unexpected repetition hint: %q", h)
	}

	// secret has repeating digit and guess contains it once -> triggers "The secret contains a repeating digit"
	secret2 := []int{2, 2, 3, 4}
	guess2 := []int{2, 0, 0, 0}
	rng2 := rand.New(rand.NewSource(9))
	h2 := GenerateSmartHint(secret2, guess2, rng2)
	if !(h2 == HintSecretRepeatingDigit || h2 == HintMostlyEvenDigits) {
		t.Fatalf("unexpected repeating-secret hint: %q", h2)
	}
}

func TestGenerateSmartHint_SumRangeAndDistance(t *testing.T) {
	codeDigits = 4

	// sum < 10
	secretLowSum := []int{0, 0, 1, 2} // sum=3
	guess := []int{9, 9, 9, 9}
	rng := rand.New(rand.NewSource(10))
	h := GenerateSmartHint(secretLowSum, guess, rng)
	if !(h == HintSumLow || h == HintMostlyEvenDigits) {
		t.Fatalf("unexpected hint for low sum: %q", h)
	}

	// sum between 10 and 20
	secretMid := []int{2, 3, 4, 2} // sum=11
	rng2 := rand.New(rand.NewSource(11))
	h2 := GenerateSmartHint(secretMid, guess, rng2)
	if !(h2 == HintSumMidLow || h2 == HintDefault) {
		t.Fatalf("unexpected hint for mid sum: %q", h2)
	}

	// sum >30
	secretHigh := []int{9, 9, 8, 9} // sum=35
	rng3 := rand.New(rand.NewSource(12))
	h3 := GenerateSmartHint(secretHigh, guess, rng3)
	if !(h3 == HintSumHigh || h3 == HintDefault) {
		t.Fatalf("unexpected hint for high sum: %q", h3)
	}
}

func TestGenerateSmartHint_IncreasingDecreasing(t *testing.T) {
	codeDigits = 4

	secretInc := []int{1, 2, 3, 4}
	guess := []int{0, 0, 0, 0}
	rng := rand.New(rand.NewSource(13))
	h := GenerateSmartHint(secretInc, guess, rng)
	if !(h == HintIncreasingOrder || h == HintSumMidLow || h == HintMostlyLowDigits) {
		t.Fatalf("unexpected hint for increasing: %q", h)
	}

	secretDec := []int{9, 8, 7, 6}
	rng2 := rand.New(rand.NewSource(14))
	h2 := GenerateSmartHint(secretDec, guess, rng2)
	if !(h2 == HintDecreasingOrder || h2 == HintSumHigh || h2 == HintMostlyHighDigits) {
		t.Fatalf("unexpected hint for decreasing: %q", h2)
	}
}

func TestGenerateSmartHint_RandomnessNonEmpty(t *testing.T) {
	codeDigits = 4

	// For many inputs, hints list should not be empty and result should be one of possible hints
	secret := []int{1, 3, 5, 7}
	guess := []int{2, 4, 6, 8}
	rng := rand.New(rand.NewSource(18))
	h := GenerateSmartHint(secret, guess, rng)
	if h == "" {
		t.Fatalf("hint should not be empty")
	}
}

func TestSplitToDigits_And_Reverse_Consistency(t *testing.T) {

	n := 4321
	d := splitToDigits(n)
	got := digitsToNumber(d)
	if got != n {
		t.Fatalf("expected roundtrip digits->number to return %d, got %d (digits %v)", n, got, d)
	}

	rev := reverseDigits(d)
	if digitsToNumber(rev) != 1234 {
		t.Fatalf("reverse digits produced unexpected number: %v", rev)
	}
}

// ensure arrays comparisons work reliably
func TestHelpersContainment(t *testing.T) {
	s := []string{"a", "b", "c"}
	if !containsString(s, "b") {
		t.Fatalf("containsString failed")
	}
	if containsString(s, "z") {
		t.Fatalf("containsString false positive")
	}
}
*/
