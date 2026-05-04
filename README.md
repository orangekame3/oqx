# oqx

<p align="center">
  <img src="docs/assets/oqx-logo.webp" alt="oqx logo" width="128">
</p>

`oqx` is an unofficial Go CLI for the OQTOPUS Cloud User API, designed first for
coding agents and automation.

It provides compact machine-readable commands for discovering devices,
submitting jobs, waiting for completion, and reading results without requiring an
agent to understand the full OpenAPI schema or large API payloads.

This project is not an official OQTOPUS project. The API surface is based on
`oqtopus-team/oqtopus-cloud` `backend/oas/user/openapi.yaml`, and authentication
is compatible with the Python SDK `quri-parts-oqtopus`.

## Install

```sh
go install ./cmd/oqx
```

After releases are published:

```sh
brew install orangekame3/tap/oqx
```

## Quick Start

If you already use the Python SDK, `oqx` can read `~/.oqtopus`:

```ini
[default]
url=https://<your-oqtopus-api-host>
api_token=your-api-token
```

You can also configure `oqx` directly:

```sh
oqx auth login --base-url https://<your-oqtopus-api-host> --token-stdin
oqx auth status
```

Then inspect the API context and available devices:

```sh
oqx --output json context
oqx --output json devices summary
```

## Agent Workflow

Recommended read-submit-wait-result flow:

```sh
oqx --output json context
oqx --output json devices summary
oqx examples submit-job --device qulacs --shots 1000 > job.json
oqx --output json jobs submit --file job.json
oqx --output json jobs wait "$JOB_ID" --timeout 10m
oqx --output json jobs result "$JOB_ID"
```

For a simple OPENQASM 3 sampling job:

```sh
oqx --output json jobs submit-sampling \
  --device qulacs \
  --program bell.qasm \
  --shots 1000 \
  --name "Bell sampling"
```

Agents should prefer `qulacs` unless the user explicitly asks to use QPU
hardware. Avoid high shot counts, job deletion, API token mutation, and QPU jobs
unless the user clearly requests them.

See [docs/agent.md](docs/agent.md) for the fuller agent playbook.

## Configuration

Configuration precedence:

```text
flags > OQX_* env > OQTOPUS_* env > oqx config > ~/.oqtopus > defaults
```

Supported environment variables:

```sh
export OQX_BASE_URL=https://<your-oqtopus-api-host>
export OQX_API_TOKEN='your-api-token'

export OQTOPUS_URL=https://<your-oqtopus-api-host>
export OQTOPUS_API_TOKEN='your-api-token'
```

`oqx` sends SDK-compatible API tokens with the `q-api-token` header. A bearer
token can still be used when needed:

```sh
oqx --base-url https://<your-oqtopus-api-host> --api-token "$TOKEN" devices summary
oqx auth login --base-url https://<your-oqtopus-api-host> --token-stdin
oqx auth login --base-url https://<your-oqtopus-api-host> --bearer --token-stdin
```

For coding agents, prefer environment variables or explicit flags over persisted
credentials.

## Output

Use `--output json` for compact machine-readable JSON.

```sh
oqx --output json devices summary
oqx --output json jobs status "$JOB_ID"
```

Output modes:

```text
pretty  human-readable JSON, default
json    compact JSON for agents
raw     response body as-is
```

Exit codes:

```text
0 success
1 CLI/input error
2 authentication or authorization error
3 not found
4 other API 4xx error
5 API 5xx error
6 network, timeout, or cancellation error
```

## Commands

Agent-oriented commands:

```sh
oqx context
oqx devices summary
oqx examples submit-job --device qulacs --shots 1000
oqx jobs submit-sampling --device qulacs --program bell.qasm --shots 1000
oqx jobs wait JOB_ID --interval 5s --timeout 10m
oqx jobs result JOB_ID
oqx raw GET /jobs/JOB_ID/status
```

General User API commands:

```sh
oqx auth login --base-url https://<your-oqtopus-api-host> --token-stdin
oqx auth status
oqx auth logout

oqx devices list
oqx devices get DEVICE_ID

oqx jobs list --status submitted --page 1 --size 10 --order ASC
oqx jobs submit --file job.json
oqx jobs get JOB_ID
oqx jobs status JOB_ID
oqx jobs cancel JOB_ID
oqx jobs delete JOB_ID
oqx jobs sselog --output sselog.zip JOB_ID

oqx api-token create
oqx api-token status
oqx api-token delete

oqx announcements list --limit 10 --order DESC
oqx announcements get ANNOUNCEMENT_ID

oqx user get
oqx user update --name "Alice" --organization "Example Org"
oqx user update --file user.json
oqx user delete

oqx settings get
```

`oqx raw METHOD PATH` is an escape hatch for endpoints that do not yet have a
first-class command:

```sh
oqx --output json raw GET /devices
oqx --output json raw POST /jobs --file job.json
oqx --output json raw GET /jobs/"$JOB_ID"/status
```

Use repeated `--query key=value` flags for query parameters.

## Development

The CLI is built with Cobra. Release automation follows the same pattern as
`arq`: tagpr updates `cmd/version.go`, tags trigger GoReleaser, and GoReleaser
publishes GitHub release assets plus a Homebrew formula.

```sh
make build
make test
make lint
make ci
```

## License

The CLI source code is licensed under the Apache License 2.0. See [LICENSE](LICENSE).

The logo and artwork assets are separate from the CLI code and are licensed
under Creative Commons Attribution 4.0 International (CC BY 4.0). See
[NOTICE](NOTICE) and [docs/assets/README.md](docs/assets/README.md).

## Artwork

The logo image is cropped and adapted from
`orangekame3/codex-pets/oqtopus/spritesheet.webp`, which is derived from the
original OQTOPUS artwork. Artwork is licensed under CC BY 4.0.
