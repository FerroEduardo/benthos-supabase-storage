# Benthos Supabase Storage plugin

[Benthos](https://www.benthos.dev/) plugin for saving files in Supabase Storage.

The plugin/output file can be found [here](output/supabase_storage.go).

## How to run

```bash
go run main.go -c config/example.yaml
```

> Set the Supabase token before running

## Example

> [example.yml](config/example.yaml)
```yml
input:
  stdin:
    codec: lines

pipeline:
  threads: 1
  processors:
  - mapping: |
      root.data = content()
      root.filename = "test.txt"
      root.contentType = "text/plain"

output:
  supabase_storage:
    bucket: lotus
    baseUrl: http://localhost:8000/storage/v1
    token: SUPABASE_API_TOKEN
```