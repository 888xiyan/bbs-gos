from __future__ import annotations

import logging


def configure_logging(level: int = logging.INFO) -> logging.Logger:
    logger = logging.getLogger("bbs_go_mcp")
    logger.setLevel(level)
    if not logger.handlers:
        handler = logging.StreamHandler()
        handler.setFormatter(logging.Formatter("%(asctime)s %(levelname)s %(message)s"))
        logger.addHandler(handler)
    return logger
