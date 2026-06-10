#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INSTALL_DIR="${HOME}/.local/share/kiro-notifications"
BIN_DIR="${HOME}/.local/bin"
AGENT_DIR="${HOME}/.kiro/agents"
AGENT_FILE="${AGENT_DIR}/default.json"

echo "=== Kiro Notifications 安装 ==="

# 检查 Go
if ! command -v go &>/dev/null; then
    echo "错误: 需要 Go 编译工具"
    echo "  请安装 Go: https://go.dev/dl/"
    exit 1
fi

# 编译
echo "→ 编译 kiro-notifications..."
cd "$SCRIPT_DIR"
go build -o kiro-notifications ./cmd/kiro-notifications/

# 安装二进制
mkdir -p "$BIN_DIR" "$INSTALL_DIR/sounds" "$INSTALL_DIR/config"
cp kiro-notifications "$BIN_DIR/"
chmod +x "$BIN_DIR/kiro-notifications"

# 安装资源
cp sounds/*.mp3 "$INSTALL_DIR/sounds/" 2>/dev/null || true
cp config/config.json "$INSTALL_DIR/config/"
echo "✓ 已安装到 ${BIN_DIR}/kiro-notifications"

# 检查 PATH
if [[ ":$PATH:" != *":${BIN_DIR}:"* ]]; then
    echo "⚠ ${BIN_DIR} 不在 PATH 中，请添加到 shell 配置："
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

# 配置 Kiro agent hook
mkdir -p "$AGENT_DIR"
HOOK_CMD="${BIN_DIR}/kiro-notifications handle-hook stop"

if [ -f "$AGENT_FILE" ]; then
    if grep -q "kiro-notifications" "$AGENT_FILE" 2>/dev/null; then
        echo "✓ Kiro hook 已配置，跳过"
    else
        python3 -c "
import json
with open('$AGENT_FILE') as f:
    cfg = json.load(f)
hook = {'command': '$HOOK_CMD', 'timeout_ms': 30000}
cfg.setdefault('hooks', {}).setdefault('stop', []).append(hook)
with open('$AGENT_FILE', 'w') as f:
    json.dump(cfg, f, indent=2, ensure_ascii=False)
" && echo "✓ 已添加 stop hook 到 ${AGENT_FILE}"
    fi
else
    cat > "$AGENT_FILE" <<EOF
{
  "name": "default",
  "allowedTools": ["@builtin"],
  "hooks": {
    "stop": [
      {
        "command": "${HOOK_CMD}",
        "timeout_ms": 30000
      }
    ]
  }
}
EOF
    echo "✓ 已创建 ${AGENT_FILE}"
fi

# 清理编译产物
rm -f "$SCRIPT_DIR/kiro-notifications"

echo ""
echo "=== 安装完成 ==="
echo "请重启 Kiro (新开 kiro chat) 以加载配置"
echo ""
echo "配置文件: ${INSTALL_DIR}/config/config.json"
echo "  - 修改 webhook.enabled=true 和 webhook.url 启用 webhook"
echo "  - 支持 preset: slack / discord / lark / custom"
echo ""
echo "设置环境变量 KIRO_NOTIFICATIONS_ROOT=${INSTALL_DIR} 以使用安装目录的资源"
