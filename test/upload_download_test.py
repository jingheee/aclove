#!/usr/bin/env python3
"""
测试文件上传和下载功能（使用标准库）
"""

import http.client
import json
import mimetypes
import os
import sys
import uuid
from http.cookiejar import CookieJar
from pathlib import Path
from urllib.parse import urlencode, urlparse

# 配置
BASE_URL = "localhost:8888"  # 根据实际情况修改
TEST_FILE_PATH = "/root/GolandProjects/aclove/test/category.png"

# 全局Cookie存储
cookies = {}


def encode_multipart_formdata(fields, files):
    """
    编码multipart/form-data
    Returns (content_type, body)
    """
    boundary = uuid.uuid4().hex
    lines = []

    # 添加字段
    for key, value in fields.items():
        lines.append(f"--{boundary}")
        lines.append(f'Content-Disposition: form-data; name="{key}"')
        lines.append("")
        lines.append(str(value))

    # 添加文件
    for key, (filename, filepath, content_type) in files.items():
        lines.append(f"--{boundary}")
        lines.append(
            f'Content-Disposition: form-data; name="{key}"; filename="{filename}"'
        )
        lines.append(f"Content-Type: {content_type}")
        lines.append("")
        with open(filepath, "rb") as f:
            lines.append(f.read().decode("latin-1"))

    lines.append(f"--{boundary}--")
    lines.append("")

    body = "\r\n".join(lines).encode("latin-1")
    content_type = f"multipart/form-data; boundary={boundary}"
    return content_type, body


def get_cookie_header():
    """获取Cookie请求头"""
    if cookies:
        return "; ".join([f"{k}={v}" for k, v in cookies.items()])
    return ""


def save_cookies(response_headers):
    """保存响应中的Cookie"""
    global cookies
    set_cookie = response_headers.get("Set-Cookie", "")
    if set_cookie:
        # 解析简单的Cookie
        for cookie in set_cookie.split(","):
            parts = cookie.strip().split(";")[0].split("=")
            if len(parts) == 2:
                cookies[parts[0].strip()] = parts[1].strip()


def make_request(method, path, body=None, headers=None, stream=False):
    """发送HTTP请求"""
    global cookies
    try:
        conn = http.client.HTTPConnection(BASE_URL, timeout=60)
        all_headers = headers or {}

        # 添加Cookie
        cookie_header = get_cookie_header()
        if cookie_header:
            all_headers["Cookie"] = cookie_header

        conn.request(method, path, body, all_headers)
        response = conn.getresponse()

        # 保存Cookie
        save_cookies(dict(response.getheaders()))

        if stream:
            # 流式读取，用于大文件下载
            data = b""
            while True:
                chunk = response.read(8192)
                if not chunk:
                    break
                data += chunk
        else:
            data = response.read()

        conn.close()

        return {
            "status": response.status,
            "headers": dict(response.getheaders()),
            "body": data,
        }
    except Exception as e:
        print(f"❌ 请求失败: {e}")
        return None


def init_session():
    """初始化会话 - 访问一个需要session的接口来获取cookie"""
    print("🔑 初始化会话...")

    # 先访问 /api/users/me 来获取 session cookie
    response = make_request("GET", "/api/users/me")

    if response and response["status"] == 401:
        # 401是正常的，说明我们需要获取session
        # 检查是否获取到了cookie
        if "aclove_session" in cookies:
            print(
                f"✅ 会话初始化成功，获取到Cookie: {cookies['aclove_session'][:20]}..."
            )
            return True

    # 如果没有获取到cookie，尝试直接访问上传接口
    # 因为中间件会自动创建session
    print("⚠️  尝试直接上传（中间件会自动创建session）")
    return True


def test_upload_image(file_path: str) -> dict:
    """测试上传图片"""
    url = "/api/upload/image"

    if not os.path.exists(file_path):
        print(f"❌ 文件不存在: {file_path}")
        sys.exit(1)

    file_size = os.path.getsize(file_path)
    filename = os.path.basename(file_path)
    content_type = mimetypes.guess_type(file_path)[0] or "application/octet-stream"

    print(f"📁 准备上传文件: {file_path}")
    print(f"📊 文件大小: {file_size / 1024:.2f} KB")

    # 构建multipart请求
    content_type_header, body = encode_multipart_formdata(
        fields={}, files={"file": (filename, file_path, content_type)}
    )

    headers = {
        "Content-Type": content_type_header,
        "Content-Length": str(len(body)),
    }

    print(f"\n📤 正在上传...")
    response = make_request("POST", url, body, headers)

    if response is None:
        print("❌ 无法连接到服务器")
        print("请确保服务器已启动并检查 BASE_URL 配置")
        sys.exit(1)

    print(f"📤 上传响应状态: {response['status']}")

    if response["status"] == 200:
        try:
            result = json.loads(response["body"].decode("utf-8"))
            data = result.get("data", {})
            print(f"✅ 上传成功!")
            print(f"   - URL: {data.get('url', 'N/A')}")
            print(f"   - Key: {data.get('key', 'N/A')}")
            print(f"   - Size: {data.get('size', 'N/A')} bytes")
            print(f"   - Content-Type: {data.get('content_type', 'N/A')}")
            print(f"   - Filename: {data.get('filename', 'N/A')}")
            print(f"   - Attachment ID: {data.get('attachment_id', 'N/A')}")
            print(f"   - Download URL: {data.get('download_url', 'N/A')}")
            return data
        except Exception as e:
            print(f"❌ 解析响应失败: {e}")
            print(f"   响应内容: {response['body'][:500]}")
            sys.exit(1)
    else:
        print(f"❌ 上传失败!")
        print(f"   状态码: {response['status']}")
        try:
            error_data = json.loads(response["body"].decode("utf-8"))
            print(f"   错误信息: {error_data}")
        except:
            print(f"   响应内容: {response['body'][:500]}")
        sys.exit(1)


def test_download_file(download_url: str, original_file_path: str) -> bool:
    """测试下载文件并验证完整性"""
    # 处理相对URL
    if download_url.startswith("/"):
        path = download_url
    else:
        parsed = urlparse(download_url)
        path = parsed.path
        if parsed.query:
            path += "?" + parsed.query

    print(f"\n📥 开始下载文件...")
    print(f"   Path: {path}")

    response = make_request("GET", path, stream=True)

    if response is None:
        print(f"❌ 无法连接到服务器")
        return False

    print(f"📥 下载响应状态: {response['status']}")

    if response["status"] != 200:
        print(f"❌ 下载失败!")
        try:
            error_data = json.loads(response["body"].decode("utf-8"))
            print(f"   错误信息: {error_data}")
        except:
            print(f"   响应内容: {response['body'][:500]}")
        return False

    # 保存下载的文件
    download_dir = Path("/tmp/aclove_downloads")
    download_dir.mkdir(exist_ok=True)

    original_name = os.path.basename(original_file_path)
    downloaded_file_path = download_dir / f"downloaded_{original_name}"

    with open(downloaded_file_path, "wb") as f:
        f.write(response["body"])

    total_size = len(response["body"])
    print(f"✅ 下载完成!")
    print(f"   保存路径: {downloaded_file_path}")
    print(f"   文件大小: {total_size / 1024:.2f} KB")
    print(f"   Content-Type: {response['headers'].get('Content-Type', 'N/A')}")

    # 验证文件完整性
    original_size = os.path.getsize(original_file_path)
    if total_size == original_size:
        print(f"✅ 文件大小验证通过 ({total_size} bytes)")
    else:
        print(f"⚠️  文件大小不一致!")
        print(f"   原始大小: {original_size} bytes")
        print(f"   下载大小: {total_size} bytes")
        return False

    # 对比文件内容
    with open(original_file_path, "rb") as f1, open(downloaded_file_path, "rb") as f2:
        original_content = f1.read()
        downloaded_content = f2.read()

        if original_content == downloaded_content:
            print(f"✅ 文件内容完全一致!")
            return True
        else:
            print(f"❌ 文件内容不一致!")
            return False


def test_get_attachment_info(attachment_id: int):
    """测试获取附件信息接口"""
    path = f"/api/attachments/{attachment_id}"

    print(f"\nℹ️  获取附件信息...")
    print(f"   Path: {path}")

    response = make_request("GET", path)

    if response is None:
        print(f"❌ 无法连接到服务器")
        return

    print(f"📋 响应状态: {response['status']}")

    if response["status"] == 200:
        try:
            result = json.loads(response["body"].decode("utf-8"))
            info = result.get("data", {})
            print(f"✅ 获取信息成功!")
            print(f"   - ID: {info.get('id', 'N/A')}")
            print(f"   - Filename: {info.get('filename', 'N/A')}")
            print(f"   - Size: {info.get('size', 'N/A')} bytes")
            print(f"   - Content-Type: {info.get('content_type', 'N/A')}")
            print(f"   - File Type: {info.get('file_type', 'N/A')}")
            print(f"   - Created At: {info.get('created_at', 'N/A')}")
        except Exception as e:
            print(f"❌ 解析响应失败: {e}")
            print(f"   响应内容: {response['body'][:500]}")
    else:
        print(f"❌ 获取信息失败!")
        try:
            error_data = json.loads(response["body"].decode("utf-8"))
            print(f"   错误信息: {error_data}")
        except:
            print(f"   响应内容: {response['body'][:500]}")


def main():
    print("=" * 60)
    print("🧪 文件上传下载测试")
    print("=" * 60)

    # 初始化会话
    init_session()

    # 1. 测试上传
    upload_result = test_upload_image(TEST_FILE_PATH)

    # 2. 测试获取附件信息
    attachment_id = upload_result.get("attachment_id")
    if attachment_id:
        test_get_attachment_info(attachment_id)

    # 3. 测试下载
    download_url = upload_result.get("download_url")
    if download_url:
        success = test_download_file(download_url, TEST_FILE_PATH)

        print("\n" + "=" * 60)
        if success:
            print("🎉 所有测试通过!")
        else:
            print("⚠️  测试未完全通过，请检查日志")
        print("=" * 60)
    else:
        print("\n❌ 未获取到下载URL，跳过下载测试")


if __name__ == "__main__":
    main()
