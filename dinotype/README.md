# dinotype

Terminal-based typing test written in TypeScript (using Deno).

## Status

Currently displays a hello world message. Full typing test implementation coming soon.

## Building

From the project root:

```bash
make build
```

Or build just dinotype:

```bash
cd dinotype
deno compile --output dinotype main.ts
```

## Running

```bash
./bin/dinotype
```

Or with the zootype wrapper:

```bash
zootype -b dinotype
```

## Dependencies

- [Deno](https://deno.land/) runtime
