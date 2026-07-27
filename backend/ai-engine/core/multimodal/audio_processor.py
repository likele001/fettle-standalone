"""语音处理模块"""
import logging
import os
import tempfile
from typing import Dict, Any

logger = logging.getLogger(__name__)


class AudioProcessor:
    """语音处理模块"""

    def __init__(self):
        self.supported_formats = ["mp3", "wav", "ogg", "flac", "m4a"]

    def is_supported(self, filename: str) -> bool:
        """检查文件格式是否支持"""
        ext = filename.lower().split(".")[-1]
        return ext in self.supported_formats

    async def process_audio(self, audio_data: bytes, filename: str = "audio.mp3", **kwargs) -> Dict[str, Any]:
        """处理音频"""
        try:
            with tempfile.NamedTemporaryFile(suffix=f".{filename.split('.')[-1]}", delete=False) as f:
                f.write(audio_data)
                temp_path = f.name

            result = {
                "size_bytes": len(audio_data),
                "format": filename.split(".")[-1],
                "success": True
            }

            if kwargs.get("transcribe"):
                result["transcription"] = await self._transcribe(temp_path)

            if kwargs.get("duration"):
                result["duration"] = await self._get_duration(temp_path)

            os.unlink(temp_path)
            return result

        except Exception as e:
            logger.error(f"Audio processing failed: {e}")
            return {"success": False, "error": str(e)}

    async def _transcribe(self, audio_path: str) -> str:
        """转录音频为文本"""
        try:
            import whisper
            
            model = whisper.load_model("base")
            result = model.transcribe(audio_path)
            return result.get("text", "")
        except ImportError:
            logger.warning("whisper not installed, skipping transcription")
            return ""
        except Exception as e:
            logger.error(f"Audio transcription failed: {e}")
            return ""

    async def _get_duration(self, audio_path: str) -> float:
        """获取音频时长"""
        try:
            from pydub import AudioSegment
            
            audio = AudioSegment.from_file(audio_path)
            return len(audio) / 1000.0
        except ImportError:
            logger.warning("pydub not installed, skipping duration")
            return 0.0
        except Exception as e:
            logger.error(f"Get duration failed: {e}")
            return 0.0

    async def text_to_speech(self, text: str, **kwargs) -> bytes:
        """文本转语音"""
        try:
            import edge_tts
            
            voice = kwargs.get("voice", "zh-CN-XiaoxiaoNeural")
            output_format = kwargs.get("format", "mp3")
            
            communicate = edge_tts.Communicate(text, voice)
            temp_file = tempfile.mktemp(suffix=f".{output_format}")
            
            await communicate.save(temp_file)
            
            with open(temp_file, "rb") as f:
                audio_data = f.read()
            
            os.unlink(temp_file)
            return audio_data

        except ImportError:
            logger.warning("edge_tts not installed, skipping TTS")
            return b""
        except Exception as e:
            logger.error(f"TTS failed: {e}")
            return b""


class AudioTranscribeTool:
    """音频转写工具"""

    @property
    def name(self) -> str:
        return "audio_transcribe"

    @property
    def description(self) -> str:
        return "将音频文件转写为文本"

    @property
    def parameters(self) -> Dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "audio_path": {
                    "type": "string",
                    "description": "音频文件路径"
                },
                "language": {
                    "type": "string",
                    "description": "语言（zh, en, ja等）",
                    "default": "zh"
                }
            },
            "required": ["audio_path"]
        }

    async def execute(self, **kwargs) -> str:
        audio_path = kwargs.get("audio_path", "")
        language = kwargs.get("language", "zh")
        
        try:
            import whisper
            
            model = whisper.load_model("base")
            result = model.transcribe(audio_path, language=language)
            return result.get("text", "")
        except ImportError:
            return "whisper 未安装，请安装后使用"
        except Exception as e:
            return f"音频转写失败: {str(e)}"


# 全局实例
audio_processor = AudioProcessor()