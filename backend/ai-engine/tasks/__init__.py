"""任务模块"""
from .celery_app import celery_app
from . import async_tasks

__all__ = ['celery_app', 'async_tasks']
