-- V017: 为所有表字段添加注释，方便数据库管理工具查看

-- ===== tenants 租户表 =====
COMMENT ON COLUMN tenants.id IS '租户ID';
COMMENT ON COLUMN tenants.name IS '租户名称/公司名称';
COMMENT ON COLUMN tenants.plan_type IS '套餐类型: free/pro/enterprise';
COMMENT ON COLUMN tenants.plan_expires_at IS '套餐到期时间';
COMMENT ON COLUMN tenants.config IS '扩展配置，JSON格式';
COMMENT ON COLUMN tenants.status IS '状态: active-正常, banned-封禁';
COMMENT ON COLUMN tenants.created_at IS '创建时间';
COMMENT ON COLUMN tenants.updated_at IS '更新时间';
COMMENT ON COLUMN tenants.deleted_at IS '删除时间（软删除）';
COMMENT ON COLUMN tenants.audit_status IS '审核状态: pending-待审核, approved-已通过, rejected-已驳回';
COMMENT ON COLUMN tenants.audit_remark IS '审核备注';
COMMENT ON COLUMN tenants.audit_at IS '审核时间';
COMMENT ON COLUMN tenants.audit_by IS '审核人ID';
COMMENT ON COLUMN tenants.banned_at IS '封禁时间';
COMMENT ON COLUMN tenants.banned_reason IS '封禁原因';
COMMENT ON COLUMN tenants.banned_by IS '封禁操作人ID';
COMMENT ON COLUMN tenants.brand_name IS '品牌名称（白标）';
COMMENT ON COLUMN tenants.brand_logo_url IS '品牌Logo URL';
COMMENT ON COLUMN tenants.brand_favicon_url IS '品牌Favicon URL';
COMMENT ON COLUMN tenants.brand_primary_color IS '品牌主色调';
COMMENT ON COLUMN tenants.brand_secondary_color IS '品牌辅助色';
COMMENT ON COLUMN tenants.brand_footer_text IS '品牌页脚文字';
COMMENT ON COLUMN tenants.white_label_enabled IS '是否启用白标';
COMMENT ON COLUMN tenants.code IS '租户编码（唯一标识）';

-- ===== users 用户表 =====
COMMENT ON COLUMN users.id IS '用户ID';
COMMENT ON COLUMN users.tenant_id IS '所属租户ID';
COMMENT ON COLUMN users.username IS '用户名（唯一）';
COMMENT ON COLUMN users.phone IS '手机号';
COMMENT ON COLUMN users.email IS '邮箱';
COMMENT ON COLUMN users.password_hash IS '密码哈希（bcrypt）';
COMMENT ON COLUMN users.wx_open_id IS '微信OpenID（小程序绑定）';
COMMENT ON COLUMN users.name IS '用户姓名';
COMMENT ON COLUMN users.avatar_url IS '头像URL';
COMMENT ON COLUMN users.role IS '角色: super_admin/admin/member';
COMMENT ON COLUMN users.status IS '状态: active-正常, banned-禁用';
COMMENT ON COLUMN users.last_login_at IS '最后登录时间';
COMMENT ON COLUMN users.created_at IS '创建时间';
COMMENT ON COLUMN users.updated_at IS '更新时间';
COMMENT ON COLUMN users.deleted_at IS '删除时间（软删除）';

-- ===== agents 智能体表 =====
COMMENT ON COLUMN agents.id IS '智能体ID';
COMMENT ON COLUMN agents.tenant_id IS '所属租户ID';
COMMENT ON COLUMN agents.name IS '智能体名称';
COMMENT ON COLUMN agents.description IS '智能体描述';
COMMENT ON COLUMN agents.avatar_url IS '头像URL';
COMMENT ON COLUMN agents.status IS '状态: idle-空闲, processing-处理中, error-异常';
COMMENT ON COLUMN agents.agent_type IS '智能体类型: chat/task/assistant';
COMMENT ON COLUMN agents.personality_config IS '人设配置，JSON格式';
COMMENT ON COLUMN agents.workflow_config IS '工作流配置，JSON格式';
COMMENT ON COLUMN agents.welcome_message IS '欢迎语';
COMMENT ON COLUMN agents.work_hours IS '工作时间，JSON格式';
COMMENT ON COLUMN agents.model_id IS '绑定的AI模型ID';
COMMENT ON COLUMN agents.max_concurrent IS '最大并发会话数';
COMMENT ON COLUMN agents.current_sessions IS '当前会话数';
COMMENT ON COLUMN agents.total_conversations IS '累计对话数';
COMMENT ON COLUMN agents.total_messages IS '累计消息数';
COMMENT ON COLUMN agents.created_at IS '创建时间';
COMMENT ON COLUMN agents.updated_at IS '更新时间';
COMMENT ON COLUMN agents.deleted_at IS '删除时间（软删除）';
COMMENT ON COLUMN agents.knowledge_base_ids IS '关联知识库ID列表，JSON数组';
COMMENT ON COLUMN agents.is_public IS '是否公开';
COMMENT ON COLUMN agents.required_plan IS '所需最低套餐: free/pro/enterprise';

-- ===== conversations 会话表 =====
COMMENT ON COLUMN conversations.id IS '会话ID';
COMMENT ON COLUMN conversations.tenant_id IS '租户ID';
COMMENT ON COLUMN conversations.user_id IS '用户ID';
COMMENT ON COLUMN conversations.agent_id IS '智能体ID';
COMMENT ON COLUMN conversations.channel IS '渠道: web, wechat, feishu, dingtalk, wecom, douyin';
COMMENT ON COLUMN conversations.channel_id IS '渠道标识';
COMMENT ON COLUMN conversations.customer_id IS '客户ID';
COMMENT ON COLUMN conversations.customer_name IS '客户名称';
COMMENT ON COLUMN conversations.customer_avatar IS '客户头像URL';
COMMENT ON COLUMN conversations.title IS '会话标题';
COMMENT ON COLUMN conversations.status IS '状态: active-进行中, closed-已关闭';
COMMENT ON COLUMN conversations.priority IS '优先级';
COMMENT ON COLUMN conversations.tags IS '标签，JSON数组';
COMMENT ON COLUMN conversations.assigned_to IS '分配给的坐席ID';
COMMENT ON COLUMN conversations.last_message_at IS '最后消息时间';
COMMENT ON COLUMN conversations.message_count IS '消息数量';
COMMENT ON COLUMN conversations.created_at IS '创建时间';
COMMENT ON COLUMN conversations.updated_at IS '更新时间';
COMMENT ON COLUMN conversations.deleted_at IS '删除时间（软删除）';

-- ===== messages 消息表 =====
COMMENT ON COLUMN messages.id IS '消息ID';
COMMENT ON COLUMN messages.conversation_id IS '会话ID';
COMMENT ON COLUMN messages.tenant_id IS '租户ID';
COMMENT ON COLUMN messages.sender_id IS '发送者ID';
COMMENT ON COLUMN messages.sender_type IS '发送者类型: user, agent, assistant, system';
COMMENT ON COLUMN messages.sender_name IS '发送者名称';
COMMENT ON COLUMN messages.content_type IS '内容类型: text, image, file, card';
COMMENT ON COLUMN messages.content IS '消息内容';
COMMENT ON COLUMN messages.metadata IS '元数据，JSON格式';
COMMENT ON COLUMN messages.status IS '状态: sending-发送中, sent-已发送, failed-失败, read-已读';
COMMENT ON COLUMN messages.model_used IS '使用的AI模型';
COMMENT ON COLUMN messages.tokens_used IS '消耗的token数';
COMMENT ON COLUMN messages.latency_ms IS '响应延迟（毫秒）';
COMMENT ON COLUMN messages.created_at IS '创建时间';
COMMENT ON COLUMN messages.updated_at IS '更新时间';

-- ===== channels 渠道表 =====
COMMENT ON COLUMN channels.id IS '渠道ID';
COMMENT ON COLUMN channels.tenant_id IS '租户ID';
COMMENT ON COLUMN channels.name IS '渠道名称';
COMMENT ON COLUMN channels.type IS '渠道类型: wechat, feishu, dingtalk, wecom, douyin, web';
COMMENT ON COLUMN channels.config IS '渠道配置，JSON格式';
COMMENT ON COLUMN channels.status IS '状态: active-启用, inactive-停用';
COMMENT ON COLUMN channels.webhook_secret IS 'Webhook密钥';
COMMENT ON COLUMN channels.created_at IS '创建时间';
COMMENT ON COLUMN channels.updated_at IS '更新时间';
COMMENT ON COLUMN channels.deleted_at IS '删除时间（软删除）';

-- ===== knowledge_bases 知识库表 =====
COMMENT ON COLUMN knowledge_bases.id IS '知识库ID';
COMMENT ON COLUMN knowledge_bases.tenant_id IS '所属租户ID';
COMMENT ON COLUMN knowledge_bases.name IS '知识库名称';
COMMENT ON COLUMN knowledge_bases.description IS '知识库描述';
COMMENT ON COLUMN knowledge_bases.status IS '状态: active-正常, indexing-索引中';
COMMENT ON COLUMN knowledge_bases.doc_count IS '文档数量';
COMMENT ON COLUMN knowledge_bases.chunk_count IS '分段数量';
COMMENT ON COLUMN knowledge_bases.total_size IS '总大小（字节）';
COMMENT ON COLUMN knowledge_bases.config IS '配置，JSON格式';
COMMENT ON COLUMN knowledge_bases.created_at IS '创建时间';
COMMENT ON COLUMN knowledge_bases.updated_at IS '更新时间';
COMMENT ON COLUMN knowledge_bases.deleted_at IS '删除时间（软删除）';

-- ===== plans 套餐表 =====
COMMENT ON COLUMN plans.id IS '套餐ID';
COMMENT ON COLUMN plans.name IS '套餐名称';
COMMENT ON COLUMN plans.description IS '套餐描述';
COMMENT ON COLUMN plans.price IS '价格';
COMMENT ON COLUMN plans.period IS '计费周期: month/year';
COMMENT ON COLUMN plans.max_agents IS '最大智能体数量';
COMMENT ON COLUMN plans.max_messages IS '最大消息数';
COMMENT ON COLUMN plans.features IS '功能列表，JSON数组';
COMMENT ON COLUMN plans.is_active IS '是否启用';
COMMENT ON COLUMN plans.created_at IS '创建时间';
COMMENT ON COLUMN plans.updated_at IS '更新时间';
COMMENT ON COLUMN plans.deleted_at IS '删除时间（软删除）';
COMMENT ON COLUMN plans.type IS '套餐类型: free/pro/enterprise';
COMMENT ON COLUMN plans.price_monthly IS '月付价格（元）';
COMMENT ON COLUMN plans.price_yearly IS '年付价格（元）';
COMMENT ON COLUMN plans.max_knowledge_bases IS '最大知识库数量';
COMMENT ON COLUMN plans.max_documents_per_kb IS '每个知识库最大文档数';
COMMENT ON COLUMN plans.max_messages_per_month IS '每月最大消息数';
COMMENT ON COLUMN plans.max_concurrent_sessions IS '最大并发会话数';
COMMENT ON COLUMN plans.status IS '状态: active-上架, inactive-下架';
COMMENT ON COLUMN plans.sort_order IS '排序权重';
COMMENT ON COLUMN plans.allowed_channels IS '允许的渠道类型，JSON数组';

-- ===== billing_plans 计费套餐表（旧） =====
COMMENT ON COLUMN billing_plans.id IS '套餐ID';
COMMENT ON COLUMN billing_plans.name IS '套餐名称';
COMMENT ON COLUMN billing_plans.description IS '套餐描述';
COMMENT ON COLUMN billing_plans.plan_type IS '套餐类型';
COMMENT ON COLUMN billing_plans.price IS '价格（元）';
COMMENT ON COLUMN billing_plans.currency IS '货币: CNY';
COMMENT ON COLUMN billing_plans.features IS '功能列表，JSON数组';
COMMENT ON COLUMN billing_plans.token_limit IS 'Token额度';
COMMENT ON COLUMN billing_plans.max_concurrent IS '最大并发数';
COMMENT ON COLUMN billing_plans.max_agents IS '最大智能体数';
COMMENT ON COLUMN billing_plans.max_knowledge_bases IS '最大知识库数';
COMMENT ON COLUMN billing_plans.is_active IS '是否启用';
COMMENT ON COLUMN billing_plans.is_default IS '是否默认套餐';
COMMENT ON COLUMN billing_plans.priority IS '优先级';
COMMENT ON COLUMN billing_plans.created_at IS '创建时间';
COMMENT ON COLUMN billing_plans.updated_at IS '更新时间';
COMMENT ON COLUMN billing_plans.price_yearly IS '年付价格（元）';
COMMENT ON COLUMN billing_plans.max_documents_per_kb IS '每个知识库最大文档数';

-- ===== billing_records 计费记录表 =====
COMMENT ON COLUMN billing_records.id IS '记录ID';
COMMENT ON COLUMN billing_records.tenant_id IS '租户ID';
COMMENT ON COLUMN billing_records.subscription_id IS '订阅ID';
COMMENT ON COLUMN billing_records.type IS '记录类型: usage/subscription/payment';
COMMENT ON COLUMN billing_records.amount IS '金额（元）';
COMMENT ON COLUMN billing_records.cost IS '成本（元）';
COMMENT ON COLUMN billing_records.record_date IS '记录日期';
COMMENT ON COLUMN billing_records.created_at IS '创建时间';
COMMENT ON COLUMN billing_records.record_type IS '记录类型';
COMMENT ON COLUMN billing_records.currency IS '货币: CNY';
COMMENT ON COLUMN billing_records.description IS '描述';
COMMENT ON COLUMN billing_records.tokens_used IS 'Token使用量';
COMMENT ON COLUMN billing_records.input_tokens IS '输入Token数';
COMMENT ON COLUMN billing_records.output_tokens IS '输出Token数';
COMMENT ON COLUMN billing_records.model_id IS '模型ID';
COMMENT ON COLUMN billing_records.provider_id IS '提供商ID';
COMMENT ON COLUMN billing_records.status IS '状态: pending/completed/failed';
COMMENT ON COLUMN billing_records.transaction_id IS '交易ID';

-- ===== subscriptions 订阅表 =====
COMMENT ON COLUMN subscriptions.id IS '订阅ID';
COMMENT ON COLUMN subscriptions.tenant_id IS '租户ID';
COMMENT ON COLUMN subscriptions.plan_id IS '套餐ID';
COMMENT ON COLUMN subscriptions.status IS '状态: active/cancelled/expired';
COMMENT ON COLUMN subscriptions.start_date IS '开始日期';
COMMENT ON COLUMN subscriptions.end_date IS '结束日期';
COMMENT ON COLUMN subscriptions.auto_renew IS '是否自动续费';
COMMENT ON COLUMN subscriptions.messages_used IS '已使用消息数';
COMMENT ON COLUMN subscriptions.created_at IS '创建时间';
COMMENT ON COLUMN subscriptions.updated_at IS '更新时间';
COMMENT ON COLUMN subscriptions.deleted_at IS '删除时间（软删除）';
COMMENT ON COLUMN subscriptions.next_billing_date IS '下次计费日期';
COMMENT ON COLUMN subscriptions.current_period_start IS '当前周期开始';
COMMENT ON COLUMN subscriptions.current_period_end IS '当前周期结束';
COMMENT ON COLUMN subscriptions.total_tokens_used IS '累计Token使用量';
COMMENT ON COLUMN subscriptions.token_limit IS 'Token额度';

-- ===== skills 技能表 =====
COMMENT ON COLUMN skills.id IS '技能ID';
COMMENT ON COLUMN skills.name IS '技能名称';
COMMENT ON COLUMN skills.description IS '技能描述';
COMMENT ON COLUMN skills.category IS '分类';
COMMENT ON COLUMN skills.icon IS '图标';
COMMENT ON COLUMN skills.version IS '版本号';
COMMENT ON COLUMN skills.author IS '作者';
COMMENT ON COLUMN skills.is_public IS '是否公开';
COMMENT ON COLUMN skills.is_built_in IS '是否内置';
COMMENT ON COLUMN skills.code IS '技能代码';
COMMENT ON COLUMN skills.config IS '配置，JSON格式';
COMMENT ON COLUMN skills.install_count IS '安装次数';
COMMENT ON COLUMN skills.created_at IS '创建时间';
COMMENT ON COLUMN skills.updated_at IS '更新时间';
COMMENT ON COLUMN skills.deleted_at IS '删除时间（软删除）';
COMMENT ON COLUMN skills.skill_type IS '技能类型: tool/workflow/agent';
COMMENT ON COLUMN skills.schema IS '技能Schema，JSON格式';
COMMENT ON COLUMN skills.provider IS '提供商';
COMMENT ON COLUMN skills.status IS '状态: active/inactive';

-- ===== tenant_ai_configs 租户AI配置表 =====
COMMENT ON COLUMN tenant_ai_configs.id IS '配置ID';
COMMENT ON COLUMN tenant_ai_configs.tenant_id IS '租户ID';
COMMENT ON COLUMN tenant_ai_configs.default_chat_model_id IS '默认对话模型ID，如未设置则使用平台默认';
COMMENT ON COLUMN tenant_ai_configs.default_embedding_model_id IS '默认Embedding模型ID';
COMMENT ON COLUMN tenant_ai_configs.default_provider_id IS '默认提供商ID';
COMMENT ON COLUMN tenant_ai_configs.ai_enabled IS '是否启用AI功能';
COMMENT ON COLUMN tenant_ai_configs.streaming_enabled IS '是否启用流式输出';
COMMENT ON COLUMN tenant_ai_configs.vision_enabled IS '是否启用视觉功能';
COMMENT ON COLUMN tenant_ai_configs.function_call_enabled IS '是否启用函数调用';
COMMENT ON COLUMN tenant_ai_configs.monthly_token_limit IS '每月token额度，0表示无限制（由套餐决定）';
COMMENT ON COLUMN tenant_ai_configs.monthly_token_used IS '本月已用token数';
COMMENT ON COLUMN tenant_ai_configs.token_limit_reset_at IS '额度重置时间';
COMMENT ON COLUMN tenant_ai_configs.rate_limit_per_minute IS '每分钟请求限制';
COMMENT ON COLUMN tenant_ai_configs.rate_limit_per_day IS '每天请求限制';
COMMENT ON COLUMN tenant_ai_configs.config IS '扩展配置，JSON格式，可存储自定义参数';
COMMENT ON COLUMN tenant_ai_configs.created_at IS '创建时间';
COMMENT ON COLUMN tenant_ai_configs.updated_at IS '更新时间';

-- ===== tenant_members 租户成员表 =====
COMMENT ON COLUMN tenant_members.id IS '成员记录ID';
COMMENT ON COLUMN tenant_members.tenant_id IS '租户ID';
COMMENT ON COLUMN tenant_members.user_id IS '用户ID';
COMMENT ON COLUMN tenant_members.role IS '角色: owner/admin/member';
COMMENT ON COLUMN tenant_members.status IS '状态: active-在职, resigned-离职';
COMMENT ON COLUMN tenant_members.invited_by IS '邀请人ID';
COMMENT ON COLUMN tenant_members.joined_at IS '加入时间';
COMMENT ON COLUMN tenant_members.resigned_at IS '离职时间';
COMMENT ON COLUMN tenant_members.created_at IS '创建时间';
COMMENT ON COLUMN tenant_members.updated_at IS '更新时间';

-- ===== tenant_invitations 邀请码表 =====
COMMENT ON COLUMN tenant_invitations.id IS '邀请记录ID';
COMMENT ON COLUMN tenant_invitations.tenant_id IS '租户ID';
COMMENT ON COLUMN tenant_invitations.code IS '邀请码（唯一）';
COMMENT ON COLUMN tenant_invitations.invitee_phone IS '受邀人手机号';
COMMENT ON COLUMN tenant_invitations.invitee_email IS '受邀人邮箱';
COMMENT ON COLUMN tenant_invitations.role IS '邀请角色: admin/member';
COMMENT ON COLUMN tenant_invitations.invited_by IS '邀请人ID';
COMMENT ON COLUMN tenant_invitations.used IS '是否已使用';
COMMENT ON COLUMN tenant_invitations.created_at IS '创建时间';

-- 记录迁移
INSERT INTO schema_migrations (version, description) VALUES ('V017', 'column_comments') ON CONFLICT (version) DO NOTHING;
