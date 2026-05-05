#!/bin/bash
# openspec-notify.sh（放在项目根目录，与项目走）

#FEISHU_WEBHOOK="https://open.feishu.cn/open-apis/bot/v2/hook/xxxx"
FEISHU_WEBHOOK="https://open.feishu.cn/open-apis/bot/v2/hook/b0c236ca-9487-41fc-a0c9-90c32fdaeb04"

# 读取覆盖率
COVERAGE=$(cat openspec/reports/coverage-summary.json | jq '.total.lines.pct')
REPORT_URL="https://your-ci-server.com/reports/$(basename openspec/reports/test-report-*.html)"

# 构建卡片消息
curl -X POST $FEISHU_WEBHOOK \
  -H "Content-Type: application/json" \
  -d "{
    \"msg_type\": \"interactive\",
    \"card\": {
      \"header\": {
        \"title\": {\"tag\": \"plain_text\", \"content\": \"🚀 OpenSpec + Superpowers 变更完成\"},
        \"template\": \"green\"
      },
      \"elements\": [
        {\"tag\": \"div\", \"text\": {\"tag\": \"lark_md\", \"content\": \"**覆盖率**: ${COVERAGE}%\"}},
        {\"tag\": \"div\", \"text\": {\"tag\": \"lark_md\", \"content\": \"**审查**: 两阶段审查通过\"}},
        {\"tag\": \"action\", \"actions\": [{\"tag\": \"button\", \"text\": {\"tag\": \"plain_text\", \"content\": \"查看报告\"}, \"url\": \"${REPORT_URL}\", \"type\": \"primary\"}]}
      ]
    }
  }"
