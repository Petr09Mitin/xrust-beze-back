import grpc
import nsfw_pb2
import nsfw_pb2_grpc


def run():
    # Подключение к gRPC-серверу
    channel = grpc.insecure_channel('localhost:8199')  # замени на нужный адрес
    stub = nsfw_pb2_grpc.S3ProcessorStub(channel)

    # Создание запроса
    request = nsfw_pb2.FileRequest(
        bucket_name="testbacket",
        file_id="porn1.webp"
    )

    # Вызов метода ProcessFile
    response = stub.ProcessFile(request)

    # Вывод результата
    print("NSFW:", response.is_nsfw)
    print("Probability:", response.prob)


if __name__ == '__main__':
    run()
