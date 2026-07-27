-- V013: Industry bundles (行业套餐)
-- Pre-configured agent + knowledge base packages for different industries

-- ============================================
-- 1. Create industry_bundles table
-- ============================================
CREATE TABLE IF NOT EXISTS industry_bundles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    industry VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    icon VARCHAR(100),
    config JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_industry_bundles_industry ON industry_bundles(industry);
CREATE INDEX IF NOT EXISTS idx_industry_bundles_active ON industry_bundles(is_active);

-- ============================================
-- 2. Seed 5 industry bundles
-- ============================================

-- 电商行业
INSERT INTO industry_bundles (industry, name, description, icon, config) VALUES (
    'ecommerce',
    '电商行业套餐',
    '为电商企业打造的智能客服与运营助手组合，覆盖售前咨询、订单处理、售后管理等场景',
    'shopping-cart',
    '{
        "agents": [
            {
                "name": "电商客服",
                "description": "专业电商客服，处理商品咨询、订单查询、物流跟踪、退换货等常见问题",
                "agent_type": "chat",
                "welcome_message": "您好！欢迎来到我们的客服中心。我可以帮您查询订单、了解商品信息、处理退换货等，请问有什么可以帮您？",
                "personality_config": {"role": "电商客服", "tone": "热情专业", "language": "中文", "expertise": ["商品咨询", "订单查询", "物流跟踪", "退换货处理"]},
                "required_plan": "free",
                "kb_template_names": ["电商常见问题库"]
            },
            {
                "name": "营销文案助手",
                "description": "电商营销文案专家，擅长商品描述、促销活动文案、社交媒体推广内容创作",
                "agent_type": "chat",
                "welcome_message": "嗨！我是您的电商营销文案助手，可以帮您写商品描述、促销文案、推广内容。告诉我您想推广什么商品？",
                "personality_config": {"role": "营销文案", "tone": "创意活泼", "language": "中文", "expertise": ["商品描述", "促销文案", "社媒推广", "活动策划"]},
                "required_plan": "pro",
                "kb_template_names": []
            },
            {
                "name": "售后专员",
                "description": "专业售后处理，负责投诉处理、退换货流程指导、客户满意度维护",
                "agent_type": "chat",
                "welcome_message": "您好，我是售后专员。如果您的订单有任何问题，请告诉我订单号和具体情况，我会尽快帮您解决。",
                "personality_config": {"role": "售后专员", "tone": "耐心友善", "language": "中文", "expertise": ["投诉处理", "退换货", "补偿方案", "客户维护"]},
                "required_plan": "pro",
                "kb_template_names": ["售后政策库"]
            }
        ],
        "kb_templates": [
            {"name": "电商常见问题库", "description": "包含电商场景常见问答：下单流程、支付方式、配送时效、退换货政策等"},
            {"name": "售后政策库", "description": "退换货政策、保修条款、补偿标准、投诉处理流程等"}
        ]
    }'
);

-- 餐饮行业
INSERT INTO industry_bundles (industry, name, description, icon, config) VALUES (
    'dining',
    '餐饮行业套餐',
    '为餐饮企业打造的智能点餐、推荐和客服组合，提升顾客用餐体验',
    'utensils',
    '{
        "agents": [
            {
                "name": "点餐助手",
                "description": "智能点餐助手，帮助顾客浏览菜单、推荐菜品、处理特殊饮食需求",
                "agent_type": "chat",
                "welcome_message": "您好！我是点餐助手，可以帮您推荐菜品、介绍口味、处理特殊饮食需求。请问今天想吃点什么？",
                "personality_config": {"role": "点餐助手", "tone": "亲切热情", "language": "中文", "expertise": ["菜品推荐", "口味介绍", "过敏原提示", "营养搭配"]},
                "required_plan": "free",
                "kb_template_names": ["菜单知识库"]
            },
            {
                "name": "餐饮客服",
                "description": "处理预约订位、外卖咨询、投诉建议、营业时间查询等",
                "agent_type": "chat",
                "welcome_message": "您好！我可以帮您预约订位、查询外卖信息、解答用餐相关问题。请问有什么需要？",
                "personality_config": {"role": "餐饮客服", "tone": "礼貌周到", "language": "中文", "expertise": ["预约订位", "外卖咨询", "投诉处理", "营业信息"]},
                "required_plan": "pro",
                "kb_template_names": ["餐饮服务指南"]
            },
            {
                "name": "营养顾问",
                "description": "专业营养顾问，提供菜品营养分析、健康饮食建议、特殊人群饮食指导",
                "agent_type": "chat",
                "welcome_message": "您好！我是营养顾问，可以为您提供菜品营养分析、健康饮食搭配建议。有什么饮食方面的疑问吗？",
                "personality_config": {"role": "营养顾问", "tone": "专业温和", "language": "中文", "expertise": ["营养分析", "健康搭配", "过敏原", "特殊饮食"]},
                "required_plan": "enterprise",
                "kb_template_names": []
            }
        ],
        "kb_templates": [
            {"name": "菜单知识库", "description": "菜品名称、价格、口味、食材、过敏原、热量等详细信息"},
            {"name": "餐饮服务指南", "description": "预约流程、外卖配送范围、营业时间、门店地址、投诉处理流程"}
        ]
    }'
);

-- 法律行业
INSERT INTO industry_bundles (industry, name, description, icon, config) VALUES (
    'legal',
    '法律行业套餐',
    '为法律从业者打造的智能法律咨询、合同审查助手，提升法律服务效率',
    'scale-balanced',
    '{
        "agents": [
            {
                "name": "法律咨询助手",
                "description": "提供基础法律咨询，涵盖合同法、劳动法、公司法等常见法律问题的初步分析",
                "agent_type": "chat",
                "welcome_message": "您好！我是法律咨询助手，可以为您提供基础法律问题的初步分析和法律知识解答。请注意，我的建议仅供参考，具体法律问题建议咨询专业律师。请问有什么法律问题？",
                "personality_config": {"role": "法律咨询", "tone": "严谨专业", "language": "中文", "expertise": ["合同法", "劳动法", "公司法", "知识产权", "民事纠纷"]},
                "required_plan": "pro",
                "kb_template_names": ["法律常见问题库"]
            },
            {
                "name": "合同审查助手",
                "description": "协助审查合同条款，识别风险点，提供修改建议",
                "agent_type": "chat",
                "welcome_message": "您好！我是合同审查助手，可以帮您分析合同条款、识别潜在风险点、提供修改建议。请提供您需要审查的合同内容。",
                "personality_config": {"role": "合同审查", "tone": "严谨细致", "language": "中文", "expertise": ["条款分析", "风险识别", "修改建议", "合规检查"]},
                "required_plan": "enterprise",
                "kb_template_names": ["合同模板库"]
            },
            {
                "name": "法规查询助手",
                "description": "快速查询法律法规、司法解释、典型案例，提供法规条文解读",
                "agent_type": "chat",
                "welcome_message": "您好！我是法规查询助手，可以帮您查询法律法规、解读法条含义、检索相关案例。请问需要查询什么法规？",
                "personality_config": {"role": "法规查询", "tone": "准确客观", "language": "中文", "expertise": ["法规检索", "条文解读", "案例检索", "法律更新"]},
                "required_plan": "pro",
                "kb_template_names": []
            }
        ],
        "kb_templates": [
            {"name": "法律常见问题库", "description": "常见法律问题解答：合同纠纷、劳动争议、消费者权益、房产纠纷等"},
            {"name": "合同模板库", "description": "常用合同模板要点：劳动合同、租赁合同、买卖合同、服务合同等"}
        ]
    }'
);

-- 教育行业
INSERT INTO industry_bundles (industry, name, description, icon, config) VALUES (
    'education',
    '教育行业套餐',
    '为教育机构打造的智能教学辅助、课程咨询和学员管理组合',
    'graduation-cap',
    '{
        "agents": [
            {
                "name": "课程顾问",
                "description": "提供课程咨询、班级介绍、学费说明、报名指导等服务",
                "agent_type": "chat",
                "welcome_message": "您好！我是课程顾问，可以为您介绍我们的课程体系、班级设置、学费标准，也可以指导您完成报名流程。请问想了解什么课程？",
                "personality_config": {"role": "课程顾问", "tone": "亲切专业", "language": "中文", "expertise": ["课程介绍", "班级设置", "学费说明", "报名指导"]},
                "required_plan": "free",
                "kb_template_names": ["课程信息库"]
            },
            {
                "name": "学习助手",
                "description": "辅助学生学习，提供知识点讲解、作业辅导、学习方法建议",
                "agent_type": "chat",
                "welcome_message": "你好！我是学习助手，可以帮你讲解知识点、辅导作业、推荐学习方法。今天想学什么？",
                "personality_config": {"role": "学习助手", "tone": "耐心鼓励", "language": "中文", "expertise": ["知识讲解", "作业辅导", "学习方法", "考试技巧"]},
                "required_plan": "pro",
                "kb_template_names": ["学科知识库"]
            },
            {
                "name": "教务助手",
                "description": "处理课表查询、请假申请、考试安排、校园通知等教务事务",
                "agent_type": "chat",
                "welcome_message": "您好！我是教务助手，可以帮您查询课表、处理请假、了解考试安排和校园通知。请问有什么教务问题？",
                "personality_config": {"role": "教务助手", "tone": "高效规范", "language": "中文", "expertise": ["课表查询", "请假管理", "考试安排", "通知发布"]},
                "required_plan": "pro",
                "kb_template_names": ["教务常见问题库"]
            },
            {
                "name": "培训讲师",
                "description": "企业培训助手，协助制定培训计划、生成培训材料、评估培训效果",
                "agent_type": "chat",
                "welcome_message": "您好！我是培训讲师助手，可以帮您制定培训计划、设计课程内容、编写培训材料。请问有什么培训需求？",
                "personality_config": {"role": "培训讲师", "tone": "专业引导", "language": "中文", "expertise": ["培训规划", "课程设计", "材料编写", "效果评估"]},
                "required_plan": "enterprise",
                "kb_template_names": []
            }
        ],
        "kb_templates": [
            {"name": "课程信息库", "description": "课程目录、班级设置、学费标准、上课时间、教师介绍等"},
            {"name": "学科知识库", "description": "各学科核心知识点、常见题型、学习方法总结"},
            {"name": "教务常见问题库", "description": "选课退课、请假制度、考试安排、成绩单获取、校园设施等"}
        ]
    }'
);

-- 医疗行业
INSERT INTO industry_bundles (industry, name, description, icon, config) VALUES (
    'medical',
    '医疗行业套餐',
    '为医疗机构打造的智能导诊、健康咨询和患者服务组合',
    'hospital',
    '{
        "agents": [
            {
                "name": "智能导诊",
                "description": "根据患者症状描述，推荐合适的科室和医生，指导挂号流程",
                "agent_type": "chat",
                "welcome_message": "您好！我是智能导诊助手。请描述您的症状，我会为您推荐合适的科室。注意：紧急情况请直接拨打120急救电话。",
                "personality_config": {"role": "导诊助手", "tone": "温和关切", "language": "中文", "expertise": ["症状分析", "科室推荐", "挂号指导", "就医流程"]},
                "required_plan": "free",
                "kb_template_names": ["科室与就诊指南"]
            },
            {
                "name": "健康咨询助手",
                "description": "提供健康知识科普、常见病护理建议、用药常识、体检报告解读指导",
                "agent_type": "chat",
                "welcome_message": "您好！我是健康咨询助手，可以为您提供健康知识科普、日常护理建议、体检报告解读指导。请注意，我的建议不能替代医生的诊断。请问有什么健康问题？",
                "personality_config": {"role": "健康咨询", "tone": "专业关怀", "language": "中文", "expertise": ["健康科普", "护理建议", "用药常识", "体检解读"]},
                "required_plan": "pro",
                "kb_template_names": ["健康科普知识库"]
            },
            {
                "name": "预约随访助手",
                "description": "协助患者预约挂号、查询检查报告、提醒复诊随访、处理医保咨询",
                "agent_type": "chat",
                "welcome_message": "您好！我可以帮您预约挂号、查询检查报告、设置复诊提醒、解答医保相关问题。请问需要什么帮助？",
                "personality_config": {"role": "预约随访", "tone": "高效贴心", "language": "中文", "expertise": ["预约挂号", "报告查询", "复诊提醒", "医保咨询"]},
                "required_plan": "enterprise",
                "kb_template_names": []
            },
            {
                "name": "患者关怀助手",
                "description": "术后随访、用药提醒、康复指导、满意度调查",
                "agent_type": "chat",
                "welcome_message": "您好！我是患者关怀助手，会关注您的康复进展、提醒您按时用药、提供康复指导。请问您目前恢复情况如何？",
                "personality_config": {"role": "患者关怀", "tone": "温暖关怀", "language": "中文", "expertise": ["术后随访", "用药提醒", "康复指导", "满意度调查"]},
                "required_plan": "enterprise",
                "kb_template_names": []
            }
        ],
        "kb_templates": [
            {"name": "科室与就诊指南", "description": "医院科室介绍、常见疾病对应科室、挂号流程、就诊须知"},
            {"name": "健康科普知识库", "description": "常见疾病科普、健康生活方式、季节性防病知识、营养健康指南"}
        ]
    }'
);
