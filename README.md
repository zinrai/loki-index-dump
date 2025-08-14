# loki-index-dump

A tool to dump Loki labels and their values to a JSON file.

## What it does

`loki-index-dump` retrieves all available labels and their corresponding values from a Loki instance and saves them as a JSON file. This is useful for:

- Understanding what labels and values exist in your Loki instance
- Creating offline references of your log metadata
- Generating cache files for other tools (like [loqui](https://github.com/zinrai/loqui))
- Debugging label cardinality issues

## Installation

```bash
$ go install github.com/zinrai/loki-index-dump@latest
```

## Prerequisites

- [logcli](https://grafana.com/docs/loki/latest/query/logcli/)
- `LOKI_ADDR` environment variable set to your Loki instance

## Usage

Set Loki server address (required by logcli):

```bash
$ export LOKI_ADDR=http://localhost:3100
```

Dump labels from last 30 days to metadata.json:

```bash
$ loki-index-dump
```

Dump labels from last 7 days:

```bash
$ loki-index-dump -days 7
```

Specify output file:

```bash
$ loki-index-dump -output /tmp/loki-labels.json
```

Combine options:

```bash
$ loki-index-dump -days 7 -output weekly-index.json
```

## Options

- `-days` - Number of days to look back (default: 30)
- `-output` - Output file path (default: metadata.json)
- `-help` - Show help message
- `-version` - Show version

## Output Format

The tool generates a JSON file with the following structure:

```json
{
  "labels": [
    "app",
    "env",
    "namespace"
  ],
  "values": {
    "app": ["nginx", "apache", "tomcat"],
    "env": ["production", "staging", "development"],
    "namespace": ["default", "kube-system", "monitoring"]
  },
  "dumped_at": "2025-08-14T15:30:00+09:00",
  "days": 30
}
```

## How it works

1. Executes `logcli labels` to get all available labels
2. For each label, executes `logcli labels <label>` to get all values
3. Saves the collected data to a JSON file
4. Shows progress during execution with the actual commands being run

## Notes

- The tool requires `logcli` to be in your PATH
- Loki instances with many labels and label values may take some time to dump
- If fetching values for a specific label fails, the tool continues with a warning
- The time range (`-days`) affects which labels and values are discovered

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.
