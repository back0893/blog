---
title: flutter学习(二)
date: 2026-08-07
tags: [dart,flutter]
draft: false
---

# 学习记录
## 构造函数
dart中的构造函数是一个特殊的方法 虽然是在类定义的作用域内部 但是它不是普通方法 需要使用className.functionName()来调用
是为了通知编译器 这是一个构造函数 而不是普通方法

## 盒子和widget
这部分可以参考前端的盒子模型，本质是一样的 拥有大小和外边框等属性 本质ui组合就是一个一个的盒子堆叠组合
不同点是 盒子的父盒子大小可能被子盒子影响 而widget的父widget大小决定了子widget的大小

## 响应式设计
使用LayoutBuilder时会发现 就算没有使用setState 但是widget也会被重建 这是应为父widget的布局时 约束变化自动触发，这样就实现了响应式设计
但是注意这不能代表可以不使用setState更新

## 数据变更和监听
1. setState方法可以更新状态,需要手动的触发，vue3也有类似的方法
2. ListenableBuilder 可以监听数据变化，当数据变化时，会自动触发build方法
这2个方法都是用来变更ui的 都可以使用只是适用的范围和方便程度不一样 ListenableBuilder更适合数据被多个widget依赖或者需要夸多个widget传递数据的情况
