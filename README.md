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
- User directories (replicated across nodes)
- A user shall access a file directory only if it's theirs

Auth flow  (see [here](https://chatgpt.com/share/68f3cb5b-08e4-8006-909f-28d660354c7b))

| Step | Description                            |
| ---- | -------------------------------------- |
| 1–2  | User authenticates with SSO            |
| 3–4  | Backend validates and issues own JWT   |
| 5–6  | Frontend uses JWT to request file      |
| 7    | Backend verifies JWT and decrypts file |
| 8    | File is streamed securely to user      |

- rate limiting (with Nginx)
## UI
- upload multiple files form (make super optimized)
- file explorer

# Resources
- Series of articles on [replication](https://www.enjoyalgorithms.com/blog/storage-and-redundancy)
- https://github.com/google/go-cloud
- https://medium.com/@kamal.maiti/object-based-storage-architecture-b841e5842124
