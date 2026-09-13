#!/usr/bin/env python3

import argparse
import asyncio
import hashlib
import json
import mimetypes
import os
import sys
from pathlib import Path
from urllib.parse import urlparse

import aiohttp
from pinscrape import Pinterest


def sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def extension_from(url: str, data: bytes) -> str | None:
    """Return an extension accepted by the image API, based on bytes first."""
    if data.startswith(b"\xff\xd8\xff"):
        return ".jpg"
    if data.startswith(b"\x89PNG\r\n\x1a\n"):
        return ".png"
    if data.startswith((b"GIF87a", b"GIF89a")):
        return ".gif"

    extension = Path(urlparse(url).path).suffix.lower()
    if extension in {".jpg", ".jpeg", ".png", ".gif"}:
        return ".jpg" if extension == ".jpeg" else extension
    return None


async def download_image(
    session: aiohttp.ClientSession,
    url: str,
    semaphore: asyncio.Semaphore,
) -> tuple[str, bytes] | None:

    async with semaphore:
        try:
            headers = {
                "User-Agent": (
                    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
                    "AppleWebKit/537.36 "
                    "(KHTML, like Gecko) "
                    "Chrome/139.0 Safari/537.36"
                ),
                "Accept": "image/avif,image/webp,image/apng,image/*,*/*;q=0.8",
            }

            async with session.get(
                url,
                headers=headers,
                timeout=aiohttp.ClientTimeout(total=60),
            ) as response:
                if response.status != 200:
                    print(
                        f"[download] HTTP {response.status}: {url}",
                        file=sys.stderr,
                    )
                    return None

                data = await response.read()

                if not data:
                    return None

                # Avoid accidentally downloading HTML.
                content_type = response.headers.get("Content-Type", "")

                if (
                    not content_type.startswith("image/")
                    and extension_from(url, data) is None
                ):
                    return None

                return url, data

        except Exception as exc:
            print(
                f"[download] failed {url}: {exc}",
                file=sys.stderr,
            )
            return None


async def upload_batch(
    session: aiohttp.ClientSession,
    endpoint: str,
    images: list[tuple[str, bytes]],
    tags: list[str],
    token: str | None,
) -> None:
    form = aiohttp.FormData()

    metadata = {
        "data": [],
        "common_tags": tags,
    }

    for source_url, data in images:
        ext = extension_from(source_url, data)
        if ext is None:
            print(f"[upload] unsupported image format: {source_url}", file=sys.stderr)
            continue

        # Batch indexes repeat between batches; the content hash is stable and
        # avoids duplicate filenames across an entire seed run.
        filename = f"pinterest_{sha256(data)[:16]}{ext}"

        metadata["data"].append(
            {
                "name": filename,
                "tags": [],
            }
        )

        content_type = mimetypes.guess_type(filename)[0] or "image/jpeg"

        form.add_field(
            "files",
            data,
            filename=filename,
            content_type=content_type,
        )

    # Must be a JSON string because the endpoint declares:
    # @Param metadata formData string
    if not metadata["data"]:
        return

    form.add_field(
        "metadata",
        json.dumps(metadata),
        content_type="application/json",
    )

    headers = {}

    if token:
        headers["Authorization"] = f"Bearer {token}"

    try:
        async with session.post(
            endpoint,
            data=form,
            headers=headers,
            timeout=aiohttp.ClientTimeout(total=180),
        ) as response:
            body = await response.text()

            if response.status >= 400:
                print(
                    f"[upload] HTTP {response.status}: {body}",
                    file=sys.stderr,
                )
                return

            print(f"[upload] {len(images)} images -> HTTP {response.status}")

            print(body)

    except Exception as exc:
        print(
            f"[upload] failed: {exc}",
            file=sys.stderr,
        )


p = Pinterest(proxies={}, sleep_time=2)


async def seed(args):
    connector = aiohttp.TCPConnector(
        limit=args.concurrency,
        ttl_dns_cache=300,
    )

    image_urls = []
    async with aiohttp.ClientSession(
        connector=connector,
    ) as session:
        processed = 0
        while processed < args.limit:
            # ------------------------------------------------------------
            # 1. Crawl Pinterest pages
            # ------------------------------------------------------------

            urls = p.search(args.query, args.pinterest_limit)
            urls_count = len(urls)
            if urls_count == 0:
                break
            processed += urls_count
            image_urls = image_urls + urls
            print(f"Found {len(urls)} images for {args.query} query")

            # ------------------------------------------------------------
            # 2. Download
            # ------------------------------------------------------------

            semaphore = asyncio.Semaphore(args.concurrency)

            tasks = [
                download_image(
                    session,
                    url,
                    semaphore,
                )
                for url in urls
            ]

            results = await asyncio.gather(*tasks)

            downloaded = [result for result in results if result is not None]

            print(f"[download] downloaded {len(downloaded)} images")

            # ------------------------------------------------------------
            # 3. Local SHA-256 deduplication
            # ------------------------------------------------------------

            unique: list[tuple[str, bytes]] = []
            seen_hashes: set[str] = set()

            for source_url, data in downloaded:
                image_hash = sha256(data)

                if image_hash in seen_hashes:
                    print(f"[dedup] duplicate: {source_url}")
                    continue

                seen_hashes.add(image_hash)

                unique.append((source_url, data))

            print(f"[dedup] {len(unique)} unique images")

            # ------------------------------------------------------------
            # 4. Upload in batches
            # ------------------------------------------------------------

            for start in range(
                0,
                len(unique),
                args.batch_size,
            ):
                batch = unique[start : start + args.batch_size]

                print(f"[batch] {start + 1}-{start + len(batch)} / {len(unique)}")

                await upload_batch(
                    session=session,
                    endpoint=args.endpoint,
                    images=batch,
                    tags=args.tags,
                    token=args.token,
                )


def main():
    parser = argparse.ArgumentParser(
        description="Seed Kartonko from public Pinterest pages."
    )

    parser.add_argument(
        "--query",
        type=str,
        required=True,
        help="search query",
    )

    parser.add_argument(
        "--endpoint",
        default=os.getenv(
            "KARTONKO_BATCH_ENDPOINT",
            "http://localhost:3000/image/upload/batch",
        ),
    )

    parser.add_argument(
        "--token",
        default=os.getenv("KARTONKO_TOKEN"),
    )

    parser.add_argument(
        "--pinterest-limit",
        type=int,
        default=20,
    )

    parser.add_argument(
        "--limit",
        type=int,
        default=100,
    )

    parser.add_argument(
        "--batch-size",
        type=int,
        default=20,
    )

    parser.add_argument(
        "--concurrency",
        type=int,
        default=8,
    )

    parser.add_argument(
        "--tag",
        dest="tags",
        action="append",
        default=[],
        help="Tag applied to imported images. Can be repeated.",
    )

    args = parser.parse_args()
    asyncio.run(seed(args))


if __name__ == "__main__":
    main()
