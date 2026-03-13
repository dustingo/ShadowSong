# 需求文档

## 介绍

游戏运维 AI 告警系统是一个智能化的告警管理平台，用于统一接收、处理、聚合和分发来自多个数据源的告警信息。系统通过 AI 技术自动分析告警、推断根因、提供处置建议，并通过多种渠道智能推送给相关运维人员。

## 术语表

- **System**: 游戏运维 AI 告警系统
- **Alert**: 告警，来自外部监控系统的异常事件通知
- **Data_Source**: 数据源，外部告警系统（如 Prometheus、Zabbix 等）
- **Template**: Go 模板，用于将原始告警数据转换为统一格式
- **Fingerprint**: 指纹，用于告警去重的唯一标识
- **Group_Key**: 分组键，用于告警聚合的标识
- **Aggregated_Alert**: 聚合告警组，窗口期内同组告警的集合
- **Silence_Rule**: 静默规则，用于临时屏蔽特定告警
- **Route_Rule**: 路由规则，用于将告警分发到不同通知渠道
- **Channel**: 推送渠道，如飞书、钉钉、企业微信等
- **Ack**: 确认，运维人员对告警的处置确认操作
- **On_Duty**: 值班，特定时间段的值班人员配置

## 需求

### 需求 1：数据源管理

**用户故事：** 作为系统管理员，我希望能够配置多个告警数据源并自定义输入输出模板，以便灵活接收任意格式的告警数据并自定义推送格式。

#### 验收标准

1. THE System SHALL 为每个数据源生成唯一的 Webhook 接收地址 `/webhook/{source_name}`
2. WHEN 创建数据源时，THE System SHALL 要求提供唯一的 name、display_name、input_template、output_template 和 group_by_labels
3. WHEN 创建数据源时，THE System SHALL 允许用户使用 Go Template 语法自定义 input_template 以接收任意格式的原始告警数据
4. WHEN 创建数据源时，THE System SHALL 允许用户使用 Go Template 语法自定义 output_template 以定制发送到不同终端的消息格式
5. WHEN 数据源创建后，THE System SHALL 禁止修改 name 字段
6. WHEN 数据源被禁用时，THE System SHALL 对该数据源的 webhook 请求返回 403 状态码
7. WHEN 保存数据源配置时，THE System SHALL 验证 input_template 能够渲染出 alert_name、severity 和 message 三个必填字段
8. WHEN input_template 渲染 severity 字段时，THE System SHALL 验证其值为 P0、P1、P2 或 P3 之一
9. WHERE group_by_labels 为空，THE System SHALL 默认使用 alert_name 作为分组维度
10. WHEN group_by_labels 中指定的 key 在 labels 中不存在时，THE System SHALL 将该 key 的值视为空字符串参与分组计算
11. WHERE output_template 为空，THE System SHALL 使用系统默认的消息格式模板

### 需求 2：模板处理与测试

**用户故事：** 作为系统管理员，我希望能够在线编辑和测试告警转换模板，以便确保模板配置正确。

#### 验收标准

1. THE System SHALL 支持使用 Go Template 语法编写数据源的 input_template 和 output_template
2. THE System SHALL 提供以下自定义模板函数：uuidv4、mapValue、timeFormat、defaultVal、regexExtract、toLower、toUpper
3. WHEN 用户测试 input_template 时，THE System SHALL 使用提供的样本数据渲染模板并返回解析后的统一结构体或错误信息
4. WHEN 用户测试 output_template 时，THE System SHALL 使用提供的告警数据渲染模板并返回格式化后的消息内容或错误信息
5. WHEN input_template 渲染失败时，THE System SHALL 生成降级告警（severity=P1，message=原始数据前500字节，labels为空）
6. WHEN input_template 渲染失败时，THE System SHALL 将原始数据写入死信日志
7. WHEN 保存数据源配置时，THE System SHALL 要求至少通过一次 input_template 测试验证

### 需求 3：告警接收与标准化

**用户故事：** 作为系统，我需要接收来自不同数据源的告警并转换为统一格式，以便后续统一处理。

#### 验收标准

1. WHEN 接收到 webhook 请求时，THE System SHALL 检查对应数据源是否启用
2. WHEN 数据源未启用时，THE System SHALL 返回 403 状态码
3. WHEN 数据源已启用时，THE System SHALL 使用配置的 input_template 渲染原始告警数据
4. WHEN 渲染成功时，THE System SHALL 生成包含以下字段的统一告警结构：alert_id、source、alert_name、severity、message、labels、fingerprint、trigger_time、received_at、status、raw
5. WHEN 生成统一告警时，THE System SHALL 自动生成 alert_id（UUID）、fingerprint、received_at 和 status 字段
6. WHEN 生成统一告警时，THE System SHALL 将原始告警数据完整保存在 raw 字段中
7. WHEN 统一告警生成后，THE System SHALL 将其写入 Redis Stream（alerts:normalized）
8. WHEN 告警写入成功后，THE System SHALL 返回 200 状态码和 alert_id


### 需求 4：告警指纹计算与去重

**用户故事：** 作为系统，我需要对相同的告警进行去重，以避免重复处理和推送。

#### 验收标准

1. WHEN 生成统一告警时，THE System SHALL 计算告警指纹为 SHA256(source + "|" + alert_name + "|" + 按字母序排列的 group_by_labels 键值对)
2. WHEN 计算指纹时，THE System SHALL 仅使用 group_by_labels 中指定的 label 键值对
3. WHEN 计算指纹时，THE System SHALL 按字母序排列参与计算的 label 键值对
4. WHEN 接收到新告警时，THE System SHALL 检查相同指纹的告警是否在去重 TTL 时间内已存在
5. WHEN 相同指纹的告警在去重 TTL 内已存在时，THE System SHALL 将新告警标记为 deduplicated 状态并入库
6. WHEN 告警被去重时，THE System SHALL 将对应活跃告警的触发计数加 1
7. WHEN 告警被去重时，THE System SHALL 不将其送入后续聚合和 AI 处理流程
8. WHERE 去重 TTL 未配置，THE System SHALL 使用默认值 5 分钟

### 需求 5：静默规则匹配

**用户故事：** 作为运维人员，我希望能够临时静默特定的告警，以便在维护期间或已知问题期间减少噪音。

#### 验收标准

1. WHEN 告警通过去重检查后，THE System SHALL 检查是否命中任何活跃的静默规则
2. WHEN 检查静默规则时，THE System SHALL 验证规则的生效时间范围（starts_at 到 ends_at）
3. WHEN 静默规则配置了 source 条件时，THE System SHALL 验证告警的 source 是否精确匹配
4. WHEN 静默规则配置了 alert_name_pattern 条件时，THE System SHALL 验证告警的 alert_name 是否匹配正则表达式
5. WHEN 静默规则配置了 severity 条件时，THE System SHALL 验证告警的 severity 是否在列表中
6. WHEN 静默规则配置了 label_matchers 条件时，THE System SHALL 验证告警的对应 label 值是否匹配正则表达式
7. WHEN 静默规则的所有配置条件都满足时，THE System SHALL 判定告警命中该静默规则
8. WHEN 告警命中静默规则时，THE System SHALL 将告警状态标记为 silenced 并入库
9. WHEN 告警被静默时，THE System SHALL 不将其送入后续聚合、AI 处理和通知流程
10. THE System SHALL 每 30 秒刷新一次静默规则缓存

### 需求 6：窗口聚合

**用户故事：** 作为系统，我需要将短时间内的同类告警聚合在一起，以便批量分析和减少通知频率。

#### 验收标准

1. WHEN 告警通过去重和静默检查后，THE System SHALL 计算分组键为 source + "|" + alert_name + "|" + 按字母序排列的 group_by_labels 键值对
2. WHEN 计算分组键时，THE System SHALL 使用与指纹计算相同的逻辑
3. WHEN 告警到达时，THE System SHALL 将其加入对应分组的聚合窗口
4. WHEN 单个分组的告警数量超过 100 条时，THE System SHALL 丢弃超出部分的告警
5. WHEN 分组最后一条告警到达后超过 10 秒没有新的同组告警时，THE System SHALL 触发该分组的窗口关闭
6. WHEN 分组第一条告警到达后已等待超过 30 秒时，THE System SHALL 触发该分组的窗口关闭
7. WHEN 窗口关闭时，THE System SHALL 将聚合告警组送入 AI 处理队列
8. WHEN 窗口关闭后，THE System SHALL 为新到达的同组告警开启新窗口


### 需求 7：AI 自动分析

**用户故事：** 作为运维人员，我希望系统能够自动分析告警并提供根因推断和处置建议，以便快速响应和处理。

#### 验收标准

1. WHEN 聚合窗口触发时，THE System SHALL 调用 AI 模型分析聚合告警组
2. WHEN 调用 AI 时，THE System SHALL 传入当前聚合告警组的完整信息
3. WHEN 调用 AI 时，THE System SHALL 传入同类告警（同 source + alert_name）最近 10 条已处置记录
4. WHEN AI 调用超过 15 秒未返回时，THE System SHALL 终止调用并执行降级处理
5. WHEN AI 调用失败或超时时，THE System SHALL 使用聚合组中最高 severity 作为最终级别
6. WHEN AI 调用失败或超时时，THE System SHALL 使用第一条告警的 message 作为摘要
7. WHEN AI 返回结果时，THE System SHALL 验证 AI 输出的 severity 不低于原始最高 severity 两个档位
8. WHEN AI 分析完成时，THE System SHALL 输出一句话摘要、根因推断、最终级别、处置建议步骤、是否自愈型告警和标签列表
9. WHEN AI 处理完成时，THE System SHALL 将完整的输入和输出数据入库
10. THE System SHALL 确保 AI 对告警数据只有读权限

### 需求 8：通知路由

**用户故事：** 作为系统管理员，我希望能够配置路由规则，以便将不同类型的告警推送到不同的通知渠道。

#### 验收标准

1. WHEN AI 处理完成后，THE System SHALL 按 priority 从小到大顺序匹配路由规则
2. WHEN 路由规则配置了 severity 条件时，THE System SHALL 验证告警的最终 severity 是否在列表中
3. WHEN 路由规则配置了 source 条件时，THE System SHALL 验证告警的 source 是否在列表中
4. WHEN 路由规则配置了 labels 条件时，THE System SHALL 验证告警的对应 label 值是否匹配正则表达式
5. WHEN 路由规则的所有配置条件都满足时，THE System SHALL 判定告警命中该路由规则
6. WHEN 告警命中路由规则后，THE System SHALL 停止匹配后续规则
7. WHEN 告警未命中任何路由规则且 severity 为 P0 或 P1 时，THE System SHALL 推送到系统默认渠道
8. WHEN 告警未命中任何路由规则且 severity 为 P2 或 P3 时，THE System SHALL 仅入库不推送
9. WHEN 路由规则配置了通知时间段且当前时间不在时间段内时，THE System SHALL 仅入库不推送
10. WHEN 存在当前值班人员时，THE System SHALL 将值班人员的渠道叠加到路由匹配结果上

### 需求 9：推送渠道管理

**用户故事：** 作为系统管理员，我希望能够配置多种推送渠道，以便通过不同的即时通讯工具发送告警通知。

#### 验收标准

1. THE System SHALL 支持飞书、钉钉、企业微信和自定义 Webhook 四种渠道类型
2. WHEN 配置飞书渠道时，THE System SHALL 要求提供 webhook_url 和 secret
3. WHEN 配置钉钉渠道时，THE System SHALL 要求提供 webhook_url 和 secret
4. WHEN 配置企业微信渠道时，THE System SHALL 要求提供 webhook_url
5. WHEN 配置自定义 Webhook 时，THE System SHALL 要求提供 url、method、headers 和 body_template
6. WHEN 推送到飞书时，THE System SHALL 使用 Interactive Card 格式发送消息
7. WHEN 推送到钉钉时，THE System SHALL 使用 ActionCard 格式发送消息
8. WHEN 推送到企业微信时，THE System SHALL 使用 Markdown 格式发送消息
9. WHEN 推送消息时，THE System SHALL 使用数据源配置的 output_template 渲染消息内容
10. WHERE output_template 为空，THE System SHALL 使用系统默认格式包含告警级别、来源、触发时间、AI摘要、根因推断、建议步骤、关联告警数量
11. WHEN 推送到飞书或钉钉时，THE System SHALL 包含「确认处理」和「查看详情」操作按钮

11. WHEN 推送失败时，THE System SHALL 重试 3 次，每次间隔 5 秒
12. WHEN 推送失败 3 次后，THE System SHALL 记录日志并停止重试
13. WHEN 渠道被禁用时，THE System SHALL 跳过该渠道的推送
14. WHEN 自定义 Webhook 的 body_template 为空时，THE System SHALL 发送标准 JSON 格式
15. WHEN 调用自定义 Webhook 超过 10 秒未返回时，THE System SHALL 终止调用

### 需求 10：告警确认

**用户故事：** 作为运维人员，我希望能够确认告警并填写处置备注，以便记录处理过程和结果。

#### 验收标准

1. WHEN 运维人员确认告警时，THE System SHALL 要求提供处置备注
2. WHEN 告警被确认时，THE System SHALL 将状态更新为 acked
3. WHEN 告警被确认时，THE System SHALL 记录操作人、操作时间和备注
4. WHEN 告警被确认时，THE System SHALL 从大盘活跃列表中移除该告警
5. WHEN 告警已被确认时，THE System SHALL 拒绝重复确认并返回提示
6. WHEN 通过消息按钮确认时，THE System SHALL 验证带签名的一次性 token
7. WHEN 确认 token 超过 24 小时时，THE System SHALL 拒绝确认并返回错误
8. THE System SHALL 永久保留告警的确认信息

### 需求 11：静默规则管理

**用户故事：** 作为运维人员，我希望能够创建和管理静默规则，以便在特定时间段屏蔽特定告警。

#### 验收标准

1. WHEN 创建静默规则时，THE System SHALL 要求提供 name、comment、starts_at 和 ends_at
2. WHEN 创建静默规则时，THE System SHALL 允许配置 source、alert_name_pattern、severity 和 label_matchers 作为匹配条件
3. WHEN 静默规则创建后，THE System SHALL 立即生效
4. WHEN 静默规则到期时，THE System SHALL 自动失效
5. WHEN AI 推荐静默规则被忽略时，THE System SHALL 在 7 天内不重复推荐同一规则
6. THE System SHALL 提供 1小时、4小时、今天结束和自定义四种快捷时长选项

### 需求 12：值班管理

**用户故事：** 作为系统管理员，我希望能够配置值班人员和时间段，以便在值班期间自动推送告警给值班人员。

#### 验收标准

1. WHEN 配置值班时，THE System SHALL 要求提供值班人员、时间段和通知渠道
2. THE System SHALL 允许同一时间段配置多个值班人员
3. WHEN 查询当前值班人员时，THE System SHALL 返回所有当前时间段内的值班人员
4. WHEN 推送 P0 或 P1 告警时，THE System SHALL 将当前值班人员的渠道叠加到推送目标中
5. THE System SHALL 缓存当前值班人员查询结果 60 秒

### 需求 13：告警状态流转

**用户故事：** 作为系统，我需要管理告警的生命周期状态，以便跟踪告警的处理进度。

#### 验收标准

1. WHEN 告警进入系统时，THE System SHALL 将状态设置为 pending
2. WHEN 告警推送成功时，THE System SHALL 将状态更新为 firing
3. WHEN 告警命中静默规则时，THE System SHALL 将状态更新为 silenced
4. WHEN 告警被人工确认时，THE System SHALL 将状态更新为 acked
5. WHEN 接收到数据源的 resolved 事件时，THE System SHALL 将对应活跃告警状态更新为 resolved
6. WHEN 告警状态更新为 resolved 时，THE System SHALL 发送恢复通知


### 需求 14：权限控制

**用户故事：** 作为系统管理员，我希望能够控制不同角色的用户权限，以便保护系统配置和数据安全。

#### 验收标准

1. THE System SHALL 支持 admin、editor 和 viewer 三种角色
2. WHEN 用户角色为 admin 时，THE System SHALL 允许所有操作包括用户管理和系统配置
3. WHEN 用户角色为 editor 时，THE System SHALL 允许增删改数据源、渠道、路由规则、静默规则、值班配置和确认告警
4. WHEN 用户角色为 viewer 时，THE System SHALL 仅允许只读查询和确认告警
5. WHEN 用户角色为 viewer 时，THE System SHALL 禁止修改任何配置

### 需求 15：告警大盘

**用户故事：** 作为运维人员，我希望能够在大盘上实时查看当前活跃告警，以便快速了解系统状态。

#### 验收标准

1. THE System SHALL 在大盘顶部展示 P0、P1、P2、P3 各级别当前活跃数量
2. THE System SHALL 通过 WebSocket 实时推送新告警到前端
3. WHEN WebSocket 连接断开时，THE System SHALL 每 3 秒自动重连
4. THE System SHALL 在活跃告警列表中按 severity（P0优先）和 trigger_time（新的优先）排序
5. THE System SHALL 将 P0 告警固定在列表顶部并红色高亮显示
6. THE System SHALL 仅在大盘展示 status=firing 的告警
7. THE System SHALL 展示最近 24 小时告警趋势折线图
8. THE System SHALL 在列表中提供快速确认操作

### 需求 16：告警管理

**用户故事：** 作为运维人员，我希望能够查询和管理所有告警，以便进行详细分析和处理。

#### 验收标准

1. THE System SHALL 展示全量告警列表
2. THE System SHALL 支持按级别、来源、状态、时间和 labels 筛选告警
3. WHEN 点击告警时，THE System SHALL 展开显示原始告警详情和 AI 分析结果
4. THE System SHALL 在告警详情中展示摘要、根因、建议和标签
5. THE System SHALL 在告警详情中展示告警触发计数
6. THE System SHALL 在告警详情中提供确认和快速静默操作
7. WHEN 快速静默时，THE System SHALL 预填充匹配条件

### 需求 17：数据源管理界面

**用户故事：** 作为系统管理员，我希望能够通过界面管理数据源配置，以便方便地添加和维护告警来源。

#### 验收标准

1. THE System SHALL 在数据源列表中展示名称、webhook 地址、分组维度、状态和最近触发时间
2. WHEN 新增或编辑数据源时，THE System SHALL 提供名称、显示名称、input_template 编辑框、output_template 编辑框和 group_by_labels 配置
3. WHEN 编辑 input_template 或 output_template 时，THE System SHALL 提供代码高亮功能
4. THE System SHALL 提供测试区域用于输入原始 JSON 样本并展示 input_template 解析结果
5. THE System SHALL 提供测试区域用于预览 output_template 渲染的消息格式
6. THE System SHALL 提供启用/禁用开关
7. WHEN 禁用数据源时，THE System SHALL 显示提示信息说明 webhook 将停止接收告警

### 需求 18：推送渠道管理界面

**用户故事：** 作为系统管理员，我希望能够通过界面管理推送渠道，以便配置和测试通知方式。

#### 验收标准

1. THE System SHALL 在渠道列表中展示名称、类型和状态
2. WHEN 新增或编辑渠道时，THE System SHALL 根据渠道类型显示对应的配置表单
3. WHEN 展示敏感信息时，THE System SHALL 进行脱敏处理
4. THE System SHALL 提供测试发送按钮用于验证渠道配置


### 需求 19：路由规则管理界面

**用户故事：** 作为系统管理员，我希望能够通过界面管理路由规则，以便灵活配置告警分发策略。

#### 验收标准

1. THE System SHALL 在路由规则列表中展示所有规则
2. THE System SHALL 支持通过拖拽调整规则优先级
3. WHEN 拖拽规则后，THE System SHALL 自动重新编号优先级
4. WHEN 新增或编辑路由规则时，THE System SHALL 提供 severity 多选、source 多选和 labels 条件配置
5. WHEN 新增或编辑路由规则时，THE System SHALL 提供目标渠道多选和通知时间段配置

### 需求 20：静默管理界面

**用户故事：** 作为运维人员，我希望能够通过界面管理静默规则，以便快速创建和取消静默。

#### 验收标准

1. THE System SHALL 提供活跃规则列表和历史规则列表的 Tab 切换
2. WHEN 新增静默时，THE System SHALL 提供名称、备注、匹配条件和时间范围配置
3. THE System SHALL 提供 1小时、4小时、今天结束三个快捷时长按钮
4. THE System SHALL 支持提前取消静默规则

### 需求 21：值班管理界面

**用户故事：** 作为系统管理员，我希望能够通过界面管理值班安排，以便可视化配置值班计划。

#### 验收标准

1. THE System SHALL 以日历视图展示值班排班
2. WHEN 新增排班时，THE System SHALL 提供用户选择、时间段选择和通知渠道选择

### 需求 22：AI 智能问答

**用户故事：** 作为运维人员，我希望能够通过自然语言查询告警信息，以便快速获取统计和分析结果。

#### 验收标准

1. THE System SHALL 提供页面内嵌对话框用于自然语言提问
2. THE System SHALL 支持查询统计、趋势、历史处置记录和跨服务对比
3. THE System SHALL 保留最近 10 轮对话上下文
4. WHEN AI 回答涉及数字结果时，THE System SHALL 优先以图表展示
5. WHEN AI 回答时，THE System SHALL 在末尾注明数据范围
6. THE System SHALL 确保 AI 只读不执行任何写操作

### 需求 23：静默规则推荐

**用户故事：** 作为运维人员，我希望系统能够自动推荐静默规则，以便减少重复性告警的干扰。

#### 验收标准

1. THE System SHALL 每天定时分析历史数据并生成静默规则推荐
2. WHEN 同一指纹告警最近 7 天出现 5 次或以上时，THE System SHALL 考虑推荐静默
3. WHEN 告警历史自愈率达到 80% 或以上时，THE System SHALL 考虑推荐静默
4. WHEN 告警触发时间有明显规律时，THE System SHALL 考虑推荐静默
5. WHEN 推荐条件同时满足时，THE System SHALL 生成静默规则推荐
6. THE System SHALL 以卡片形式展示推荐并预填匹配条件和时间范围
7. THE System SHALL 允许运维人员修改后一键采纳或忽略推荐
8. WHEN 推荐被忽略时，THE System SHALL 在 7 天内不重复推荐同一规则

### 需求 24：处置建议查询

**用户故事：** 作为运维人员，我希望能够查询同类告警的历史处置建议，以便参考过往经验快速处理。

#### 验收标准

1. THE System SHALL 在告警详情页提供「问 AI」按钮
2. WHEN 点击「问 AI」时，THE System SHALL 检索同类告警（同 source + alert_name）最近 20 条已处置记录
3. WHEN 历史记录不足 3 条时，THE System SHALL 注明"历史数据较少"并给出通用建议
4. THE System SHALL 以有序步骤列表展示处置建议


### 需求 25：AI 处理日志

**用户故事：** 作为系统管理员，我希望能够查看和评估 AI 处理的准确性，以便持续优化 AI 模型。

#### 验收标准

1. THE System SHALL 展示每次 AI 自动处理的输入和输出
2. THE System SHALL 允许运维人员标记「准确」或「不准确」并填写反馈
3. THE System SHALL 展示近 7 天和近 30 天的 AI 判断准确率统计
4. THE System SHALL 高亮展示标记为「不准确」的记录

### 需求 26：数据保留

**用户故事：** 作为系统管理员，我希望系统能够自动管理历史数据，以便保持系统性能和存储空间。

#### 验收标准

1. THE System SHALL 保留原始告警数据 90 天
2. THE System SHALL 保留 AI 处理日志 90 天
3. THE System SHALL 保留死信日志 30 天
4. WHEN 已处置告警超过 90 天时，THE System SHALL 将其归档至历史表
5. THE System SHALL 确保归档操作不影响主表查询性能

### 需求 27：消息视觉规范

**用户故事：** 作为运维人员，我希望不同级别的告警有明显的视觉区分，以便快速识别告警严重程度。

#### 验收标准

1. WHEN 展示 P0 级别告警时，THE System SHALL 使用 🔴 Emoji 和"核心服务故障，立即处理"描述
2. WHEN 展示 P1 级别告警时，THE System SHALL 使用 🟠 Emoji 和"重要服务异常，尽快处理"描述
3. WHEN 展示 P2 级别告警时，THE System SHALL 使用 🟡 Emoji 和"一般告警，关注"描述
4. WHEN 展示 P3 级别告警时，THE System SHALL 使用 🔵 Emoji 和"低优先级，了解即可"描述

### 需求 28：前端界面设计

#### 验收标准

1. THE System SHALL 采用浅色系配色方案
2. THE System SHALL 参考现代化 UI 设计风格
3. THE System SHALL 提供响应式布局适配不同屏幕尺寸
4. THE System SHALL 提供流畅的交互动画和过渡效果
