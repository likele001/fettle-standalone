-- V018: 为遗漏的表字段补充注释

-- ===== admin_audit_log 管理员审计日志表 =====
COMMENT ON COLUMN admin_audit_log.id IS '日志ID';
COMMENT ON COLUMN admin_audit_log.admin_id IS '管理员ID';
COMMENT ON COLUMN admin_audit_log.detail IS '操作详情，JSON格式';
COMMENT ON COLUMN admin_audit_log.ip_address IS '操作IP地址';
COMMENT ON COLUMN admin_audit_log.created_at IS '创建时间';

-- ===== admin_users 管理员用户表 =====
COMMENT ON COLUMN admin_users.id IS '管理员ID';
COMMENT ON COLUMN admin_users.username IS '用户名（唯一）';
COMMENT ON COLUMN admin_users.phone IS '手机号';
COMMENT ON COLUMN admin_users.email IS '邮箱';
COMMENT ON COLUMN admin_users.password_hash IS '密码哈希（bcrypt）';
COMMENT ON COLUMN admin_users.name IS '姓名';
COMMENT ON COLUMN admin_users.avatar_url IS '头像URL';
COMMENT ON COLUMN admin_users.role IS '角色: super_admin/admin';
COMMENT ON COLUMN admin_users.status IS '状态: active-正常, banned-禁用';
COMMENT ON COLUMN admin_users.last_login_at IS '最后登录时间';
COMMENT ON COLUMN admin_users.created_at IS '创建时间';
COMMENT ON COLUMN admin_users.updated_at IS '更新时间';
COMMENT ON COLUMN admin_users.deleted_at IS '删除时间（软删除）';

-- ===== ai_models AI模型表 =====
COMMENT ON COLUMN ai_models.id IS '模型ID';
COMMENT ON COLUMN ai_models.provider_id IS '提供商ID';
COMMENT ON COLUMN ai_models.model_code IS '模型代码（唯一标识）';
COMMENT ON COLUMN ai_models.model_name IS '模型名称';
COMMENT ON COLUMN ai_models.model_type IS '模型类型: chat/embedding/image/vision/tts/stt/video';
COMMENT ON COLUMN ai_models.max_input_tokens IS '最大输入Token数';
COMMENT ON COLUMN ai_models.max_output_tokens IS '最大输出Token数';
COMMENT ON COLUMN ai_models.input_price_per1k IS '输入价格（元/千Token）';
COMMENT ON COLUMN ai_models.output_price_per1k IS '输出价格（元/千Token）';
COMMENT ON COLUMN ai_models.description IS '模型描述';
COMMENT ON COLUMN ai_models.capabilities IS '能力列表，JSON数组';
COMMENT ON COLUMN ai_models.status IS '状态: active/inactive';
COMMENT ON COLUMN ai_models.is_default IS '是否默认模型';
COMMENT ON COLUMN ai_models.priority IS '优先级';
COMMENT ON COLUMN ai_models.created_at IS '创建时间';
COMMENT ON COLUMN ai_models.updated_at IS '更新时间';

-- ===== ai_providers AI提供商表 =====
COMMENT ON COLUMN ai_providers.id IS '提供商ID';
COMMENT ON COLUMN ai_providers.code IS '提供商代码: openai/aliyun/baidu等';
COMMENT ON COLUMN ai_providers.name IS '提供商名称（中文）';
COMMENT ON COLUMN ai_providers.name_en IS '提供商名称（英文）';
COMMENT ON COLUMN ai_providers.logo_url IS 'Logo URL';
COMMENT ON COLUMN ai_providers.description IS '描述';
COMMENT ON COLUMN ai_providers.api_base_url IS 'API基础地址';
COMMENT ON COLUMN ai_providers.auth_type IS '认证方式: api_key/bearer';
COMMENT ON COLUMN ai_providers.status IS '状态: active/inactive';
COMMENT ON COLUMN ai_providers.support_streaming IS '是否支持流式输出';
COMMENT ON COLUMN ai_providers.support_vision IS '是否支持视觉';
COMMENT ON COLUMN ai_providers.support_function_call IS '是否支持函数调用';
COMMENT ON COLUMN ai_providers.created_at IS '创建时间';
COMMENT ON COLUMN ai_providers.updated_at IS '更新时间';

-- ===== ai_usage_logs AI使用日志表 =====
COMMENT ON COLUMN ai_usage_logs.id IS '日志ID';
COMMENT ON COLUMN ai_usage_logs.tenant_id IS '租户ID';
COMMENT ON COLUMN ai_usage_logs.user_id IS '用户ID';
COMMENT ON COLUMN ai_usage_logs.conversation_id IS '会话ID';
COMMENT ON COLUMN ai_usage_logs.message_id IS '消息ID';
COMMENT ON COLUMN ai_usage_logs.provider_id IS '提供商ID';
COMMENT ON COLUMN ai_usage_logs.model_id IS '模型ID';
COMMENT ON COLUMN ai_usage_logs.input_tokens IS '输入Token数';
COMMENT ON COLUMN ai_usage_logs.output_tokens IS '输出Token数';
COMMENT ON COLUMN ai_usage_logs.total_tokens IS '总Token数';
COMMENT ON COLUMN ai_usage_logs.input_cost IS '输入成本（元）';
COMMENT ON COLUMN ai_usage_logs.output_cost IS '输出成本（元）';
COMMENT ON COLUMN ai_usage_logs.latency_ms IS '延迟（毫秒）';
COMMENT ON COLUMN ai_usage_logs.success IS '是否成功';
COMMENT ON COLUMN ai_usage_logs.error_message IS '错误信息';
COMMENT ON COLUMN ai_usage_logs.created_at IS '创建时间';

-- ===== industry_bundles 行业套餐表 =====
COMMENT ON COLUMN industry_bundles.id IS '套餐ID';
COMMENT ON COLUMN industry_bundles.industry IS '行业: ecommerce/dining/legal/education/medical';
COMMENT ON COLUMN industry_bundles.name IS '套餐名称';
COMMENT ON COLUMN industry_bundles.description IS '套餐描述';
COMMENT ON COLUMN industry_bundles.icon IS '图标';
COMMENT ON COLUMN industry_bundles.config IS '套餐配置（智能体+知识库模板），JSON格式';
COMMENT ON COLUMN industry_bundles.is_active IS '是否启用';
COMMENT ON COLUMN industry_bundles.created_at IS '创建时间';
COMMENT ON COLUMN industry_bundles.updated_at IS '更新时间';

-- ===== knowledge_documents 知识文档表 =====
COMMENT ON COLUMN knowledge_documents.id IS '文档ID';
COMMENT ON COLUMN knowledge_documents.knowledge_base_id IS '所属知识库ID';
COMMENT ON COLUMN knowledge_documents.tenant_id IS '租户ID';
COMMENT ON COLUMN knowledge_documents.file_name IS '文件名';
COMMENT ON COLUMN knowledge_documents.file_type IS '文件类型: pdf/md/txt/docx';
COMMENT ON COLUMN knowledge_documents.file_size IS '文件大小（字节）';
COMMENT ON COLUMN knowledge_documents.file_url IS '文件存储URL';
COMMENT ON COLUMN knowledge_documents.status IS '状态: pending/processing/completed/failed';
COMMENT ON COLUMN knowledge_documents.chunk_count IS '分段数量';
COMMENT ON COLUMN knowledge_documents.error_message IS '错误信息';
COMMENT ON COLUMN knowledge_documents.processed_at IS '处理完成时间';
COMMENT ON COLUMN knowledge_documents.created_at IS '创建时间';
COMMENT ON COLUMN knowledge_documents.updated_at IS '更新时间';

-- ===== payment_configs 支付配置表 =====
COMMENT ON COLUMN payment_configs.id IS '配置ID';
COMMENT ON COLUMN payment_configs.app_id IS '应用ID';
COMMENT ON COLUMN payment_configs.app_secret IS '应用密钥';
COMMENT ON COLUMN payment_configs.gateway_url IS '支付网关地址';
COMMENT ON COLUMN payment_configs.notify_url IS '异步通知地址';
COMMENT ON COLUMN payment_configs.return_url IS '同步返回地址';
COMMENT ON COLUMN payment_configs.enabled IS '是否启用';
COMMENT ON COLUMN payment_configs.created_at IS '创建时间';
COMMENT ON COLUMN payment_configs.updated_at IS '更新时间';
COMMENT ON COLUMN payment_configs.deleted_at IS '删除时间（软删除）';

-- ===== payment_orders 支付订单表 =====
COMMENT ON COLUMN payment_orders.id IS '订单ID';
COMMENT ON COLUMN payment_orders.tenant_id IS '租户ID';
COMMENT ON COLUMN payment_orders.plan_id IS '套餐ID';
COMMENT ON COLUMN payment_orders.plan_name IS '套餐名称';
COMMENT ON COLUMN payment_orders.amount IS '金额（元）';
COMMENT ON COLUMN payment_orders.trade_order_id IS '交易订单号（支付平台）';
COMMENT ON COLUMN payment_orders.status IS '状态: pending/paid/failed/refunded';
COMMENT ON COLUMN payment_orders.payment_url IS '支付链接';
COMMENT ON COLUMN payment_orders.transaction_id IS '交易流水号';
COMMENT ON COLUMN payment_orders.paid_at IS '支付时间';
COMMENT ON COLUMN payment_orders.created_at IS '创建时间';
COMMENT ON COLUMN payment_orders.updated_at IS '更新时间';

-- ===== permissions 权限表 =====
COMMENT ON COLUMN permissions.id IS '权限ID';
COMMENT ON COLUMN permissions.code IS '权限代码: user:read, agent:write等';
COMMENT ON COLUMN permissions.name IS '权限名称';
COMMENT ON COLUMN permissions.module IS '所属模块';
COMMENT ON COLUMN permissions.description IS '权限描述';
COMMENT ON COLUMN permissions.created_at IS '创建时间';
COMMENT ON COLUMN permissions.updated_at IS '更新时间';

-- ===== platform_config 平台配置表 =====
COMMENT ON COLUMN platform_config.id IS '配置ID';
COMMENT ON COLUMN platform_config.description IS '配置描述';
COMMENT ON COLUMN platform_config.created_at IS '创建时间';
COMMENT ON COLUMN platform_config.updated_at IS '更新时间';
COMMENT ON COLUMN platform_config.updated_by IS '最后修改人ID';

-- ===== role_permissions 角色权限关联表 =====
COMMENT ON COLUMN role_permissions.role_id IS '角色ID';
COMMENT ON COLUMN role_permissions.permission_id IS '权限ID';
COMMENT ON COLUMN role_permissions.created_at IS '创建时间';

-- ===== roles 角色表 =====
COMMENT ON COLUMN roles.id IS '角色ID';
COMMENT ON COLUMN roles.tenant_id IS '租户ID';
COMMENT ON COLUMN roles.name IS '角色名称';
COMMENT ON COLUMN roles.code IS '角色代码';
COMMENT ON COLUMN roles.description IS '角色描述';
COMMENT ON COLUMN roles.created_at IS '创建时间';
COMMENT ON COLUMN roles.updated_at IS '更新时间';

-- ===== skill_installations 技能安装表 =====
COMMENT ON COLUMN skill_installations.id IS '安装记录ID';
COMMENT ON COLUMN skill_installations.tenant_id IS '租户ID';
COMMENT ON COLUMN skill_installations.skill_id IS '技能ID';
COMMENT ON COLUMN skill_installations.status IS '状态: active/inactive';
COMMENT ON COLUMN skill_installations.config IS '安装配置，JSON格式';
COMMENT ON COLUMN skill_installations.installed_at IS '安装时间';
COMMENT ON COLUMN skill_installations.updated_at IS '更新时间';
COMMENT ON COLUMN skill_installations.deleted_at IS '删除时间（软删除）';

-- ===== sys_config 系统配置表 =====
COMMENT ON COLUMN sys_config.id IS '配置ID';
COMMENT ON COLUMN sys_config.key IS '配置键';
COMMENT ON COLUMN sys_config.config IS '配置值，JSON格式';
COMMENT ON COLUMN sys_config.created_at IS '创建时间';
COMMENT ON COLUMN sys_config.updated_at IS '更新时间';

-- ===== tenant_api_keys 租户API密钥表 =====
COMMENT ON COLUMN tenant_api_keys.id IS 'API密钥ID';
COMMENT ON COLUMN tenant_api_keys.tenant_id IS '租户ID';
COMMENT ON COLUMN tenant_api_keys.provider_id IS '提供商ID';
COMMENT ON COLUMN tenant_api_keys.api_key_name IS '密钥名称';
COMMENT ON COLUMN tenant_api_keys.api_key_encrypted IS '加密后的API密钥';
COMMENT ON COLUMN tenant_api_keys.monthly_quota IS '每月配额（Token数）';
COMMENT ON COLUMN tenant_api_keys.monthly_used IS '本月已用（Token数）';
COMMENT ON COLUMN tenant_api_keys.status IS '状态: active/inactive/expired';
COMMENT ON COLUMN tenant_api_keys.expired_at IS '过期时间';
COMMENT ON COLUMN tenant_api_keys.custom_headers IS '自定义请求头，JSON格式';
COMMENT ON COLUMN tenant_api_keys.created_at IS '创建时间';
COMMENT ON COLUMN tenant_api_keys.updated_at IS '更新时间';
COMMENT ON COLUMN tenant_api_keys.last_used_at IS '最后使用时间';

-- 记录迁移
INSERT INTO schema_migrations (version, description) VALUES ('V018', 'column_comments_remaining') ON CONFLICT (version) DO NOTHING;
