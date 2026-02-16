// Modern TypeScript syntax for tree-sitter parse validation

// Generics
function identity<T>(arg: T): T {
  return arg;
}

// Generic interfaces
interface Container<T> {
  value: T;
  map<U>(fn: (v: T) => U): Container<U>;
}

// Type assertions
const canvas = document.getElementById("main") as HTMLCanvasElement;

// Enums
enum Direction {
  Up = "UP",
  Down = "DOWN",
  Left = "LEFT",
  Right = "RIGHT",
}

// Const assertions
const config = {
  endpoint: "/api",
  timeout: 3000,
} as const;

// Decorators
function log(target: any, key: string, descriptor: PropertyDescriptor) {
  return descriptor;
}

class Service {
  @log
  fetchData(): Promise<void> {
    return Promise.resolve();
  }
}

// Satisfies operator
type Color = "red" | "green" | "blue";
const palette = {
  primary: "red",
  secondary: "blue",
} satisfies Record<string, Color>;

// Conditional types
type IsString<T> = T extends string ? "yes" : "no";

// Mapped types
type Readonly2<T> = {
  readonly [P in keyof T]: T[P];
};

// Template literal types
type EventName = `on${Capitalize<string>}`;

// Intersection and union types
type Result<T> = { success: true; data: T } | { success: false; error: Error };
type WithTimestamp<T> = T & { createdAt: Date; updatedAt: Date };

// Utility types usage
type PartialUser = Partial<{ name: string; age: number }>;
type RequiredConfig = Required<{ host?: string; port?: number }>;

// Index signatures
interface StringMap {
  [key: string]: string;
}

// Tuple types
type Pair = [string, number];
type Rest = [first: string, ...rest: number[]];

// Async generators
async function* asyncRange(start: number, end: number): AsyncGenerator<number> {
  for (let i = start; i < end; i++) {
    await new Promise((r) => setTimeout(r, 100));
    yield i;
  }
}

// Non-null assertion (for parse validation, not a recommendation)
const el = document.querySelector(".item")!;

// Type guards
function isString(value: unknown): value is string {
  return typeof value === "string";
}

// Infer keyword
type ReturnType2<T> = T extends (...args: any[]) => infer R ? R : never;
