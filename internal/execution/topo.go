package execution

import "errors"
import wf "github.com/Rithvik-C18/go-flow/internal/workflow"

func topologicalSort(w *wf.Workflow) ([]string, error) {
	inDegree := make(map[string]int)

	for id := range w.Nodes {
		inDegree[id] = 0
	}

	for _, neighbors := range w.Edges {
		for _, to := range neighbors {
			inDegree[to]++
		}
	}

	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	var order []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		order = append(order, current)

		for _, neighbor := range w.Edges[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(order) != len(w.Nodes) {
		return nil, errors.New("workflow contains a cycle")
	}

	return order, nil
}

func getExecutionLevels(w *wf.Workflow) ([][]string, error) {
	inDegree := make(map[string]int)
	level := make(map[string]int)

	for id := range w.Nodes {
		inDegree[id] = 0
		level[id] = 0
	}

	for _, neighbors := range w.Edges {
		for _, to := range neighbors {
			inDegree[to]++
		}
	}

	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	processed := 0
	for len(queue) > 0 {
		currentLevel := len(queue)
		var currentLevelNodes []string

		for i := 0; i < currentLevel; i++ {
			current := queue[0]
			queue = queue[1:]

			currentLevelNodes = append(currentLevelNodes, current)
			processed++

			for _, neighbor := range w.Edges[current] {
				inDegree[neighbor]--
				if inDegree[neighbor] == 0 {
					level[neighbor] = level[current] + 1
					queue = append(queue, neighbor)
				}
			}
		}

		if len(currentLevelNodes) > 0 {
			// Sort nodes within the level for consistent execution
			sortStringSlice(currentLevelNodes)
		}
	}

	if processed != len(w.Nodes) {
		return nil, errors.New("workflow contains a cycle")
	}

	// Group nodes by their levels
	maxLevel := 0
	for _, l := range level {
		if l > maxLevel {
			maxLevel = l
		}
	}

	levels := make([][]string, maxLevel+1)
	for id, l := range level {
		levels[l] = append(levels[l], id)
	}

	// Sort nodes within each level
	for i := range levels {
		sortStringSlice(levels[i])
	}

	return levels, nil
}

func sortStringSlice(slice []string) {
	for i := 0; i < len(slice); i++ {
		for j := i + 1; j < len(slice); j++ {
			if slice[i] > slice[j] {
				slice[i], slice[j] = slice[j], slice[i]
			}
		}
	}
}
