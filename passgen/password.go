package passgen

import (
	"errors"
	"math"
	"strings"
)

var (
	lowerCase = "abcdefghijklmnopqrstuvwxyz"                                                                                              // level 1
	upperCase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"                                                                                              // level 2
	digits    = "0123456789"                                                                                                              // level 3
	space     = " "                                                                                                                       // level 4
	brackets  = "(){}[]<>"                                                                                                                // level 5
	points    = ".¿?!¡,;:"                                                                                                                // level 6
	special   = "$%&|/=*#@~€^"                                                                                                            // level 7
	extended  = "ƒ„…†‡ˆ‰Š‹›ŒŽ‘’“”•-+_—˜™šœžŸ¢£¤¥¦§¨©ª«¬®¯°±²³´µ¶·¸¹º»¼½¾ÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖ×ØÙÚÛÜÝÞßàáâãäåæçèéêëìíîïðñòóôõö÷øùúûüýþÿ" // level 8
)

// Password represents a sequence of characters required for access to a computer system.
type Password struct {
	Chars   string
	Length  uint64
	Format  []int
	Include string
}

// Generate generates a random password with the length and format given.
func (p *Password) Generate() (string, error) {
	if p.Length < 1 {
		return "", errors.New("password length must be equal to or higher than 1")
	}

	if len(p.Include) > int(p.Length) {
		return "", errors.New("characters to include exceed the password length")
	}

	// Generate <include length> random numbers
	// These are going to represent a slot in the password
	inclSlots := make([]int, len(p.Include))
	inclChars := []rune(p.Include)

	if len(p.Include) > 0 {
		inclSlots = p.includeSlots(inclSlots)
	}

	pool, err := p.generatePool()
	if err != nil {
		return "", err
	}

	password := make([]rune, p.Length)

	for i := range password {
		var skip bool
		// Compare i and random numbers, if they are equal, a random char from "inclChars"
		// will be appended to the password until "inclChars" is empty
		// In this case, selecting one from the pool will be skipped
		if len(p.Include) != 0 {
			for _, n := range inclSlots {
				if i == n {
					char := randInt(len(inclChars))
					password[i] = inclChars[char]
					// Remove the char used from inclParts
					inclChars = append(inclChars[:char], inclChars[char+1:]...)
					skip = true
				}
			}
		}

		if !skip {
			// Generate a random number to take a character from the pool and put it into the password
			password[i] = pool[randInt(len(pool))]
		}
	}

	p.Chars = string(password)

	return p.Chars, nil
}

// Entropy is a measure of what the password could have been so it does not really relate to the password
// itself, but to the selection process. It takes the length of the password and all the possible characters
// from the pool and returns the entropy bits of the password.
func (p *Password) Entropy() float64 {
	poolLength := p.poolLength()

	pow := math.Pow(float64(poolLength), float64(p.Length))
	entropy := math.Log2(pow)

	// This is a generic way to measure the number of attempts, dictionary attacks, social engineering
	// and other techniques are left out of consideration
	// nAttempts := pow / 2

	return entropy
}

// generatePool takes the format specified by the user and creates the pool to generate a random
// password.
func (p *Password) generatePool() ([]rune, error) {
	var characters []string

	// If the format is not specified, set default value
	if p.Format == nil || p.Format[0] == 0 {
		p.Format = []int{1, 2, 3, 4}
	}

	levels := make(map[int]struct{}, len(p.Format))

	for _, v := range p.Format {
		levels[v] = struct{}{}
	}

	for key := range levels {
		if key > 8 {
			return nil, errors.New("password level must be equal to or lower than 8")
		}

		switch key {
		case 1:
			characters = append(characters, lowerCase)
		case 2:
			characters = append(characters, upperCase)
		case 3:
			characters = append(characters, digits)
		case 4:
			characters = append(characters, space)
		case 5:
			characters = append(characters, brackets)
		case 6:
			characters = append(characters, points)
		case 7:
			characters = append(characters, special)
		case 8:
			characters = append(characters, extended)
		}
	}

	pool := []rune(strings.Join(characters, ""))

	return pool, nil
}

// includeSlots returns an array with the positions that each include character will ocupy
// in the password.
func (p *Password) includeSlots(inclSlots []int) []int {
	// Create an array with the length of the password and make
	// indexes and values the same
	pwPositions := make([]int, p.Length)
	for i := 0; i < len(pwPositions); i++ {
		pwPositions[i] = i
	}

	// Select a slot in the password for each character of "include"
	// For example: password[3] = include char. The 4th element of the password will be a inclChar
	for j := range inclSlots {
		// Take a random element (position) from the array
		pos := randInt(len(pwPositions))
		randN := pwPositions[pos]

		// Remove the position selected from the array
		pwPositions = append(pwPositions[:pos], pwPositions[pos+1:]...)

		// Add the random number to the array
		inclSlots[j] = randN
	}

	return inclSlots
}

// poolLength returns the length of the pool used.
func (p *Password) poolLength() int {
	var l int

	for _, level := range p.Format {
		switch level {
		case 1:
			l += len(lowerCase)
		case 2:
			l += len(upperCase)
		case 3:
			l += len(digits)
		case 4:
			l += len(space)
		case 5:
			l += len(brackets)
		case 6:
			l += len(points)
		case 7:
			l += len(special)
		case 8:
			l += len(extended)
		}
	}

	if len(p.Include) != 0 {
		l += len(p.Include)
	}

	return l
}
