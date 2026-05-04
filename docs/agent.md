# Agent Usage

`oqx` is designed to be usable by coding agents without prior OQTOPUS context.

Start by reading:

```sh
oqx --output json context
oqx --output json devices summary
```

Use `--output json` for machine-readable output. Use the simulator device
`qulacs` by default unless the user explicitly asks to use QPU hardware.

## Submit A Sampling Job

Generate a request body:

```sh
oqx examples submit-job --device qulacs --shots 1000 > job.json
```

Or submit an OPENQASM 3 file directly:

```sh
oqx --output json jobs submit-sampling \
  --device qulacs \
  --program bell.qasm \
  --shots 1000 \
  --name "Bell sampling"
```

For explicit JSON bodies:

```sh
oqx --output json jobs submit --file job.json
```

Then wait and read the result:

```sh
oqx --output json jobs wait "$JOB_ID" --timeout 10m
oqx --output json jobs result "$JOB_ID"
```

## Safety

Do not use QPU devices, high shot counts, job deletion, or token mutation unless
the user clearly asks for that action.

Use `oqx raw METHOD PATH` only when the first-class command is missing:

```sh
oqx --output json raw GET /jobs/"$JOB_ID"/status
```
