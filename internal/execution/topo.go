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
