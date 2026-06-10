# kiro-notifications-go

Kiro CLI 任务完成后自动发送桌面通知 + 声音提醒。支持 macOS 和 Linux。

Fork from [claude-notifications-go](https://github.com/777genius/claude-notifications-go)，适配 Kiro CLI hook 系统。

## 效果预览

![通知效果](notify_test.png)

## 功能

- ✅ 任务完成通知（区分类型：完成 / 提问 / 审查 / 错误）
- 🔊 不同状态播放不同声音
- 🔔 桌面通知（macOS / Linux 跨平台）
- 🌐 Webhook 推送（Lark / Slack / Discord / 自定义）
- 🧠 自动解析 Kiro session JSONL 判断任务状态

## 安装

```bash
git clone git@github.com:zijingcpp/kiro-notifications-go.git
cd kiro-notifications-go
./install.sh
```

安装脚本会：
1. 编译二进制到 `~/.local/bin/kiro-notifications`
2. 复制声音文件和配置到 `~/.local/share/kiro-notifications/`
3. 在 `~/.kiro/agents/default.json` 中添加 `stop` hook

安装完成后重启 Kiro（新开 `kiro chat`）即可生效。

### 依赖

- Go 1.21+（编译）
- Linux: `libnotify-bin`（notify-send）、音频播放器（mpv / ffplay / gst-play-1.0 任一）
- macOS: 无额外依赖

## 工作原理

```
Kiro 任务结束 → stop hook 触发
    ↓
kiro-notifications handle-hook stop
    ↓
查找最新 ~/.kiro/sessions/cli/*.jsonl
    ↓
解析 session → 分析状态（task_complete / question / review_complete / error）
    ↓
桌面通知 + 声音 + webhook
```

### 状态判断逻辑

| 状态 | 条件 |
|------|------|
| `task_complete` | 使用了 write / shell / use_aws 等 active tool |
| `question` | 最后一条 assistant 消息以 `?` 结尾 |
| `review_complete` | 仅使用 read / grep / glob 等 passive tool，且回复 >200 字符 |
| `error` | ToolResults 中出现非 success 状态 |

## 配置

配置文件位于 `~/.local/share/kiro-notifications/config/config.json`：

```json
{
  "desktop": {
    "enabled": true,
    "sound": true
  },
  "webhook": {
    "enabled": false,
    "url": "",
    "preset": "lark"
  }
}
```

### Webhook Preset

| Preset | 说明 |
|--------|------|
| `slack` | Slack Incoming Webhook |
| `discord` | Discord Webhook |
| `lark` | 飞书 / Lark Bot Webhook |
| `custom` | 自定义 JSON 格式 |

### 飞书 Webhook 示例

```json
{
  "webhook": {
    "enabled": true,
    "url": "https://open.larksuite.com/open-apis/bot/v2/hook/your-token",
    "preset": "lark"
  }
}
```

## 调试

```bash
# 启用 debug 日志手动触发
KIRO_NOTIFY_DEBUG=1 KIRO_NOTIFICATIONS_ROOT=~/.local/share/kiro-notifications \
  kiro-notifications handle-hook stop
```

## 卸载

```bash
rm -f ~/.local/bin/kiro-notifications
rm -rf ~/.local/share/kiro-notifications
# 从 ~/.kiro/agents/default.json 的 hooks.stop 数组中移除 kiro-notifications 条目
```

## 项目结构

```
├── cmd/kiro-notifications/   # CLI 入口
├── config/                   # 默认配置
├── internal/
│   ├── analyzer/             # Session 状态分析
│   ├── config/               # 配置加载
│   ├── dedup/                # 通知去重
│   ├── hooks/                # Kiro stop hook 处理
│   ├── logging/              # 日志
│   ├── notifier/             # 桌面通知 + 声音播放
│   ├── summary/              # 消息摘要生成
│   └── webhook/              # Webhook 发送
├── pkg/jsonl/                # Kiro session JSONL 解析器
├── sounds/                   # 通知声音（mp3）
├── install.sh                # 一键安装
└── README.md
```

## License

MIT
