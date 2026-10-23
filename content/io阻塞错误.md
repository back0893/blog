---
title: io阻塞错误
date: 2026-08-11
tags: [Go]
draft: false
---

# 本次案例
输入一个坏的mp4文件，转码时 ffmpge持续的输出 stderr信息，使用一个带有大小的buffer(512k)，来存储这些信息。有一个看门狗，观测转码文件大小，当转码进行中并且文件大小一直不变时，认为转码被阻塞，会发送一个kill信号给ffmpge进程，强制结束转码

- 现象：
  - 「命令输出」日志在 512KB 边界处**半行截断**；
  - 重试产出的 mp4 停在 48 字节（只有容器头）；
  - 没有任何 ffmpeg 致命错误；
  - 任务最后被 `monitorFileSize` 的停滞检测发 SIGINT 终止。
- 误判方向：先怀疑音频滤镜 `pan`、硬解回退逻辑。实际把 stderr 直接重定向到文件后，同一条命令 10.87s 就正常报错退出，加 `pan` 后 150s 正常完成——问题不在 ffmpeg 参数。

# 本次错误的原因
```
func (b *threadSafeBuffer) Write(p []byte) (int, error) {
    b.mu.Lock()
    defer b.mu.Unlock()

    remaining := b.max - b.b.Len()
    if remaining > 0 {
        // 能塞多少塞多少，塞不下的丢弃
        return b.b.Write(p[:remaining])
    }
    return len(p), nil
}
```
其中的 `return b.b.Write(p[:remaining])` 破坏了 `io.Copy`的约定

## `io.Copy` 的约定
标准库的复制循环（简化）：

```go
nr, er := src.Read(buf)
nw, ew := dst.Write(buf[:nr])
if ew != nil {
    return ew
}
if nw != nr {
    return io.ErrShortWrite   // 没写满又没有错误，视为短写并停止
}
```

对应的 `io.Writer` 契约：

> Write 必须返回 `nw == len(p)`；否则必须返回非 nil 错误。

两种情况同时成立就是 bug：

```go
nw < len(p) && err == nil
```

## 正确写法

限制保留量，但必须持续消费：

```go
func (b *threadSafeBuffer) Write(p []byte) (int, error) {
    b.mu.Lock()
    defer b.mu.Unlock()

    remaining := b.max - b.b.Len()
    if remaining > 0 {
        // 能塞多少塞多少，塞不下的丢弃
        _, _ = b.b.Write(p[:remaining])
    }
    // 无论实际写入多少，都必须报告整块已处理，
    // 否则 io.Copy 会以 io.ErrShortWrite 停止读取
    return len(p), nil
}
```

## 问题还原
1. 因为写入的数据量大于保留量，所以会返回 `nw < len(p)`，导致 `io.Copy` 停止读取
2. 因为导致 `io.Copy` 停止读取，所以会导致进行阻塞,导致后续的流程被阻塞