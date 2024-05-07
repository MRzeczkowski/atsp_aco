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
	"time"

	"golang.org/x/exp/maps"
)

type ACO struct {
	alpha, beta, evaporation, q        float64
	ants, iterations, currentIteration int
	distances, pheromone, mst          [][]float64
	bestLength                         float64
	bestPath                           []int
}

func NewACO(alpha, beta, evaporation, q float64, ants, iterations int, distances [][]float64) *ACO {
	dimension := len(distances)
	pheromone := make([][]float64, dimension)
	for i := range pheromone {
		pheromone[i] = make([]float64, dimension)
	}

	return &ACO{
		alpha:       alpha,
		beta:        beta,
		evaporation: evaporation,
		q:           q,
		ants:        ants,
		iterations:  iterations,
		distances:   distances,
		pheromone:   pheromone,
		bestLength:  math.Inf(1),
	}
}

func isSpanningTree(mst [][]float64) bool {
	size := len(mst)
	edges := 0
	for i := 0; i < size; i++ {
		for j := i + 1; j < size; j++ {
			if mst[i][j] >= 1 {
				edges++
			}
		}
	}
	return edges == size-1
}

func totalEdges(mst [][]float64) int {
	size := len(mst)
	edges := 0
	for i := 0; i < size; i++ {
		for j := i + 1; j < size; j++ {
			edges += int(mst[i][j])
		}
	}
	return edges
}

func totalWeight(mst [][]float64, distances [][]float64) float64 {
	size := len(mst)
	weight := 0.0
	for i := 0; i < size; i++ {
		for j := i + 1; j < size; j++ {
			if mst[i][j] == 1 {
				weight += distances[i][j]
			}
		}
	}
	return weight
}

func validateMST(mst [][]float64, distances [][]float64) bool {
	size := len(mst)
	if !isSpanningTree(mst) {
		fmt.Println("The generated MST is not a spanning tree.")
		return false
	}

	expectedEdges := size - 1
	actualEdges := totalEdges(mst)
	if actualEdges != expectedEdges {
		fmt.Printf("The MST has %d edges instead of %d.\n", actualEdges, expectedEdges)
		return false
	}

	expectedWeight := totalWeight(mst, distances)
	if expectedWeight != 0 {
		fmt.Println("The total weight of the MST is", expectedWeight)
	}

	return true
}

func (aco *ACO) Run() {

	// https://ieeexplore.ieee.org/document/5522700
	//aco.mst = aco.constructMST()

	// Convert the matrix to the required format
	vertices, edges, weights := convertToEdges(aco.distances)

	// Compute the MSA
	msa := findMSA(vertices, edges, 0, weights)

	// Convert the result back to an adjacency matrix
	aco.mst = convertToMatrix(msa, len(aco.distances))

	for aco.currentIteration = 0; aco.currentIteration < aco.iterations; aco.currentIteration++ {
		paths := make([][]int, aco.ants)
		lengths := make([]float64, aco.ants)

		for i := 0; i < aco.ants; i++ {
			paths[i], lengths[i] = aco.constructPath(i)
		}

		aco.updatePheromone(paths, lengths)
	}
}

func (aco *ACO) constructMST() [][]float64 {
	dimension := len(aco.distances)
	parent := make([]int, dimension)
	keys := make([]float64, dimension)
	mstSet := make([]bool, dimension)

	for i := 0; i < dimension; i++ {
		keys[i] = math.MaxFloat64
	}

	keys[0] = 0
	parent[0] = -1

	for count := 0; count < dimension-1; count++ {
		u := minKey(keys, mstSet)
		mstSet[u] = true

		for v := 0; v < dimension; v++ {
			if !mstSet[v] && aco.distances[u][v] < keys[v] {
				parent[v] = u
				keys[v] = aco.distances[u][v]
			}
		}
	}

	mst := make([][]float64, dimension)
	for i := range mst {
		mst[i] = make([]float64, dimension)
	}

	for i := 1; i < dimension; i++ {
		mst[parent[i]][i] = 1
		mst[i][parent[i]] = 1
	}

	return mst
}

func minKey(keys []float64, mstSet []bool) int {
	min := math.MaxFloat64
	minIndex := -1

	for v := 0; v < len(keys); v++ {
		if !mstSet[v] && keys[v] < min {
			min = keys[v]
			minIndex = v
		}
	}

	return minIndex
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
			// This should not happen, not sure how handle it now if at all.
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
		//fmt.Printf("Iteration:%d; Ant:%d; %.0f;\n", aco.currentIteration, antNumber, aco.bestLength)
	}

	return path, length
}

func pow(base, exp float64) float64 {
	switch exp {
	case 0.25:
		return math.Sqrt(math.Sqrt(base)) // Fourth root
	case 0.5:
		return math.Sqrt(base) // Square root
	case 0.75:
		return math.Sqrt(base) * math.Sqrt(math.Sqrt(base)) // Square root of square root times square root
	case 1:
		return base
	case 1.25:
		return base * math.Sqrt(math.Sqrt(base)) // Base times fourth root
	case 1.5:
		return base * math.Sqrt(base) // Base times square root
	case 1.75:
		return base * math.Sqrt(base) * math.Sqrt(math.Sqrt(base)) // Base times square root of square root times square root
	case 2:
		return base * base // Square
	case 2.25:
		return base * base * math.Sqrt(math.Sqrt(base)) // Square times fourth root
	case 2.5:
		return base * base * math.Sqrt(base) // Square times square root
	case 2.75:
		return base * base * math.Sqrt(base) * math.Sqrt(math.Sqrt(base)) // Square times square root of square root times square root
	case 3:
		return base * base * base // Cube
	case 3.25:
		return base * base * base * math.Sqrt(math.Sqrt(base)) // Cube times fourth root
	case 3.5:
		return base * base * base * math.Sqrt(base) // Cube times square root
	case 3.75:
		return base * base * base * math.Sqrt(base) * math.Sqrt(math.Sqrt(base)) // Cube times square root of square root times square root
	case 4:
		return base * base * base * base // Fourth power
	case 4.25:
		return base * base * base * base * math.Sqrt(math.Sqrt(base)) // Fourth power times fourth root
	case 4.5:
		return base * base * base * base * math.Sqrt(base) // Fourth power times square root
	case 4.75:
		return base * base * base * base * math.Sqrt(base) * math.Sqrt(math.Sqrt(base)) // Fourth power times square root of square root times square root
	case 5:
		return base * base * base * base * base // Fifth power
	default:
		return math.Pow(base, exp) // Fallback for other exponents
	}
}

func (aco *ACO) selectNextCity(current int, visited []bool) int {
	dimension := len(aco.distances)
	probabilities := make([]float64, dimension)
	total := 0.0

	// This should make ants use better paths in the beginning.
	// https://ieeexplore.ieee.org/document/5522700
	adaptiveMstProbability := 0.5 * (1.0 - float64(aco.currentIteration)/float64(aco.iterations))
	if rand.Float64() < adaptiveMstProbability {
		for i := 0; i < dimension; i++ {
			if aco.mst[current][i] == 1 && !visited[i] {
				return i
			}
		}
	}

	for i := 0; i < dimension; i++ {
		if !visited[i] {

			// https://ieeexplore.ieee.org/document/6972311
			if aco.pheromone[current][i] == 0 {
				sum := 0.0
				for j := 0; j < dimension; j++ {
					if j != current {
						sum += aco.distances[current][j]
					}
				}

				aco.pheromone[current][i] = 1 / sum
			}

			pheromone := pow(aco.pheromone[current][i], aco.alpha)
			invDistance := 1.0 / float64(aco.distances[current][i])
			desirability := pow(invDistance, aco.beta)
			probabilities[i] = pheromone * desirability
			total += probabilities[i]
		}
	}

	for i := 0; i < dimension; i++ {
		if !visited[i] {
			probabilities[i] /= total
		}
	}

	r := rand.Float64()
	for i, cumulativeProbability := 0, 0.0; i < dimension; i++ {
		if !visited[i] {
			cumulativeProbability += probabilities[i]
			if r < cumulativeProbability || math.IsNaN(probabilities[i]) {
				return i
			}
		}
	}

	return -1 // Fallback, should not happen
}

func (aco *ACO) updatePheromone(paths [][]int, lengths []float64) {

	// Evaporate pheromone first
	for i := range aco.pheromone {
		for j := range aco.pheromone[i] {
			aco.pheromone[i][j] *= (1 - aco.evaporation)
		}
	}

	for i, path := range paths {
		p := len(path)

		delta := aco.q / lengths[i]

		for j := 0; j < p-1; j++ {
			start, end := path[j], path[j+1]
			aco.pheromone[start][end] += delta
		}

		// Handle the wrap-around from the last to the first node separately
		if p > 0 {
			last, first := path[p-1], path[0]
			aco.pheromone[last][first] += delta
		}
	}
}

func (aco *ACO) pathLength(path []int) float64 {
	sum := 0.0
	p := len(path)

	for i := 0; i < p-1; i++ {
		start, end := path[i], path[i+1]
		sum += float64(aco.distances[start][end])
	}

	// Handle the wrap-around from the last node back to the first node
	if p > 0 {
		last, first := path[p-1], path[0]
		sum += float64(aco.distances[last][first])
	}

	return sum
}

func startProfiling() {
	// Start profiling
	f, err := os.Create("aco.prof")
	if err != nil {

		fmt.Println(err)
		return
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
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
			if matrix[i][j] != 0 { // Assuming non-zero entries are valid edges
				edge := Edge{from: i, to: j}
				edges = append(edges, edge)
				weights[edge] = matrix[i][j]
			}
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

func main() {

	// Define the directory containing the TSP files
	dir := "tsp_files"

	// Use filepath.Glob to find all .atsp files
	files, err := filepath.Glob(filepath.Join(dir, "*.atsp"))
	if err != nil {
		fmt.Println("Error finding files:", err)
		return
	}

	// Check if there are any files to process
	if len(files) == 0 {
		fmt.Println("No files found in the directory.")
		return
	}

	optimalSolutions := map[string]float64{
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

	fmt.Println("| Name | Iterations | Dimension | Ants | Found Result | Known Optimal | Deviation (%) | Time (ms) |")
	fmt.Println("|-|-|-|-|-|-|-|")

	// Process each file
	for _, file := range files {

		if strings.Contains(file, "rbg443") {
			name, dimension, matrix, err := parsing.ParseTSPLIBFile(file)
			if err != nil {
				fmt.Println("Error parsing file:", file, err)
				continue
			}

			vertices, edges, weights := convertToEdges(matrix)

			for i := 0; i < len(matrix); i++ {
				msa := findMSA(vertices, edges, i, weights)

				sum := 0.0
				for j := 0; j < len(msa); j++ {
					x := msa[j].from
					y := msa[j].to

					sum += matrix[x][y]
				}

				fmt.Printf("%s MSA(%d) = %f\n", name, i, sum)
			}

			return

			// Parameters set in accordance to these articles:
			// https://ieeexplore.ieee.org/document/8820263
			// https://ieeexplore.ieee.org/document/5522700
			alpha := 0.25
			beta := 5.0
			evaporation := 0.95
			q := 1.0
			ants := dimension //int(math.Ceil(float64(dimension) / 1.5))
			var iterations int

			if dimension <= 50 {
				iterations = 100
			}
			if 50 < dimension && dimension <= 100 {
				iterations = 500
			}
			if 100 < dimension {
				iterations = 5000
			}

			aco := NewACO(alpha, beta, evaporation, q, ants, iterations, matrix)
			start := time.Now()
			aco.Run()
			elapsed := time.Since(start)

			knownOptimal := optimalSolutions[name]

			deviation := 100 * (aco.bestLength - knownOptimal) / knownOptimal

			fmt.Printf("| %s | %d | %d | %d | %.0f | %.0f | %.2f | %v |\n", name, iterations, dimension, ants, aco.bestLength, knownOptimal, deviation, elapsed.Milliseconds())
			//fmt.Println(aco.bestPath)
		}
	}
}
