# restricted-runner 设计文档

## 状态

Draft

## 1. 项目目标

`restricted-runner` 是一个面向受限执行场景的独立工具，目标是在 self-hosted runner 所在环境中，提供一条**可验证、可审计、可收敛权限边界**的宿主机执行路径。

它要解决的核心问题不是“怎么远程执行命令”，而是：

- 如何避免把宿主机的高权限直接交给 CI runner
- 如何把宿主机执行限制为一组明确允许的操作
- 如何让这组操作具备严格参数校验、可审计日志和结构化结果
- 如何让上层系统，例如 HomeCloud，将它作为一个独立组件接入，而不是把这些安全边界逻辑散落在业务仓库里

## 2. 背景

在 HomeCloud 的 site deploy 场景里，计划从旧的 timer-driven deploy agent 迁移到 GitHub Actions self-hosted runner 模式。

但直接让 runner 容器拥有宿主机广泛执行权限是不可接受的。runner 可能受到 workflow 输入影响，如果它能直接拿到高权限宿主执行能力，风险边界会非常差。

因此，需要一个单独的受限执行组件来承接如下职责：

- 接收受限格式的执行请求
- 校验请求是否合法
- 校验请求是否命中允许的策略
- 将允许的请求映射到宿主机上的有限动作
- 输出统一的结构化结果
- 保留完整的审计线索

## 3. 非目标

当前阶段，这个项目**不负责**：

- 实现完整的 GitHub Actions runner 容器
- 实现业务仓库自己的 deploy 策略细节
- 提供通用远程管理 shell
- 成为任意命令执行代理
- 直接处理复杂编排系统的全部生命周期

换句话说，它不是远控框架，也不是通用 SSH 包装器。

## 4. 技术栈

### 选型

项目使用 **Golang**，并跟随当前最新稳定版。

当前仓库初始化目标版本：

- Go `1.24.0`

### 选择 Go 的原因

- 单二进制发布，适合部署到宿主机和受限环境
- 运行时依赖少，便于以最小依赖方式落地
- 适合写 CLI、协议校验、策略检查和结构化日志
- 测试与交叉编译体验较好
- 更适合把这个项目做成长期独立工具，而不是业务仓库内的脚本集合

## 5. 架构总览

系统中计划存在四类边界角色：

1. **上层编排方**
   - 例如 GitHub Actions workflow
   - 负责发起“意图”，不直接拥有宿主机权限

2. **runner 执行环境**
   - 例如 self-hosted runner 容器
   - 负责接收 workflow 任务，调用 restricted-runner
   - 本身不应持有广泛宿主机权限

3. **restricted-runner**
   - 本项目
   - 负责请求解析、策略校验、命令分发、结果输出、审计记录

4. **宿主机受限动作层**
   - 被 restricted-runner 调用的本地执行动作
   - 必须是显式白名单动作，而不是任意 shell

## 6. 核心设计原则

### 6.1 默认拒绝

所有请求默认拒绝，只有显式允许的 operation 才可以执行。

### 6.2 参数必须结构化

不接受拼接 shell 字符串作为协议输入。
必须使用结构化字段描述请求，例如：

- operation
- target
- resource
- revision
- metadata

### 6.3 执行与策略分层

“能不能做”与“怎么做”是两层：

- policy layer 负责判断请求是否允许
- executor layer 负责把允许的请求映射为受控动作

### 6.4 结果必须结构化

所有执行结果必须输出统一 JSON 结构，至少包含：

- ok
- operation
- target
- resource
- exit_code
- stdout
- stderr
- audit metadata

### 6.5 可审计

每次请求都必须具备日志和关联信息，至少能追踪：

- 谁发起
- 请求了什么
- 命中了哪条策略
- 最终是否执行
- 执行结果是什么

## 7. 逻辑分层

计划中的代码结构建议如下：

- `cmd/restricted-runner/`
  - CLI 入口
- `internal/protocol/`
  - 请求与响应结构
  - JSON decode/encode
  - 字段合法性校验
- `internal/policy/`
  - allowlist / deny by default
  - target、resource、operation 匹配规则
- `internal/dispatch/`
  - 将请求路由到具体 handler
- `internal/executor/`
  - 受控执行器
  - 不暴露任意 shell
- `internal/audit/`
  - 审计日志输出
- `internal/config/`
  - 策略和运行配置加载

## 8. 请求模型

第一版请求模型建议至少包含：

```json
{
  "operation": "resource.apply",
  "target": "homecloud-server",
  "resource": "sites/homes/ruyi/hass",
  "revision": "abcdef123456",
  "metadata": {
    "request_id": "...",
    "workflow_run_id": "...",
    "actor": "..."
  }
}
```

### 字段说明

- `operation`
  - 请求的动作类型
  - 例如 `resource.validate`、`resource.apply`、`repo.checkout`
- `target`
  - 目标主机或逻辑目标
- `resource`
  - 被操作的受控资源标识
- `revision`
  - 版本、提交号或其他可验证引用
- `metadata`
  - 审计和关联信息，不直接决定权限

## 9. 策略模型

第一版策略建议采用**静态配置 + 默认拒绝**。

例如：

- 允许哪些 `operation`
- 某个 `target` 允许哪些 `resource`
- 某个 `resource` 允许哪些动作
- 某些动作是否必须携带 `revision`

策略配置不应允许任意 shell 模板。
策略的职责应是：

- 允许或拒绝请求
- 选出受控 handler
- 提供有限的执行参数映射

## 10. 执行模型

restricted-runner 第一版建议只支持**显式注册的 handler**，例如：

- `repo.checkout`
- `resource.validate`
- `resource.apply`
- `resource.status`
- `resource.logs`

这些 handler 背后可以调用：

- 固定二进制
- 固定脚本路径
- 固定子命令

但不能直接把用户输入拼成 shell 再执行。

## 11. CLI 形态

第一版 CLI 可以先提供：

### `restricted-runner dispatch`

输入一个 JSON payload，执行完整流程：

1. parse
2. validate
3. policy check
4. dispatch
5. emit structured result

### `restricted-runner validate`

只校验请求与策略，不真正执行。

### `restricted-runner version`

输出版本信息。

## 12. 配置形态

建议支持：

- `--config <path>` 指定配置文件
- `--payload <json>` 直接传入请求
- `--payload-file <path>` 从文件读取请求
- stdin 读取请求

配置文件格式优先考虑：

- YAML
- 或 JSON

倾向 YAML，便于人工维护。

## 13. 安全要求

### 必须满足

- 默认拒绝
- 不支持任意 shell passthrough
- 不允许路径逃逸
- 不允许未注册 operation
- 不允许 target/resource 绕过 allowlist
- 日志中避免无意泄露敏感信息

### 后续增强

- 审计日志持久化
- 更严格的 target identity 绑定
- 细粒度 policy match reason
- 结果签名或更强的完整性机制

## 14. 与 HomeCloud 的关系

HomeCloud 是第一批消费者之一，但 `restricted-runner` 不应在模型上强绑定 HomeCloud。

这意味着：

- 协议字段要尽量中性
- 不把 `site` 作为唯一一等概念
- 不把 HomeCloud 路径结构写死在协议层
- HomeCloud 特定的资源命名和 handler 映射可以在接入层实现

## 15. 第一阶段交付目标

第一阶段不追求功能完整，而追求把边界做对。

### Phase 1

- 初始化 Go 项目
- 定义请求/结果协议结构
- 定义 policy 配置结构
- 实现最小 CLI 框架
- 实现 parse + validate
- 实现最小 dispatcher
- 为协议和策略写单元测试

### Phase 2

- 实现静态 policy allowlist
- 实现受控 handler 注册机制
- 实现 dry-run 模式
- 输出结构化 JSON 结果
- 增加审计日志

### Phase 3

- 接入 HomeCloud 进行真实 PoC
- 明确 target/resource/revision 映射
- 评估 transport 和部署方式

## 16. 当前结论

我们现在先不急着实现复杂执行逻辑。
第一步要把以下东西做扎实：

- Go 项目基础结构
- 设计文档
- 中性协议模型
- policy 与 dispatch 的边界

这样后面再接具体执行器、SSH restricted command、HomeCloud 接入时，才不会反复推倒重来。
