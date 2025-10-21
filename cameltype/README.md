# cameltype

Terminal-based typing test written in OCaml.

## Status

Currently displays a hello world message. Full typing test implementation coming soon.

## Building

From the project root:

```bash
make build
```

Or build just cameltype:

```bash
cd cameltype
opam exec -- dune build
```

## Running

```bash
./bin/cameltype
```

Or with the zootype wrapper:

```bash
zootype -b cameltype
```

## Dependencies

- OCaml (>= 4.08)
- dune (>= 3.0)

Install via opam:

```bash
opam install dune
```
