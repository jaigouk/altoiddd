# DuckDuckGo Search Library Evaluation

**Date:** 2026-03-06
**Type:** Research Spike
**Context:** WebSearchPort adapter for alto's knowledge base or research features

---

## 1. Package Identity and Status

### Important: Package Rename

The original `duckduckgo-search` package (PyPI: `duckduckgo-search`) has been **renamed and superseded** by `ddgs` (PyPI: `ddgs`). They are by the same author (deedy5) and the `ddgs` package is the actively maintained successor.

| Property | `duckduckgo-search` (legacy) | `ddgs` (current) |
|----------|------------------------------|-------------------|
| **Latest version** | 8.1.1 (July 6, 2025) | 9.11.2 (March 5, 2026) |
| **Status** | Deprecated / frozen | Actively maintained |
| **Python** | >= 3.9 | >= 3.10 |
| **License** | MIT | MIT |
| **Install** | `pip install duckduckgo-search` | `pip install ddgs` |
| **Repo** | github.com/deedy5/duckduckgo_search | github.com/deedy5/ddgs |

**Source:** [PyPI ddgs](https://pypi.org/project/ddgs/), [GitHub releases](https://github.com/deedy5/ddgs/releases)

**Recommendation: Use `ddgs`, not `duckduckgo-search`.** The `ddgs` package has 835+ commits, 5 releases in the last 4 months, and is the only one receiving fixes.

### What DDGS Actually Is

DDGS stands for "Dux Distributed Global Search." It is a **metasearch library**, not a DuckDuckGo wrapper. It aggregates results from multiple search backends:
- **Text search:** Bing, Brave, DuckDuckGo, Google, and others
- **Images/Videos:** DuckDuckGo backend
- **News:** Bing, DuckDuckGo, Yahoo
- **Books:** Anna's Archive

**Source:** [GitHub README](https://github.com/deedy5/ddgs)

---

## 2. Dependencies

From `pyproject.toml` ([source](https://github.com/deedy5/ddgs/blob/main/pyproject.toml)):

```
primp >= 0.15.0          # HTTP client (Rust-based, anti-fingerprint)
lxml >= 4.9.4            # HTML parsing
httpx[http2,socks,brotli] >= 0.28.1  # HTTP client with proxy support
fake-useragent >= 2.2.0  # User agent rotation
```

**Concern:** `primp` is a Rust-based HTTP client library that requires compilation. This adds build complexity compared to pure-Python alternatives. However, it provides pre-built wheels for major platforms.

---

## 3. API: Text Search

### Minimal Code Example

```python
from ddgs import DDGS

results = DDGS().text("python programming", max_results=5)
for r in results:
    print(r["title"], r["href"])
```

### Method Signature (current `ddgs` v9.x)

```python
class DDGS:
    def __init__(self, proxy: str | None = None, timeout: int = 5, verify: bool = True): ...

    def text(
        self,
        query: str,
        *,
        region: str = "us-en",
        safesearch: str = "moderate",
        timelimit: str | None = None,    # "d", "w", "m", "y"
        backend: str = "auto",           # "auto", "html", "lite"
        max_results: int | None = None,
        page: int = 1,
    ) -> list[dict[str, Any]]: ...
```

**Source:** [GitHub ddgs/ddgs.py](https://github.com/deedy5/ddgs/blob/main/ddgs/ddgs.py)

### Return Fields

Each result is a `dict[str, str]` with these keys:

| Field | Description | Example |
|-------|-------------|---------|
| `title` | Page title | `"News, sport and gossip \| The Sun"` |
| `href` | URL | `"https://www.thesun.co.uk/"` |
| `body` | Snippet text | `"Get the latest news..."` |

**Note:** There is **no `date` field** in text search results. The `news()` method does return a `date` field. If you need dates, you would need to either use `news()` or extract dates separately.

**Source:** [Context7 docs](https://github.com/deedy5/duckduckgo_search/blob/main/README.md)

---

## 4. API Key Requirement

**No API key required.** DDGS scrapes public search engine pages. It does not use any official API that requires authentication.

**Source:** [PyPI page](https://pypi.org/project/ddgs/)

---

## 5. Async Support

### Current State (ddgs v9.x)

The `ddgs` package does **NOT** have an `AsyncDDGS` class. The old `duckduckgo-search` package had `AsyncDDGS`, but it was removed in the rewrite.

The current `DDGS` class uses `ThreadPoolExecutor` for concurrent operations internally but exposes only synchronous methods.

### Wrapping in Async

Since the library is synchronous and I/O-bound, it can be wrapped with `asyncio.to_thread()` (Python 3.9+):

```python
import asyncio
from ddgs import DDGS

async def async_text_search(query: str, max_results: int = 5) -> list[dict]:
    return await asyncio.to_thread(
        DDGS().text, query, max_results=max_results
    )
```

Or with `loop.run_in_executor()` for more control:

```python
import asyncio
from concurrent.futures import ThreadPoolExecutor
from ddgs import DDGS

executor = ThreadPoolExecutor(max_workers=2)

async def async_text_search(query: str, max_results: int = 5) -> list[dict]:
    loop = asyncio.get_running_loop()
    return await loop.run_in_executor(
        executor, lambda: DDGS().text(query, max_results=max_results)
    )
```

**Recommendation:** `asyncio.to_thread()` is sufficient for alto's use case (infrequent, non-concurrent searches).

---

## 6. Exceptions

From `ddgs/exceptions.py` ([source](https://github.com/deedy5/ddgs/blob/main/ddgs/exceptions.py)):

```python
class DDGSException(Exception):
    """Base exception class for ddgs."""

class RatelimitException(DDGSException):
    """Raised for rate limit exceeded errors during API requests."""

class TimeoutException(DDGSException):
    """Raised for timeout errors during API requests."""
```

### Error Handling Strategy for WebSearchPort

```python
from ddgs import DDGS
from ddgs.exceptions import DDGSException, RatelimitException, TimeoutException

try:
    results = DDGS(timeout=10).text(query, max_results=max_results)
except TimeoutException:
    # Network timeout -- retry or return empty
    ...
except RatelimitException:
    # Rate limited -- back off, use proxy, or return cached results
    ...
except DDGSException:
    # Any other library error
    ...
except Exception:
    # Unexpected errors (network failures from httpx/primp)
    ...
```

**Note:** Network-level errors from `httpx` (e.g., `httpx.ConnectError`, `httpx.ReadTimeout`) may also propagate if not caught internally by DDGS. Catching `DDGSException` as a base plus a broad `Exception` fallback is safest.

---

## 7. Rate Limiting Concerns

DuckDuckGo does **not publish** official rate limits. Based on community reports:

| Observed Behavior | Source |
|--------------------|--------|
| Rate limited after ~6-7 requests in quick succession | [GitHub issue #213](https://github.com/deedy5/duckduckgo_search/issues/213) |
| Rate limited after ~50 searches at 1 per 3-5 seconds | [crewAI issue #136](https://github.com/crewAIInc/crewAI/issues/136) |
| 202 status code indicates rate limiting | [open-webui discussion #6624](https://github.com/open-webui/open-webui/discussions/6624) |

### Mitigations

1. **Backend rotation:** `backend="auto"` tries multiple backends in random order, distributing load.
2. **Proxy support:** `DDGS(proxy="socks5://...")` to rotate IPs.
3. **Backoff:** Implement exponential backoff on `RatelimitException`.
4. **Caching:** Cache results for identical queries (most important mitigation for alto's use case).
5. **Since ddgs is metasearch:** It aggregates across Bing, Google, Brave, DuckDuckGo -- spreading load across providers.

### Risk Assessment for Alto

**Low risk.** Alto's web search usage would be:
- Spike research (infrequent, human-initiated)
- Knowledge base updates (periodic, can be throttled)
- Not bulk scraping or real-time high-volume queries

A simple in-memory or file-based cache with 15-minute TTL plus exponential backoff on `RatelimitException` should be sufficient.

---

## 8. Installation

```bash
# Using uv (alto's package manager)
uv add ddgs

# Using pip
pip install ddgs

# With API server support (not needed for alto)
pip install ddgs[api]
```

**Package name for pyproject.toml:** `ddgs`

---

## 9. Adapter Design Sketch

Based on the findings, here is how a `WebSearchPort` adapter would look:

```python
# src/application/ports/web_search.py
from __future__ import annotations

from dataclasses import dataclass
from typing import Protocol


@dataclass(frozen=True)
class SearchResult:
    """Value object for a web search result."""
    url: str
    title: str
    snippet: str
    date: str | None = None  # Only available from news(), not text()


class WebSearchPort(Protocol):
    """Port for web search operations."""

    async def search(
        self, query: str, *, max_results: int = 5
    ) -> list[SearchResult]:
        """Search the web and return structured results."""
        ...
```

```python
# src/infrastructure/external/ddgs_web_search.py
from __future__ import annotations

import asyncio
import logging

from ddgs import DDGS
from ddgs.exceptions import DDGSException, RatelimitException, TimeoutException

from src.application.ports.web_search import SearchResult, WebSearchPort

logger = logging.getLogger(__name__)


class DdgsWebSearch:
    """WebSearchPort adapter using the ddgs metasearch library."""

    def __init__(self, timeout: int = 10, proxy: str | None = None) -> None:
        self._timeout = timeout
        self._proxy = proxy

    async def search(
        self, query: str, *, max_results: int = 5
    ) -> list[SearchResult]:
        try:
            raw = await asyncio.to_thread(
                self._sync_search, query, max_results
            )
            return [
                SearchResult(
                    url=r.get("href", ""),
                    title=r.get("title", ""),
                    snippet=r.get("body", ""),
                )
                for r in raw
            ]
        except TimeoutException:
            logger.warning("DDGS timeout for query: %s", query)
            return []
        except RatelimitException:
            logger.warning("DDGS rate limited for query: %s", query)
            return []
        except DDGSException as e:
            logger.error("DDGS error for query '%s': %s", query, e)
            return []

    def _sync_search(self, query: str, max_results: int) -> list[dict]:
        return DDGS(
            timeout=self._timeout,
            proxy=self._proxy,
        ).text(query, max_results=max_results)
```

---

## 10. Summary

| Question | Answer |
|----------|--------|
| Still maintained? | Yes -- `ddgs` v9.11.2 released March 5, 2026 |
| License? | MIT (permissive) |
| API key required? | No |
| Package name? | `ddgs` (not `duckduckgo-search`) |
| Async native? | No -- sync only, wrap with `asyncio.to_thread()` |
| Return fields? | `title`, `href`, `body` (no date in text search) |
| Exceptions? | `DDGSException`, `RatelimitException`, `TimeoutException` |
| Rate limits? | Undocumented; ~6-50 requests before throttling; mitigate with caching + backoff |
| Python version? | >= 3.10 (compatible with alto's 3.12+ requirement) |
| Dependencies? | primp, lxml, httpx, fake-useragent |

### Recommendation

**Use `ddgs` (v9.11.2+) behind a `WebSearchPort` protocol.** It is:
- Actively maintained (5 releases in 4 months)
- MIT licensed
- Zero API key required
- Adequate for alto's low-volume, human-initiated search use case

### Key Risk

Rate limiting is the primary risk. DuckDuckGo does not publish limits, and the library can be blocked after as few as 6-7 rapid requests. Mitigation: implement result caching and exponential backoff. For alto's expected usage pattern (infrequent spike research), this is low risk.

### Follow-Up Items

1. **Implementation ticket:** Create `WebSearchPort` protocol and `DdgsWebSearch` adapter
2. **Caching layer:** Add a TTL-based result cache (in-memory or file-based)
3. **Test strategy:** Unit tests with mocked DDGS; integration test with real search (marked slow)
