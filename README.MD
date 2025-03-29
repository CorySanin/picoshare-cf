# PicoShare-CF

[![GitHub Action](https://img.shields.io/github/actions/workflow/status/CorySanin/picoshare-cf/docker.yml)](https://github.com/CorySanin/picoshare-cf/actions/workflows/docker.yml)
[![GitHub Tag](https://img.shields.io/github/v/tag/CorySanin/picoshare-cf)](https://github.com/CorySanin/picoshare-cf/releases/latest)
[![Docker Pulls](https://img.shields.io/docker/pulls/corysanin/picoshare-cf)](https://hub.docker.com/r/corysanin/picoshare-cf)
[![License](https://img.shields.io/badge/license-agpl-blue)](https://github.com/CorySanin/picoshare-cf/blob/master/LICENSE)

This is a fork of [PicoShare](https://github.com/mtlynch/picoshare) with modifications that suit CloudFlare customers, including free tier.

## PicoShare-CF Features

- **IPs of proxied connections**: Real IPs of downloads over proxied connections are logged.
- **Files uploaded in chunks**: No problem uploading large files over proxied connections.

## PicoShare Features

Everything else is the same as the incredible upstream project you know and love.

- **Direct download links**: PicoShare gives you a direct download link you can share with anyone. They can view or download the file with no ads or signups.
- **No file restrictions**: Unlike sites like imgur, Vimeo, or SoundCloud that only allow you to share specific types of files, PicoShare lets you share any file of any size.
- **No resizing/re-encoding**: If you upload media like images, video, or audio, PicoShare never forces you to wait on re-encoding. You get a direct download link as soon as you upload the file, and PicoShare never resizes or re-encodes your file.

## Run PicoShare-CF

### From source

```bash
PS_SHARED_SECRET=somesecretpass PORT=4001 \
  go run cmd/picoshare/main.go
```

### From Docker

To run PicoShare-CF within a Docker container, mount a volume from your local system to store the PicoShare-CF sqlite database.

```bash
docker run \
  --env "PORT=4001" \
  --env "PS_SHARED_SECRET=somesecretpass" \
  --publish 4001:4001/tcp \
  --volume "${PWD}/data:/data" \
  --name picoshare-cf \
  corysanin/picoshare-cf
```

### From Docker + cloud data replication

If you specify settings for a [Litestream](https://litestream.io/)-compatible cloud storage location, PicoShare-CF will automatically replicate your data.

You can kill the container and start it later, and PicoShare-CF will restore your data from the cloud storage location and continue as if there was no interruption.

```bash
PORT=4001
PS_SHARED_SECRET="somesecretpass"
LITESTREAM_BUCKET=YOUR-LITESTREAM-BUCKET
LITESTREAM_ENDPOINT=YOUR-LITESTREAM-ENDPOINT
LITESTREAM_ACCESS_KEY_ID=YOUR-ACCESS-ID
LITESTREAM_SECRET_ACCESS_KEY=YOUR-SECRET-ACCESS-KEY

docker run \
  --publish "${PORT}:${PORT}/tcp" \
  --env "PORT=${PORT}" \
  --env "PS_SHARED_SECRET=${PS_SHARED_SECRET}" \
  --env "LITESTREAM_ACCESS_KEY_ID=${LITESTREAM_ACCESS_KEY_ID}" \
  --env "LITESTREAM_SECRET_ACCESS_KEY=${LITESTREAM_SECRET_ACCESS_KEY}" \
  --env "LITESTREAM_BUCKET=${LITESTREAM_BUCKET}" \
  --env "LITESTREAM_ENDPOINT=${LITESTREAM_ENDPOINT}" \
  --name picoshare-cf \
  corysanin/picoshare-cf
```

Notes:

- Only run one Docker container for each Litestream location.
  - PicoShare-CF can't sync writes across multiple instances.

### Using Docker Compose

To run PicoShare-CF under docker-compose, copy the following to a file called `docker-compose.yml` and then run `docker-compose up`.

```yaml
version: "3.2"
services:
  picoshare-cf:
    image: corysanin/picoshare-cf
    environment:
      - PORT=4001
      - PS_SHARED_SECRET=dummypass # Change to any password
    ports:
      - 4001:4001
    command: -db /data/store.db
    volumes:
      - ./data:/data
```

## Parameters

### Command-line flags

| Flag  | Meaning                 | Default Value     |
| ----- | ----------------------- | ----------------- |
| `-db` | Path to SQLite database | `"data/store.db"` |

### Environment variables

| Environment Variable | Meaning                                                                                 |
| -------------------- | --------------------------------------------------------------------------------------- |
| `PORT`               | TCP port on which to listen for HTTP connections (defaults to 4001).                    |
| `PS_BEHIND_PROXY`    | Set to `"true"` for better logging when PicoShare-cf is running behind a reverse proxy. |
| `PS_SHARED_SECRET`   | (required) Specifies a passphrase for the admin user to log in to PicoShare-cf.         |

### Docker environment variables

You can adjust behavior of the Docker container by specifying these Docker-specific variables with `docker run -e`:

| Environment Variable           | Meaning                                                                                               |
| ------------------------------ | ----------------------------------------------------------------------------------------------------- |
| `LITESTREAM_BUCKET`            | Litestream-compatible cloud storage bucket where Litestream should replicate data.                    |
| `LITESTREAM_ENDPOINT`          | Litestream-compatible cloud storage endpoint where Litestream should replicate data.                  |
| `LITESTREAM_ACCESS_KEY_ID`     | Litestream-compatible cloud storage access key ID to the bucket where you want to replicate data.     |
| `LITESTREAM_SECRET_ACCESS_KEY` | Litestream-compatible cloud storage secret access key to the bucket where you want to replicate data. |
| `LITESTREAM_RETENTION`         | The amount of time Litestream snapshots & WAL files will be kept (defaults to 72h).                   |

### Docker build args

If you rebuild the Docker image from source, you can adjust the build behavior with `docker build --build-arg`:

| Build Arg            | Meaning                                                                     | Default Value |
| -------------------- | --------------------------------------------------------------------------- | ------------- |
| `litestream_version` | Version of [Litestream](https://litestream.io/) to use for data replication | `0.3.9`       |

## Tips and tricks

### Reclaiming reserved database space

When you delete files, PicoShare-CF reserves the space for future uploads. If you'd like to reduce PicoShare-CF's usage of your filesystem, you can manually force PicoShare-CF to give up the space by performing the following steps:

1. Shut down PicoShare-CF.
1. Run `sqlite3 data/store.db 'VACUUM'` where `data/store.db` is the path to your PicoShare-CF database.

You should find that the `data/store.db` should shrink in file size, as it relinquishes the space dedicated to previously deleted files. If you start PicoShare-CF again, the System Information screen will show the smaller size of PicoShare-CF files.
