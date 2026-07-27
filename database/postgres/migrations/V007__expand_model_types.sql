-- ==========================================
-- Version: V007
-- Description: 扩展 AI 模型类型（image/video/tts/stt/vision）
-- Date: 2026-07-11
-- 功能说明：
--   1. 为各厂商添加图片生成、视频生成、语音合成、语音识别、视觉理解等模型
--   2. 使用 ON CONFLICT DO NOTHING 保证幂等性
-- ==========================================

BEGIN;

-- 阿里云 - 扩展模型
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='aliyun'), 'wanx-v1', '通义万相-文生图', 'image', 0, 0, 0.02, 0, '["image_generation"]', false, 10),
((SELECT id FROM ai_providers WHERE code='aliyun'), 'wanx-2.1-imageedit', '通义万相-图像编辑', 'image', 0, 0, 0.02, 0, '["image_edit"]', false, 11),
((SELECT id FROM ai_providers WHERE code='aliyun'), 'cosyvoice-v1', 'CosyVoice语音合成', 'tts', 0, 0, 0.002, 0, '["tts"]', false, 20),
((SELECT id FROM ai_providers WHERE code='aliyun'), 'qwen-vl-max', 'Qwen-VL-Max', 'vision', 32768, 2048, 0.003, 0.009, '["chat", "vision"]', false, 5),
((SELECT id FROM ai_providers WHERE code='aliyun'), 'qwen-vl-plus', 'Qwen-VL-Plus', 'vision', 8192, 2048, 0.0015, 0.005, '["chat", "vision"]', false, 6),
((SELECT id FROM ai_providers WHERE code='aliyun'), 'paraformer-realtime-v2', 'Paraformer实时语音识别', 'stt', 0, 0, 0.001, 0, '["stt"]', false, 21)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- 百度 - 扩展模型
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='baidu'), 'ernie-vilg-v2', '文心一格', 'image', 0, 0, 0.02, 0, '["image_generation"]', false, 10),
((SELECT id FROM ai_providers WHERE code='baidu'), 'ernie-vi', '文心千帆视觉理解', 'vision', 8192, 2048, 0.003, 0.009, '["chat", "vision"]', false, 5)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- 腾讯 - 扩展模型
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='tencent'), 'hunyuan-image', '混元图像生成', 'image', 0, 0, 0.02, 0, '["image_generation"]', false, 10),
((SELECT id FROM ai_providers WHERE code='tencent'), 'hunyuan-vision', '混元视觉理解', 'vision', 8192, 2048, 0.003, 0.009, '["chat", "vision"]', false, 5),
((SELECT id FROM ai_providers WHERE code='tencent'), 'hunyuan-video', '混元视频生成', 'video', 0, 0, 0.05, 0, '["video_generation"]', false, 15)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- OpenAI - 扩展模型
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='openai'), 'dall-e-3', 'DALL-E 3', 'image', 0, 0, 0.04, 0, '["image_generation"]', false, 10),
((SELECT id FROM ai_providers WHERE code='openai'), 'whisper-1', 'Whisper语音识别', 'stt', 0, 0, 0.006, 0, '["stt"]', false, 20),
((SELECT id FROM ai_providers WHERE code='openai'), 'tts-1', 'TTS-1语音合成', 'tts', 0, 0, 0.015, 0, '["tts"]', false, 21),
((SELECT id FROM ai_providers WHERE code='openai'), 'sora', 'Sora视频生成', 'video', 0, 0, 0.10, 0, '["video_generation"]', false, 15)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- Anthropic - 视觉模型（已有的chat模型带vision能力，此处只加vision类型条目）
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='anthropic'), 'claude-3-5-sonnet-vision', 'Claude 3.5 Sonnet (Vision)', 'vision', 200000, 8192, 0.003, 0.015, '["chat", "vision"]', false, 5),
((SELECT id FROM ai_providers WHERE code='anthropic'), 'claude-3-opus-vision', 'Claude 3 Opus (Vision)', 'vision', 200000, 4096, 0.015, 0.075, '["chat", "vision"]', false, 6),
((SELECT id FROM ai_providers WHERE code='anthropic'), 'claude-3-haiku-vision', 'Claude 3 Haiku (Vision)', 'vision', 200000, 4096, 0.00025, 0.00125, '["chat", "vision"]', false, 7)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- Google - 扩展模型
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='google'), 'gemini-1.5-pro-vision', 'Gemini 1.5 Pro (Vision)', 'vision', 1000000, 8192, 0.0035, 0.0105, '["chat", "vision"]', false, 5),
((SELECT id FROM ai_providers WHERE code='google'), 'gemini-1.5-flash-vision', 'Gemini 1.5 Flash (Vision)', 'vision', 1000000, 8192, 0.0008, 0.003, '["chat", "vision"]', false, 6),
((SELECT id FROM ai_providers WHERE code='google'), 'imagen-3', 'Imagen 3', 'image', 0, 0, 0.03, 0, '["image_generation"]', false, 10)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- MiniMax - 扩展模型
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='minimax'), 'speech-01', 'Speech-01语音合成', 'tts', 0, 0, 0.002, 0, '["tts"]', false, 20),
((SELECT id FROM ai_providers WHERE code='minimax'), 'video-01', 'Video-01视频生成', 'video', 0, 0, 0.05, 0, '["video_generation"]', false, 15)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- 智谱AI - 扩展模型
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='zhipu'), 'cogview-3', 'CogView-3', 'image', 0, 0, 0.02, 0, '["image_generation"]', false, 10),
((SELECT id FROM ai_providers WHERE code='zhipu'), 'glm-4v', 'GLM-4V (Vision)', 'vision', 8192, 2048, 0.003, 0.009, '["chat", "vision"]', false, 5),
((SELECT id FROM ai_providers WHERE code='zhipu'), 'cogvideox', 'CogVideoX', 'video', 0, 0, 0.05, 0, '["video_generation"]', false, 15)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- Moonshot - 扩展模型
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='moonshot'), 'moonshot-v1-128k-vision', 'Moonshot V1 128K (Vision)', 'vision', 128000, 4096, 0.004, 0.012, '["chat", "vision"]', false, 5)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- DeepSeek - 扩展模型
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='deepseek'), 'deepseek-vl', 'DeepSeek-VL', 'vision', 16000, 4096, 0.002, 0.006, '["chat", "vision"]', false, 5)
ON CONFLICT (provider_id, model_code) DO NOTHING;

COMMIT;
