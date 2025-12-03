package game

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSecretCode_Easy_NoRepeatingDigits(t *testing.T) {
	for i := 0; i < 5_000; i++ {
		code := GenerateSecretCodeWithDifficulty(DifficultyEasy)
		assert.False(t, hasRepeatingDigit(code), "easy mode must not contain repeating digits")
	}
}

func TestGenerateSecretCode_Medium_StillValidRange(t *testing.T) {
	for i := 0; i < 5_000; i++ {
		code := GenerateSecretCodeWithDifficulty(DifficultyMedium)
		assert.GreaterOrEqual(t, code, 0)
		assert.LessOrEqual(t, code, MaxCode)
	}
}

func TestGenerateSecretCode_Hard_Constraints(t *testing.T) {
	for i := 0; i < 5_000; i++ {
		code := GenerateSecretCodeWithDifficulty(DifficultyHard)

		assert.True(t, hasRepeatingDigit(code), "hard mode must contain a repeated digit")

		sum := digitSum(splitToDigits(code))
		assert.True(t, isPrime(sum), "hard mode digit sum must be prime")
	}
}

func TestIsPrime(t *testing.T) {
	assert.False(t, isPrime(0))
	assert.False(t, isPrime(1))
	assert.True(t, isPrime(2))
	assert.True(t, isPrime(3))
	assert.False(t, isPrime(4))
	assert.True(t, isPrime(5))
	assert.False(t, isPrime(9))
	assert.True(t, isPrime(13))
}

func TestHasRepeatingDigit(t *testing.T) {
	assert.True(t, hasRepeatingDigit(1123))
	assert.True(t, hasRepeatingDigit(9009))
	assert.False(t, hasRepeatingDigit(1234))
	assert.False(t, hasRepeatingDigit(9876))
}

func TestGenerateSecretCode_Range(t *testing.T) {
	for i := 0; i < 10_000; i++ {
		code := GenerateSecretCode()

		assert.GreaterOrEqual(t, code, 0)
		assert.LessOrEqual(t, code, MaxCode)
	}
}

func TestSplitToDigits_Min(t *testing.T) {
	d := splitToDigits(MinCode)
	assert.Equal(t, CodeLength, len(d))
	assert.Equal(t, []int{1, 0, 0, 0}, d)
}

func TestSplitToDigits_Max(t *testing.T) {
	d := splitToDigits(MaxCode)
	assert.Equal(t, CodeLength, len(d))
	assert.Equal(t, []int{9, 9, 9, 9}, d)
}

func TestDigitSum_Even(t *testing.T) {
	assert.Equal(t, 10, digitSum([]int{1, 2, 3, 4}))
}

func TestDigitSum_Odd(t *testing.T) {
	assert.Equal(t, 9, digitSum([]int{1, 2, 3, 3}))
}

func TestReverseDigits(t *testing.T) {
	in := []int{1, 2, 3, 4}
	out := reverseDigits(in)

	assert.Equal(t, CodeLength, len(out))
	assert.Equal(t, []int{4, 3, 2, 1}, out)
}

func TestIncrementDigits_NoWrap(t *testing.T) {
	in := []int{1, 2, 3, 4}
	out := incrementDigits(in)

	assert.Equal(t, CodeLength, len(out))
	assert.Equal(t, []int{2, 3, 4, 5}, out)
}

func TestIncrementDigits_WithWrap(t *testing.T) {
	in := []int{9, 9, 9, 9}
	out := incrementDigits(in)

	assert.Equal(t, CodeLength, len(out))
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

	// Must be a valid integer
	_, err := strconv.ParseInt(trimmed, 10, 64)
	assert.NoError(t, err, "timestamp part must be a valid integer")
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

func TestGenerateFeedback_ExactMatch(t *testing.T) {
	secret := 1234
	guess := 1234

	f := GenerateFeedback(secret, guess)

	assert.Equal(t, 4, f.CorrectPlace)
	assert.Equal(t, 0, f.WrongPlace)
	assert.Equal(t, HintFirstHalf, f.Hint)
}

func TestGenerateFeedback_AllMisplaced(t *testing.T) {
	secret := 1234
	guess := 4321

	f := GenerateFeedback(secret, guess)

	assert.Equal(t, 0, f.CorrectPlace)
	assert.Equal(t, 4, f.WrongPlace)
	assert.Equal(t, HintNone, f.Hint)
}

func TestGenerateFeedback_FirstHalfHint(t *testing.T) {
	secret := 1234
	guess := 1299 // 1 and 2 correct, first half

	f := GenerateFeedback(secret, guess)

	assert.Equal(t, 2, f.CorrectPlace)
	assert.Equal(t, 0, f.WrongPlace)
	assert.Equal(t, HintFirstHalf, f.Hint)
}

func TestGenerateFeedback_SecondHalfHint(t *testing.T) {
	secret := 1234
	guess := 9934 // 3 and 4 correct, second half

	f := GenerateFeedback(secret, guess)

	assert.Equal(t, 2, f.CorrectPlace)
	assert.Equal(t, 0, f.WrongPlace)
	assert.Equal(t, HintSecondHalf, f.Hint)
}

func TestGenerateFeedback_Mixed(t *testing.T) {
	secret := 1234
	guess := 1243

	f := GenerateFeedback(secret, guess)

	assert.Equal(t, 2, f.CorrectPlace) // 1 and 2
	assert.Equal(t, 2, f.WrongPlace)   // 3 and 4 swapped
	assert.Equal(t, HintFirstHalf, f.Hint)
}

func TestGenerateFeedback_NoMatches(t *testing.T) {
	secret := 1234
	guess := 9999

	f := GenerateFeedback(secret, guess)

	assert.Equal(t, 0, f.CorrectPlace)
	assert.Equal(t, 0, f.WrongPlace)
	assert.Equal(t, HintNone, f.Hint)
}
