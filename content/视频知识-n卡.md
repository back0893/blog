---
title: 视频相关知识-n卡
date: 2026-10-16
tags: [ffmpeg]
draft: false
---
n硬件解码的相关知识

- NVENC（NVIDIA Encoder）硬件编码器，负责将视频数据编码为指定的视频格式。比如 `h264_nvenc`
- NVDEC（NVIDIA Decoder）硬件解码器，负责将视频数据解码为原始视频帧。比如 `h264_cuvid`
- CUVID 是一个更上层的 API 框架，会调用编码器还包含容器格式的解码等功能
# nvidia 硬件解码
[nvidia ffmpeg使用](https://docs.nvidia.com/video-technologies/video-codec-sdk/13.1/ffmpeg-with-nvidia-gpu/index.html)

文档中没有具体的介绍命令，可以使用命令查询来了解

```
# 查询cuda的硬件相关
ffmpeg -hide_banner -filters  | grep cuda

# 查询硬件编码器
ffmpeg -hide_banner -encoders | grep nvenc
ffmpeg -hide_banner -decoders | grep cuvid

# 查询如何使用
ffmpeg -h filter=hwupload_cuda

```
其实和amd的硬件使用流程差不多，只是在使用时，需要使用n卡相关的参数

# 10bit和8 bit转化的问题
也是使用滤镜 `vf`来处理
```
-vf "hwdownload,format=p010le,format=yuv420p,scale=1920:1076,hwupload_cuda"
```
第一个format表示从卡上搬回来的像素格式，第二个format表示主机处理后的像素格式，这样就可以在主机上处理10bit的视频，然后在卡上编码8bit的视频
