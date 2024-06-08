# Max-Min Ant System for Asymmetric Traveling Salesman Problem

## Overview
This project is and implementation of the Max-Min Ant System (MMAS), an optimization algorithm derived from the Ant Colony Optimization (ACO) metaheuristic.
MMSA is used in this project to solve the Asymmetric Traveling Salesman Problem (ATSP). It seems to fair better than the classic ACO that was used in this project previously.

MMAS focuses on enhancing solution quality by intensifying search around promising areas while preserving solution diversity. The approach is grounded in strengthening pheromone trails associated with the most promising solutions, which theoretically guides subsequent searches towards globally optimal solutions.

## Features
- **Dynamic Pheromone Update Mechanism**: Implements constraints on pheromone concentrations, setting minimum and maximum bounds to prevent premature convergence on suboptimal solutions and to sustain exploration capabilities throughout the search process.
- **Adaptive Parameter Tuning**: Employs a systematic exploration of parameter spaces for `Alpha` (influence of pheromone trails), `Beta` (influence of heuristic information), `Evaporation` (rate of decreasing pheromone levels), and `Exploration` (increases exploratory aspect by decreasing the minimum bound of pheromone level) to optimize the performance of the algorithm across different problem instances.

## Algorithmic Framework
The Max-Min Ant System introduces modifications to the traditional pheromone updating rules used in ACO algorithms by focusing on the following two key mechanisms:
1. **Limited pheromone evaporation**: At each iteration, pheromone values on all paths are reduced by a predetermined evaporation rate, simulating the natural decay of pheromone trails over time and preventing unbounded growth of pheromone values. The pheromone value cannot be lower than a lower bound.
2. **Elitist pheromone reinforcement**: Exclusively reinforces pheromone trails that are part of the best solution found, thereby directing the search towards regions of the search space that are likely to contain near-optimal or optimal solutions. The pheromone concentration on these paths is subject to an upper bound.

`maxPheromone` and `minPheromone` bounds help in maintaining a balance between exploration of new areas and exploitation of known good solutions.

## Methodology
### Test data
Part of the [TSPLib](http://comopt.ifi.uni-heidelberg.de/software/TSPLIB95/atsp/) problem set is used and since it also provides optimal solutions for each graph we can calculate the quality of calculated results. By default the `ftv170` graph is used.

### Initialization
The algorithm initializes pheromone trails to a value of 1.0, allowing equal initial exploration probability across all edges in the ATSP instance.

`Alpha`, `Beta`, `Evaporation` and `Exploration` parameter values are varied with the following scheme:
| Parameter | Start value | End value | Step |
|-|-|-|-|
| `Alpha` | 0.75 | 1.25 | 0.25 |
| `Beta` | 3.0 | 5.0 | 1.0 |
| `Evaporation` | 0.5 | 0.8 | 0.1 |
| `Exploration` | 8.0 | 10.0 | 1.0 |

For each set of values 10 runs of the algorithm are performed and following metrics are returned as the average from those runs: `Average Result`, `Deviation` and `Success rate`.
`Deviation` shows how close is the `Average Result` to the optimal solution provided by TSPLib.

The amount of iterations performed by the MMAS is dependant on the size of the problem. Below is the logic that determines the amount of iterations:
```
if size < 50 {
	iterations = 100
}

if 50 <= size && size < 100 {
	iterations = 500
}

if size >= 100 {
	iterations = 1000
}
```

### Construction of Solutions
Ants construct solutions by probabilistically selecting the next city to visit based on a combination of pheromone intensity and heuristic desirability (inverse of travel cost). This process is repeated until all cities are visited.

### Pheromone Update
After all ants have constructed their tours, pheromone levels are updated according to the quality of the solutions found. Only the best-performing ant contributes to the pheromone update, enhancing trails associated with more promising solutions.

### Parameter Adaptation
The algorithm adaptively adjusts the pheromone bounds based on the quality of solutions found, which is designed to dynamically tailor the search intensity and exploration extent as needed.

These values are calculated in accordance to this formula and changed before updating pheromone levels:
```
maxPheromone = 1.0 / ((1 - evaporation) * bestPathLength)
minPheromone = maxPheromone / (exploration * numberOfAnts)
```

## Results
Below is the output of the program testing various parameter values for `Alpha`, `Beta`, `Evaporation` and `Exploration` for the `ftv170` graph:

| Instance | Alpha | Beta | Evaporation | Exploration | Ants | Iterations | Average Result | Best found | Best found at iteration | Known Optimal | Deviation (%) | Success rate (%) |
|-|-|-|-|-|-|-|-|-|-|-|-|-|
| ftv170 | 0.75 | 3.00 | 0.50 | 8.00 | 171 | 1000 | 2854 | 2790 | 342 | 2755 | 3.61 | 0.00 |
| ftv170 | 0.75 | 3.00 | 0.50 | 9.00 | 171 | 1000 | 2942 | 2794 | 926 | 2755 | 6.78 | 0.00 |
| ftv170 | 0.75 | 3.00 | 0.50 | 10.00 | 171 | 1000 | 2918 | 2824 | 820 | 2755 | 5.90 | 0.00 |
| ftv170 | 0.75 | 3.00 | 0.60 | 8.00 | 171 | 1000 | 2885 | 2786 | 725 | 2755 | 4.70 | 0.00 |
| ftv170 | 0.75 | 3.00 | 0.60 | 9.00 | 171 | 1000 | 2900 | 2814 | 926 | 2755 | 5.27 | 0.00 |
| ftv170 | 0.75 | 3.00 | 0.60 | 10.00 | 171 | 1000 | 2892 | 2770 | 875 | 2755 | 4.97 | 0.00 |
| ftv170 | 0.75 | 3.00 | 0.70 | 8.00 | 171 | 1000 | 2852 | 2768 | 419 | 2755 | 3.53 | 0.00 |
| ftv170 | 0.75 | 3.00 | 0.70 | 9.00 | 171 | 1000 | 2892 | 2850 | 370 | 2755 | 4.97 | 0.00 |
| ftv170 | 0.75 | 3.00 | 0.70 | 10.00 | 171 | 1000 | 2871 | 2788 | 597 | 2755 | 4.20 | 0.00 |
| ftv170 | 0.75 | 3.00 | 0.80 | 8.00 | 171 | 1000 | 2935 | 2822 | 584 | 2755 | 6.53 | 0.00 |
| ftv170 | 0.75 | 3.00 | 0.80 | 9.00 | 171 | 1000 | 2941 | 2865 | 932 | 2755 | 6.75 | 0.00 |
| ftv170 | 0.75 | 3.00 | 0.80 | 10.00 | 171 | 1000 | 2899 | 2808 | 992 | 2755 | 5.24 | 0.00 |
| ftv170 | 0.75 | 4.00 | 0.50 | 8.00 | 171 | 1000 | 2833 | 2777 | 866 | 2755 | 2.83 | 0.00 |
| ftv170 | 0.75 | 4.00 | 0.50 | 9.00 | 171 | 1000 | 2861 | 2788 | 430 | 2755 | 3.86 | 0.00 |
| ftv170 | 0.75 | 4.00 | 0.50 | 10.00 | 171 | 1000 | 2898 | 2828 | 333 | 2755 | 5.19 | 0.00 |
| ftv170 | 0.75 | 4.00 | 0.60 | 8.00 | 171 | 1000 | 2891 | 2821 | 386 | 2755 | 4.94 | 0.00 |
| ftv170 | 0.75 | 4.00 | 0.60 | 9.00 | 171 | 1000 | 2854 | 2796 | 846 | 2755 | 3.61 | 0.00 |
| ftv170 | 0.75 | 4.00 | 0.60 | 10.00 | 171 | 1000 | 2838 | 2765 | 930 | 2755 | 3.00 | 0.00 |
| ftv170 | 0.75 | 4.00 | 0.70 | 8.00 | 171 | 1000 | 2891 | 2844 | 665 | 2755 | 4.95 | 0.00 |
| ftv170 | 0.75 | 4.00 | 0.70 | 9.00 | 171 | 1000 | 2857 | 2761 | 344 | 2755 | 3.71 | 0.00 |
| ftv170 | 0.75 | 4.00 | 0.70 | 10.00 | 171 | 1000 | 2876 | 2791 | 725 | 2755 | 4.37 | 0.00 |
| ftv170 | 0.75 | 4.00 | 0.80 | 8.00 | 171 | 1000 | 2856 | 2780 | 744 | 2755 | 3.67 | 0.00 |
| ftv170 | 0.75 | 4.00 | 0.80 | 9.00 | 171 | 1000 | 2868 | 2798 | 952 | 2755 | 4.08 | 0.00 |
| ftv170 | 0.75 | 4.00 | 0.80 | 10.00 | 171 | 1000 | 2834 | 2785 | 937 | 2755 | 2.86 | 0.00 |
| ftv170 | 0.75 | 5.00 | 0.50 | 8.00 | 171 | 1000 | 2853 | 2765 | 702 | 2755 | 3.55 | 0.00 |
| ftv170 | 0.75 | 5.00 | 0.50 | 9.00 | 171 | 1000 | 2887 | 2764 | 754 | 2755 | 4.80 | 0.00 |
| ftv170 | 0.75 | 5.00 | 0.50 | 10.00 | 171 | 1000 | 2900 | 2762 | 749 | 2755 | 5.27 | 0.00 |
| ftv170 | 0.75 | 5.00 | 0.60 | 8.00 | 171 | 1000 | 2890 | 2794 | 972 | 2755 | 4.90 | 0.00 |
| ftv170 | 0.75 | 5.00 | 0.60 | 9.00 | 171 | 1000 | 2852 | 2762 | 972 | 2755 | 3.52 | 0.00 |
| ftv170 | 0.75 | 5.00 | 0.60 | 10.00 | 171 | 1000 | 2866 | 2792 | 763 | 2755 | 4.05 | 0.00 |
| ftv170 | 0.75 | 5.00 | 0.70 | 8.00 | 171 | 1000 | 2857 | 2768 | 844 | 2755 | 3.69 | 0.00 |
| ftv170 | 0.75 | 5.00 | 0.70 | 9.00 | 171 | 1000 | 2859 | 2768 | 823 | 2755 | 3.78 | 0.00 |
| ftv170 | 0.75 | 5.00 | 0.70 | 10.00 | 171 | 1000 | 2866 | 2822 | 541 | 2755 | 4.04 | 0.00 |
| ftv170 | 0.75 | 5.00 | 0.80 | 8.00 | 171 | 1000 | 2836 | 2775 | 815 | 2755 | 2.95 | 0.00 |
| ftv170 | 0.75 | 5.00 | 0.80 | 9.00 | 171 | 1000 | 2870 | 2825 | 242 | 2755 | 4.16 | 0.00 |
| ftv170 | 0.75 | 5.00 | 0.80 | 10.00 | 171 | 1000 | 2845 | 2765 | 433 | 2755 | 3.25 | 0.00 |
| ftv170 | 1.00 | 3.00 | 0.50 | 8.00 | 171 | 1000 | 3008 | 2882 | 785 | 2755 | 9.20 | 0.00 |
| ftv170 | 1.00 | 3.00 | 0.50 | 9.00 | 171 | 1000 | 2997 | 2897 | 979 | 2755 | 8.78 | 0.00 |
| ftv170 | 1.00 | 3.00 | 0.50 | 10.00 | 171 | 1000 | 3005 | 2886 | 763 | 2755 | 9.06 | 0.00 |
| ftv170 | 1.00 | 3.00 | 0.60 | 8.00 | 171 | 1000 | 3005 | 2917 | 434 | 2755 | 9.07 | 0.00 |
| ftv170 | 1.00 | 3.00 | 0.60 | 9.00 | 171 | 1000 | 2997 | 2922 | 988 | 2755 | 8.78 | 0.00 |
| ftv170 | 1.00 | 3.00 | 0.60 | 10.00 | 171 | 1000 | 2931 | 2799 | 737 | 2755 | 6.39 | 0.00 |
| ftv170 | 1.00 | 3.00 | 0.70 | 8.00 | 171 | 1000 | 2950 | 2776 | 640 | 2755 | 7.09 | 0.00 |
| ftv170 | 1.00 | 3.00 | 0.70 | 9.00 | 171 | 1000 | 2936 | 2811 | 818 | 2755 | 6.56 | 0.00 |
| ftv170 | 1.00 | 3.00 | 0.70 | 10.00 | 171 | 1000 | 2952 | 2820 | 906 | 2755 | 7.16 | 0.00 |
| ftv170 | 1.00 | 3.00 | 0.80 | 8.00 | 171 | 1000 | 2910 | 2796 | 805 | 2755 | 5.61 | 0.00 |
| ftv170 | 1.00 | 3.00 | 0.80 | 9.00 | 171 | 1000 | 2980 | 2898 | 476 | 2755 | 8.19 | 0.00 |
| ftv170 | 1.00 | 3.00 | 0.80 | 10.00 | 171 | 1000 | 2926 | 2850 | 669 | 2755 | 6.22 | 0.00 |
| ftv170 | 1.00 | 4.00 | 0.50 | 8.00 | 171 | 1000 | 2893 | 2802 | 878 | 2755 | 4.99 | 0.00 |
| ftv170 | 1.00 | 4.00 | 0.50 | 9.00 | 171 | 1000 | 2906 | 2775 | 917 | 2755 | 5.50 | 0.00 |
| ftv170 | 1.00 | 4.00 | 0.50 | 10.00 | 171 | 1000 | 2885 | 2761 | 690 | 2755 | 4.72 | 0.00 |
| ftv170 | 1.00 | 4.00 | 0.60 | 8.00 | 171 | 1000 | 2915 | 2818 | 794 | 2755 | 5.82 | 0.00 |
| ftv170 | 1.00 | 4.00 | 0.60 | 9.00 | 171 | 1000 | 2964 | 2888 | 383 | 2755 | 7.58 | 0.00 |
| ftv170 | 1.00 | 4.00 | 0.60 | 10.00 | 171 | 1000 | 2968 | 2849 | 714 | 2755 | 7.71 | 0.00 |
| ftv170 | 1.00 | 4.00 | 0.70 | 8.00 | 171 | 1000 | 2863 | 2783 | 537 | 2755 | 3.92 | 0.00 |
| ftv170 | 1.00 | 4.00 | 0.70 | 9.00 | 171 | 1000 | 2902 | 2788 | 661 | 2755 | 5.34 | 0.00 |
| ftv170 | 1.00 | 4.00 | 0.70 | 10.00 | 171 | 1000 | 2907 | 2778 | 652 | 2755 | 5.51 | 0.00 |
| ftv170 | 1.00 | 4.00 | 0.80 | 8.00 | 171 | 1000 | 2868 | 2761 | 406 | 2755 | 4.11 | 0.00 |
| ftv170 | 1.00 | 4.00 | 0.80 | 9.00 | 171 | 1000 | 2888 | 2799 | 820 | 2755 | 4.81 | 0.00 |
| ftv170 | 1.00 | 4.00 | 0.80 | 10.00 | 171 | 1000 | 2860 | 2770 | 348 | 2755 | 3.82 | 0.00 |
| ftv170 | 1.00 | 5.00 | 0.50 | 8.00 | 171 | 1000 | 2894 | 2780 | 894 | 2755 | 5.06 | 0.00 |
| ftv170 | 1.00 | 5.00 | 0.50 | 9.00 | 171 | 1000 | 2904 | 2828 | 728 | 2755 | 5.39 | 0.00 |
| ftv170 | 1.00 | 5.00 | 0.50 | 10.00 | 171 | 1000 | 2866 | 2758 | 330 | 2755 | 4.02 | 0.00 |
| ftv170 | 1.00 | 5.00 | 0.60 | 8.00 | 171 | 1000 | 2916 | 2815 | 966 | 2755 | 5.85 | 0.00 |
| ftv170 | 1.00 | 5.00 | 0.60 | 9.00 | 171 | 1000 | 2878 | 2806 | 826 | 2755 | 4.46 | 0.00 |
| ftv170 | 1.00 | 5.00 | 0.60 | 10.00 | 171 | 1000 | 2890 | 2758 | 978 | 2755 | 4.90 | 0.00 |
| ftv170 | 1.00 | 5.00 | 0.70 | 8.00 | 171 | 1000 | 2883 | 2814 | 421 | 2755 | 4.63 | 0.00 |
| ftv170 | 1.00 | 5.00 | 0.70 | 9.00 | 171 | 1000 | 2885 | 2795 | 476 | 2755 | 4.73 | 0.00 |
| ftv170 | 1.00 | 5.00 | 0.70 | 10.00 | 171 | 1000 | 2894 | 2783 | 326 | 2755 | 5.04 | 0.00 |
| ftv170 | 1.00 | 5.00 | 0.80 | 8.00 | 171 | 1000 | 2855 | 2761 | 598 | 2755 | 3.64 | 0.00 |
| ftv170 | 1.00 | 5.00 | 0.80 | 9.00 | 171 | 1000 | 2834 | 2778 | 183 | 2755 | 2.88 | 0.00 |
| ftv170 | 1.00 | 5.00 | 0.80 | 10.00 | 171 | 1000 | 2841 | 2765 | 719 | 2755 | 3.11 | 0.00 |
| ftv170 | 1.25 | 3.00 | 0.50 | 8.00 | 171 | 1000 | 3138 | 2981 | 990 | 2755 | 13.91 | 0.00 |
| ftv170 | 1.25 | 3.00 | 0.50 | 9.00 | 171 | 1000 | 3224 | 3053 | 988 | 2755 | 17.03 | 0.00 |
| ftv170 | 1.25 | 3.00 | 0.50 | 10.00 | 171 | 1000 | 3255 | 3080 | 661 | 2755 | 18.16 | 0.00 |
| ftv170 | 1.25 | 3.00 | 0.60 | 8.00 | 171 | 1000 | 3072 | 2934 | 985 | 2755 | 11.52 | 0.00 |
| ftv170 | 1.25 | 3.00 | 0.60 | 9.00 | 171 | 1000 | 3140 | 2924 | 663 | 2755 | 13.98 | 0.00 |
| ftv170 | 1.25 | 3.00 | 0.60 | 10.00 | 171 | 1000 | 3124 | 2985 | 626 | 2755 | 13.40 | 0.00 |
| ftv170 | 1.25 | 3.00 | 0.70 | 8.00 | 171 | 1000 | 3089 | 2852 | 738 | 2755 | 12.11 | 0.00 |
| ftv170 | 1.25 | 3.00 | 0.70 | 9.00 | 171 | 1000 | 3063 | 2775 | 673 | 2755 | 11.18 | 0.00 |
| ftv170 | 1.25 | 3.00 | 0.70 | 10.00 | 171 | 1000 | 3154 | 2975 | 685 | 2755 | 14.50 | 0.00 |
| ftv170 | 1.25 | 3.00 | 0.80 | 8.00 | 171 | 1000 | 3018 | 2886 | 704 | 2755 | 9.53 | 0.00 |
| ftv170 | 1.25 | 3.00 | 0.80 | 9.00 | 171 | 1000 | 3043 | 2818 | 874 | 2755 | 10.46 | 0.00 |
| ftv170 | 1.25 | 3.00 | 0.80 | 10.00 | 171 | 1000 | 3051 | 2970 | 951 | 2755 | 10.73 | 0.00 |
| ftv170 | 1.25 | 4.00 | 0.50 | 8.00 | 171 | 1000 | 3065 | 2948 | 439 | 2755 | 11.25 | 0.00 |
| ftv170 | 1.25 | 4.00 | 0.50 | 9.00 | 171 | 1000 | 3043 | 2911 | 694 | 2755 | 10.46 | 0.00 |
| ftv170 | 1.25 | 4.00 | 0.50 | 10.00 | 171 | 1000 | 3081 | 2963 | 826 | 2755 | 11.85 | 0.00 |
| ftv170 | 1.25 | 4.00 | 0.60 | 8.00 | 171 | 1000 | 2987 | 2838 | 940 | 2755 | 8.41 | 0.00 |
| ftv170 | 1.25 | 4.00 | 0.60 | 9.00 | 171 | 1000 | 2987 | 2838 | 674 | 2755 | 8.42 | 0.00 |
| ftv170 | 1.25 | 4.00 | 0.60 | 10.00 | 171 | 1000 | 3040 | 2946 | 941 | 2755 | 10.34 | 0.00 |
| ftv170 | 1.25 | 4.00 | 0.70 | 8.00 | 171 | 1000 | 2996 | 2876 | 806 | 2755 | 8.73 | 0.00 |
| ftv170 | 1.25 | 4.00 | 0.70 | 9.00 | 171 | 1000 | 3035 | 2907 | 917 | 2755 | 10.16 | 0.00 |
| ftv170 | 1.25 | 4.00 | 0.70 | 10.00 | 171 | 1000 | 3001 | 2861 | 555 | 2755 | 8.93 | 0.00 |
| ftv170 | 1.25 | 4.00 | 0.80 | 8.00 | 171 | 1000 | 2942 | 2786 | 549 | 2755 | 6.80 | 0.00 |
| ftv170 | 1.25 | 4.00 | 0.80 | 9.00 | 171 | 1000 | 2900 | 2841 | 957 | 2755 | 5.26 | 0.00 |
| ftv170 | 1.25 | 4.00 | 0.80 | 10.00 | 171 | 1000 | 2980 | 2864 | 895 | 2755 | 8.19 | 0.00 |
| ftv170 | 1.25 | 5.00 | 0.50 | 8.00 | 171 | 1000 | 2988 | 2839 | 669 | 2755 | 8.45 | 0.00 |
| ftv170 | 1.25 | 5.00 | 0.50 | 9.00 | 171 | 1000 | 2905 | 2829 | 977 | 2755 | 5.44 | 0.00 |
| ftv170 | 1.25 | 5.00 | 0.50 | 10.00 | 171 | 1000 | 2945 | 2871 | 988 | 2755 | 6.90 | 0.00 |
| ftv170 | 1.25 | 5.00 | 0.60 | 8.00 | 171 | 1000 | 2907 | 2829 | 623 | 2755 | 5.51 | 0.00 |
| ftv170 | 1.25 | 5.00 | 0.60 | 9.00 | 171 | 1000 | 2980 | 2817 | 787 | 2755 | 8.18 | 0.00 |
| ftv170 | 1.25 | 5.00 | 0.60 | 10.00 | 171 | 1000 | 2976 | 2905 | 949 | 2755 | 8.03 | 0.00 |
| ftv170 | 1.25 | 5.00 | 0.70 | 8.00 | 171 | 1000 | 2923 | 2798 | 299 | 2755 | 6.08 | 0.00 |
| ftv170 | 1.25 | 5.00 | 0.70 | 9.00 | 171 | 1000 | 2897 | 2783 | 480 | 2755 | 5.15 | 0.00 |
| ftv170 | 1.25 | 5.00 | 0.70 | 10.00 | 171 | 1000 | 2949 | 2815 | 916 | 2755 | 7.05 | 0.00 |
| ftv170 | 1.25 | 5.00 | 0.80 | 8.00 | 171 | 1000 | 2938 | 2809 | 445 | 2755 | 6.64 | 0.00 |
| ftv170 | 1.25 | 5.00 | 0.80 | 9.00 | 171 | 1000 | 2917 | 2784 | 850 | 2755 | 5.87 | 0.00 |
| ftv170 | 1.25 | 5.00 | 0.80 | 10.00 | 171 | 1000 | 2927 | 2850 | 842 | 2755 | 6.23 | 0.00 |

Best parameters:
 - Alpha: 0.75
 - Beta: 4.00
 - Evaporation: 0.50
 - Exploration: 8.00
 - Average length: 2833
 - Deviation: 2.83
 - Success rate: 0.00

### Analysis

Best configuration achieved the best balance between exploration and exploitation among the tested settings, resulting in the lowest deviation from the optimal path length.

Parameter Impact:
 - `Alpha` (Pheromone Importance): A lower alpha (0.75) appears to yield better results in several configurations, suggesting that less emphasis on pheromone strength (more on heuristic information) helps in this scenario.
 - `Beta` (Heuristic Information Importance): Increasing beta generally improves performance, with beta = 4.00 providing the best results. This indicates a higher reliance on heuristic information (inverse of distance) is beneficial for this problem.
 - `Evaporation`: A lower evaporation rate (0.50) has shown better performance, potentially because it maintains useful pheromone trails longer, aiding in the convergence to good solutions.
 - `Exploration`: An exploration factor of 8.00 gave the best results, striking a balance between exploring new paths and exploiting known good paths.


All configurations report a 0.00% success rate, indicating that none of the runs achieved the known optimal path length. This highlights a potential area for further tuning or the need for algorithmic enhancements.

The deviation from the optimal varies significantly across different settings, suggesting sensitivity to parameter changes. Some configurations, particularly those with higher evaporation rates and exploration factors, lead to poorer performance, possibly due to excessive exploration or too rapid loss of pheromone information.

## Conclusion
The MMAS represents a robust approach to solving the ATSP, characterized by its ability to dynamically adjust search strategies in response to the evolving landscape of solution quality. Future work may explore hybrid strategies combining MMAS with other heuristic or exact methods to further enhance its performance on complex ATSP instances.
