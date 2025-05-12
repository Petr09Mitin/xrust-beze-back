import boto3
from botocore.exceptions import ClientError
import os

s3 = boto3.client(
        's3',
        aws_access_key_id=os.getenv('MINIO_ROOT_USER'),
        aws_secret_access_key=os.getenv('MINIO_ROOT_PASSWORD'),
        endpoint_url=os.getenv('S3_ENDPOINT_URL')
)


def download_file_from_s3(bucket_name: str, file_key: str, download_path: str) -> bool:
    try:
        s3.download_file(bucket_name, file_key, download_path)
        return True
    except ClientError as e:
        print(f"Failed to download file: {e}")
        return False
