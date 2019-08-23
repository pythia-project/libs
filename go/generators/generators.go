// Pythia utilities for unit testing-based tasks
// Author: Sébastien Combéfis <sebastien@combefis.be>
//
// Copyright (C) 2019, Computer Science and IT in Education ASBL
// Copyright (C) 2019, ECAM Brussels Engineering School
//
// This program is free software: you can redistribute it and/or modify
// under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 2 of the License, or
//  (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package generators

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

type RandomGenerator interface {
	Generate() string
}

type ArrayGenerator struct {
	Generators []RandomGenerator
}

func (g ArrayGenerator) Generate() string {
	result := make([]string, len(g.Generators))
	for i, generator := range g.Generators {
		result[i] = generator.Generate()
	}
	fmt.Println(result)
	return ""
}

////////////////////////////////////////////////////////////////////////////////
// int

type IntRandomGenerator struct {
	Min int
	Max int
}

// Generates a random integer number comprised between two bounds.
func (g IntRandomGenerator) Generate() string {
	return fmt.Sprintf("%d", randint(g.Min, g.Max))
}

func randint(min int, max int) int {
	return min + (int(rand.Int63()) % (max - min + 1))
}

////////////////////////////////////////////////////////////////////////////////
// bool

type BoolRandomGenerator struct {
}

// Generates a random boolean value.
func (g BoolRandomGenerator) Generate() string {
	if rand.Intn(2) == 0 {
		return "true"
	}
	return "false"
}

////////////////////////////////////////////////////////////////////////////////
// float

type FloatRandomGenerator struct {
	Min float64
	Max float64
}

// Generates a random floating-point number comprised between two bounds.
func (g FloatRandomGenerator) Generate() string {
	return fmt.Sprintf("%f", g.Min+(rand.Float64()*(g.Max-g.Min)))
}

////////////////////////////////////////////////////////////////////////////////
// string

type StringRandomGenerator struct {
	MinLength int
	MaxLength int
}

// Generates a random string with a random number of characters comprised between two bounds.
func (g StringRandomGenerator) Generate() string {
	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := randint(g.MinLength, g.MaxLength)

	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteString(string(alphabet[randint(0, len(alphabet)-1)]))
	}
	return sb.String()
}

////////////////////////////////////////////////////////////////////////////////
// enum

type EnumRandomGenerator struct {
	Values []string
}

// Generates a random value from an enumeration.
func (g EnumRandomGenerator) Generate() string {
	return g.Values[randint(0, len(g.Values)-1)]
}

////////////////////////////////////////////////////////////////////////////////
// Utility functions

const (
	intPattern   = `-{0,1}[1-9][0-9]*`
	floatPattern = `-{0,1}[1-9][0-9]*(?:\.[0-9]*[1-9]){0,1}`
)

func buildGenerator(desc string) RandomGenerator {
	var regex *regexp.Regexp

	// int(min,max)
	regex, _ = regexp.Compile(fmt.Sprintf(`^int\((%[1]s),(%[1]s)\)$`, intPattern))
	if matches := regex.FindStringSubmatch(desc); matches != nil {
		min, _ := strconv.Atoi(matches[1])
		max, _ := strconv.Atoi(matches[2])
		return IntRandomGenerator{min, max}
	}

	// bool
	if desc == "bool" {
		return BoolRandomGenerator{}
	}

	// float(min,max)
	regex, _ = regexp.Compile(fmt.Sprintf(`^float\((%[1]s),(%[1]s)\)$`, floatPattern))
	if matches := regex.FindStringSubmatch(desc); matches != nil {
		min, _ := strconv.ParseFloat(matches[1], 64)
		max, _ := strconv.ParseFloat(matches[2], 64)
		return FloatRandomGenerator{min, max}
	}

	// str(min,max)
	regex, _ = regexp.Compile(fmt.Sprintf(`^str\((%[1]s),(%[1]s)\)$`, intPattern))
	if matches := regex.FindStringSubmatch(desc); matches != nil {
		minLength, _ := strconv.Atoi(matches[1])
		maxLength, _ := strconv.Atoi(matches[2])
		return StringRandomGenerator{minLength, maxLength}
	}

	// enum(list)
	regex, _ = regexp.Compile(`^enum\((.+)\)$`)
	if matches := regex.FindStringSubmatch(desc); matches != nil {
		return EnumRandomGenerator{strings.Split(matches[1], ",")}
	}

	return nil
}

func BuildGenerators(descs ...string) []RandomGenerator {
	generators := make([]RandomGenerator, len(descs))
	for i, desc := range descs {
		generators[i] = buildGenerator(desc)
	}

	return generators
}
