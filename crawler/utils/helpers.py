import re


def parse_price(raw):
    if not raw:
        return None
    try:
        return float(raw)
    except ValueError:
        cleaned = re.sub(r"[^\d]", "", raw)
        return float(cleaned) if cleaned else None


def parse_discount(raw):
    """Returns discount as int (e.g. 20 for 20%), or None."""
    if not raw:
        return None
    try:
        val = int(raw.strip())
        return val if val > 0 else None
    except ValueError:
        match = re.search(r"\d+", raw)
        return int(match.group()) if match else None


def clean_text(text):
    if not text:
        return None
    return " ".join(text.strip().split())


def parse_rating(raw):
    if not raw:
        return None
    match = re.search(r"\d+\.?\d*", raw)
    return float(match.group()) if match else None
