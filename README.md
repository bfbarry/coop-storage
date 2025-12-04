# README

```bash
├── cli-client # a command line tool to test uploading files
└── server     # file API and storage server
```

To get started:
```bash
$ cd server
$ docker compose up -d --force-recreate --build 
```

# TODO:
## Security
- encrypt files
- User buckets (replicated across nodes)
    - A user shall access a bucket only if it's theirs
- Auth flow:
    - Client pings Metadata server
    - Metadata server issues token, mappings etc
    - Client pings OSD server

- rate limiting (with Nginx)
## UI
- upload multiple files form (make super optimized)
- file explorer

# Resources
- Series of articles on [replication](https://www.enjoyalgorithms.com/blog/storage-and-redundancy)
- https://github.com/google/go-cloud
- https://medium.com/@kamal.maiti/object-based-storage-architecture-b841e5842124
