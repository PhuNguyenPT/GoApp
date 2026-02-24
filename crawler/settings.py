import os

from dotenv import load_dotenv

load_dotenv()

BOT_NAME = "crawler"
SPIDER_MODULES = ["spiders"]
NEWSPIDER_MODULE = "spiders"

ROBOTSTXT_OBEY = False
CONCURRENT_REQUESTS = 8
CONCURRENT_REQUESTS_PER_DOMAIN = 2
DOWNLOAD_DELAY = 1.5
RANDOMIZE_DOWNLOAD_DELAY = True

RETRY_ENABLED = True
RETRY_TIMES = 3
RETRY_HTTP_CODES = [500, 502, 503, 504, 408, 429]

DOWNLOAD_TIMEOUT = 30

DOWNLOADER_MIDDLEWARES = {
    "middlewares.RandomUserAgentMiddleware": 400,
}

ITEM_PIPELINES = {
    "pipelines.JsonWriterPipeline": 100,
    "pipelines.PostgresPipeline": 200,
}
DOWNLOAD_HANDLERS = {
    "http": "scrapy_playwright.handler.ScrapyPlaywrightDownloadHandler",
    "https": "scrapy_playwright.handler.ScrapyPlaywrightDownloadHandler",
}
TWISTED_REACTOR = "twisted.internet.asyncioreactor.AsyncioSelectorReactor"
OUTPUT_DIR = os.getenv("OUTPUT_DIR", "output")
DATABASE_URL = os.getenv("DATABASE_URL", "")
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")

_ABORT_TYPES = {"image", "media", "font", "ping", "stylesheet"}
_ABORT_DOMAINS = {
    "google-analytics.com", "googletagmanager.com", "facebook.net",
    "facebook.com", "hotjar.com", "zalo.me", "clarity.ms",
    "doubleclick.net", "adservice.google.com",
}

def _should_abort_request(req):
    if req.resource_type in _ABORT_TYPES:
        return True
    if req.resource_type in {"xhr", "fetch", "script"}:
        return any(d in req.url for d in _ABORT_DOMAINS)
    return False

PLAYWRIGHT_ABORT_REQUEST = _should_abort_request
