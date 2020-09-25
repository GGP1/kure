package passgen

import (
	"math"
	"strings"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	// Some points, brackets and specials modify the length of the password,
	// if they are included, the test will fail despite everything is working as expected
	p := &Password{
		Length:  14,
		Format:  []int{1, 2, 3, 4},
		Include: "kure",
	}

	password, err := p.Generate()
	if err != nil {
		t.Errorf("Test failed, error: %v", err)
	}

	entropy := p.Entropy()

	if len(password) != 14 {
		t.Errorf("Wrong password length, expected 14 characters, got %d", len(password))
	}

	exclude := strings.Join([]string{brackets, points, extended}, "")

	if strings.ContainsAny(password, exclude) {
		t.Error("Test failed, password included unspecified levels")
	}

	if !strings.ContainsAny(password, p.Include) {
		t.Error("Test failed, include chars weren't added to the password")
	}

	expected := 83.68

	if math.Floor(entropy) != math.Floor(expected) {
		t.Errorf("Calculate entropy failed, expected: %f, got: %f", expected, entropy)
	}
}

func TestGeneratePassphrase(t *testing.T) {
	p := &Passphrase{
		Length:    10,
		Separator: "/",
	}

	passphrase, _ := p.Generate()
	entropy := p.Entropy()

	// min word length: 3 - max word length: 12
	phraseLength := int(p.Length) + len(p.Separator)
	min := 3 * phraseLength
	max := 12 * phraseLength

	if len(passphrase) < min || len(passphrase) > max {
		t.Errorf("Wrong passphrase length, minimum expected: %d, maximum expected: %d, got: %d", min, max, len(passphrase))
	}

	if !strings.ContainsAny(passphrase, p.Separator) {
		t.Errorf("Passphrase does not include the separator (%s) as expected", p.Separator)
	}

	pool := len(vowels) + len(constants) + len(p.Separator)
	pow := math.Pow(float64(pool), float64(len(passphrase)))

	expected := math.Log2(pow)

	if math.Floor(entropy) != math.Floor(expected) {
		t.Errorf("Calculate entropy failed, expected: %f, got: %f", expected, entropy)
	}
}

func TestCalculateEntropy(t *testing.T) {
	testCases := []struct {
		poolLength int
		length     uint64
		expected   float64
	}{
		{40, 10, 53},
		{82, 20, 127},
		{25, 17, 79},
		{52, 12, 68},
		{100, 22, 146},
		{132, 30, 211},
	}

	for _, tC := range testCases {
		pow := math.Pow(float64(tC.poolLength), float64(tC.length))
		entropy := math.Log2(pow)

		got := math.Round(entropy)

		if got != tC.expected {
			t.Errorf("Test failed, expected: %f, got: %f", tC.expected, got)
		}
	}
}
