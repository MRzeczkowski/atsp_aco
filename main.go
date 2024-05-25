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
	alpha, beta, evaporation                float64
	minPheromone, maxPheromone, exploration float64
	ants, iterations, currentIteration      int
	distances, pheromone                    [][]float64
	bestLength                              float64
	bestPath                                []int
}

// NewACO initializes a new ACO instance with initial pheromone levels set to an estimated best value
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
		ants:         ants,
		iterations:   iterations,
		distances:    distances,
		pheromone:    pheromone,
		bestLength:   math.Inf(1),
		maxPheromone: initialPheromone,
		minPheromone: initialPheromone / (exploration * float64(ants)),
		exploration:  exploration,
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

		aco.updatePheromoneLevels() // Recalculate pheromone limits based on the new best solution
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
		if !visited[i] {
			pheromone := pow(aco.pheromone[current][i], aco.alpha)
			invDistance := 1.0 / float64(aco.distances[current][i])
			desirability := pow(invDistance, aco.beta)
			probabilities[i] = pheromone * desirability
			total += probabilities[i]
		}
	}

	r := rand.Float64()
	for i, cumulativeProbability := 0, 0.0; i < dimension; i++ {
		if !visited[i] {
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
	// Find the best path of this iteration
	bestIdx := 0
	for i := 1; i < len(lengths); i++ {
		if lengths[i] < lengths[bestIdx] {
			bestIdx = i
		}
	}

	// Evaporate pheromone first
	for i := range aco.pheromone {
		for j := range aco.pheromone[i] {
			aco.pheromone[i][j] *= (1 - aco.evaporation)
			aco.pheromone[i][j] = math.Max(aco.pheromone[i][j], aco.minPheromone) // Enforce minimum pheromone level
		}
	}

	// Strengthen pheromone trail for the best ant's path
	path := paths[bestIdx]
	delta := 1.0 / lengths[bestIdx]
	for i := 0; i < len(path)-1; i++ {
		start, end := path[i], path[i+1]
		aco.pheromone[start][end] += delta
		aco.pheromone[start][end] = math.Min(aco.pheromone[start][end], aco.maxPheromone) // Enforce maximum pheromone level
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
	length                                float64
	deviation                             float64
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

func runExperiment(file string, iterations, numRuns int, alpha, beta, evaporation, exploration float64) {

	name, dimension, matrix, err := parsing.ParseTSPLIBFile(file)
	if err != nil {
		fmt.Println("Error parsing file:", file, err)
		return
	}

	var totalBestLength float64
	var totalElapsedTime time.Duration

	ants := dimension

	for i := 0; i < numRuns; i++ {
		aco := NewACO(alpha, beta, evaporation, exploration, ants, iterations, matrix)
		start := time.Now()
		aco.Run()
		elapsed := time.Since(start)

		totalBestLength += aco.bestLength
		totalElapsedTime += elapsed
	}

	averageBestLength := totalBestLength / float64(numRuns)
	averageTime := totalElapsedTime / time.Duration(numRuns)
	knownOptimal := optimalSolutions[name]
	deviation := 100 * (averageBestLength - knownOptimal) / knownOptimal

	if bestParams.length == 0 || averageBestLength < bestParams.length {
		bestParams = struct {
			alpha, beta, evaporation, exploration float64
			length                                float64
			deviation                             float64
		}{alpha, beta, evaporation, exploration, averageBestLength, deviation}
	}

	fmt.Printf("| %s | %.2f | %.2f | %.2f | %.2f | %d | %d | %.0f | %.0f | %.2f | %v |\n", name, alpha, beta, evaporation, exploration, ants, iterations, averageBestLength, knownOptimal, deviation, averageTime.Milliseconds())
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

	iterations := 1000
	numRuns := 10

	fmt.Println("| Instance | Alpha | Beta | Evaporation | Exploration | Ants | Iterations | Average Result | Known Optimal | Deviation (%) | Time (ms) |")
	fmt.Println("|-|-|-|-|-|-|-|-|-|-|-|")

	for _, file := range files {

		if !strings.Contains(file, "170") {
			continue
		}

		for _, alpha := range generateRange(0.5, 3.0, 0.25) {
			for _, beta := range generateRange(2.0, 5.0, 0.25) {
				for _, evaporation := range generateRange(0.2, 0.8, 0.1) {
					for _, exploration := range generateRange(2.0, 10.0, 1.0) {
						runExperiment(file, iterations, numRuns, alpha, beta, evaporation, exploration)
					}
				}
			}
		}
	}

	//fmt.Printf("\nBest Parameters: Alpha: %.2f, Beta: %.2f, Evaporation: %.2f, Exploration: %.2f, Best Length: %.0f, Deviation: %.2f%%\n", bestParams.alpha, bestParams.beta, bestParams.evaporation, bestParams.exploration, bestParams.length, bestParams.deviation)
}
