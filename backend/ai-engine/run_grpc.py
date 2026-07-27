"""AI Engine gRPC 服务启动入口"""
import logging
from concurrent import futures

import grpc

from config.settings import settings
from proto import ai_engine_pb2_grpc
from grpc_server.server import AIEngineServicer

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    ai_engine_pb2_grpc.add_AIEngineServicer_to_server(AIEngineServicer(), server)
    port = settings.app_grpc_port
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logger.info(f'AI Engine gRPC server started on port {port}')
    server.wait_for_termination()


if __name__ == '__main__':
    serve()
