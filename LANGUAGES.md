# Language Comparison

## Binary Sizes

```
$ ls -lh bin/
35K   rattype      (C++)
397K  crabtype     (Rust)
464K  cameltype    (OCaml)
2.5M  gophertype   (Go)
69M   dinotype     (Deno/TypeScript)
97B   pythontype   (Python wrapper)
```

## Compilation & Runtime

| Language   | Size | Compilation          | Runtime                            |
| ---------- | ---- | -------------------- | ---------------------------------- |
| C++        | 35K  | AOT → native code    | Minimal (stack unwinding)          |
| Rust       | 397K | AOT → native code    | Minimal (panic handler, no GC)     |
| OCaml      | 464K | AOT → native code    | Bundled (GC + runtime system)      |
| Go         | 2.5M | AOT → native code    | Bundled (GC + goroutine scheduler) |
| TypeScript | 69M  | JIT (V8)             | Bundled (entire V8 + Deno runtime) |
| Python     | 97B  | Interpreted bytecode | External (requires Python 3.10+)   |

## Dynamic Dependencies & Syscalls

### MacOS

Everything links to libSystem under the hood.

```bash
$ otool -L bin/rattype
/usr/lib/libc++.1.dylib
/usr/lib/libSystem.B.dylib

$ otool -L bin/crabtype
/usr/lib/libSystem.B.dylib

$ otool -L bin/gophertype
/usr/lib/libSystem.B.dylib
/usr/lib/libresolv.9.dylib
```

### Linux

On Linux, most languages use libc but there are ways around it:

- **C++/Rust/OCaml**: Link to libc (glibc or musl)
- **Go**: No libc dependency, makes syscalls directly. Enables truly static binaries.
- **Deno**: Self-contained (V8 embedded)

### Why is Deno 69MB?

`deno compile` bundles the entire Deno runtime into a standalone executable:

- V8 JavaScript engine (~30MB)
- TypeScript compiler
- Tokio async runtime
- Standard library

Trade-off: massive binary size for zero external dependencies. The binary runs on any machine without requiring Deno or Node.js installed.

## Why Each Language

- **C++**: Maximum control, manual memory management
- **Rust**: Memory safety without GC via ownership system
- **OCaml**: Functional programming, GC for safety
- **Go**: Built for concurrency, goroutines and channels
- **TypeScript**: Familiar syntax, massive ecosystem, async/await
- **Python**: Rapid development, dynamic typing
