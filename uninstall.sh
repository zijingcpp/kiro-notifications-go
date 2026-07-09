#!/bin/bash
set -e

BIN_DIR="${HOME}/.local/bin"
INSTALL_DIR="${HOME}/.local/share/kiro-notifications"
AGENT_DIR="${HOME}/.kiro/agents"

echo "=== Kiro Notifications 卸载 ==="

# 删除二进制
if [ -f "${BIN_DIR}/kiro-notifications" ]; then
    rm "${BIN_DIR}/kiro-notifications"
    echo "✓ 已删除 ${BIN_DIR}/kiro-notifications"
else
    echo "- 二进制未安装，跳过"
fi

# 删除资源
if [ -d "$INSTALL_DIR" ]; then
    rm -rf "$INSTALL_DIR"
    echo "✓ 已删除 ${INSTALL_DIR}"
fi

# 从所有 agent 中移除 stop hook
for f in "$AGENT_DIR"/*.json; do
    [ -f "$f" ] || continue
    if grep -q "kiro-notifications" "$f" 2>/dev/null; then
        python3 -c "
import json
with open('$f') as fp:
    cfg = json.load(fp)
hooks = cfg.get('hooks', {}).get('stop', [])
cfg['hooks']['stop'] = [h for h in hooks if 'kiro-notifications' not in h.get('command', '')]
if not cfg['hooks']['stop']:
    del cfg['hooks']['stop']
if not cfg['hooks']:
    del cfg['hooks']
with open('$f', 'w') as fp:
    json.dump(cfg, fp, indent=2, ensure_ascii=False)
" && echo "✓ 已从 $(basename $f) 移除 hook"
    fi
done

echo ""
echo "=== 卸载完成 ==="
echo "请重启 Kiro 以生效"
