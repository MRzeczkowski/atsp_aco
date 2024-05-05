package parsing

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseTSPLIBFile(filename string) (name string, dimension int, matrix [][]float64, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", 0, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	readMatrix := false

	// Initialize variables to keep track of matrix data as it's read
	var valuesInMatrix []float64

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "EOF") {
			break
		}

		if strings.HasPrefix(line, "NAME") {
			parts := strings.Split(line, ":")
			name = strings.TrimSpace(parts[1])
		}

		if strings.HasPrefix(line, "DIMENSION") {
			parts := strings.Split(line, ":")
			dimension, err = strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return "", 0, nil, err
			}
			matrix = make([][]float64, dimension)
			for i := range matrix {
				matrix[i] = make([]float64, dimension)
			}
		}

		if readMatrix {
			// Add to a continuous list of values
			rowValues := strings.Fields(line)
			for _, val := range rowValues {
				num, err := strconv.ParseFloat(val, 64)
				if err != nil {
					return "", 0, nil, err
				}
				valuesInMatrix = append(valuesInMatrix, num)
			}
		}

		if strings.HasPrefix(line, "EDGE_WEIGHT_SECTION") {
			readMatrix = true
		}
	}

	if err := scanner.Err(); err != nil {
		return "", 0, nil, err
	}

	// Populate the matrix from the list of values
	if len(valuesInMatrix) != dimension*dimension {
		return "", 0, nil, fmt.Errorf("the total numbers in matrix (%d) does not match expected dimension squared (%d)", len(valuesInMatrix), dimension*dimension)
	}
	for i := 0; i < dimension; i++ {
		for j := 0; j < dimension; j++ {
			matrix[i][j] = valuesInMatrix[i*dimension+j]
		}
	}

	return name, dimension, matrix, nil
}
