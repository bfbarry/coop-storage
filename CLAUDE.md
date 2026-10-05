# CLAUDE.md

This app is a storage system like google drive. See the docker compose in `/devops` to understand the architecture. There is an API (metadata-server), a postgres server to store object metadata, and the object storage device (rustFS).

`/metadata-server` is the main backend that the client interacts with. It contains the logic for the metadata associated with objects, as well as the logic to interface with rustFS.  Clerk is used for authentication.

Here is one example flow: uploading a file:
1. User pings metadata server for a presigned URL to upload to rustFS
2. User sends data to rustFS
3. If successful, client pings metadata server to create new metadata entry about this file 
