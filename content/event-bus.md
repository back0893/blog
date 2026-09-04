---
title: event-bus设计并且实现
date: 2026-09-04
tags: [Go, event-bus]
draft: false
---

# event-bus 设计并且实现
我也不记不住我是在那里看到`event-bus`并且记住其设计原理。可能是在vue那里把？
目标：需要将事件广播给监听者，监听者接收到广播后处理
事件总线是将发送方和接收方进行了一个解耦，让扩展变得极其简单。

## 为什么不用channel？
对啊 为什么不用呢？因为channel的处理是一对一的，多个监听者订阅了同一个事件后，channel也需要发送多个数据才能保证
每个监听者接受到数据，这显然是不合理，也不好实现。

## 简单的实现
```
// subscriber 一条订阅：id 用于退订定位，fn 为处理函数。
type subscriber struct {
	id int64
	fn func(Event)
}

// Bus 进程内事件总线：按事件类型分发类型化事件。
// 订阅者之间互不阻塞：单个订阅者 panic 会被捕获，不中断其他订阅者。
type Bus struct {
	mu     sync.RWMutex
	subs   map[EventType][]*subscriber
	nextID int64
}
// Subscribe 为指定事件类型注册订阅者，返回取消订阅的函数。
func (b *Bus) Subscribe(t EventType, fn func(Event)) func() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextID++
	s := &subscriber{id: b.nextID, fn: fn}
	b.subs[t] = append(b.subs[t], s)
	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		list := b.subs[t]
		for i, x := range list {
			if x.id == s.id {
				b.subs[t] = append(list[:i], list[i+1:]...)
				return
			}
		}
	}
}

// Publish 同步地把事件分发给该类型的所有订阅者。
// 在锁外调用处理函数：订阅者可安全地在处理函数内订阅 / 退订。
// 某个订阅者 panic 只影响它自己，不影响其余订阅者与调用方。
func (b *Bus) Publish(e Event) {
	b.mu.RLock()
	fns := make([]func(Event), 0, len(b.subs[e.Type]))
	for _, s := range b.subs[e.Type] {
		fns = append(fns, s.fn)
	}
	b.mu.RUnlock()
	for _, fn := range fns {
		func() {
			defer func() { _ = recover() }()
			fn(e)
		}()
	}
}
```
非常简单的实现，但是可以发现，如果想要进行扩展，同一个事件类型，可以有多个处理函数。
比如同一个事件可以为数据库插入订阅一个事件，也可以为前端推送订阅一个事件，而且前端推送的事件可以作为一个懒事件，让前端连接后才订阅。
不需要修改`event-bus`的实现或者其他订阅事件就可以扩展,关于订阅处理实现了对扩展开放，对修改关闭。