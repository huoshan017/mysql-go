package mysql_base

import (
	"sync"
)

type ListNodePool struct {
	pool   *sync.Pool
	inited bool
}

var listnode_pool ListNodePool

func (p *ListNodePool) Init() {
	p.pool = &sync.Pool{
		New: func() interface{} {
			return &ListNode{}
		},
	}
	p.inited = true
}

func (p *ListNodePool) Get(data interface{}) *ListNode {
	if !p.inited {
		p.Init()
	}
	node := p.pool.Get().(*ListNode)
	node.next = nil
	node.prev = nil
	node.data = data
	return node
}

func (p *ListNodePool) Put(m *ListNode) {
	if !p.inited {
		return
	}
	p.pool.Put(m)
}

type ListNode struct {
	data interface{}
	next *ListNode
	prev *ListNode
}

func (n *ListNode) GetData() interface{} {
	return n.data
}

func (n *ListNode) GetNext() *ListNode {
	return n.next
}

func (n *ListNode) GetPrev() *ListNode {
	return n.prev
}

type List struct {
	head     *ListNode
	tail     *ListNode
	data_map map[interface{}]*ListNode
}

func (l *List) Append(data interface{}) *ListNode {
	node := listnode_pool.Get(data)
	if l.head == nil {
		l.head = node
	}
	if l.tail != nil {
		l.tail.next = node
		node.prev = l.tail
	}
	l.tail = node
	if l.data_map == nil {
		l.data_map = make(map[interface{}]*ListNode)
	}
	l.data_map[data] = node
	return node
}

func (l *List) HasData(data interface{}) bool {
	_, o := l.data_map[data]
	return o
}

func (l *List) IsHead(data interface{}) bool {
	head, o := l.data_map[data]
	if !o {
		return false
	}
	return l.head == head
}

func (l *List) IsTail(data interface{}) bool {
	tail, o := l.data_map[data]
	if !o {
		return false
	}
	return l.tail == tail
}

func (l *List) GetHeadNode() *ListNode {
	return l.head
}

func (l *List) GetTailNode() *ListNode {
	return l.tail
}

func (l *List) GetLength() int {
	var length int
	if l.data_map != nil {
		length = len(l.data_map)
	}
	return length
}

func (l *List) Clear() {
	n := l.head
	for n != nil {
		listnode_pool.Put(n)
		n = n.next
	}
	l.head = nil
	l.tail = nil
	l.data_map = nil
}

func (l *List) _delete_node(node *ListNode) {
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}
	if l.head == node {
		l.head = node.next
	}
}

func (l *List) _delete(data interface{}) *ListNode {
	node, o := l.data_map[data]
	if !o {
		return nil
	}
	l._delete_node(node)
	return node
}

func (l *List) Delete(data interface{}) bool {
	node := l._delete(data)
	if node == nil {
		return false
	}
	if l.tail == node {
		l.tail = node.prev
	}
	delete(l.data_map, data)
	return true
}

func (l *List) MoveToLast(data interface{}) bool {
	node, o := l.data_map[data]
	if !o {
		return false
	}

	if l.tail == node {
		return true
	}

	l._delete_node(node)

	l.tail.next = node
	node.prev = l.tail
	node.next = nil
	l.tail = node

	return true
}
