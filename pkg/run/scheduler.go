package run

type Scheduler struct {
	Graph TaskGraph
}

func (s Scheduler) Ready(completed, enqueued map[string]struct{}) []TaskNode {
	var ready []TaskNode
	for _, node := range s.Graph.Nodes {
		if _, ok := completed[node.ID]; ok {
			continue
		}
		if _, ok := enqueued[node.ID]; ok {
			continue
		}
		if depsComplete(node.DependsOn, completed) {
			ready = append(ready, node)
		}
	}
	return ready
}

func depsComplete(deps []string, completed map[string]struct{}) bool {
	for _, dep := range deps {
		if _, ok := completed[dep]; !ok {
			return false
		}
	}
	return true
}
