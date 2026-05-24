package lifeline

import "github.com/devlikebear/tessera/pkg/queue"

type Recovery struct {
	Recovered []queue.Task
}

func (r Recovery) Empty() bool {
	return len(r.Recovered) == 0
}
