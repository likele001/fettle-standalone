"""图像处理模块"""
import logging
import base64
from typing import Dict, Any, Optional
from PIL import Image
from io import BytesIO

logger = logging.getLogger(__name__)


class ImageProcessor:
    """图像处理模块"""

    def __init__(self):
        self.supported_formats = ["jpg", "jpeg", "png", "gif", "bmp", "webp"]

    def is_supported(self, filename: str) -> bool:
        """检查文件格式是否支持"""
        ext = filename.lower().split(".")[-1]
        return ext in self.supported_formats

    async def process_image(self, image_data: bytes, **kwargs) -> Dict[str, Any]:
        """处理图像"""
        try:
            image = Image.open(BytesIO(image_data))
            
            result = {
                "width": image.width,
                "height": image.height,
                "mode": image.mode,
                "format": image.format,
                "size_bytes": len(image_data),
                "success": True
            }

            if kwargs.get("resize"):
                result.update(await self._resize_image(image, kwargs["resize"]))

            if kwargs.get("convert_to_base64"):
                result["base64"] = await self._to_base64(image)

            if kwargs.get("extract_text"):
                result["extracted_text"] = await self._extract_text(image_data)

            return result

        except Exception as e:
            logger.error(f"Image processing failed: {e}")
            return {"success": False, "error": str(e)}

    async def _resize_image(self, image: Image.Image, size: Dict[str, int]) -> Dict[str, Any]:
        """调整图像大小"""
        try:
            width = size.get("width", image.width)
            height = size.get("height", image.height)
            resized = image.resize((width, height))
            
            buffer = BytesIO()
            resized.save(buffer, format=image.format or "PNG")
            
            return {
                "resized_width": width,
                "resized_height": height,
                "resized_size_bytes": len(buffer.getvalue())
            }
        except Exception as e:
            logger.error(f"Image resize failed: {e}")
            return {}

    async def _to_base64(self, image: Image.Image) -> str:
        """转换为 Base64"""
        try:
            buffer = BytesIO()
            image.save(buffer, format=image.format or "PNG")
            return base64.b64encode(buffer.getvalue()).decode("utf-8")
        except Exception as e:
            logger.error(f"Image to base64 failed: {e}")
            return ""

    async def _extract_text(self, image_data: bytes) -> str:
        """提取图像中的文本（OCR）"""
        try:
            import pytesseract
            
            image = Image.open(BytesIO(image_data))
            text = pytesseract.image_to_string(image, lang="chi_sim+eng")
            return text.strip()
        except ImportError:
            logger.warning("pytesseract not installed, skipping OCR")
            return ""
        except Exception as e:
            logger.error(f"OCR failed: {e}")
            return ""

    async def analyze_image_for_llm(self, image_path: str, detail: str = "low") -> Dict[str, Any]:
        """分析图像供 LLM 使用"""
        try:
            with open(image_path, "rb") as f:
                image_data = f.read()
            
            image = Image.open(BytesIO(image_data))
            base64_str = await self._to_base64(image)
            
            return {
                "success": True,
                "base64": base64_str,
                "width": image.width,
                "height": image.height,
                "detail": detail
            }
        except Exception as e:
            logger.error(f"Image analysis failed: {e}")
            return {"success": False, "error": str(e)}


class ImageToTextTool:
    """图像转文本工具"""

    @property
    def name(self) -> str:
        return "image_to_text"

    @property
    def description(self) -> str:
        return "提取图像中的文本内容"

    @property
    def parameters(self) -> Dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "image_path": {
                    "type": "string",
                    "description": "图像文件路径"
                },
                "lang": {
                    "type": "string",
                    "description": "语言（chi_sim+eng）",
                    "default": "chi_sim+eng"
                }
            },
            "required": ["image_path"]
        }

    async def execute(self, **kwargs) -> str:
        image_path = kwargs.get("image_path", "")
        try:
            import pytesseract
            from PIL import Image
            
            image = Image.open(image_path)
            text = pytesseract.image_to_string(image, lang=kwargs.get("lang", "chi_sim+eng"))
            return text.strip()
        except ImportError:
            return "pytesseract 未安装，请安装后使用"
        except Exception as e:
            return f"图像转文本失败: {str(e)}"


# 全局实例
image_processor = ImageProcessor()