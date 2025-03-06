package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"
)

type Attempt struct {
	openedBoxes []int
	success     bool
}

type Game struct {
	boxes            []int
	studentsAttempts []Attempt
	strategy         func(g *Game) bool
	random           *rand.Rand
}

func NewGame(random *rand.Rand, Strategy func(g *Game) bool) *Game {
	g := Game{
		boxes:            randSlice(50, random),
		studentsAttempts: make([]Attempt, 50),
		random:           random,
		strategy:         Strategy,
	}

	return &g
}

func (g *Game) Start() bool {
	return g.strategy(g)
}

func NoContract(g *Game) (win bool) {
	var winsCount int

	for studentIndex := range g.studentsAttempts {
		for _, openingBoxIndex := range randSlice(25, g.random) {
			g.studentsAttempts[studentIndex].openedBoxes = append(
				g.studentsAttempts[studentIndex].openedBoxes,
				openingBoxIndex,
			)

			if g.boxes[openingBoxIndex] == studentIndex {
				winsCount += 1
				g.studentsAttempts[studentIndex].success = true
				break
			}
		}
	}

	return winsCount == 50
}

func WithContract(g *Game) (win bool) {
	var winsCount int

	for studentIndex := range g.studentsAttempts {
		openingBoxIndex := studentIndex
		for range 25 {
			g.studentsAttempts[studentIndex].openedBoxes = append(
				g.studentsAttempts[studentIndex].openedBoxes,
				openingBoxIndex,
			)

			if g.boxes[openingBoxIndex] == studentIndex {
				winsCount += 1
				g.studentsAttempts[studentIndex].success = true
				break
			}

			openingBoxIndex = g.boxes[openingBoxIndex]
		}
	}

	return winsCount == 50
}

func main() {
	t := time.Now()
	var noContractWins atomic.Int64
	var withContractWins atomic.Int64
	gamesCount := 1_000_000

	var wg sync.WaitGroup
	sem := make(chan struct{}, 100)

	for i := range gamesCount {
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			src := rand.NewPCG(uint64(i+1), uint64(i))
			random := rand.New(src)

			game := NewGame(random, NoContract)
			if game.Start() {
				noContractWins.Add(1)
			}

			game = NewGame(random, WithContract)
			if game.Start() {
				withContractWins.Add(1)
			}
		}()
	}

	wg.Wait()

	fmt.Printf("wins with NO contact: %d\n", noContractWins.Load())
	fmt.Printf("probability of win with NO contract: %f\n", float64(noContractWins.Load())/float64(gamesCount))

	fmt.Printf("wins with contact: %d\n", withContractWins.Load())
	fmt.Printf("probability of win with contract: %f\n", float64(withContractWins.Load())/float64(gamesCount))

	fmt.Printf("takes time: %v\n", time.Since(t))
}

// randSlice returns shuffled slice with numbers in range [0, length)
func randSlice(length int, random *rand.Rand) []int {
	slice := make([]int, length)

	for i := range length {
		slice[i] = i
	}

	random.Shuffle(length, func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})

	return slice
}
