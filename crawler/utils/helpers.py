import re
from typing import Optional


def parse_price(raw):
    if not raw:
        return None
    cleaned = re.sub(r"[^\d]", "", raw)
    return float(cleaned) if cleaned else None


def parse_discount(raw):
    if not raw:
        return None
    match = re.search(r"\d+", raw)
    return int(match.group()) if match else None


def clean_text(text):
    if not text:
        return None
    return " ".join(text.strip().split())


def parse_rating(raw):
    if not raw:
        return None
    match = re.search(r"[\d.]+", raw)
    return float(match.group()) if match else None
