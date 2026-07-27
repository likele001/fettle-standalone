-- V011: 补充企业实用技能（平台预置）
-- 适用于企业和个体户的核心业务场景

-- 1. 文档生成
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '文档生成', '自动生成合同、报告、方案等商业文档', 'utility', '📄', 'doc_generator', 'api',
    '{"input": {"type": "string", "title": "string", "content": "string"}, "output": {"document_url": "string", "format": "string"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 2. 邮件撰写
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '邮件撰写', '智能撰写商务邮件，支持多种场景和语气', 'utility', '📧', 'email_writer', 'api',
    '{"input": {"recipient": "string", "subject": "string", "tone": "string", "key_points": "array"}, "output": {"email_body": "string"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 3. 图片生成
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '图片生成', 'AI生成营销海报、产品图、社交媒体配图', 'marketing', '🎨', 'image_generator', 'api',
    '{"input": {"prompt": "string", "style": "string", "size": "string"}, "output": {"image_url": "string"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 4. 网页搜索
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '网页搜索', '实时搜索互联网获取最新资讯和信息', 'utility', '🔍', 'web_search', 'api',
    '{"input": {"query": "string", "max_results": "number"}, "output": {"results": "array"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 5. 文件转换
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '文件转换', '支持PDF/Word/Excel/PPT等格式互转', 'utility', '🔄', 'file_converter', 'api',
    '{"input": {"file_url": "string", "target_format": "string"}, "output": {"converted_url": "string"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 6. PDF解析
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    'PDF解析', '提取PDF文档中的文字、表格和图片内容', 'utility', '📑', 'pdf_parser', 'api',
    '{"input": {"file_url": "string", "extract_type": "string"}, "output": {"text": "string", "tables": "array"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 7. 数据可视化
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '数据可视化', '将数据自动转换为图表和可视化报告', 'data', '📊', 'data_visualizer', 'api',
    '{"input": {"data": "array", "chart_type": "string", "title": "string"}, "output": {"chart_url": "string"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 8. 语音转文字
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '语音转文字', '将语音/录音文件转换为文字记录', 'utility', '🎙️', 'speech_to_text', 'api',
    '{"input": {"audio_url": "string", "language": "string"}, "output": {"text": "string", "duration": "number"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 9. 短信发送
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '短信发送', '批量发送营销短信、通知短信和验证码', 'marketing', '📱', 'sms_sender', 'api',
    '{"input": {"phone_numbers": "array", "message": "string", "template_id": "string"}, "output": {"sent_count": "number", "failed": "array"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 10. 二维码生成
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '二维码生成', '生成带Logo的营销二维码、支付码和名片码', 'marketing', '', 'qr_generator', 'api',
    '{"input": {"content": "string", "style": "string", "logo_url": "string"}, "output": {"qr_url": "string"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 11. 表单处理
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '表单处理', '创建和管理在线表单，自动汇总分析数据', 'utility', '📋', 'form_handler', 'api',
    '{"input": {"form_fields": "array", "title": "string"}, "output": {"form_url": "string", "responses": "array"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 12. 审批流程
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '审批流程', '创建和管理企业审批流程，支持多级审批', 'management', '✅', 'approval_flow', 'api',
    '{"input": {"flow_name": "string", "approvers": "array", "rules": "object"}, "output": {"flow_id": "string", "status": "string"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 13. 库存管理
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '库存管理', '商品库存查询、预警和出入库记录管理', 'management', '📦', 'inventory_manager', 'api',
    '{"input": {"product_id": "string", "action": "string", "quantity": "number"}, "output": {"stock": "number", "alerts": "array"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 14. 客户管理
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '客户管理', '客户信息录入、跟进记录和分类管理', 'management', '', 'crm_manager', 'api',
    '{"input": {"customer_name": "string", "action": "string", "data": "object"}, "output": {"customer_id": "string", "status": "string"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;

-- 15. 合同管理
INSERT INTO skills (name, description, category, icon, code, skill_type, schema, is_public, is_built_in, status)
VALUES (
    '合同管理', '合同模板管理、签署跟踪和到期提醒', 'management', '📝', 'contract_manager', 'api',
    '{"input": {"contract_type": "string", "parties": "array", "terms": "object"}, "output": {"contract_id": "string", "status": "string"}}',
    true, true, 'active'
) ON CONFLICT (name) DO NOTHING;
