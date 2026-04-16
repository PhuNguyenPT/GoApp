"""
Reimport scraped JSONL files back into the database.
Usage:
    python reimport.py                        # imports all files in OUTPUT_DIR
    python reimport.py output/fpt_*.jsonl     # imports specific files
"""

import json
import logging
import os
import sys

import psycopg2
from dotenv import load_dotenv

from db import save_item

load_dotenv()

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

BATCH_SIZE = 2000


def get_files():
    if len(sys.argv) > 1:
        return sys.argv[1:]
    output_dir = os.getenv("OUTPUT_DIR", "output")
    files = sorted(
        os.path.join(output_dir, f) for f in os.listdir(output_dir) if f.endswith(".jsonl")
    )
    logger.info(f"Found {len(files)} JSONL file(s): {[os.path.basename(f) for f in files]}")
    return files


def reimport_file(conn, filepath):
    cur = conn.cursor()
    count = 0
    errors = 0

    with open(filepath, encoding="utf-8") as f:
        items = [json.loads(line) for line in f if line.strip()]

    logger.info(f"  Importing {len(items)} items from {os.path.basename(filepath)} ...")

    for data in items:
        try:
            save_item(cur, data)
            count += 1
            if count % BATCH_SIZE == 0:
                conn.commit()
                logger.info(f"  ... {count} items committed")
        except Exception as e:
            conn.rollback()
            errors += 1
            logger.warning(f"  Skipped item (url={data.get('url')}): {e}")

    conn.commit()
    cur.close()
    logger.info(f"  Done: {count} imported, {errors} skipped.")
    return count, errors


def main():
    files = get_files()
    if not files:
        logger.error("No JSONL files found.")
        sys.exit(1)

    db_url = os.getenv(
        "DATABASE_URL"
    ) or "postgresql://{user}:{password}@{host}:{port}/{db}".format(
        user=os.getenv("POSTGRES_USERNAME", ""),
        password=os.getenv("POSTGRES_PASSWORD", ""),
        host=os.getenv("POSTGRES_HOST", "localhost"),
        port=os.getenv("POSTGRES_PORT", "5432"),
        db=os.getenv("POSTGRES_DATABASE", "go"),
    )

    conn = psycopg2.connect(db_url)
    total_imported, total_errors = 0, 0

    for filepath in files:
        logger.info(f"Processing: {filepath}")
        imported, errors = reimport_file(conn, filepath)
        total_imported += imported
        total_errors += errors

    conn.close()
    logger.info(f"All done — total imported: {total_imported}, total skipped: {total_errors}")


if __name__ == "__main__":
    main()
