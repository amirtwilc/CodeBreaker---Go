package game

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSecretCode_Easy_NoRepeatingDigits(t *testing.T) {
	for i := 2; i <= 8; i++ {
		for j := 0; j < 5000; j++ {
			code := GenerateSecretCodeWithDifficulty(i, DifficultyEasy)
			assert.False(t, hasRepeatingDigit(code), "easy mode must not contain repeating digits")
		}
	}
}

func TestGenerateSecretCode_Medium_StillValidRange(t *testing.T) {
	for i := 2; i <= 8; i++ {
		minCode, maxCode = computeCodeBounds(i)
		for j := 0; j < 5000; j++ {
			code := GenerateSecretCodeWithDifficulty(i, DifficultyMedium)
			assert.GreaterOrEqual(t, code, minCode)
			assert.LessOrEqual(t, code, maxCode)
		}
	}
}

func TestGenerateSecretCode_Hard_Constraints(t *testing.T) {
	for i := 3; i <= 8; i++ {
		for j := 0; j < 5000; j++ {
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

func TestSplitToDigits_MinMax(t *testing.T) {
	for digits := 2; digits <= 8; digits++ {
		SetCodeDigits(digits)
		dmin := splitToDigits(minCode)
		require.Equal(t, digits, len(dmin))
		expected := make([]int, digits)
		expected[0] = 1
		assert.Equal(t, expected, dmin)

		dmax := splitToDigits(maxCode)
		require.Equal(t, digits, len(dmax))
		expMax := make([]int, digits)
		for i := range expMax {
			expMax[i] = 9
		}
		assert.Equal(t, expMax, dmax)
	}
}

func TestDigitSumReverseIncrement(t *testing.T) {
	codeDigits = 4
	assert.Equal(t, 10, digitSum([]int{1, 2, 3, 4}))
	assert.Equal(t, 9, digitSum([]int{1, 2, 3, 3}))

	in := []int{1, 2, 3, 4}
	assert.Equal(t, []int{4, 3, 2, 1}, reverseDigits(in))
	assert.Equal(t, []int{2, 3, 4, 5}, incrementDigits(in))

	in2 := []int{9, 9, 9, 9}
	assert.Equal(t, []int{0, 0, 0, 0}, incrementDigits(in2))
}

func TestDigitsToNumberPalindrome(t *testing.T) {
	d := []int{4, 3, 2, 1}
	assert.Equal(t, 4321, digitsToNumber(d))
	assert.True(t, isPalindrome(1221))
	assert.False(t, isPalindrome(1234))
}

func TestValidateGuess(t *testing.T) {
	codeDigits = 4
	tests := []struct {
		name        string
		input       string
		expected    int
		expectError bool
	}{
		{"Valid 4 digits", "2000", 2000, false},
		{"Valid with leading space", " 9931", 9931, false},
		{"Valid with trailing space", "1234 ", 1234, false},
		{"Valid with both", " 5678 ", 5678, false},
		{"With CR", "5678\r", 5678, false},
		{"Too short", "123", 0, true},
		{"Too long", "12345", 0, true},
		{"Empty", "", 0, true},
		{"Spaces only", "    ", 0, true},
		{"Letters", "12a4", 0, true},
		{"All letters", "abcd", 0, true},
		{"Special chars", "12#4", 0, true},
		{"Internal space", "1 23", 0, true},
		{"Negative", "-123", 0, true},
		{"Decimal", "12.3", 0, true},
		{"Unicode digits", "１２３４", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ValidateGuess(tt.input)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestGenerateTimestampPrefix_Format(t *testing.T) {
	prefix := GenerateTimestampPrefix()
	require.NotEmpty(t, prefix)
	expectedPrefixStart := TimePrefixLabel + ": "
	assert.True(t, strings.HasPrefix(prefix, expectedPrefixStart))
	assert.True(t, strings.HasSuffix(prefix, " - "))
	trimmed := strings.TrimPrefix(prefix, expectedPrefixStart)
	trimmed = strings.TrimSuffix(trimmed, " - ")
	trimmed = strings.TrimSpace(trimmed)
	_, err := time.Parse(TimeLayout, trimmed)
	assert.NoError(t, err)
}

func TestGenerateFeedbackScenarios(t *testing.T) {
	codeDigits = 4
	// All correct
	secret := 1234
	guess := 1234
	r := rand.New(rand.NewSource(1))
	f := GenerateFeedback(secret, guess, r)
	assert.Equal(t, codeDigits, f.CorrectPlace)
	assert.Equal(t, 0, f.WrongPlace)
	if f.Hint == "" {
		t.Fatal("hint must not be empty")
	}

	// None correct
	secret = 1234
	guess = 5678
	r = rand.New(rand.NewSource(2))
	f2 := GenerateFeedback(secret, guess, r)
	assert.Equal(t, 0, f2.CorrectPlace)
	assert.Equal(t, 0, f2.WrongPlace)
	assert.NotEmpty(t, f2.Hint)

	// Misplaced + repeated
	secret = 1234
	guess = 1393
	r = rand.New(rand.NewSource(3))
	f3 := GenerateFeedback(secret, guess, r)
	assert.Equal(t, 1, f3.CorrectPlace)
	assert.Equal(t, 1, f3.WrongPlace)
	assert.NotEmpty(t, f3.Hint)
}

func TestGenerateSmartHintVarious(t *testing.T) {
	codeDigits = 4
	r := rand.New(rand.NewSource(4))
	h := GenerateSmartHint([]int{1, 9, 8, 7}, []int{1, 0, 0, 0}, r)
	assert.NotEmpty(t, h)

	r2 := rand.New(rand.NewSource(5))
	h2 := GenerateSmartHint([]int{0, 1, 2, 9}, []int{0, 0, 0, 9}, r2)
	assert.NotEmpty(t, h2)

	r3 := rand.New(rand.NewSource(6))
	h3 := GenerateSmartHint([]int{2, 4, 6, 1}, []int{0, 0, 0, 0}, r3)
	assert.NotEmpty(t, h3)
}

func TestIsStrictlyDecreasing(t *testing.T) {
	require.True(t, isStrictlyDecreasing([]int{9, 8, 7, 1}))
	require.False(t, isStrictlyDecreasing([]int{9, 9, 7, 1})) // equal digit
	require.False(t, isStrictlyDecreasing([]int{1, 2, 3, 4})) // increasing
}
func TestGenerateSecret_HardDifficulty_PanicsIfTooShort(t *testing.T) {

	require.PanicsWithValue(
		t,
		"Hard difficulty must have at least 3 digits",
		func() {
			// Attempting to generate a Hard secret with <3 digits must panic
			GenerateSecretCodeWithDifficulty(2, DifficultyHard)
		},
		"Expected panic when using Hard difficulty with a code length < 3",
	)
}
