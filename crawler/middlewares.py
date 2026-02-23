from fake_useragent import UserAgent


class RandomUserAgentMiddleware:
    def __init__(self, crawler):
        self.ua = UserAgent()
        self.crawler = crawler

    @classmethod
    def from_crawler(cls, crawler):
        return cls(crawler)

    def process_request(self, request):
        request.headers["User-Agent"] = self.ua.random
