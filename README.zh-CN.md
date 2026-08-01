<p align="center">
  <img src="assets/readme-hero.svg" alt="展示 Alfred 翻译结果的 Alfrenslate 产品示意图" width="100%">
</p>

# Alfrenslate

**Translation at Alfred speed.**

Alfrenslate 是原生集成于 Alfred 5 的多 provider 翻译 Workflow。只需记住 `ts`：中文自动译为自然英文，英文和其他语言自动译为简体中文，并可在 Alfred 中比较多个 provider 的结果。

```text
ts embodied intelligence  →  具身智能
ts 多模态智能体           →  multimodal agent
```

[下载最新 Release](https://github.com/qinchonghanzuibang/Alfrenslate/releases/latest) · 需要 **Alfred 5 Powerpack**

## 为什么选择 Alfrenslate

- 日常只需一个 keyword：`ts`。
- 七个一等 provider，并发请求、稳定排序、互不阻塞。
- provider 完全由用户选择；可以只开一个，也可以同时比较多个。
- 同时包含 macOS arm64 与 x86_64 原生二进制；安装后不需要 Python、Node.js、Homebrew package 或额外 runtime。
- Alfrenslate 没有自己的服务器，不收集 telemetry 或使用统计。

## 工作方式

Alfrenslate 在本机进行 Unicode 级分析，忽略 URL、路径、标点、代码符号、数字与 emoji 等技术噪声，再综合中文字符数、拉丁字母数、中文占比和连续中文片段判断方向。所有 provider 共享同一个目标语言判断；provider 仍负责自动识别源语言，不会额外调用一次付费 API。

所有已启用且配置完整的 provider 并发执行。成功结果优先并按配置顺序显示；失败项随后按同一顺序分别显示。单个 provider 变慢、超时或失败不会隐藏其他结果。

## 安装

1. 从官方 [GitHub Release](https://github.com/qinchonghanzuibang/Alfrenslate/releases/latest) 下载 `Alfrenslate.alfredworkflow`。
2. 建议将 SHA-256 与同一 Release 中的 `Alfrenslate.alfredworkflow.sha256` 对比。
3. 双击导入 Alfred 5，并确认已拥有 Powerpack。
4. 如果 macOS Gatekeeper 首次运行时拦截，请按下面限定范围的安全步骤放行。
5. 打开 Workflow Configuration，至少配置并启用一个 provider。
6. 在 Alfred 输入 `ts hello`。

配置入口：

```text
Alfred Settings
→ Workflows
→ Alfrenslate
→ Configure Workflow
```

### macOS Gatekeeper

当前 Release 二进制没有 Apple Developer ID 签名，Release 也没有通过 Apple notarization（公证）。因此 Apple 并未验证这些二进制，macOS 首次运行时可能显示 **“alfrenslate-arm64” Not Opened**。以下操作只应用于从 Alfrenslate 官方 GitHub Release 下载的 Workflow。

放行前，建议先在 Terminal 校验下载文件：

```bash
shasum -a 256 Alfrenslate.alfredworkflow
```

将完整输出与同一 Release 附带的 `Alfrenslate.alfredworkflow.sha256` 对比。

#### 方法一：通过系统设置放行

1. 运行一次 `ts`，触发 Gatekeeper 拦截提示。
2. 点击 **完成（Done）**，不要点击 **移到废纸篓（Move to Trash）**。
3. 打开 **系统设置 → 隐私与安全性 → 安全性**。
4. 找到 Alfrenslate 被阻止的信息。
5. 点击 **仍要打开（Open Anyway）**。
6. 按系统提示再次确认。

#### 方法二：只移除 Alfrenslate Workflow 文件夹的 quarantine 属性

1. 打开 Alfred Settings，进入 **Workflows → Alfrenslate**。
2. 右键 Alfrenslate，选择 **Open in Finder**。
3. 在 Terminal 输入下面的命令，然后在末尾再输入一个空格：

```bash
xattr -dr com.apple.quarantine
```

4. 将 Finder 中的 Alfrenslate Workflow 文件夹拖入 Terminal。
5. 确认补全后的路径只指向 Alfrenslate Workflow 文件夹，再按回车执行。

完整命令形式如下：

```bash
xattr -dr com.apple.quarantine "/path/to/Alfrenslate"
```

只能对来自官方 Release 的 Alfrenslate Workflow 文件夹执行。不要对整个 Downloads 目录、Home 目录或系统目录递归执行；不要对来源未知的文件执行；不要粘贴或执行自己不理解的路径。

## 配置 provider

只有启用的 provider 才会被调用。credential 保存在本机 Alfred Workflow Configuration。

### DeepSeek

开启 **DeepSeek**，填写 API Key。Base URL 和 Model 通常保持默认：官方 `https://api.deepseek.com` 与低延迟非思考模型 `deepseek-v4-flash`。两者均可修改，以适配未来模型。参见 [DeepSeek API 文档](https://api-docs.deepseek.com/)。

### OpenAI-Compatible

开启 **OpenAI-Compatible**，填写 Base URL、Model 以及服务需要的 API Key；本地免认证服务可以留空 API Key。下列两类地址都会被安全规范化，不会重复拼接 `/v1` 或 `/chat/completions`：

```text
https://api.example.com/v1
https://api.example.com/v1/chat/completions
```

可用于 OpenRouter、SiliconFlow、自托管 vLLM、Ollama 兼容 endpoint、LM Studio 等 Chat Completions 兼容服务；实际兼容性取决于服务和模型。

### DeepL

开启 **DeepL** 并填写 API Key。Free key 选择 `free`，Pro key 选择 `pro`；自定义代理或兼容 endpoint 选择 `custom` 并填写 Custom Base URL。参见 [DeepL Translate API](https://developers.deepl.com/api-reference/translate/request-translation)。

### Youdao

开启 **Youdao**，填写 App Key、App Secret，并选择 `general`、`computers`、`medicine`、`finance` 或 `game` domain。专业领域需要账号开通相应服务。参见 [有道文本翻译文档](https://ai.youdao.com/DOCSIRMA/html/trans/api/wbfy/index.html)。

### Baidu

开启 **Baidu**，填写百度翻译开放平台的 App ID 和 App Secret。参见 [百度通用文本翻译文档](https://fanyi-api.baidu.com/doc/21)。

### Google

开启 **Google**，填写 Google Cloud Translation API Key，并确保项目已启用 Cloud Translation API。Alfrenslate 使用官方 Basic v2 REST API，不使用网页抓取或非公开 endpoint。参见 [Google v2 文档](https://cloud.google.com/translate/docs/reference/rest/v2/translate)。

### Microsoft

开启 **Microsoft Translator**，填写 Azure Subscription Key、资源类型要求的 Region 以及 Endpoint。标准 endpoint 已预填。参见 [Microsoft Translator v3 文档](https://learn.microsoft.com/en-us/azure/ai-services/translator/text-translation/reference/v3/translate)。

## 支持的 provider

| Provider | 认证 | 目标语言 |
|---|---|---|
| DeepSeek | API Key | 简体中文、英文 |
| OpenAI-Compatible | 可选 Bearer Key | 简体中文、英文 |
| DeepL | API Key；Free/Pro/Custom | `ZH-HANS`、`EN-US` |
| Youdao | App Key + App Secret，v3 SHA-256 | `zh-CHS`、`en` |
| Baidu Translate | App ID + App Secret，MD5 签名 | `zh`、`en` |
| Google Cloud Translation | API Key | `zh-CN`、`en` |
| Microsoft Translator | Subscription Key + 可选 Region | `zh-Hans`、`en` |

## 使用

在 Alfred 输入：

```text
ts embodied intelligence
ts 这个实验结果支持我们的主要结论
```

成功项的 title 和 arg 均为纯译文；subtitle 显示 provider、检测到的源语言、目标语言、耗时或缓存状态。

### 高级强制方向

自动检测是主要用法。必要时可强制目标：

```text
ts --to zh Your text
ts --to en 你的文字
ts -t zh Your text
ts -t en 你的文字
```

非法参数会显示可操作的 Alfred 错误项。

## 快捷操作

- **Enter**：复制译文。
- **Command + Enter**：用 Alfred Large Type 显示译文。
- **Option + Enter**：复制译文，并尝试粘贴到当前前台应用。

自动粘贴需要 macOS Accessibility 权限：**System Settings → Privacy & Security → Accessibility**。某些应用会阻止模拟粘贴；Alfrenslate 会先写入剪贴板，因此译文仍可手动粘贴。

## Universal Action

在任意应用选中文字，打开 Alfred Universal Actions，选择 **Translate with Alfrenslate**。它与 `ts` 使用完全相同的方向检测和 provider 比较逻辑。

## Provider 管理

输入 `alfrenslate` 可以查看版本与配置状态、测试所有或单个启用的 provider、清除本地缓存、打开配置、诊断、仓库、Releases 和文档。

状态只显示 Enabled/Disabled、Configured/Missing credentials 和 Reachable/Failed，绝不显示 credential。Provider test 会发送一条很短的真实翻译请求，可能消耗额度。

Workflow Configuration 可调整 provider 顺序、timeout、重试次数、缓存开关与缓存寿命。

## 隐私

Alfrenslate 没有自己的后端服务器，不收集 telemetry 或使用统计。待翻译文本从用户 Mac 直接发送给已启用的第三方 provider；请自行阅读相应隐私政策。

API credential 保存在本机 Alfred Workflow Configuration。开启缓存时，译文可能保存在 Alfred 的本地 Workflow cache directory；cache key 和文件都不包含 credential。可关闭 **Enable local cache**，或通过 `alfrenslate` → **Clear local translation cache** 清除。Alfrenslate 不会将本地缓存上传到自己的服务。

## 故障排除

- **macOS 显示“Not Opened”：**先校验 Release checksum，再按上面的 [Gatekeeper 说明](#macos-gatekeeper)限定范围处理。
- 未启用 provider：至少启用并配置一个。
- 认证失败：重新填写或轮换 key，并核对 endpoint、tier、region。
- Endpoint 或 model 不存在：检查完整 URL 和精确 model ID。
- 限流或额度不足：稍后重试、调整额度或暂时禁用该 provider。
- DeepL tier 错误：Free key 用 `free`，Pro key 用 `pro`。
- 无法自动粘贴：授予 Accessibility 权限；剪贴板通常仍已有译文。
- 本地服务不可达：确认服务已启动且 macOS 可访问对应 URL。

Debug logging 默认关闭，并且不会记录完整 query、译文或 credential。

## 参与贡献

[CONTRIBUTING.md](CONTRIBUTING.md) 包含架构、开发、mock 测试、打包和可选 live test。安全问题请按 [SECURITY.md](SECURITY.md) 私下报告。

## 致谢与声明

单 keyword 交互受到 [alfred-translate-it](https://github.com/yinan-c/alfred-translate-it) 启发。Alfrenslate 是 clean-room 独立实现，没有复制参考项目的代码、Workflow 定义、README 文案、图标、截图或其他资产。

Alfrenslate is an independent community project and is not affiliated with or endorsed by Alfred or Running with Crayons Ltd.

## 许可证

[MIT](LICENSE) © Chonghan Qin
