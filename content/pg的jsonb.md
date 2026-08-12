---
title: PostgreSQL JSONB 操作学习
date: 2026-08-11
tags: [PGgreSQL, JSONB]
draft: false
---
# GORM 中 `datatypes.JSON` 与 `serializer:json` 深度对比及 JSONB 原子操作指南

在构建 AI 用户画像、动态配置等半结构化数据场景时，选择正确的存储策略至关重要。本文基于实战经验，深入对比 GORM 中的两种方案，并详解如何使用 `gorm.Expr` 实现高性能的数据库级原子更新。

---

## 一、核心差异：`datatypes.JSON` vs `serializer:json`

| 特性 | `datatypes.JSON` (推荐) | `serializer:json` (慎用) |
| :--- | :--- | :--- |
| **底层存储** | PostgreSQL `JSONB` / MySQL `JSON` | `TEXT` / `VARCHAR` |
| **数据库感知** | **是**，数据库理解 JSON 内部结构 | **否**，仅视为普通字符串 |
| **查询能力** | 支持原生 JSON 函数 (如 `@>`, `?`) | 仅支持全文本匹配 (`LIKE`) |
| **索引支持** | ✅ 支持 **GIN 索引**，查询极快 | ❌ 无法建立有效索引，全表扫描 |
| **更新机制** | **原子局部更新** (数据库内部执行) | **全量覆盖** (应用层 Read-Modify-Write) |
| **并发安全** | ✅ 原子操作，无竞争条件 | ❌ 高并发下必然导致数据覆盖 |
| **适用场景** | 用户画像、标签系统、动态表单 | 简单配置、无需查询的静态 JSON |

**结论**：只要涉及查询、局部修改或并发写，必须选择 `datatypes.JSON`。`serializer:json` 仅适用于“存进去，读出来，永不修改”的简单场景。

---

## 二、致命误区：全量更新 vs 原子更新

### ❌ 错误示例：全量覆盖（并发下的灾难）

很多开发者误以为直接赋值是局部更新，实际上这是**全量覆盖**，在并发下会导致数据丢失。

```go
// ❌ 错误：这会导致其他字段丢失！
var p PersonaExtra
db.First(&p, 1)
// 假设原始数据: {"mood":"sad", "hobbies":["reading"]}
p.PersonalityTags = datatypes.JSON(`{"mood":"happy"}`) 
db.Save(&p) 
// 结果: {"mood":"happy"} 
// 😱 "hobbies" 字段被整条覆盖，数据丢失！
```

### ✅ 正确姿势：`gorm.Expr` + `jsonb_set`

`gorm.Expr` 允许我们将 SQL 表达式原样发送给数据库，利用数据库的 `jsonb_set` 函数实现真正的**局部更新**。

**场景**：仅更新用户画像中的 `mood` 字段。

```go
// ✅ 正确：原子局部更新
db.Model(&PersonaExtra{}).
    Where("id = ?", 1).
    Update("personality_tags", 
        gorm.Expr(
            "jsonb_set(personality_tags, '{mood}', ?, true)", 
            "happy", // GORM 会自动处理 JSON 字符串的引号
        ),
    )
```

**生成的 SQL**：
```sql
UPDATE persona_extras 
SET personality_tags = jsonb_set(personality_tags, '{mood}', '"happy"', true) 
WHERE id = 1;
```
**结果**：`{"mood":"happy", "hobbies":["reading"]}` (其他字段完好无损)

---

## 三、实战：`gorm.Expr` 操作 JSONB 数组

数组操作在标签系统（如兴趣爱好、性格特点）中极为常见。

### 1. 追加元素 (`||` 操作符)
向 `hobbies` 数组中添加新元素。

```go
// 追加单个元素
db.Model(&PersonaExtra{}).
    Where("id = ?", 1).
    Update("personality_tags",
        gorm.Expr("personality_tags || ?", `{"photography"}`),
    )

// 追加多个元素
db.Model(&PersonaExtra{}).
    Update("personality_tags",
        gorm.Expr("personality_tags || ?", `["reading","gaming"]`),
    )
```

### 2. 删除元素 (`-` 或 `-#` 操作符)
从数组中移除指定元素。

```go
// 删除指定元素 (根据值)
db.Model(&PersonaExtra{}).
    Update("personality_tags",
        gorm.Expr("personality_tags - ?", "reading"), // 删除键名为 reading，或数组中的 "reading" 元素
    )

// 删除指定索引 (针对纯数组)
db.Model(&PersonaExtra{}).
    Update("hobbies", // 假设 hobbies 是纯数组字段 ['a','b','c']
        gorm.Expr("hobbies - ?", 0), // 删除第一个元素
    )
```

### 3. 检查包含关系 (查询示例)
在 `WHERE` 条件中判断数组是否包含某元素。

```go
// 查找包含 "理性" 标签的用户
var personas []PersonaExtra
db.Where(gorm.Expr("personality_tags @> ?", `["理性"]`)).Find(&personas)

// 查找 mood 字段值为 "happy" 的记录
db.Where(gorm.Expr("personality_tags ->> 'mood' = ?", "happy")).Find(&personas)
```

### 4. 更新嵌套数组或对象
假设 JSON 结构如下：
```json
{
  "chat_style": {
    "tone": "serious",
    "topics": ["tech", "news"]
  }
}
```

**更新嵌套对象**：
```go
db.Model(&PersonaExtra{}).
    Update("personality_tags",
        gorm.Expr("jsonb_set(personality_tags, '{chat_style,tone}', ?)", "friendly"),
    )
```

**向嵌套数组追加**：
```go
db.Model(&PersonaExtra{}).
    Update("personality_tags",
        gorm.Expr("jsonb_set(personality_tags, '{chat_style,topics}', personality_tags->'chat_style'->'topics' || ?)", `["sports"]`),
    )
```

---

## 四、最佳实践与注意事项

1.  **索引是性能关键**：对于高频查询的 JSONB 字段，务必创建 **GIN 索引**。
    ```sql
    CREATE INDEX idx_persona_tags ON persona_extras USING GIN (personality_tags);
    ```
2.  **类型必须匹配**：`jsonb_set` 的 `new_value` 参数必须是有效的 JSON。`gorm.Expr` 的 `?` 占位符会自动为字符串添加引号，但数字、布尔值需特殊处理。
    ```go
    // 数字
    gorm.Expr("jsonb_set(data, '{age}', ?)", "25") // 字符串形式的数字
    // 布尔值
    gorm.Expr("jsonb_set(data, '{is_vip}', ?)", "true")
    ```
3.  **处理 NULL**：如果字段可能为 `NULL`，使用 `COALESCE` 防止报错。
    ```go
    gorm.Expr("jsonb_set(COALESCE(personality_tags, '{}'), '{mood}', ?)", "happy")
    ```
4.  **避免滥用**：高频更新或需要强一致性的字段（如余额、状态），不应存放在 JSONB 中，应使用独立列。

---

**总结**：在 GORM 中操作 PostgreSQL JSONB，应始终遵循 **“数据库原生操作”** 原则。利用 `datatypes.JSON` 配合 `gorm.Expr` 调用 `jsonb_set` 及数组操作符，是实现高性能、并发安全、可查询的半结构化数据管理的最佳实践。

需要我为你补充关于 **JSONB GIN 索引的优化策略**，或者演示如何在 **事务** 中安全地组合多个 JSONB 操作吗？