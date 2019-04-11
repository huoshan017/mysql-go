package mysql_base

import (
	"sync"
)

type ListNodePool struct {
	pool   *sync.Pool
	inited bool
}

var listnode_pool ListNodePool

func (this *ListNodePool) Init() {
	this.pool = &sync.Pool{
		New: func() interface{} {
			return &ListNode{}
		},
	}
	this.inited = true
}

func (this *ListNodePool) Get(data interface{}) *ListNode {
	if !this.inited {
		this.Init()
	}
	node := this.pool.Get().(*ListNode)
	node.next = nil
	node.prev = nil
	node.data = data
	return node
}

func (this *ListNodePool) Put(m *ListNode) {
	if !this.inited {
		return
	}
	this.pool.Put(m)
}

type ListNode struct {
	data interface{}
	next *ListNode
	prev *ListNode
}

func NewListNode(data interface{}) *ListNode {
	return &ListNode{
		data: data,
	}
}

func (this *ListNode) GetData() interface{} {
	return this.data
}

func (this *ListNode) GetNext() *ListNode {
	return this.next
}

func (this *ListNode) GetPrev() *ListNode {
	return this.prev
}

type List struct {
	head     *ListNode
	tail     *ListNode
	data_map map[interface{}]*ListNode
}

func (this *List) Append(data interface{}) *ListNode {
	node := listnode_pool.Get(data)
	if this.head == nil {
		this.head = node
	}
	if this.tail != nil {
		this.tail.next = node
		node.prev = this.tail
	}
	this.tail = node
	if this.data_map == nil {
		this.data_map = make(map[interface{}]*ListNode)
	}
	this.data_map[data] = node
	return node
}

func (this *List) HasData(data interface{}) bool {
	_, o := this.data_map[data]
	if !o {
		return false
	}
	return true
}

func (this *List) IsHead(data interface{}) bool {
	head, o := this.data_map[data]
	if !o {
		return false
	}
	return this.head == head
}

func (this *List) IsTail(data interface{}) bool {
	tail, o := this.data_map[data]
	if !o {
		return false
	}
	return this.tail == tail
}

func (this *List) GetHeadNode() *ListNode {
	return this.head
}

func (this *List) GetTailNode() *ListNode {
	return this.tail
}

func (this *List) GetLength() int {
	var length int
	if this.data_map != nil {
		length = len(this.data_map)
	}
	return length
}

func (this *List) Clear() {
	n := this.head
	for n != nil {
		listnode_pool.Put(n)
		n = n.next
	}
	this.head = nil
	this.tail = nil
	this.data_map = nil
}

func (this *List) _delete_node(node *ListNode) {
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}
	if this.head == node {
		this.head = node.next
	}
}

func (this *List) _delete(data interface{}) *ListNode {
	node, o := this.data_map[data]
	if !o {
		return nil
	}
	this._delete_node(node)
	return node
}

func (this *List) Delete(data interface{}) bool {
	node := this._delete(data)
	if node == nil {
		return false
	}
	if this.tail == node {
		this.tail = node.prev
	}
	delete(this.data_map, data)
	return true
}

func (this *List) MoveToLast(data interface{}) bool {
	node, o := this.data_map[data]
	if !o {
		return false
	}

	if this.tail == node {
		return true
	}

	this._delete_node(node)

	this.tail.next = node
	node.prev = this.tail
	node.next = nil
	this.tail = node

	return true
}
