---
title: flutter学习(一)
date: 2026-08-06
tags: [dart,flutter]
draft: false
---

# 学习记录
## widget 是否可以变？
- 不能 这和vue差不多 widget只要创建了 就不能改变了，每次都是重建
1. stateless的widget能否变ui？
它本省不能 但是如果子组件是有状态的，当子组件的状态改变时，子组件会重新build，从而改变ui。
这点和vue有直观的不同 vue中如果不绑定属性,那么组件的ui不会改变。