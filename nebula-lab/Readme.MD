# How to add a new host

Example below adds `host2`. For each new host, change the name, the Nebula IP, and the Docker IP.

Run everything from the `nebula-lab` folder in PowerShell.

## 1. Make a folder for the host

```powershell
mkdir host2
```

## 2. Sign a cert for the host

The CA already exists, so this only signs a new host certificate. Pick a unique Nebula IP.

```powershell
docker run --rm -v "${PWD}/certs:/certs" -w /certs --entrypoint /nebula-cert nebulaoss/nebula sign -name host2 -ip 192.168.100.9/24
```

## 3. Copy the certs into the folder

Never copy `ca.key` into a host folder.

```powershell
cp certs\ca.crt host2\
cp certs\host2.crt host2\host.crt
cp certs\host2.key host2\host.key
```

## 4. Write the config

Copy the existing host's config. The file must be named exactly `config.yml`, not `.yaml` or `.txt`.

```powershell
cp host1\config.yml host2\config.yml
```

No edits needed: the lighthouse address in `static_host_map` stays the same.

## 5. Add the host to docker-compose.yml

Use a new container name, folder, and an unused Docker IP.

```yaml
  host2:
    image: nebulaoss/nebula
    container_name: host2
    cap_add: [NET_ADMIN]
    devices: ["/dev/net/tun"]
    volumes: ["./host2:/config"]
    depends_on: [lighthouse]
    networks:
      nebnet:
        ipv4_address: 172.28.0.4
```

## 6. Start it

```powershell
docker compose up -d
```

## 7. Verify

The Nebula image has no `ping`, so use a throwaway container on the host's network:

```powershell
docker run --rm --network container:host2 alpine ping -c 3 192.168.100.1
```

You don't need to change or restart the lighthouse or other hosts.

## IP reference

| Host        | Nebula IP        | Docker IP    |
|-------------|------------------|--------------|
| lighthouse1 | 192.168.100.1/24 | 172.28.0.2   |
| host1       | 192.168.100.5/24 | 172.28.0.3   |
| host2       | 192.168.100.9/24 | 172.28.0.4   |