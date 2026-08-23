---
title: flutter学习(三)
date: 2026-08-22
tags: [dart,flutter]
draft: false
---
# Flutter 状态同步问题总结：修改设备名称后 UI 不更新

## 1. 为什么会有这个问题

在 Flutter 中，**页面之间传递参数是"快照"传递**。当通过 `Navigator.push` 打开子页面时，传递的 `device` 对象被"冻结"在子页面中。即使父页面数据源刷新了，子页面也不会感知到变化，因为：

- `Navigator.push` 是一次性路由，不会自动重建已打开的页面
- Flutter 是**声明式 UI**，UI 不会自动随数据变化，除非显式监听数据变化
- `DeviceItem` 是一个**不可变对象**（所有字段都是 `final`），不能直接修改属性

## 2. 这个问题是如何产生的

具体流程如下：

```
用户点击"修改设备名称" → 弹出对话框 → 输入新名称 → 确认
  ↓
调用 API 成功更新到服务器
  ↓
子页面直接修改了 item.deviceName = newName   ← ❌ 不可变对象不能直接修改
  ↓
调用 onRefresh 回调通知父页面
  ↓
父页面更新了 refreshTrigger.value = item     ← ❌ 同一个引用，ValueNotifier 不会触发
  ↓
父页面 _refresh() 从服务器重新获取数据        ← ✅ 父页面列表更新了
  ↓
但子页面没有变化                            ← ❌ 子页面 widget.device 还是旧值，且没有监听变化
```

关键问题有三个：

1. **直接修改了不可变对象的属性** — `DeviceItem` 字段都是 `final`
2. **`ValueNotifier` 需要新引用才能触发** — 同一个对象赋值不会触发 listener
3. **子页面没有监听 `ValueNotifier`** — 用了 `widget.device.value` 读取值，但没有监听变化

## 3. 这个问题的解决办法

**核心思路**：使用 `ValueNotifier<T>` 作为共享状态源，让父页面和子页面监听同一个 notifier。

### 解决办法分三步

#### 第一步：父页面创建 `ValueNotifier` 并传递给子页面

```dart
// 父页面
final refreshTrigger = ValueNotifier(DeviceItem.empty());

// 选择设备时
refreshTrigger.value = device; // 设置初始值
Navigator.push(
  MaterialPageRoute(
    builder: (_) => DeviceConfigMenuPage(
      device: refreshTrigger, // 传递同一个 notifier
    ),
  ),
);
```

#### 第二步：子页面用 `ValueListenableBuilder` 监听变化

```dart
// 子页面
@override
Widget build(BuildContext context) {
  return ValueListenableBuilder<DeviceItem>(
    valueListenable: widget.device,
    builder: (context, currentDevice, child) {
      return Scaffold(
        appBar: AppBar(
          title: Text(currentDevice.deviceName), // ← 使用 builder 中的值
        ),
      );
    },
  );
}
```

#### 第三步：修改数据时创建新对象 + 更新 notifier

```dart
// 修改成功后
final updatedDevice = item.copyWith(deviceName: newName); // 创建新对象
widget.onRefresh?.call(updatedDevice); // 通知父页面
// 父页面收到后：refreshTrigger.value = updatedDevice → 触发 listener
```

## 4. 最后我是如何解决的

**最终方案**：使用 `ValueNotifier<DeviceItem>` 作为共享状态源。

### 父页面

```dart
// 定义共享的 ValueNotifier
final refreshTrigger = ValueNotifier(DeviceItem.empty());

// 监听变化，自动刷新列表
refreshTrigger.addListener(() {
  _refresh();      // 刷新设备列表
  setState(() {}); // 重建 UI
});

// 选择设备时传递同一个 notifier
void _selectDevice(DeviceItem device) {
  refreshTrigger.value = device; // 设置初始值
  Navigator.push(
    MaterialPageRoute(
      builder: (_) => DeviceConfigMenuPage(
        device: refreshTrigger, // 共享同一个 notifier
        onRefresh: (newDevice) {
          refreshTrigger.value = newDevice; // 更新触发刷新
        },
      ),
    ),
  );
}
```

### 子页面

```dart
// 参数改为 ValueNotifier
final ValueNotifier<DeviceItem> device;

// 用 ValueListenableBuilder 监听变化
@override
Widget build(BuildContext context) {
  return ValueListenableBuilder<DeviceItem>(
    valueListenable: widget.device,
    builder: (context, currentDevice, child) {
      return Scaffold(
        appBar: AppBar(
          title: Text(currentDevice.deviceName), // 自动更新
        ),
      );
    },
  );
}

// 修改名称时创建新对象
Future<void> _changeDeviceName(DeviceItem item) async {
  // ... 弹出对话框获取新名称 ...
  await api.updateDevice(item.macId, newName);
  widget.onRefresh?.call(item.copyWith(deviceName: newName)); // 传新对象
}
```

## 关键收获

- `ValueNotifier` 只有在 `value` 被**重新赋值**（新引用）时才触发 listener
- 不可变对象必须用 `copyWith` 创建新对象，不能直接修改属性
- `ValueListenableBuilder` 是监听 `ValueNotifier` 变化并自动重建 UI 的标准方式
- 共享同一个 `ValueNotifier` 实例才能在父页面和子页面之间同步状态
