# ne - 快速英语词典工具

## 下载说明

### macOS
- **Intel Mac**: 下载 `*-darwin-amd64.tar.gz` 文件
- **Apple Silicon (M1/M2/M3)**: 下载 `*-darwin-arm64.tar.gz` 文件

### Linux
- **64位 (x86_64)**: 下载 `*-linux-amd64.tar.gz` 文件
- **32位 (x86)**: 下载 `*-linux-386.tar.gz` 文件
- **ARM64**: 下载 `*-linux-arm64.tar.gz` 文件

## 安装步骤

1. **下载所需文件**：
   - `ne-darwin-[架构].tar.gz` - 主程序
   - `kvbuilder-darwin-[架构].tar.gz` - 数据库构建工具
   - `ecdict.csv.xz` - 字典数据文件

2. **解压文件**：
   ```bash
   # 解压程序（根据你的系统和架构选择）
   # macOS 示例
   tar -xzf ne-darwin-arm64.tar.gz
   tar -xzf kvbuilder-darwin-arm64.tar.gz
   
   # Linux 示例
   tar -xzf ne-linux-amd64.tar.gz
   tar -xzf kvbuilder-linux-amd64.tar.gz
   
   # 解压字典数据
   xz -d ecdict.csv.xz
   ```

3. **构建数据库**（首次使用需要）：
   ```bash
   ./kvbuilder --csv ecdict.csv
   ```
   这会在当前目录生成 `ecdict.bbolt` 数据库文件。

4. **移动到系统路径**（可选）：
   ```bash
   sudo mv ne /usr/local/bin/
   sudo mv kvbuilder /usr/local/bin/
   mv ecdict.bbolt ~/.cache/ne/
   ```

## 使用方法

```bash
# 基本查询
ne hello

# JSON 格式输出
ne --json hello

# 显示所有字段
ne --full hello

# 指定数据库路径
ne --dbpath /path/to/ecdict.bbolt hello
```

## 功能特性

- 🚀 离线查询，速度极快
- 🔍 智能模糊搜索，自动纠正拼写错误
- 📊 支持表格和 JSON 输出格式
- 🗂️ 包含 77 万+ 词条（ECDICT 数据源）

## 故障排除

### macOS
如果遇到"无法打开"的安全提示：
```bash
# 移除隔离属性
xattr -d com.apple.quarantine ne
xattr -d com.apple.quarantine kvbuilder
```

### Linux
如果遇到权限问题：
```bash
# 添加执行权限
chmod +x ne
chmod +x kvbuilder
```

## 更新日志

请查看 [Release Notes](#) 了解此版本的更新内容。