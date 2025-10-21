# Language Comparison

## Binary Sizes

```
$ ls -lh bin/
97B   pythontype   (Python wrapper)
221B  eggtype      (Shell script)
52K   rattype      (C++)
415K  crabtype     (Rust)
482K  cameltype    (OCaml)
2.5M  gophertype   (Go)
69M   dinotype     (Deno/TypeScript)
```

## Compilation & Runtime

| Language   | Size | Compilation          | Runtime                            |
| ---------- | ---- | -------------------- | ---------------------------------- |
| Python     | 97B  | Interpreted bytecode | External (requires Python 3.10+)   |
| Shell      | 221B | Interpreted script   | External (POSIX shell)             |
| C++        | 52K  | AOT → native code    | Minimal (stack unwinding)          |
| Rust       | 415K | AOT → native code    | Minimal (panic handler, no GC)     |
| OCaml      | 482K | AOT → native code    | Bundled (GC + runtime system)      |
| Go         | 2.5M | AOT → native code    | Bundled (GC + goroutine scheduler) |
| TypeScript | 69M  | JIT (V8)             | Bundled (entire V8 + Deno runtime) |

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

- **Shell**: Universal availability, scripting simplicity
- **Python**: Rapid development, dynamic typing
- **C++**: Maximum control, manual memory management
- **Rust**: Memory safety without GC via ownership system
- **OCaml**: Functional programming, GC for safety
- **Go**: Built for concurrency, goroutines and channels
- **TypeScript**: Familiar syntax, massive ecosystem, async/await
