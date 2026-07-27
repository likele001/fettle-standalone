"""内置工具实现"""
import datetime
import json
import logging
import os
from typing import Any, Dict

import httpx

from .tool_registry import BaseTool, tool_registry

logger = logging.getLogger(__name__)


class CurrentTimeTool(BaseTool):
    """获取当前时间"""

    @property
    def name(self) -> str:
        return "current_time"

    @property
    def description(self) -> str:
        return "获取当前日期和时间，支持时区参数"

    @property
    def parameters(self) -> Dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "format": {
                    "type": "string",
                    "description": "时间格式，如 '%Y-%m-%d %H:%M:%S'",
                    "default": "%Y-%m-%d %H:%M:%S"
                },
                "timezone": {
                    "type": "string",
                    "description": "时区偏移，如 '+08:00'",
                    "default": "+08:00"
                }
            }
        }

    async def execute(self, **kwargs) -> str:
        fmt = kwargs.get("format", "%Y-%m-%d %H:%M:%S")
        now = datetime.datetime.now()
        return now.strftime(fmt)


class CalculatorTool(BaseTool):
    """数学计算器"""

    @property
    def name(self) -> str:
        return "calculator"

    @property
    def description(self) -> str:
        return "执行数学计算，支持 + - * / ** % 等运算"

    @property
    def parameters(self) -> Dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "expression": {
                    "type": "string",
                    "description": "数学表达式，如 '1 + 2 * 3'"
                }
            },
            "required": ["expression"]
        }

    async def execute(self, **kwargs) -> str:
        expr = kwargs.get("expression", "")
        try:
            safe_list = {
                "abs": abs, "round": round, "max": max, "min": min,
                "sum": sum, "pow": pow, "int": int, "float": float
            }
            result = eval(expr, {"__builtins__": {}}, safe_list)
            return str(result)
        except Exception as e:
            return f"计算错误: {str(e)}"


class WebSearchTool(BaseTool):
    """网络搜索"""

    @property
    def name(self) -> str:
        return "web_search"

    @property
    def description(self) -> str:
        return "搜索网络信息，返回搜索结果摘要"

    @property
    def parameters(self) -> Dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "query": {
                    "type": "string",
                    "description": "搜索关键词"
                },
                "max_results": {
                    "type": "integer",
                    "description": "最大返回结果数",
                    "default": 5
                }
            },
            "required": ["query"]
        }

    async def execute(self, **kwargs) -> str:
        query = kwargs.get("query", "")
        max_results = kwargs.get("max_results", 5)
        try:
            url = f"https://www.baidu.com/s?wd={query}"
            async with httpx.AsyncClient(timeout=10) as client:
                resp = await client.get(url, headers={
                    "User-Agent": "Mozilla/5.0 (compatible; AIBot/1.0)"
                })
                return f"搜索'{query}'返回状态码: {resp.status_code}，内容长度: {len(resp.text)}字符"
        except Exception as e:
            return f"搜索失败: {str(e)}"


class HttpGetTool(BaseTool):
    """HTTP GET 请求"""

    @property
    def name(self) -> str:
        return "http_get"

    @property
    def description(self) -> str:
        return "发送 HTTP GET 请求获取 URL 内容"

    @property
    def parameters(self) -> Dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "url": {
                    "type": "string",
                    "description": "请求的 URL"
                },
                "timeout": {
                    "type": "integer",
                    "description": "超时时间（秒）",
                    "default": 10
                }
            },
            "required": ["url"]
        }

    async def execute(self, **kwargs) -> str:
        url = kwargs.get("url", "")
        timeout = kwargs.get("timeout", 10)
        try:
            async with httpx.AsyncClient(timeout=timeout) as client:
                resp = await client.get(url)
                return f"状态码: {resp.status_code}\n内容: {resp.text[:2000]}"
        except Exception as e:
            return f"请求失败: {str(e)}"


class HttpPostTool(BaseTool):
    """HTTP POST 请求"""

    @property
    def name(self) -> str:
        return "http_post"

    @property
    def description(self) -> str:
        return "发送 HTTP POST 请求"

    @property
    def parameters(self) -> Dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "url": {
                    "type": "string",
                    "description": "请求的 URL"
                },
                "data": {
                    "type": "object",
                    "description": "POST 数据"
                },
                "headers": {
                    "type": "object",
                    "description": "请求头"
                },
                "timeout": {
                    "type": "integer",
                    "description": "超时时间（秒）",
                    "default": 10
                }
            },
            "required": ["url"]
        }

    async def execute(self, **kwargs) -> str:
        url = kwargs.get("url", "")
        data = kwargs.get("data", {})
        headers = kwargs.get("headers", {})
        timeout = kwargs.get("timeout", 10)
        try:
            async with httpx.AsyncClient(timeout=timeout) as client:
                resp = await client.post(url, json=data, headers=headers)
                return f"状态码: {resp.status_code}\n响应: {resp.text[:2000]}"
        except Exception as e:
            return f"请求失败: {str(e)}"


class JsonParserTool(BaseTool):
    """JSON 解析器"""

    @property
    def name(self) -> str:
        return "json_parse"

    @property
    def description(self) -> str:
        return "解析 JSON 字符串并格式化输出"

    @property
    def parameters(self) -> Dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "json_string": {
                    "type": "string",
                    "description": "要解析的 JSON 字符串"
                }
            },
            "required": ["json_string"]
        }

    async def execute(self, **kwargs) -> str:
        json_str = kwargs.get("json_string", "")
        try:
            data = json.loads(json_str)
            return json.dumps(data, ensure_ascii=False, indent=2)
        except json.JSONDecodeError as e:
            return f"JSON 解析错误: {str(e)}"


class FileReadTool(BaseTool):
    """文件读取工具"""

    @property
    def name(self) -> str:
        return "file_read"

    @property
    def description(self) -> str:
        return "读取指定文件内容"

    @property
    def parameters(self) -> Dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "file_path": {
                    "type": "string",
                    "description": "文件路径"
                },
                "encoding": {
                    "type": "string",
                    "description": "文件编码",
                    "default": "utf-8"
                }
            },
            "required": ["file_path"]
        }

    async def execute(self, **kwargs) -> str:
        file_path = kwargs.get("file_path", "")
        encoding = kwargs.get("encoding", "utf-8")
        try:
            with open(file_path, 'r', encoding=encoding) as f:
                content = f.read()
            return content[:5000]
        except Exception as e:
            return f"文件读取失败: {str(e)}"


class FileWriteTool(BaseTool):
    """文件写入工具"""

    @property
    def name(self) -> str:
        return "file_write"

    @property
    def description(self) -> str:
        return "写入内容到指定文件"

    @property
    def parameters(self) -> Dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "file_path": {
                    "type": "string",
                    "description": "文件路径"
                },
                "content": {
                    "type": "string",
                    "description": "要写入的内容"
                },
                "append": {
                    "type": "boolean",
                    "description": "是否追加模式",
                    "default": False
                },
                "encoding": {
                    "type": "string",
                    "description": "文件编码",
                    "default": "utf-8"
                }
            },
            "required": ["file_path", "content"]
        }

    async def execute(self, **kwargs) -> str:
        file_path = kwargs.get("file_path", "")
        content = kwargs.get("content", "")
        append = kwargs.get("append", False)
        encoding = kwargs.get("encoding", "utf-8")
        try:
            mode = 'a' if append else 'w'
            with open(file_path, mode, encoding=encoding) as f:
                f.write(content)
            return f"文件写入成功: {file_path}"
        except Exception as e:
            return f"文件写入失败: {str(e)}"


class ListDirTool(BaseTool):
    """列出目录内容"""

    @property
    def name(self) -> str:
        return "list_dir"

    @property
    def description(self) -> str:
        return "列出指定目录的文件和子目录"

    @property
    def parameters(self) -> Dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "dir_path": {
                    "type": "string",
                    "description": "目录路径"
                },
                "show_hidden": {
                    "type": "boolean",
                    "description": "是否显示隐藏文件",
                    "default": False
                }
            },
            "required": ["dir_path"]
        }

    async def execute(self, **kwargs) -> str:
        dir_path = kwargs.get("dir_path", "")
        show_hidden = kwargs.get("show_hidden", False)
        try:
            items = []
            for item in os.listdir(dir_path):
                if not show_hidden and item.startswith('.'):
                    continue
                full_path = os.path.join(dir_path, item)
                is_dir = os.path.isdir(full_path)
                items.append(f"{'[DIR]' if is_dir else '[FILE]'} {item}")
            return "\n".join(items)
        except Exception as e:
            return f"目录读取失败: {str(e)}"


class EmailSendTool(BaseTool):
    """邮件发送工具"""

    @property
    def name(self) -> str:
        return "email_send"

    @property
    def description(self) -> str:
        return "发送邮件"

    @property
    def parameters(self) -> Dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "to": {
                    "type": "string",
                    "description": "收件人邮箱"
                },
                "subject": {
                    "type": "string",
                    "description": "邮件主题"
                },
                "body": {
                    "type": "string",
                    "description": "邮件正文"
                },
                "smtp_server": {
                    "type": "string",
                    "description": "SMTP 服务器",
                    "default": "smtp.gmail.com"
                },
                "smtp_port": {
                    "type": "integer",
                    "description": "SMTP 端口",
                    "default": 587
                },
                "username": {
                    "type": "string",
                    "description": "SMTP 用户名"
                },
                "password": {
                    "type": "string",
                    "description": "SMTP 密码"
                }
            },
            "required": ["to", "subject", "body", "username", "password"]
        }

    async def execute(self, **kwargs) -> str:
        to = kwargs.get("to", "")
        subject = kwargs.get("subject", "")
        body = kwargs.get("body", "")
        smtp_server = kwargs.get("smtp_server", "smtp.gmail.com")
        smtp_port = kwargs.get("smtp_port", 587)
        username = kwargs.get("username", "")
        password = kwargs.get("password", "")

        try:
            import smtplib
            from email.mime.text import MIMEText

            msg = MIMEText(body, 'plain', 'utf-8')
            msg['Subject'] = subject
            msg['From'] = username
            msg['To'] = to

            with smtplib.SMTP(smtp_server, smtp_port) as server:
                server.starttls()
                server.login(username, password)
                server.sendmail(username, [to], msg.as_string())

            return f"邮件发送成功: {to}"
        except Exception as e:
            return f"邮件发送失败: {str(e)}"


class ScheduleTool(BaseTool):
    """日程管理工具"""

    @property
    def name(self) -> str:
        return "schedule"

    @property
    def description(self) -> str:
        return "管理日程安排"

    @property
    def parameters(self) -> Dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "action": {
                    "type": "string",
                    "description": "操作类型: add, list, delete",
                    "enum": ["add", "list", "delete"]
                },
                "event_name": {
                    "type": "string",
                    "description": "事件名称"
                },
                "event_date": {
                    "type": "string",
                    "description": "事件日期，格式: YYYY-MM-DD"
                },
                "event_time": {
                    "type": "string",
                    "description": "事件时间，格式: HH:MM"
                },
                "event_id": {
                    "type": "string",
                    "description": "事件 ID"
                }
            },
            "required": ["action"]
        }

    async def execute(self, **kwargs) -> str:
        action = kwargs.get("action", "")
        event_name = kwargs.get("event_name", "")
        event_date = kwargs.get("event_date", "")
        event_time = kwargs.get("event_time", "")
        event_id = kwargs.get("event_id", "")

        schedule_file = os.path.expanduser("~/.fettle_schedule.json")

        try:
            if os.path.exists(schedule_file):
                with open(schedule_file, 'r') as f:
                    schedule = json.load(f)
            else:
                schedule = []

            if action == "add":
                event = {
                    "id": str(len(schedule) + 1),
                    "name": event_name,
                    "date": event_date,
                    "time": event_time,
                    "created_at": datetime.datetime.now().isoformat()
                }
                schedule.append(event)
                with open(schedule_file, 'w') as f:
                    json.dump(schedule, f, indent=2)
                return f"日程添加成功: {event_name} ({event_date} {event_time})"

            elif action == "list":
                if not schedule:
                    return "暂无日程"
                result = []
                for event in schedule:
                    result.append(f"{event['id']}. {event['name']} - {event['date']} {event['time']}")
                return "\n".join(result)

            elif action == "delete":
                schedule = [e for e in schedule if e['id'] != event_id]
                with open(schedule_file, 'w') as f:
                    json.dump(schedule, f, indent=2)
                return f"日程删除成功: {event_id}"

            else:
                return f"未知操作: {action}"

        except Exception as e:
            return f"日程操作失败: {str(e)}"


def register_builtin_tools():
    """注册所有内置工具"""
    tools = [
        CurrentTimeTool(),
        CalculatorTool(),
        WebSearchTool(),
        HttpGetTool(),
        HttpPostTool(),
        JsonParserTool(),
        FileReadTool(),
        FileWriteTool(),
        ListDirTool(),
        EmailSendTool(),
        ScheduleTool(),
    ]
    for tool in tools:
        tool_registry.register_tool_class(type(tool))
        logger.info(f"Registered builtin tool: {tool.name}")