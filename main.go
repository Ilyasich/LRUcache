// Package cache - пакет для работы с LRU кэшем
package cache

import (
	"container/list"
	"sync"
)

// Item хранит ключ и значение в очереди.
// Ключ хранится для удобного удаления из map при эвикшне.
type Item struct {
	Key   string
	Value any
}

// LRU простой потокобезопасный LRU-кэш.
// Ведёт себя следующим образом:
//   - при добавлении существующего ключа значение обновляется и элемент
//     перемещается в начало (most-recently-used)
//   - при переполнении удаляется наименее-используемый элемент
type LRU struct {
	mutex    sync.RWMutex
	capacity int
	queue    *list.List
	items    map[string]*list.Element
}

// NewLRU создаёт LRU кэш с заданной ёмкостью. Нулевая или отрицательная
// ёмкость корректируется до 1.
func NewLRU(capacity int) *LRU {
	if capacity <= 0 {
		capacity = 1
	}
	return &LRU{
		capacity: capacity,
		queue:    list.New(),
		items:    make(map[string]*list.Element, capacity),
	}
}

// Add сохраняет значение в кэш по ключу. Если ключ уже существует —
// значение обновляется и элемент помечается как недавно использованный.
// При переполнении удаляется один наименее-используемый элемент.
func (c *LRU) Add(key string, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if element, exists := c.items[key]; exists {
		// обновляем значение и помечаем как MRU
		element.Value.(*Item).Value = value
		c.queue.MoveToFront(element)
		return
	}

	if c.queue.Len() >= c.capacity {
		c.evictOne()
	}

	item := &Item{Key: key, Value: value}
	element := c.queue.PushFront(item)
	c.items[key] = element
}

// Get возвращает значение и флаг найденности. Элемент помечается как MRU.
func (c *LRU) Get(key string) (any, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	element, exists := c.items[key]
	if !exists {
		return nil, false
	}
	c.queue.MoveToFront(element)
	return element.Value.(*Item).Value, true
}

// Peek возвращает значение без изменения порядка (без пометки MRU).
func (c *LRU) Peek(key string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	element, exists := c.items[key]
	if !exists {
		return nil, false
	}
	return element.Value.(*Item).Value, true
}

// Remove удаляет ключ из кэша, возвращает true если элемент был удалён.
func (c *LRU) Remove(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if element, ok := c.items[key]; ok {
		c.deleteElement(element)
		return true
	}
	return false
}

// Len возвращает количество элементов в кэше.
func (c *LRU) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.queue.Len()
}

// Clear очищает весь кэш.
func (c *LRU) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.queue = list.New()
	c.items = make(map[string]*list.Element, c.capacity)
}

// evictOne удаляет один наименее-используемый элемент (с хвоста).
func (c *LRU) evictOne() {
	if element := c.queue.Back(); element != nil {
		c.deleteElement(element)
	}
}

// deleteElement удаляет элемент из очереди и из map. Ожидается, что
// вызывающий уже захватил Lock.
func (c *LRU) deleteElement(element *list.Element) {
	if element == nil {
		return
	}
	item, ok := c.queue.Remove(element).(*Item)
	if !ok || item == nil {
		return
	}
	delete(c.items, item.Key)
}
