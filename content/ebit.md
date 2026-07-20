---
title: Ebitengine游戏开发(一)
date: 2026-07-21
tags: [Go, 游戏]
draft: false
---

记录一下使用ebitengine游戏开发的过程
为什么要开发游戏，为什么选着Ebitengine？
其实没有那么多为什么，单存就是我想开发一个游戏，我想使用Ebitengine来开发一个游戏，然后我刚好找到了一本书教我
[Ebitengine游戏开发](https://yumenaka.github.io/ebitengine-book-cn/)

### 画图和帧
``` lang="go"
type Game struct {
	block *ebiten.Image
	y     float64
	g     float64
	vy    float64
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    screen.Fill(color.White)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.vy = -3
	}
	g.vy += g.g
	g.y += g.vy
	y := 180 + g.y
	if y >= 180 {
		y = 180
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(180, y)
	screen.DrawImage(g.block, op)
}
```
每次调用Draw时 都会去画面上重复的绘制，所以不能再其中定义需要逃逸的变量(栈逃逸)
所以如果vy 每次初始化后赋值 那么会导致 其中的黑色方框无法移动
好了你已经知道全部的基础知识了，可以开始来写一个飞天老鼠了