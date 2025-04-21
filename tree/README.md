# tree

Print directory and file (if `-f` option is specified) tree at a given path.

## Usage

```bash
go run tree.go <path string> [-f]
```

>[!IMPORTANT]
> `path` must be a directory.

## Examples

1. Include directories and files:

    ```ignorelang
    go run main.go . -f
    ├───main.go (1881b)
    ├───main_test.go (1318b)
    └───testdata
        ├───project
        │    ├───file.txt (19b)
        │    └───gopher.png (70372b)
        ├───static
        │    ├───css
        │    │    └───body.css (28b)
        │    ├───html
        │    │    └───index.html (57b)
        │    └───js
        │        └───site.js (10b)
        ├───zline
        │    └───empty.txt (empty)
        └───zzfile.txt (empty)
    ```

2. Include only directories:

    ```ignorelang
    go run main.go .
    └───testdata
        ├───project
        ├───static
        │    ├───css
        │    ├───html
        │    └───js
        └───zline
    ```
