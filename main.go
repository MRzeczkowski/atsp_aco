package main

import (
	"atsp_aco/parsing"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"sync"
	"time"
)

type ACO struct {
	alpha, beta, evaporation                            float64
	minPheromone, maxPheromone, exploration             float64
	ants, iterations, currentIteration, bestAtIteration int
	distances, pheromone                                [][]float64
	bestLength                                          float64
	bestPath                                            []int
}

func NewACO(alpha, beta, evaporation, exploration float64, ants, iterations int, distances [][]float64) *ACO {
	dimension := len(distances)
	pheromone := make([][]float64, dimension)
	initialPheromone := 1.0
	for i := range pheromone {
		pheromone[i] = make([]float64, dimension)
		for j := range pheromone[i] {
			pheromone[i][j] = initialPheromone
		}
	}

	return &ACO{
		alpha:        alpha,
		beta:         beta,
		evaporation:  evaporation,
		exploration:  exploration,
		ants:         ants,
		iterations:   iterations,
		distances:    distances,
		pheromone:    pheromone,
		bestLength:   math.Inf(1),
		maxPheromone: initialPheromone,
		minPheromone: initialPheromone / (exploration * float64(ants)),
	}
}

func (aco *ACO) Run() {
	for aco.currentIteration = 0; aco.currentIteration < aco.iterations; aco.currentIteration++ {
		paths := make([][]int, aco.ants)
		lengths := make([]float64, aco.ants)

		var wg sync.WaitGroup
		wg.Add(aco.ants)
		for i := 0; i < aco.ants; i++ {
			go func(i int) {
				paths[i], lengths[i] = aco.constructPath(i)
				wg.Done()
			}(i)
		}
		wg.Wait()

		aco.updatePheromoneLevels()
		aco.updatePheromone(paths, lengths)
	}
}

func (aco *ACO) updatePheromoneLevels() {
	aco.maxPheromone = 1.0 / ((1 - aco.evaporation) * aco.bestLength)
	aco.minPheromone = aco.maxPheromone / (aco.exploration * float64(aco.ants))
}

func (aco *ACO) constructPath(antNumber int) ([]int, float64) {
	dimension := len(aco.distances)
	path := make([]int, dimension)
	visited := make([]bool, dimension)
	current := antNumber % dimension
	path[0] = current
	visited[current] = true

	for i := 1; i < dimension; i++ {
		next := aco.selectNextCity(current, visited)
		if next == -1 {
			break
		}
		path[i] = next
		visited[next] = true
		current = next
	}

	length := aco.pathLength(path)
	if length < aco.bestLength {
		aco.bestLength = length
		aco.bestPath = append([]int(nil), path...)
		aco.bestAtIteration = aco.currentIteration
	}

	return path, length
}

func pow(base, exp float64) float64 {
	switch exp {
	case 0.25:
		return math.Sqrt(math.Sqrt(base))
	case 0.5:
		return math.Sqrt(base)
	case 0.75:
		return math.Sqrt(base) * math.Sqrt(math.Sqrt(base))
	case 1:
		return base
	case 1.25:
		return base * math.Sqrt(math.Sqrt(base))
	case 1.5:
		return base * math.Sqrt(base)
	case 1.75:
		return base * math.Sqrt(base) * math.Sqrt(math.Sqrt(base))
	case 2:
		return base * base
	case 2.25:
		return base * base * math.Sqrt(math.Sqrt(base))
	case 2.5:
		return base * base * math.Sqrt(base)
	case 2.75:
		return base * base * math.Sqrt(base) * math.Sqrt(math.Sqrt(base))
	case 3:
		return base * base * base
	case 3.25:
		return base * base * base * math.Sqrt(math.Sqrt(base))
	case 3.5:
		return base * base * base * math.Sqrt(base)
	case 3.75:
		return base * base * base * math.Sqrt(base) * math.Sqrt(math.Sqrt(base))
	case 4:
		return base * base * base * base
	case 4.25:
		return base * base * base * base * math.Sqrt(math.Sqrt(base))
	case 4.5:
		return base * base * base * base * math.Sqrt(base)
	case 4.75:
		return base * base * base * base * math.Sqrt(base) * math.Sqrt(math.Sqrt(base))
	case 5:
		return base * base * base * base * base
	default:
		return math.Pow(base, exp)
	}
}

func (aco *ACO) selectNextCity(current int, visited []bool) int {
	dimension := len(aco.distances)
	probabilities := make([]float64, dimension)
	total := 0.0

	for i := 0; i < dimension; i++ {
		if !visited[i] && probabilities[i] == 0 {
			pheromone := pow(aco.pheromone[current][i], aco.alpha)
			invDistance := 1.0 / float64(aco.distances[current][i])
			desirability := pow(invDistance, aco.beta)
			probabilities[i] = pheromone * desirability
			total += probabilities[i]
		}
	}

	r := rand.Float64()
	for i, cumulativeProbability := 0, 0.0; i < dimension; i++ {
		if !visited[i] && probabilities[i] > 0.0 {
			probabilities[i] /= total
			cumulativeProbability += probabilities[i]
			if r < cumulativeProbability || math.IsNaN(probabilities[i]) {
				return i
			}
		}
	}

	return -1
}

func (aco *ACO) updatePheromone(paths [][]int, lengths []float64) {

	bestIdx := 0
	for i := 1; i < len(lengths); i++ {
		if lengths[i] < lengths[bestIdx] {
			bestIdx = i
		}
	}

	for i := range aco.pheromone {
		for j := range aco.pheromone[i] {
			aco.pheromone[i][j] *= (1 - aco.evaporation)
			aco.pheromone[i][j] = math.Max(aco.pheromone[i][j], aco.minPheromone)
		}
	}

	path := paths[bestIdx]
	delta := 1.0 / lengths[bestIdx]
	for i := 0; i < len(path)-1; i++ {
		start, end := path[i], path[i+1]
		aco.pheromone[start][end] += delta
		aco.pheromone[start][end] = math.Min(aco.pheromone[start][end], aco.maxPheromone)
	}

	if len(path) > 0 {
		last, first := path[len(path)-1], path[0]
		aco.pheromone[last][first] += delta
		aco.pheromone[last][first] = math.Min(aco.pheromone[last][first], aco.maxPheromone)
	}
}

func (aco *ACO) pathLength(path []int) float64 {
	sum := 0.0
	p := len(path)

	for i := 0; i < p-1; i++ {
		start, end := path[i], path[i+1]
		sum += float64(aco.distances[start][end])
	}

	if p > 0 {
		last, first := path[p-1], path[0]
		sum += float64(aco.distances[last][first])
	}

	return sum
}

func startProfiling() {
	f, err := os.Create("aco.prof")
	if err != nil {
		fmt.Println(err)
		return
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
}

func generateRange(start, end, step float64) []float64 {
	var rangeSlice []float64

	for i := start; i <= end; i += step {
		rangeSlice = append(rangeSlice, i)
	}

	return rangeSlice
}

var bestParams struct {
	alpha, beta, evaporation, exploration float64
	averageLength                         float64
	bestLength                            float64
	bestPath                              []int
	deviation                             float64
	successRate                           float64
}

var optimalSolutions = map[string]float64{
	"br17":   39,
	"ft53":   6905,
	"ft70":   38673,
	"ftv33":  1286,
	"ftv35":  1473,
	"ftv38":  1530,
	"ftv44":  1613,
	"ftv47":  1776,
	"ftv55":  1608,
	"ftv64":  1839,
	"ftv70":  1950,
	"ftv170": 2755,
	"p43":    5620,
	"rbg323": 1326,
	"rbg358": 1163,
	"rbg403": 2465,
	"rbg443": 2720,
	"ry48p":  14422,
}

func runExperiment(file string, numRuns int, alpha, beta, evaporation, exploration float64) {

	name, dimension, matrix, err := parsing.ParseTSPLIBFile(file)
	if err != nil {
		fmt.Println("Error parsing file:", file, err)
		return
	}

	var iterations int

	if dimension < 50 {
		iterations = 100
	}

	if 50 <= dimension && dimension < 100 {
		iterations = 500
	}

	if dimension >= 100 {
		iterations = 1000
	}

	var totalBestLength float64
	var totalElapsedTime time.Duration

	bestLength := math.MaxFloat64
	var bestPath []int
	successCounter := 0.0
	bestAtIteration := 0

	knownOptimal := optimalSolutions[name]

	ants := dimension

	for i := 0; i < numRuns; i++ {
		aco := NewACO(alpha, beta, evaporation, exploration, ants, iterations, matrix)
		start := time.Now()
		aco.Run()
		elapsed := time.Since(start)

		totalBestLength += aco.bestLength
		totalElapsedTime += elapsed

		if aco.bestLength < bestLength {
			bestLength = aco.bestLength
			bestPath = aco.bestPath
			bestAtIteration = aco.bestAtIteration
		}

		if aco.bestLength == knownOptimal {
			successCounter++
		}
	}

	averageBestLength := totalBestLength / float64(numRuns)
	deviation := 100 * (averageBestLength - knownOptimal) / knownOptimal
	successRate := 100 * successCounter / float64(numRuns)

	if bestParams.averageLength == 0 || averageBestLength < bestParams.averageLength {
		bestParams = struct {
			alpha, beta, evaporation, exploration float64
			averageLength                         float64
			bestLength                            float64
			bestPath                              []int
			deviation                             float64
			successRate                           float64
		}{alpha, beta, evaporation, exploration, averageBestLength, bestLength, bestPath, deviation, successRate}
	}

	fmt.Printf("| %s | %.2f | %.2f | %.2f | %.2f | %d | %d | %.0f | %.0f | %d | %.0f | %.2f | %.2f |\n",
		name, alpha, beta, evaporation, exploration, ants, iterations, averageBestLength, bestLength, bestAtIteration, knownOptimal, deviation, successRate)
}

func main() {
	dir := "tsp_files"
	files, err := filepath.Glob(filepath.Join(dir, "*.atsp"))
	if err != nil {
		fmt.Println("Error finding files:", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No files found in the directory.")
		return
	}

	numRuns := 10

	for _, file := range files {

		if !strings.Contains(file, "ftv170") {
			continue
		}

		fmt.Println("| Instance | Alpha | Beta | Evaporation | Exploration | Ants | Iterations | Average Result | Best found | Best found at iteration | Known Optimal | Deviation (%) | Success rate (%) |")
		fmt.Println("|-|-|-|-|-|-|-|-|-|-|-|-|-|")

		for _, alpha := range generateRange(0.75, 1.25, 0.25) {
			for _, beta := range generateRange(3.0, 5.0, 1.0) {
				for _, evaporation := range generateRange(0.5, 0.8, 0.1) {
					for _, exploration := range generateRange(8.0, 10.0, 1.0) {
						runExperiment(file, numRuns, alpha, beta, evaporation, exploration)
					}
				}
			}
		}

		fmt.Printf("\nBest parameters:")
		fmt.Printf("\n - Alpha: %.2f", bestParams.alpha)
		fmt.Printf("\n - Beta: %.2f", bestParams.beta)
		fmt.Printf("\n - Evaporation: %.2f", bestParams.evaporation)
		fmt.Printf("\n - Exploration: %.2f", bestParams.exploration)
		fmt.Printf("\n - Average length: %.0f", bestParams.averageLength)
		fmt.Printf("\n - Deviation: %.2f", bestParams.deviation)
		fmt.Printf("\n - Success rate: %.2f", bestParams.successRate)

		fmt.Println()

		bestParams.averageLength = 0.0
	}
}
