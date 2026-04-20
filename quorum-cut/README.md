# mycut

`mycut` is a distributed version of `cut` with quorum-based completion.

## Modes

- `local`: process files or stdin locally
- `node`: run a worker node over HTTP
- `search`: split input into chunks and dispatch them to nodes

## Quorum semantics

- `--consistency all`: wait for every chunk and return a complete result
- `--consistency quorum`: stop after `N/2 + 1` successful chunk responses

`all` is the compatibility mode for comparison with the original `cut`.
`quorum` is the mode that satisfies the distributed completion requirement from the assignment model, but it may return only a partial result if some chunks were not processed.

## Examples

Run three nodes:

```bash
go run . node --listen :9001 --id n1
go run . node --listen :9002 --id n2
go run . node --listen :9003 --id n3
```

Local field extraction:

```bash
printf 'a:b:c\nx:y:z\n' | go run . local -f 1,3 -d :
```

Distributed search with quorum:

```bash
printf 'a:b:c\nx:y:z\n' | go run . search --nodes localhost:9001,localhost:9002,localhost:9003 -f 2 -d :
```

Wait for all chunks:

```bash
go run . search --nodes localhost:9001,localhost:9002,localhost:9003 --consistency all -c 1-4 sample.txt
```

## Supported `cut` flags

- `-b`
- `-c`
- `-f`
- `-d`
- `-s`
- `--output-delimiter`

## Distributed flags

- `--nodes`
- `--consistency quorum|all`
- `--timeout`
- `--chunk-size`
- `node --listen`
- `node --id`

## Comparison with original `cut`

Functional comparison:

```bash
cut -d : -f 1,3 sample.txt
go run . local -d : -f 1,3 sample.txt
go run . search --nodes localhost:9001,localhost:9002,localhost:9003 --consistency all -d : -f 1,3 sample.txt
```

Timing comparison:

```bash
time cut -d : -f 1,3 big.txt
time go run . local -d : -f 1,3 big.txt
time go run . search --nodes localhost:9001,localhost:9002,localhost:9003 --consistency all -d : -f 1,3 big.txt
```

## Limitations

- this project implements the core `cut` flags listed above, not the full GNU `cut` surface
- compatibility with the original `cut` is guaranteed only for the supported flags
- `search --consistency quorum` is intentionally not byte-for-byte equivalent to original `cut`, because it may stop after majority completion
