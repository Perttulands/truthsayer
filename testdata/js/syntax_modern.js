// Modern JavaScript syntax for tree-sitter parse validation

// Optional chaining
const name = user?.profile?.name;
const length = arr?.length;

// Nullish coalescing
const val = input ?? "default";
const nested = a?.b ?? c?.d ?? "fallback";

// Class fields and static fields
class Counter {
  count = 0;
  static instances = 0;
  #privateField = "secret";

  increment() {
    this.count++;
    Counter.instances++;
  }

  get value() {
    return this.#privateField;
  }
}

// Logical assignment operators
let x = 1;
x &&= 2;
x ||= 3;
x ??= 4;

// Array/Object destructuring with defaults
const { a = 1, b: renamed = 2, ...rest } = obj;
const [first, , third, ...remaining] = arr;

// Async/await with for-await
async function fetchAll(urls) {
  for await (const response of urls.map(fetch)) {
    console.log(response.status);
  }
}

// Dynamic import
async function loadModule() {
  const mod = await import("./module.js");
  return mod.default;
}

// Template literals with nesting
const msg = `Hello ${user?.name ?? "stranger"}, you have ${count} items`;

// Spread in various positions
const merged = { ...defaults, ...overrides };
const combined = [...arr1, ...arr2, newItem];

// Computed property names
const key = "dynamic";
const obj2 = { [key]: true, [`${key}_2`]: false };

// Generator function
function* range(start, end) {
  for (let i = start; i < end; i++) {
    yield i;
  }
}

// Promise.allSettled
const results = await Promise.allSettled([
  fetch("/api/a"),
  fetch("/api/b"),
]);

// Numeric separators
const billion = 1_000_000_000;
const hex = 0xFF_FF_FF;

// BigInt
const big = 9007199254740991n;
