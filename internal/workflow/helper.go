package workflow

func (w *Workflow) hasPath(start, target string) bool {
	visited := make(map[string]bool)
	return w.dfs(start, target, visited)
}

func (w *Workflow) dfs(current, target string, visited map[string]bool) bool {
	if current == target {
		return true
	}

	visited[current] = true

	for _, neighbor := range w.Edges[current] {
		if !visited[neighbor] {
			if w.dfs(neighbor, target, visited) {
				return true
			}
		}
	}
	return false
}
