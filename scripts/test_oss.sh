#!/bin/bash

API_URL="http://localhost:9527"
TEST_FILE="test_oss_upload.txt"

# 创建一个测试文件
echo "This is a test file for OSS upload at $(date)" > "$TEST_FILE"

echo "=== 开始 OSS 上传测试 ==="
echo "1. 上传文件到 $API_URL/api/public/resources ..."

# 使用 curl 上传文件
RESPONSE=$(curl -s -X POST "$API_URL/api/public/resources" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@$TEST_FILE" \
  -F "title=OSSTestFile" \
  -F "description=Testing OSS upload" \
  -F "type=document")

# 打印原始响应
echo -e "\n服务器响应:"
if command -v jq &> /dev/null; then
    echo "$RESPONSE" | jq .
else
    echo "$RESPONSE"
fi

# 检查响应中是否包含 OSS 域名
if [[ "$RESPONSE" == *"aliyuncs.com"* ]]; then
    echo -e "\n✅ 测试通过: 响应中包含 aliyuncs.com，说明文件已上传至 OSS。"
    
    # 尝试提取 URL (简单提取)
    FILE_URL=$(echo "$RESPONSE" | grep -o 'https://[^"]*aliyuncs.com[^"]*')
    if [ -n "$FILE_URL" ]; then
        echo "文件链接: $FILE_URL"
    fi
else
    echo -e "\n⚠️  测试警告: 响应中未检测到 OSS 域名。请检查服务器日志确认是否使用了 LocalStorage。"
fi

# 清理测试文件
rm "$TEST_FILE"
