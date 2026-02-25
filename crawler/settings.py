import os

from dotenv import load_dotenv

load_dotenv()
OUTPUT_DIR = os.getenv("OUTPUT_DIR", "output")
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")
# Build DATABASE_URL from individual vars (same as GoApp .env)
_pg_host = os.getenv("POSTGRES_HOST", "localhost")
_pg_port = os.getenv("POSTGRES_PORT", "5432")
_pg_db = os.getenv("POSTGRES_DATABASE", "go")
_pg_user = os.getenv("POSTGRES_USERNAME", "")
_pg_pass = os.getenv("POSTGRES_PASSWORD", "")
DATABASE_URL = f"postgresql://{_pg_user}:{_pg_pass}@{_pg_host}:{_pg_port}/{_pg_db}"

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
