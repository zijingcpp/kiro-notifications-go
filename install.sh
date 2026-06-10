#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INSTALL_DIR="${HOME}/.local/share/kiro-notifications"
BIN_DIR="${HOME}/.local/bin"
AGENT_DIR="${HOME}/.kiro/agents"

DEFAULT_TOOLS='["read", "write", "shell", "grep", "glob", "code", "web_search", "web_fetch", "use_aws", "knowledge", "subagent", "todo_list", "introspect"]'

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

# 安装二进制和资源
mkdir -p "$BIN_DIR" "$INSTALL_DIR/sounds" "$INSTALL_DIR/config"
cp kiro-notifications "$BIN_DIR/"
chmod +x "$BIN_DIR/kiro-notifications"
cp sounds/*.mp3 "$INSTALL_DIR/sounds/" 2>/dev/null || true
cp config/config.json "$INSTALL_DIR/config/"
rm -f "$SCRIPT_DIR/kiro-notifications"
echo "✓ 已安装到 ${BIN_DIR}/kiro-notifications"

# 检查 PATH
if [[ ":$PATH:" != *":${BIN_DIR}:"* ]]; then
    echo "⚠ ${BIN_DIR} 不在 PATH 中，请添加到 shell 配置："
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

# === Agent 配置 ===
mkdir -p "$AGENT_DIR"

# 列出已有 agent
agents=()
for f in "$AGENT_DIR"/*.json; do
    [ -f "$f" ] && agents+=("$f")
done

echo ""
echo "请选择要配置 hook 的 Kiro Agent："
echo ""
i=1
for f in "${agents[@]}"; do
    name=$(python3 -c "import json; print(json.load(open('$f')).get('name','unknown'))" 2>/dev/null)
    echo "  $i) $name ($(basename $f))"
    ((i++))
done
echo "  $i) 新建 agent"
echo ""
read -p "请输入编号 [1]: " choice
choice=${choice:-1}

if [ "$choice" -eq "$i" ] 2>/dev/null; then
    read -p "Agent 名称 [default]: " agent_name
    agent_name=${agent_name:-default}
    AGENT_FILE="${AGENT_DIR}/${agent_name}.json"
    python3 -c "
import json
cfg = {
    'name': '$agent_name',
    'tools': $DEFAULT_TOOLS,
    'allowedTools': ['@builtin'],
    'toolsSettings': {'shell': {'autoAllowReadonly': True}},
    'hooks': {}
}
with open('$AGENT_FILE', 'w') as f:
    json.dump(cfg, f, indent=2, ensure_ascii=False)
"
    echo "✓ 已创建 agent: $AGENT_FILE"
elif [ "$choice" -ge 1 ] 2>/dev/null && [ "$choice" -le "${#agents[@]}" ] 2>/dev/null; then
    AGENT_FILE="${agents[$((choice-1))]}"
else
    echo "无效选择"
    exit 1
fi

# 确保 agent 有 tools 字段
python3 -c "
import json
with open('$AGENT_FILE') as f:
    cfg = json.load(f)
if 'tools' not in cfg:
    cfg['tools'] = $DEFAULT_TOOLS
    with open('$AGENT_FILE', 'w') as f:
        json.dump(cfg, f, indent=2, ensure_ascii=False)
    print('✓ 已为 agent 补充 tools 字段')
"

# 添加 stop hook
HOOK_CMD="${BIN_DIR}/kiro-notifications handle-hook stop &"
python3 -c "
import json
with open('$AGENT_FILE') as f:
    cfg = json.load(f)
hooks = cfg.setdefault('hooks', {}).setdefault('stop', [])
for h in hooks:
    if 'kiro-notifications' in h.get('command', ''):
        print('✓ Hook 已存在，跳过')
        exit(0)
hooks.append({'command': '$HOOK_CMD', 'timeout_ms': 5000})
with open('$AGENT_FILE', 'w') as f:
    json.dump(cfg, f, indent=2, ensure_ascii=False)
print('✓ 已添加 stop hook 到 ' + '$AGENT_FILE')
"

echo ""
echo "=== 安装完成 ==="
echo "请重启 Kiro (新开 kiro chat) 以加载配置"
echo ""
echo "配置文件: ${INSTALL_DIR}/config/config.json"
echo "  - 修改 webhook.enabled=true 和 webhook.url 启用 webhook"
echo "  - 支持 preset: slack / discord / lark / custom"
