"""网页爬取器"""
import logging
import requests
from bs4 import BeautifulSoup
import html2text
from typing import Optional, Dict, Any

logger = logging.getLogger(__name__)


class WebCrawler:
    """网页爬取器"""

    def __init__(self, timeout: int = 30, max_depth: int = 1):
        self.timeout = timeout
        self.max_depth = max_depth
        self.headers = {
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
            "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
            "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
        }

    async def crawl(self, url: str, depth: int = 0) -> Optional[str]:
        """爬取网页内容"""
        if depth > self.max_depth:
            return None

        try:
            response = requests.get(url, headers=self.headers, timeout=self.timeout)
            response.raise_for_status()

            soup = BeautifulSoup(response.text, "html.parser")

            title = soup.title.string if soup.title else ""
            content = self._extract_content(soup)

            return f"# {title}\n\n{content}"

        except Exception as e:
            logger.error(f"Failed to crawl {url}: {e}")
            return None

    def _extract_content(self, soup: BeautifulSoup) -> str:
        """提取网页正文内容"""
        for tag in soup(["script", "style", "nav", "footer", "header", "aside"]):
            tag.decompose()

        main_content = soup.find("main") or soup.find("article") or soup.find("div", class_="content")
        if main_content:
            html_content = str(main_content)
        else:
            html_content = str(soup.body)

        h2t = html2text.HTML2Text()
        h2t.ignore_links = False
        h2t.ignore_images = True
        h2t.body_width = 0

        return h2t.handle(html_content).strip()


class DocumentFetcher:
    """在线文档获取器"""

    SUPPORTED_TYPES = ["url", "pdf", "doc", "docx"]

    def __init__(self):
        self.crawler = WebCrawler()

    async def fetch(self, source: str, source_type: str = "url") -> Dict[str, Any]:
        """
        获取在线文档内容
        
        Args:
            source: 文档来源（URL 或文件路径）
            source_type: 来源类型
            
        Returns:
            包含文档内容的字典
        """
        if source_type == "url":
            content = await self.crawler.crawl(source)
            if content:
                return {
                    "success": True,
                    "content": content,
                    "source": source,
                    "type": "url"
                }
            return {"success": False, "error": "Failed to fetch URL"}

        return {"success": False, "error": f"Unsupported source type: {source_type}"}