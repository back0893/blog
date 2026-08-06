---
title: dart学习记录
date: 2026-07-29
tags: [dart]
draft: false
---
记录一下学习dart过程中和其他语言不一样的点

1. futureOr 可以返回一个future，也可以返回一个值
2. package 定义文件可以统一返回多个文件
3. bin目录不能直接引用lib目录下的文件，需要通过package引用
4. 一个项目文件夹 可以管理多个sub-package
 - 根目录下的pubspec.yaml文件可以定义多个sub-package
 - 每个sub-package的pubspec.yaml添加依赖的工作区

5. 对json进行解析，有了新的解析写法
```
static Summary fromJson(Map<String, Object?> json) {
  // 需要手动检查每个字段是否存在
  if (json.containsKey('titles') && 
      json.containsKey('pageid') && 
      json.containsKey('description')) {
    // 还要手动类型转换
    final titles = json['titles'] as Map<String, Object?>;
    final pageid = json['pageid'] as int;
    // ... 很多重复代码
    return Summary(...);
  } else if (...) {
    // 另一种情况
  } else {
    throw FormatException('...');
  }
}
```
每次解析json都需要手动检查每个字段是否存在，手动类型转换，手动返回值。

新的写法
```
static Summary fromJson(Map<String, Object?> json) {
  return switch (json) {
    // 模式1：包含 description 字段的 JSON
    {
      'titles': final Map<String, Object?> titles,
      'pageid': final int pageid,
      'extract': final String extract,
      'extract_html': final String extractHtml,
      'lang': final String lang,
      'dir': final String dir,
      'content_urls': {
        'desktop': {'page': final String url},
        'mobile': {'page': String _},
      },
      'description': final String description,
    } =>
      Summary(
        titles: TitlesSet.fromJson(titles),
        pageid: pageid,
        extract: extract,
        extractHtml: extractHtml,
        lang: lang,
        dir: dir,
        url: url,
        description: description,
      ),
  }
}
```
和if-case也有类似的写法
```
if (json case {
  'canonical': final String canonical,
  'normalized': final String normalized,
  'display': final String display,
}) {
  return TitlesSet(
    canonical: canonical,
    normalized: normalized,
    display: display,
  );
}
```
都是可以加速解析json的

6. 测试文件和go一样 都是_test 结尾，不同是使用main和group包裹测试代码