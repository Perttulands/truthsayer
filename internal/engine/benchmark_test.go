package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/rules"
)

// generateJSCorpus creates ~10k lines of realistic JS/TS code across multiple files.
func generateJSCorpus(dir string) error {
	// Each file ~200 lines, 50 files = ~10k lines
	for i := range 25 {
		ext := ".js"
		if i%5 == 0 {
			ext = ".ts"
		}
		name := fmt.Sprintf("module_%03d%s", i, ext)
		content := generateJSFile(i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			return err
		}
	}
	for i := 25; i < 50; i++ {
		ext := ".js"
		if i%5 == 0 {
			ext = ".ts"
		}
		name := fmt.Sprintf("service_%03d%s", i, ext)
		content := generateJSServiceFile(i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func generateJSFile(seed int) string {
	return fmt.Sprintf(`// Module %d - generated benchmark corpus
import { Logger } from './logger';
import { Database } from './database';
import { validate } from './utils';

const logger = new Logger('module_%d');

export class Service%d {
  constructor(private db: Database, private config: Config) {
    this.db = db;
    this.config = config;
  }

  async findById(id: string): Promise<Result> {
    try {
      const result = await this.db.query('SELECT * FROM items WHERE id = $1', [id]);
      if (!result.rows.length) {
        throw new Error('Item not found: ' + id);
      }
      logger.info('Found item', { id });
      return this.mapRow(result.rows[0]);
    } catch (err) {
      logger.error('Failed to find item', { id, error: err });
      throw new Error('Database query failed: ' + err.message);
    }
  }

  async create(data: CreateInput): Promise<Result> {
    validate(data);
    const normalized = this.normalize(data);
    try {
      const result = await this.db.query(
        'INSERT INTO items (name, value, created_at) VALUES ($1, $2, NOW()) RETURNING *',
        [normalized.name, normalized.value]
      );
      logger.info('Created item', { id: result.rows[0].id });
      return this.mapRow(result.rows[0]);
    } catch (err) {
      if (err.code === '23505') {
        throw new Error('Duplicate item: ' + normalized.name);
      }
      throw err;
    }
  }

  async update(id: string, data: UpdateInput): Promise<Result> {
    const existing = await this.findById(id);
    if (!existing) {
      throw new Error('Cannot update non-existent item: ' + id);
    }
    const merged = { ...existing, ...data, updatedAt: new Date() };
    try {
      await this.db.query(
        'UPDATE items SET name = $1, value = $2, updated_at = $3 WHERE id = $4',
        [merged.name, merged.value, merged.updatedAt, id]
      );
      logger.info('Updated item', { id });
      return merged;
    } catch (err) {
      logger.error('Update failed', { id, error: err });
      throw new Error('Update failed: ' + err.message);
    }
  }

  async delete(id: string): Promise<void> {
    const existing = await this.findById(id);
    if (!existing) {
      throw new Error('Cannot delete non-existent item: ' + id);
    }
    try {
      await this.db.query('DELETE FROM items WHERE id = $1', [id]);
      logger.info('Deleted item', { id });
    } catch (err) {
      logger.error('Delete failed', { id, error: err });
      throw new Error('Delete failed: ' + err.message);
    }
  }

  async list(options: ListOptions): Promise<Result[]> {
    const { limit = 100, offset = 0, orderBy = 'created_at' } = options;
    const validColumns = ['created_at', 'name', 'updated_at'];
    if (!validColumns.includes(orderBy)) {
      throw new Error('Invalid order column: ' + orderBy);
    }
    const result = await this.db.query(
      'SELECT * FROM items ORDER BY ' + orderBy + ' LIMIT $1 OFFSET $2',
      [limit, offset]
    );
    return result.rows.map(row => this.mapRow(row));
  }

  async search(query: string): Promise<Result[]> {
    if (!query || query.length < 3) {
      throw new Error('Search query too short');
    }
    const sanitized = query.replace(/[%%_]/g, '\\$&');
    const result = await this.db.query(
      'SELECT * FROM items WHERE name ILIKE $1 ORDER BY name LIMIT 50',
      ['%%' + sanitized + '%%']
    );
    return result.rows.map(row => this.mapRow(row));
  }

  private mapRow(row: DbRow): Result {
    return {
      id: row.id,
      name: row.name,
      value: row.value,
      createdAt: new Date(row.created_at),
      updatedAt: row.updated_at ? new Date(row.updated_at) : null,
    };
  }

  private normalize(data: CreateInput): NormalizedInput {
    return {
      name: data.name.trim().toLowerCase(),
      value: typeof data.value === 'string' ? data.value.trim() : data.value,
    };
  }

  async bulkCreate(items: CreateInput[]): Promise<Result[]> {
    const results = [];
    for (const item of items) {
      const result = await this.create(item);
      results.push(result);
    }
    return results;
  }

  async count(): Promise<number> {
    const result = await this.db.query('SELECT COUNT(*) as count FROM items');
    return parseInt(result.rows[0].count, 10);
  }

  async exists(id: string): Promise<boolean> {
    const result = await this.db.query(
      'SELECT 1 FROM items WHERE id = $1 LIMIT 1',
      [id]
    );
    return result.rows.length > 0;
  }

  async archive(id: string): Promise<void> {
    await this.db.query(
      'UPDATE items SET archived = true, archived_at = NOW() WHERE id = $1',
      [id]
    );
    logger.info('Archived item', { id });
  }

  async restore(id: string): Promise<void> {
    await this.db.query(
      'UPDATE items SET archived = false, archived_at = NULL WHERE id = $1',
      [id]
    );
    logger.info('Restored item', { id });
  }
}

export function createRouter(service: Service%d) {
  return {
    get: async (req, res) => {
      const result = await service.findById(req.params.id);
      res.json(result);
    },
    list: async (req, res) => {
      const results = await service.list(req.query);
      res.json(results);
    },
    create: async (req, res) => {
      const result = await service.create(req.body);
      res.status(201).json(result);
    },
    update: async (req, res) => {
      const result = await service.update(req.params.id, req.body);
      res.json(result);
    },
    delete: async (req, res) => {
      await service.delete(req.params.id);
      res.status(204).end();
    },
  };
}

export interface Config {
  dbUrl: string;
  maxRetries: number;
  timeout: number;
}

export interface CreateInput {
  name: string;
  value: unknown;
}

export interface UpdateInput {
  name?: string;
  value?: unknown;
}

export interface ListOptions {
  limit?: number;
  offset?: number;
  orderBy?: string;
}

export interface Result {
  id: string;
  name: string;
  value: unknown;
  createdAt: Date;
  updatedAt: Date | null;
}

interface NormalizedInput {
  name: string;
  value: unknown;
}

interface DbRow {
  id: string;
  name: string;
  value: unknown;
  created_at: string;
  updated_at: string | null;
}
`, seed, seed, seed, seed)
}

func generateJSServiceFile(seed int) string {
	return fmt.Sprintf(`// Service %d - generated benchmark corpus
const express = require('express');
const { EventEmitter } = require('events');

class EventProcessor%d extends EventEmitter {
  constructor(options) {
    super();
    this.queue = [];
    this.processing = false;
    this.maxRetries = options.maxRetries || 3;
    this.batchSize = options.batchSize || 10;
    this.timeout = options.timeout || 5000;
  }

  enqueue(event) {
    this.queue.push({
      ...event,
      enqueuedAt: Date.now(),
      retries: 0,
    });
    if (!this.processing) {
      this.process();
    }
  }

  async process() {
    this.processing = true;
    while (this.queue.length > 0) {
      const batch = this.queue.splice(0, this.batchSize);
      try {
        await Promise.all(batch.map(event => this.handleEvent(event)));
      } catch (err) {
        this.emit('error', err);
        for (const event of batch) {
          if (event.retries < this.maxRetries) {
            event.retries++;
            this.queue.push(event);
          } else {
            this.emit('dead-letter', event);
          }
        }
      }
    }
    this.processing = false;
  }

  async handleEvent(event) {
    const handler = this.handlers.get(event.type);
    if (!handler) {
      throw new Error('Unknown event type: ' + event.type);
    }
    const startTime = Date.now();
    try {
      const result = await handler(event.data);
      const duration = Date.now() - startTime;
      this.emit('processed', { event, result, duration });
      return result;
    } catch (err) {
      const duration = Date.now() - startTime;
      this.emit('failed', { event, error: err, duration });
      throw err;
    }
  }

  registerHandler(type, handler) {
    if (!this.handlers) {
      this.handlers = new Map();
    }
    this.handlers.set(type, handler);
  }

  getStats() {
    return {
      queueLength: this.queue.length,
      processing: this.processing,
      handlers: this.handlers ? this.handlers.size : 0,
    };
  }

  async drain() {
    while (this.queue.length > 0 || this.processing) {
      await new Promise(resolve => setTimeout(resolve, 100));
    }
  }

  clear() {
    this.queue = [];
    this.processing = false;
  }
}

class Cache%d {
  constructor(options) {
    this.store = new Map();
    this.ttl = options.ttl || 60000;
    this.maxSize = options.maxSize || 1000;
  }

  get(key) {
    const entry = this.store.get(key);
    if (!entry) return undefined;
    if (Date.now() > entry.expiresAt) {
      this.store.delete(key);
      return undefined;
    }
    entry.lastAccess = Date.now();
    entry.hits++;
    return entry.value;
  }

  set(key, value, ttl) {
    if (this.store.size >= this.maxSize) {
      this.evict();
    }
    this.store.set(key, {
      value,
      createdAt: Date.now(),
      expiresAt: Date.now() + (ttl || this.ttl),
      lastAccess: Date.now(),
      hits: 0,
    });
  }

  delete(key) {
    return this.store.delete(key);
  }

  has(key) {
    const entry = this.store.get(key);
    if (!entry) return false;
    if (Date.now() > entry.expiresAt) {
      this.store.delete(key);
      return false;
    }
    return true;
  }

  evict() {
    let oldest = null;
    let oldestKey = null;
    for (const [key, entry] of this.store) {
      if (Date.now() > entry.expiresAt) {
        this.store.delete(key);
        continue;
      }
      if (!oldest || entry.lastAccess < oldest.lastAccess) {
        oldest = entry;
        oldestKey = key;
      }
    }
    if (oldestKey) {
      this.store.delete(oldestKey);
    }
  }

  clear() {
    this.store.clear();
  }

  getStats() {
    let totalHits = 0;
    let expired = 0;
    for (const entry of this.store.values()) {
      totalHits += entry.hits;
      if (Date.now() > entry.expiresAt) expired++;
    }
    return { size: this.store.size, totalHits, expired };
  }
}

function createMiddleware%d(cache, processor) {
  return function(req, res, next) {
    const start = Date.now();
    const originalEnd = res.end;
    res.end = function(...args) {
      const duration = Date.now() - start;
      processor.enqueue({
        type: 'request',
        data: {
          method: req.method,
          path: req.path,
          status: res.statusCode,
          duration,
        },
      });
      originalEnd.apply(res, args);
    };
    next();
  };
}

function validateConfig%d(config) {
  const required = ['port', 'dbUrl', 'secret'];
  const missing = required.filter(key => !config[key]);
  if (missing.length > 0) {
    throw new Error('Missing config: ' + missing.join(', '));
  }
  if (config.port < 1 || config.port > 65535) {
    throw new Error('Invalid port: ' + config.port);
  }
  if (config.secret.length < 32) {
    throw new Error('Secret too short');
  }
  return { ...config, validated: true };
}

module.exports = { EventProcessor%d, Cache%d, createMiddleware%d, validateConfig%d };
`, seed, seed, seed, seed, seed, seed, seed, seed, seed)
}

// generatePyCorpus creates ~10k lines of realistic Python code across multiple files.
func generatePyCorpus(dir string) error {
	for i := range 25 {
		name := fmt.Sprintf("module_%03d.py", i)
		content := generatePyFile(i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			return err
		}
	}
	for i := 25; i < 50; i++ {
		name := fmt.Sprintf("service_%03d.py", i)
		content := generatePyServiceFile(i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func generatePyFile(seed int) string {
	return fmt.Sprintf(`"""Module %d - generated benchmark corpus."""
import logging
from dataclasses import dataclass, field
from typing import Optional, List, Dict, Any
from datetime import datetime, timedelta

logger = logging.getLogger(__name__)


@dataclass
class Config%d:
    host: str
    port: int
    database: str
    timeout: int = 30
    max_connections: int = 10
    retry_count: int = 3
    debug: bool = False


@dataclass
class Item%d:
    id: str
    name: str
    value: Any
    created_at: datetime = field(default_factory=datetime.utcnow)
    updated_at: Optional[datetime] = None
    tags: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)

    def is_stale(self, max_age: timedelta) -> bool:
        if self.updated_at:
            return datetime.utcnow() - self.updated_at > max_age
        return datetime.utcnow() - self.created_at > max_age

    def add_tag(self, tag: str) -> None:
        if tag not in self.tags:
            self.tags.append(tag)

    def remove_tag(self, tag: str) -> None:
        if tag in self.tags:
            self.tags.remove(tag)


class Repository%d:
    def __init__(self, db_connection, config: Config%d):
        self.db = db_connection
        self.config = config
        self._cache: Dict[str, Item%d] = {}
        logger.info("Repository initialized with config: %%s", config)

    def find_by_id(self, item_id: str) -> Optional[Item%d]:
        if item_id in self._cache:
            logger.debug("Cache hit for item %%s", item_id)
            return self._cache[item_id]

        try:
            row = self.db.execute(
                "SELECT * FROM items WHERE id = %%s", (item_id,)
            ).fetchone()
        except Exception as exc:
            logger.error("Database error looking up %%s: %%s", item_id, exc)
            raise RuntimeError(f"Failed to find item {item_id}") from exc

        if row is None:
            return None

        item = self._map_row(row)
        self._cache[item_id] = item
        return item

    def create(self, name: str, value: Any) -> Item%d:
        try:
            result = self.db.execute(
                "INSERT INTO items (name, value) VALUES (%%s, %%s) RETURNING *",
                (name, value),
            ).fetchone()
        except Exception as exc:
            logger.error("Failed to create item: %%s", exc)
            raise RuntimeError(f"Create failed for {name}") from exc

        item = self._map_row(result)
        self._cache[item.id] = item
        logger.info("Created item %%s", item.id)
        return item

    def update(self, item_id: str, **kwargs) -> Item%d:
        existing = self.find_by_id(item_id)
        if existing is None:
            raise ValueError(f"Item {item_id} not found")

        updates = {k: v for k, v in kwargs.items() if v is not None}
        if not updates:
            return existing

        set_clause = ", ".join(f"{k} = %%s" for k in updates)
        values = list(updates.values()) + [item_id]
        try:
            result = self.db.execute(
                f"UPDATE items SET {set_clause}, updated_at = NOW() WHERE id = %%s RETURNING *",
                values,
            ).fetchone()
        except Exception as exc:
            logger.error("Update failed for %%s: %%s", item_id, exc)
            raise RuntimeError(f"Update failed for {item_id}") from exc

        item = self._map_row(result)
        self._cache[item_id] = item
        return item

    def delete(self, item_id: str) -> None:
        try:
            self.db.execute("DELETE FROM items WHERE id = %%s", (item_id,))
        except Exception as exc:
            logger.error("Delete failed for %%s: %%s", item_id, exc)
            raise RuntimeError(f"Delete failed for {item_id}") from exc

        self._cache.pop(item_id, None)
        logger.info("Deleted item %%s", item_id)

    def list_items(
        self, limit: int = 100, offset: int = 0, order_by: str = "created_at"
    ) -> List[Item%d]:
        valid_columns = {"created_at", "name", "updated_at"}
        if order_by not in valid_columns:
            raise ValueError(f"Invalid order column: {order_by}")

        rows = self.db.execute(
            f"SELECT * FROM items ORDER BY {order_by} LIMIT %%s OFFSET %%s",
            (limit, offset),
        ).fetchall()
        return [self._map_row(row) for row in rows]

    def search(self, query: str) -> List[Item%d]:
        if len(query) < 3:
            raise ValueError("Search query too short")

        rows = self.db.execute(
            "SELECT * FROM items WHERE name ILIKE %%s ORDER BY name LIMIT 50",
            (f"%%%%{query}%%%%",),
        ).fetchall()
        return [self._map_row(row) for row in rows]

    def count(self) -> int:
        result = self.db.execute("SELECT COUNT(*) FROM items").fetchone()
        return result[0]

    def exists(self, item_id: str) -> bool:
        if item_id in self._cache:
            return True
        result = self.db.execute(
            "SELECT 1 FROM items WHERE id = %%s LIMIT 1", (item_id,)
        ).fetchone()
        return result is not None

    def clear_cache(self) -> None:
        self._cache.clear()
        logger.debug("Cache cleared")

    def _map_row(self, row) -> Item%d:
        return Item%d(
            id=row["id"],
            name=row["name"],
            value=row["value"],
            created_at=row["created_at"],
            updated_at=row.get("updated_at"),
        )

    def bulk_create(self, items: List[Dict[str, Any]]) -> List[Item%d]:
        results = []
        for item_data in items:
            result = self.create(item_data["name"], item_data["value"])
            results.append(result)
        return results

    def archive(self, item_id: str) -> None:
        self.db.execute(
            "UPDATE items SET archived = TRUE, archived_at = NOW() WHERE id = %%s",
            (item_id,),
        )
        self._cache.pop(item_id, None)
        logger.info("Archived item %%s", item_id)

    def restore(self, item_id: str) -> None:
        self.db.execute(
            "UPDATE items SET archived = FALSE, archived_at = NULL WHERE id = %%s",
            (item_id,),
        )
        logger.info("Restored item %%s", item_id)
`, seed, seed, seed, seed, seed, seed, seed, seed, seed, seed, seed, seed, seed, seed)
}

func generatePyServiceFile(seed int) string {
	return fmt.Sprintf(`"""Service %d - generated benchmark corpus."""
import logging
from typing import Any, Dict, List, Optional, Callable
from functools import wraps
from datetime import datetime

logger = logging.getLogger(__name__)


def retry(max_attempts: int = 3, delay: float = 1.0):
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        def wrapper(*args, **kwargs):
            last_error = None
            for attempt in range(max_attempts):
                try:
                    return func(*args, **kwargs)
                except Exception as exc:
                    last_error = exc
                    logger.warning(
                        "Attempt %%d/%%d failed for %%s: %%s",
                        attempt + 1,
                        max_attempts,
                        func.__name__,
                        exc,
                    )
            raise RuntimeError(
                f"All {max_attempts} attempts failed for {func.__name__}"
            ) from last_error
        return wrapper
    return decorator


class EventBus%d:
    def __init__(self):
        self._handlers: Dict[str, List[Callable]] = {}
        self._history: List[Dict[str, Any]] = []
        self._max_history = 1000

    def subscribe(self, event_type: str, handler: Callable) -> None:
        if event_type not in self._handlers:
            self._handlers[event_type] = []
        self._handlers[event_type].append(handler)
        logger.debug("Subscribed to %%s", event_type)

    def unsubscribe(self, event_type: str, handler: Callable) -> None:
        if event_type in self._handlers:
            self._handlers[event_type] = [
                h for h in self._handlers[event_type] if h != handler
            ]

    def publish(self, event_type: str, data: Any = None) -> None:
        record = {
            "type": event_type,
            "data": data,
            "timestamp": datetime.utcnow().isoformat(),
        }
        self._history.append(record)
        if len(self._history) > self._max_history:
            self._history = self._history[-self._max_history:]

        handlers = self._handlers.get(event_type, [])
        for handler in handlers:
            try:
                handler(data)
            except Exception as exc:
                logger.error(
                    "Handler %%s failed for %%s: %%s",
                    handler.__name__,
                    event_type,
                    exc,
                )

    def get_history(
        self, event_type: Optional[str] = None, limit: int = 100
    ) -> List[Dict[str, Any]]:
        if event_type:
            filtered = [e for e in self._history if e["type"] == event_type]
            return filtered[-limit:]
        return self._history[-limit:]

    def clear_history(self) -> None:
        self._history.clear()

    def handler_count(self, event_type: str) -> int:
        return len(self._handlers.get(event_type, []))


class Pipeline%d:
    def __init__(self):
        self._stages: List[Callable] = []
        self._error_handler: Optional[Callable] = None

    def add_stage(self, stage: Callable) -> "Pipeline%d":
        self._stages.append(stage)
        return self

    def on_error(self, handler: Callable) -> "Pipeline%d":
        self._error_handler = handler
        return self

    def execute(self, data: Any) -> Any:
        result = data
        for i, stage in enumerate(self._stages):
            try:
                result = stage(result)
                if result is None:
                    logger.warning("Stage %%d returned None, stopping pipeline", i)
                    break
            except Exception as exc:
                logger.error("Pipeline stage %%d failed: %%s", i, exc)
                if self._error_handler:
                    return self._error_handler(exc, data, i)
                raise RuntimeError(
                    f"Pipeline failed at stage {i}"
                ) from exc
        return result

    def execute_batch(self, items: List[Any]) -> List[Any]:
        results = []
        for item in items:
            try:
                result = self.execute(item)
                results.append(result)
            except Exception as exc:
                logger.error("Batch item failed: %%s", exc)
                results.append(None)
        return results


class Validator%d:
    def __init__(self):
        self._rules: Dict[str, List[Callable]] = {}

    def add_rule(self, field: str, rule: Callable, message: str = "") -> None:
        if field not in self._rules:
            self._rules[field] = []
        self._rules[field].append((rule, message))

    def validate(self, data: Dict[str, Any]) -> List[str]:
        errors = []
        for field_name, rules in self._rules.items():
            value = data.get(field_name)
            for rule_func, message in rules:
                try:
                    if not rule_func(value):
                        errors.append(
                            message or f"Validation failed for {field_name}"
                        )
                except Exception as exc:
                    errors.append(f"Validation error for {field_name}: {exc}")
        return errors

    def is_valid(self, data: Dict[str, Any]) -> bool:
        return len(self.validate(data)) == 0


def parse_config%d(raw: Dict[str, Any]) -> Dict[str, Any]:
    config = {}
    for key, value in raw.items():
        if isinstance(value, str) and value.startswith("$"):
            import os
            env_value = os.environ.get(value[1:])
            if env_value is None:
                raise ValueError(f"Missing env var: {value}")
            config[key] = env_value
        else:
            config[key] = value
    return config


def setup_logging%d(level: str = "INFO", fmt: str = None) -> None:
    log_format = fmt or "%%(asctime)s - %%(name)s - %%(levelname)s - %%(message)s"
    logging.basicConfig(level=getattr(logging, level), format=log_format)
    logger.info("Logging configured at %%s level", level)


def batch_process%d(
    items: List[Any],
    processor: Callable,
    batch_size: int = 100,
    on_error: Optional[Callable] = None,
) -> List[Any]:
    results = []
    for i in range(0, len(items), batch_size):
        batch = items[i : i + batch_size]
        try:
            batch_results = [processor(item) for item in batch]
            results.extend(batch_results)
            logger.debug("Processed batch %%d-%%d", i, i + len(batch))
        except Exception as exc:
            logger.error("Batch %%d-%%d failed: %%s", i, i + len(batch), exc)
            if on_error:
                on_error(exc, batch)
            else:
                raise
    return results
`, seed, seed, seed, seed, seed, seed, seed, seed, seed)
}

func BenchmarkJSScan10k(b *testing.B) {
	dir := b.TempDir()
	if err := generateJSCorpus(dir); err != nil {
		b.Fatalf("generate JS corpus: %v", err)
	}

	reg := rules.DefaultRegistry()
	eng := New(reg)

	// Warm up: run once to initialize parsers
	if _, err := eng.Scan(dir); err != nil {
		b.Fatalf("warmup scan failed: %v", err)
	}

	b.ResetTimer()
	for range b.N {
		result, err := eng.Scan(dir)
		if err != nil {
			b.Fatalf("scan failed: %v", err)
		}
		if result.FilesScanned == 0 {
			b.Fatal("no files scanned")
		}
	}
}

func BenchmarkPythonScan10k(b *testing.B) {
	dir := b.TempDir()
	if err := generatePyCorpus(dir); err != nil {
		b.Fatalf("generate Python corpus: %v", err)
	}

	reg := rules.DefaultRegistry()
	eng := New(reg)

	// Warm up: run once to initialize parsers
	if _, err := eng.Scan(dir); err != nil {
		b.Fatalf("warmup scan failed: %v", err)
	}

	b.ResetTimer()
	for range b.N {
		result, err := eng.Scan(dir)
		if err != nil {
			b.Fatalf("scan failed: %v", err)
		}
		if result.FilesScanned == 0 {
			b.Fatal("no files scanned")
		}
	}
}

// TestBenchmarkPerformanceTarget verifies <5s per 10k LOC target.
func TestBenchmarkPerformanceTarget(t *testing.T) {
	reg := rules.DefaultRegistry()

	t.Run("JS_10k_LOC_under_5s", func(t *testing.T) {
		dir := t.TempDir()
		if err := generateJSCorpus(dir); err != nil {
			t.Fatalf("generate JS corpus: %v", err)
		}
		eng := New(reg)
		start := time.Now()
		result, err := eng.Scan(dir)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		t.Logf("JS scan: %d files, %d findings in %v", result.FilesScanned, len(result.Findings), elapsed)
		if elapsed > 5*time.Second {
			t.Errorf("JS scan took %v, exceeds 5s target", elapsed)
		}
	})

	t.Run("Python_10k_LOC_under_5s", func(t *testing.T) {
		dir := t.TempDir()
		if err := generatePyCorpus(dir); err != nil {
			t.Fatalf("generate Python corpus: %v", err)
		}
		eng := New(reg)
		start := time.Now()
		result, err := eng.Scan(dir)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		t.Logf("Python scan: %d files, %d findings in %v", result.FilesScanned, len(result.Findings), elapsed)
		if elapsed > 5*time.Second {
			t.Errorf("Python scan took %v, exceeds 5s target", elapsed)
		}
	})
}
