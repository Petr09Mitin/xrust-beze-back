from concurrent import futures
import grpc
import time
import os

import nsfw_pb2
import nsfw_pb2_grpc
from s3_utils import download_file_from_s3
from image_detect import check_image
import logging


class S3Processor(nsfw_pb2_grpc.S3ProcessorServicer):
    def ProcessFile(self, request, context):
        local_path = f"/tmp/{request.file_id}"

        logging.info(f"Backet: {request.bucket_name} | Filename: {request.file_id}")
        logging.info(f"Downloading file: {local_path}")

        success = download_file_from_s3(request.bucket_name, request.file_id, local_path)

        if not success:
            logging.error('Something went wrong')
            context.set_code(grpc.StatusCode.UNAVAILABLE)
        else:
            is_nsfw, prob = check_image(local_path)
            logging.info(f'is_nsfw: {is_nsfw} | prob: {prob}')

            os.remove(local_path)

            logging.info(f'Delete file: {local_path} | Status: {not os.path.exists(local_path)}')

            return nsfw_pb2.ProcessResult(is_nsfw=is_nsfw, prob=prob)


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    nsfw_pb2_grpc.add_S3ProcessorServicer_to_server(S3Processor(), server)
    port = os.getenv('GRPC_PORT', '50051')
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    print(f"gRPC server started on port {port}")
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        server.stop(0)


if __name__ == '__main__':
    serve()
