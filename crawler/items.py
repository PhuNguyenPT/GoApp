from dataclasses import dataclass, field
from datetime import datetime
from typing import Optional


@dataclass
class ProductItem:
    url: str
    source: Optional[str] = None
    sku: Optional[str] = None
    name: Optional[str] = None
    brand: Optional[str] = None
    category: Optional[str] = None
    subcategory: Optional[str] = None
    description: Optional[str] = None
    images: list[str] = field(default_factory=list)
    specs: dict[str, str] = field(default_factory=dict)
    currency: str = "VND"
    price: Optional[float] = None
    original_price: Optional[float] = None
    discount_percent: Optional[int] = None
    quantity: Optional[int] = None
    in_stock: Optional[bool] = None
    rating: Optional[float] = None
    review_count: Optional[int] = None
    crawled_at: Optional[datetime] = None
