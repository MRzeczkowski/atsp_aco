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

	"golang.org/x/exp/maps"
)

type ACO struct {
	alpha, beta, evaporation                float64
	minPheromone, maxPheromone, exploration float64
	ants, iterations, currentIteration      int
	distances, pheromone                    [][]float64
	msas                                    [][][]float64
	bestLength                              float64
	bestPath                                []int
}

// NewACO initializes a new ACO instance with initial pheromone levels set to an estimated best value
func NewACO(alpha, beta, evaporation, exploration float64, ants, iterations int, distances [][]float64, msas [][][]float64) *ACO {
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
		msas:         msas,
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
		next := aco.selectNextCity(antNumber, current, visited)
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

func (aco *ACO) selectNextCity(antNumber, current int, visited []bool) int {
	dimension := len(aco.distances)
	probabilities := make([]float64, dimension)
	total := 0.0

	// This should make ants use better paths in the beginning.
	// https://ieeexplore.ieee.org/document/5522700
	// adaptiveMstProbability := 0.5 * (1.0 - float64(aco.currentIteration)/float64(aco.iterations))
	// if rand.Float64() < adaptiveMstProbability {
	// 	for i := 0; i < dimension; i++ {
	// 		if aco.msas[antNumber][current][i] == 1 && !visited[i] {
	// 			return i
	// 		}
	// 	}
	// }

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
	averageLength                         float64
	bestLength                            float64
	bestPath                              []int
	deviation                             float64
	successRate                           float64
}

var optimalSolutions = map[string]float64{
	"br17":   39,
	"ft53":   6905, // [49,52,50,48,29,28,25,27,26,3,13,11,10,12,14,41,47,42,46,43,45,44,34,32,33,31,30,0,4,2,17,16,15,37,39,38,36,35,40,21,20,24,23,22,19,18,1,8,9,7,6,5,51]
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

func findMSA(V []int, E []Edge, r int, w map[Edge]float64) []Edge {
	// Step 1: Removing all edges leading back to the root and adjusting edge set
	var _E []Edge
	_w := make(map[Edge]float64, len(w))
	for _, e := range E {
		if e.to != r {
			_E = append(_E, e)
			_w[e] = w[e]
		}
	}

	// Step 2: Finding minimum incoming edge for each vertex
	pi := make(map[int]int)
	for _, v := range V {
		if v == r {
			continue
		}
		minCost := math.MaxFloat64
		for _, e := range _E {
			if e.to == v && _w[e] < minCost {
				minCost = _w[e]
				pi[v] = e.from
			}
		}
	}

	// Step 3: Finding cycles
	cycleVertex := -1
	var visited map[int]bool
	for _, v := range V {
		if cycleVertex != -1 {
			break
		}

		visited = make(map[int]bool)
		next_v, ok := pi[v]

		for ok {
			if visited[next_v] {
				cycleVertex = next_v
				break
			}

			visited[next_v] = true
			next_v, ok = pi[next_v]
		}
	}

	var result []Edge

	// Step 4: No cycle
	if cycleVertex == -1 {
		for v, u := range pi {
			result = append(result, Edge{u, v})
		}

		return result
	}

	// Step 5: Handle cycle
	cycle := make(map[int]bool)
	cycle[cycleVertex] = true
	next_v := pi[cycleVertex]
	for next_v != cycleVertex {
		cycle[next_v] = true
		next_v = pi[next_v]
	}

	// Step 6: Contract the cycle into a new node v_c
	v_c := -(cycleVertex * cycleVertex) // Unique negative squared identifier
	V_prime := []int{}
	for _, v := range V {
		if !cycle[v] {
			V_prime = append(V_prime, v)
		}
	}

	V_prime = append(V_prime, v_c)
	E_prime := make(map[Edge]bool)
	w_prime := make(map[Edge]float64)
	correspondence := make(map[Edge]Edge)

	for _, uv := range _E {
		u := uv.from
		v := uv.to

		if !cycle[u] && cycle[v] {
			e := Edge{u, v_c}
			tmpEdge := Edge{pi[v], v}
			if E_prime[e] {
				if w_prime[e] < _w[uv]-_w[tmpEdge] {
					continue
				}
			}

			w_prime[e] = _w[uv] - _w[tmpEdge]
			correspondence[e] = uv
			E_prime[e] = true
		} else if cycle[u] && !cycle[v] {
			e := Edge{v_c, v}
			if E_prime[e] {
				old_u := correspondence[e].from

				tmpEdge := Edge{old_u, v}
				if _w[tmpEdge] < _w[uv] {
					continue
				}
			}

			E_prime[e] = true
			w_prime[e] = _w[uv]
			correspondence[e] = uv
		} else if !cycle[u] && !cycle[v] {
			e := uv
			E_prime[e] = true
			w_prime[e] = _w[uv]
			correspondence[e] = uv
		}
	}

	// Recursive call
	tree := findMSA(V_prime, maps.Keys(E_prime), r, w_prime)

	// Step 8: Expanding back
	var cycle_edge Edge

	for _, e := range tree {
		u := e.from
		v := e.to

		if v == v_c {
			tmpEdge := Edge{u, v_c}
			old_v := correspondence[tmpEdge].to
			cycle_edge = Edge{pi[old_v], old_v}
			break
		}
	}

	resultSet := make(map[Edge]bool)

	for _, uv := range tree {
		resultSet[correspondence[uv]] = true
	}

	for v := range cycle {
		u := pi[v]
		tmpEdge := Edge{u, v}
		resultSet[tmpEdge] = true
	}

	delete(resultSet, cycle_edge)

	result = make([]Edge, 0)
	for e := range resultSet {
		result = append(result, e)
	}

	return result
}

type Edge struct {
	from, to int
}

func convertToEdges(matrix [][]float64) ([]int, []Edge, map[Edge]float64) {
	var vertices []int
	edges := make([]Edge, 0)
	weights := make(map[Edge]float64)

	for i := range matrix {
		vertices = append(vertices, i)
		for j := range matrix[i] {
			edge := Edge{from: i, to: j}
			edges = append(edges, edge)
			weights[edge] = matrix[i][j]
		}
	}
	return vertices, edges, weights
}

func convertToMatrix(edges []Edge, size int) [][]float64 {
	matrix := make([][]float64, size)
	for i := range matrix {
		matrix[i] = make([]float64, size)
	}

	for _, edge := range edges {
		matrix[edge.from][edge.to] = 1 // Use 1 to indicate an edge in the MSA
	}
	return matrix
}

func runExperiment(file string, iterations, numRuns int, alpha, beta, evaporation, exploration float64) {

	name, dimension, matrix, err := parsing.ParseTSPLIBFile(file)
	if err != nil {
		fmt.Println("Error parsing file:", file, err)
		return
	}

	vertices, edges, weights := convertToEdges(matrix)

	msas := make([][][]float64, dimension)

	for i := 0; i < dimension; i++ {
		msa := findMSA(vertices, edges, i, weights)

		msaMatrix := convertToMatrix(msa, dimension)

		msas[i] = msaMatrix
	}

	var totalBestLength float64
	var totalElapsedTime time.Duration

	bestLength := math.MaxFloat64
	var bestPath []int
	successCounter := 0.0

	knownOptimal := optimalSolutions[name]

	ants := dimension

	for i := 0; i < numRuns; i++ {
		aco := NewACO(alpha, beta, evaporation, exploration, ants, iterations, matrix, msas)
		start := time.Now()
		aco.Run()
		elapsed := time.Since(start)

		totalBestLength += aco.bestLength
		totalElapsedTime += elapsed

		if aco.bestLength < bestLength {
			bestLength = aco.bestLength
			bestPath = aco.bestPath
		}

		if aco.bestLength == knownOptimal {
			successCounter++
		}
	}

	averageBestLength := totalBestLength / float64(numRuns)
	averageTime := totalElapsedTime / time.Duration(numRuns)
	deviation := 100 * (averageBestLength - knownOptimal) / knownOptimal
	successRate := successCounter / float64(numRuns)

	bestPathEdges := make([]Edge, len(bestPath))

	bestPath = []int{49, 52, 50, 48, 29, 28, 25, 27, 26, 3, 13, 11, 10, 12, 14, 41, 47, 42, 46, 43, 45, 44, 34, 32, 33, 31, 30, 0, 4, 2, 17, 16, 15, 37, 39, 38, 36, 35, 40, 21, 20, 24, 23, 22, 19, 18, 1, 8, 9, 7, 6, 5, 51}

	for i := 0; i < dimension; i++ {
		bestPathEdges[i] = Edge{from: bestPath[i], to: bestPath[(i+1)%dimension]}
	}

	bestPathMatrix := convertToMatrix(bestPathEdges, dimension)

	commonalityWithMsa := 0.0

	//msaToCompare := msas[bestPath[0]]

	combinedMsas := make([][]float64, dimension)
	for i := range matrix {
		combinedMsas[i] = make([]float64, dimension)
	}

	for _, msa := range msas {
		for i := range dimension {
			for j := range dimension {
				if msa[i][j] == 1 {
					combinedMsas[i][j] = 1
				}
			}
		}
	}

	for i := 0; i < dimension; i++ {
		for j := 0; j < dimension; j++ {
			if bestPathMatrix[i][j] == 1 && combinedMsas[i][j] == 1 {
				commonalityWithMsa++
			}
		}
	}

	commonalityWithMsa /= float64(dimension)

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

	fmt.Printf("| %s | %.2f | %.2f | %.2f | %.2f | %d | %d | %.0f | %.0f | %.0f | %.2f | %.2f | %.2f | %v |\n",
		name, alpha, beta, evaporation, exploration, ants, iterations, averageBestLength, bestLength, knownOptimal, deviation, successRate, commonalityWithMsa, averageTime.Milliseconds())
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

	iterations := 500
	numRuns := 50

	fmt.Println("| Instance | Alpha | Beta | Evaporation | Exploration | Ants | Iterations | Average Result | Best found | Known Optimal | Deviation (%) | Success rate (%) | Commonality with MSA (%) | Time (ms) |")
	fmt.Println("|-|-|-|-|-|-|-|-|-|-|-|")

	for _, file := range files {

		if !strings.Contains(file, "ft53") {
			continue
		}

		for _, alpha := range generateRange(1.0, 1.0, 0.5) {
			for _, beta := range generateRange(5.0, 5.0, 0.5) {
				for _, evaporation := range generateRange(0.8, 0.8, 0.1) {
					for _, exploration := range generateRange(10.0, 10.0, 1.0) {
						runExperiment(file, iterations, numRuns, alpha, beta, evaporation, exploration)
					}
				}
			}
		}
	}

	fmt.Printf("\nBest parameters: Alpha: %.2f, Beta: %.2f, Evaporation: %.2f, Exploration: %.2f, Best average length: %.0f, Deviation: %.2f%%, Success rate: %.2f%%\n",
		bestParams.alpha, bestParams.beta, bestParams.evaporation, bestParams.exploration, bestParams.averageLength, bestParams.deviation, bestParams.successRate)

	fmt.Println("Best path:")
	for _, v := range bestParams.bestPath {
		fmt.Print(v, " ")
	}
}
