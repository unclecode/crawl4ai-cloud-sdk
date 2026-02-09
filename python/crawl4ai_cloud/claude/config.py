"""Plugin configuration for Crawl4AI Claude Code plugin."""
import json
import os
from dataclasses import dataclass, field, asdict
from pathlib import Path
from typing import Literal, Optional


CONFIG_DIR = Path.home() / ".crawl4ai"
CONFIG_FILE = CONFIG_DIR / "claude_config.json"


@dataclass
class PluginConfig:
    """Configuration for the Crawl4AI Claude Code plugin."""
    mode: Literal["cloud", "local"] = "cloud"
    api_key: Optional[str] = None
    api_base_url: str = "https://api.crawl4ai.com"
    default_timeout: float = 120.0
    headless: bool = True
    browser_type: str = "chromium"
    verbose: bool = False


def load_config() -> PluginConfig:
    """Load plugin config from ~/.crawl4ai/claude_config.json with env var overrides."""
    config = PluginConfig()

    if CONFIG_FILE.exists():
        try:
            data = json.loads(CONFIG_FILE.read_text())
            for k, v in data.items():
                if hasattr(config, k):
                    setattr(config, k, v)
        except (json.JSONDecodeError, OSError):
            pass

    # Env var overrides
    env_key = os.environ.get("CRAWL4AI_API_KEY")
    if env_key:
        config.api_key = env_key

    env_mode = os.environ.get("CRAWL4AI_MODE")
    if env_mode in ("cloud", "local"):
        config.mode = env_mode

    env_base = os.environ.get("CRAWL4AI_API_BASE_URL")
    if env_base:
        config.api_base_url = env_base

    return config


def save_config(config: PluginConfig) -> None:
    """Save plugin config to ~/.crawl4ai/claude_config.json."""
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    data = asdict(config)
    # Don't persist api_key if it came from env
    if os.environ.get("CRAWL4AI_API_KEY") and data.get("api_key") == os.environ["CRAWL4AI_API_KEY"]:
        data.pop("api_key", None)
    CONFIG_FILE.write_text(json.dumps(data, indent=2) + "\n")
