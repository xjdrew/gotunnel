//
//   date  : 2015-06-05
//   author: xjdrew
//

package tunnel

type HubItem struct {
	*ClientHub
	priority int // current link count
	index    int // index in the heap
}

func (h *HubItem) Status() {
	h.Hub.Status()
	Log("priority:%d, index:%d", h.priority, h.index)
}

type HubQueue []*HubItem

func (cq HubQueue) Len() int {
	return len(cq)
}

func (cq HubQueue) Less(i, j int) bool {
	return cq[i].priority < cq[j].priority
}

func (cq HubQueue) Swap(i, j int) {
	cq[i], cq[j] = cq[j], cq[i]
	cq[i].index = i
	cq[j].index = j
}

func (cq *HubQueue) Push(x interface{}) {
	n := len(*cq)
	hub := x.(*HubItem)
	hub.index = n
	*cq = append(*cq, hub)
}

func (cq *HubQueue) Pop() interface{} {
	old := *cq
	n := len(old)
	hub := old[n-1]
	hub.index = -1
	*cq = old[0 : n-1]
	return hub
}
