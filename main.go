package main

import (
	"atsp_aco/parsing"
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"time"
)

type ACO struct {
	alpha       float64
	beta        float64
	evaporation float64
	ants        int
	iterations  int
	pheromone   [][]float64
	distances   [][]int
	bestLength  float64
	bestPath    []int
}

func NewACO(alpha, beta, evaporation float64, ants, iterations int, distances [][]int) *ACO {
	dimension := len(distances)
	pheromone := make([][]float64, dimension)
	for i := range pheromone {
		pheromone[i] = make([]float64, dimension)
		for j := range pheromone[i] {
			pheromone[i][j] = 1.0 // initial pheromone level
		}
	}

	return &ACO{
		alpha:       alpha,
		beta:        beta,
		evaporation: evaporation,
		ants:        ants,
		iterations:  iterations,
		pheromone:   pheromone,
		distances:   distances,
		bestLength:  math.Inf(1),
	}
}

func (aco *ACO) Run() {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < aco.iterations; i++ {
		paths := make([][]int, aco.ants)
		lengths := make([]float64, aco.ants)
		for j := 0; j < aco.ants; j++ {
			paths[j], lengths[j] = aco.constructPath()
		}
		aco.updatePheromone(paths, lengths)
	}
}

func (aco *ACO) constructPath() ([]int, float64) {
	dimension := len(aco.distances)
	path := make([]int, dimension)
	visited := make([]bool, dimension)
	current := rand.Intn(dimension)
	path[0] = current
	visited[current] = true

	for i := 1; i < dimension; i++ {
		next := aco.selectNextCity(current, visited)
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

func (aco *ACO) selectNextCity(current int, visited []bool) int {
	dimension := len(aco.distances)
	probabilities := make([]float64, dimension)
	total := 0.0

	for i := 0; i < dimension; i++ {
		if !visited[i] {
			pheromone := math.Pow(aco.pheromone[current][i], aco.alpha)
			invDistance := 1.0 / float64(aco.distances[current][i])
			desirability := math.Pow(invDistance, aco.beta)
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
	for i, cum := 0, 0.0; i < dimension; i++ {
		if !visited[i] {
			cum += probabilities[i]
			if r < cum {
				return i
			}
		}
	}

	return -1 // fallback, should not happen
}

func (aco *ACO) updatePheromone(paths [][]int, lengths []float64) {
	for i, path := range paths {
		for j := 0; j < len(path)-1; j++ {
			start, end := path[j], path[j+1]
			// Increase pheromone level for this path
			delta := 1 / lengths[i]
			aco.pheromone[start][end] += delta
			aco.pheromone[end][start] += delta // if symmetric TSP
		}
	}

	// Evaporate pheromone
	for i := range aco.pheromone {
		for j := range aco.pheromone[i] {
			aco.pheromone[i][j] *= (1 - aco.evaporation)
		}
	}
}

func (aco *ACO) pathLength(path []int) float64 {
	sum := 0.0
	for i := 0; i < len(path)-1; i++ {
		sum += float64(aco.distances[path[i]][path[i+1]])
	}
	return sum
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

	// Process each file
	for _, file := range files {
		fmt.Println("Processing file:", file)
		name, dimension, distances, err := parsing.ParseTSPLIBFile(file)
		if err != nil {
			fmt.Println("Error parsing file:", file, err)
			continue
		}

		fmt.Println("Name:", name)
		fmt.Println("Dimension:", dimension)

		aco := NewACO(1.0, 5.0, 0.5, 20, 100, distances)
		aco.Run()

		fmt.Println("Best Path Length for", name, ":", aco.bestLength)
		fmt.Println("Best Path for", name, ":", aco.bestPath)
	}
}
