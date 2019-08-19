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

type IntRandomGenerator struct {
	Min int64
	Max int64
}

func (g IntRandomGenerator) Generate() string {
	return fmt.Sprintf("%d", g.Min+(rand.Int63()%(g.Max-g.Min+1)))
}

type BoolRandomGenerator struct {
}

func (g BoolRandomGenerator) Generate() string {
	if rand.Intn(2) == 0 {
		return "true"
	}
	return "false"
}

type FloatRandomGenerator struct {
	Min float64
	Max float64
}

func (g FloatRandomGenerator) Generate() string {
	return fmt.Sprintf("%f", g.Min+(rand.Float64()*(g.Max-g.Min)))
}
