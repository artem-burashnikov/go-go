# signer

Calculates the hash from a given input passing it through the asynchronous pipeline of hash functions.

## Usage

```bash
go run signer.go <salt> [integer1] ... [integerN]
```

## Examples

Salt is an empty string.

```bash
go run signer "" 0 1
```

Outputs:

```bash
29568666068035183841425683795340791879727309630931025356555_29568666068035183841425683795340791879727309630931025356555
```
