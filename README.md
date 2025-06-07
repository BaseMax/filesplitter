# FileSplitter

📁 **FileSplitter v1.0** by Max Base  
📦 Split massive files by lines, size, or pattern with style!

---

## Overview

**FileSplitter** is a fast and easy-to-use CLI tool written in Go for splitting large text files into smaller parts based on:

- Number of lines per file
- Maximum size per file (e.g., 100MB, 500KB)
- Matching a regex pattern (split whenever the pattern matches)

It supports customizable output filename prefixes, extensions, zero-padded indices, optional timestamps, dry run mode, and quiet mode for logging control.

---

## Features

- Split by line count or file size
- Regex pattern-based splitting
- Custom output directory, prefix, and file extension
- Zero-padded part indices with configurable width
- Optional timestamp appended to output filenames
- Dry run mode to preview file splits without writing files
- Colorful console logging for better UX
- Handles very large files efficiently with buffered I/O

---

## Installation

Download the latest executable from the [Releases](https://github.com/BaseMax/filesplitter/releases) page or build from source:

```bash
git clone https://github.com/BaseMax/filesplitter.git
cd filesplitter
go build -o filesplitter
````

---

## Usage

```bash
filesplitter -in <input-file> [options]
```

### Required

* `-in` : Input file path (e.g., `usernames.txt`)

### Optional

* `-lines` : Split by number of lines per file (e.g., 1000000)
* `-size` : Split by max size per file (e.g., `100MB`, `500KB`)
* `-pattern` : Regex pattern to split whenever matched
* `-prefix` : Output filename prefix (default: `part`)
* `-outdir` : Output directory (default: current directory)
* `-ext` : Output file extension (default: `txt`)
* `-pad` : Zero padding width for file indices (default: 3)
* `-ts` : Append timestamp to output filenames (default: false)
* `-dry` : Dry run mode, preview split without writing files
* `-q` : Quiet mode, suppress logs

### Example

Split a large file by 1 million lines per output part:

```bash
filesplitter -in largefile.txt -lines 1000000
```

Split a file by 100MB chunks, adding timestamps to filenames:

```bash
filesplitter -in largefile.txt -size 100MB -ts
```

Split a file whenever a pattern matches:

```bash
filesplitter -in log.txt -pattern "^ERROR"
```

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Feel free to contribute, open issues, or request features!

---

## Author

Max Base © 2025

[https://github.com/BaseMax](https://github.com/BaseMax)

**Happy splitting!** 🎉
