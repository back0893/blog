---
title: win下安装gcc
date: 2026-07-29
tags: [Go, gcc]
draft: false
---
因为MinGW-w64 已经不提供 gcc 的发布版本 导致我以往安装经验丢失了，需要重新记录下如何安装 才能让cgo使用
1. 包管理 
对没错 win下也有包管理 只是用的比较少
使用Scoop包管理器（最快捷）​​：如果你喜欢命令行，这是最现代、最简洁的方法。只需在PowerShell中执行 scoop install gcc 即可自动完成安装和环境配置

2.使用MSYS2
它提供了一个类似Linux的包管理环境，后续安装其他依赖库（如zlib、openssl）也非常方便
从 https://www.msys2.org/下载并安装MSYS2。
- 打开“MSYS2 UCRT64”终端，执行以下命令安装GCC：
```
        pacman -Syu
        pacman -S mingw-w64-ucrt-x86_64-gcc
```
- 安装完成后，将 `C:\msys64\ucrt64\bin` 路径添加到系统环境变量 `PATH` 中
