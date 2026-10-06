# Nebula lab

A Nebula mesh running in Docker: one lighthouse plus any number of hosts.

Run everything from the `nebula-lab` folder in PowerShell.

| Host        | Nebula IP        | Docker IP    |
|-------------|------------------|--------------|
| lighthouse1 | 192.168.100.1/24 | 172.28.0.2   |
| host1       | 192.168.100.5/24 | 172.28.0.3   |
| host2       | 192.168.100.9/24 | 172.28.0.4   |

Keep this table updated so you never reuse an IP.

---

# Part 1: Set up the lighthouse (one time)

## 1. Pull the image and make the folders

```powershell
docker pull nebulaoss/nebula
mkdir nebula-lab
cd nebula-lab
mkdir certs, lighthouse
```

## 2. Create the certificate authority (CA)

```powershell
docker run --rm -v "${PWD}/certs:/certs" -w /certs --entrypoint /nebula-cert nebulaoss/nebula ca -name "My Lab"
```

This creates `ca.crt` and `ca.key` in `certs/`. **`ca.key` is the master secret.** Never copy it into a host folder, send it to anyone, or commit it to git.

By default the CA (and every cert it signs) expires after one year.

## 3. Sign the lighthouse cert

```powershell
docker run --rm -v "${PWD}/certs:/certs" -w /certs --entrypoint /nebula-cert nebulaoss/nebula sign -name lighthouse1 -ip 192.168.100.1/24
```

## 4. Copy the files into the lighthouse folder

```powershell
cp certs\ca.crt lighthouse\
cp certs\lighthouse1.crt lighthouse\host.crt
cp certs\lighthouse1.key lighthouse\host.key
```

## 5. Write `lighthouse\config.yml`

The file must be named exactly `config.yml` (not `.yaml` or `.txt`) and saved as UTF-8. Use VS Code or Notepad, not PowerShell redirection.

```yaml
pki:
  ca: /config/ca.crt
  cert: /config/host.crt
  key: /config/host.key

static_host_map: {}

lighthouse:
  am_lighthouse: true

listen:
  host: 0.0.0.0
  port: 4242

punchy:
  punch: true

tun:
  dev: nebula1

firewall:
  outbound:
    - port: any
      proto: any
      host: any
  inbound:
    - port: any
      proto: icmp
      host: any
```

## 6. Create `docker-compose.yml`

```yaml
services:
  lighthouse:
    image: nebulaoss/nebula
    container_name: lighthouse
    cap_add: [NET_ADMIN]
    devices: ["/dev/net/tun"]
    volumes: ["./lighthouse:/config"]
    networks:
      nebnet:
        ipv4_address: 172.28.0.2

networks:
  nebnet:
    ipam:
      config:
        - subnet: 172.28.0.0/24
```

Hosts get added to this file later (Part 2).

## 7. Start the lighthouse

```powershell
docker compose up -d
docker compose logs lighthouse
```

The logs should show Nebula starting with no errors. If the container exits, check:
- `no config files found`: the file isn't named exactly `config.yml`
- cert or key errors: the files in `lighthouse\` must be named `ca.crt`, `host.crt`, `host.key`

Check what the container sees:

```powershell
docker run --rm -v "${PWD}/lighthouse:/config" alpine ls -la /config
```

---

# Part 2: Add a new host

Example below adds `host2`. For each new host, change the name, the Nebula IP, and the Docker IP.

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

For the first host, create `host2\config.yml` with the contents below. For later hosts, copy an existing host's config.

```yaml
pki:
  ca: /config/ca.crt
  cert: /config/host.crt
  key: /config/host.key

static_host_map:
  "192.168.100.1": ["172.28.0.2:4242"]

lighthouse:
  am_lighthouse: false
  interval: 60
  hosts:
    - "192.168.100.1"

listen:
  host: 0.0.0.0
  port: 0

punchy:
  punch: true

tun:
  dev: nebula1

firewall:
  outbound:
    - port: any
      proto: any
      host: any
  inbound:
    - port: any
      proto: icmp
      host: any
```

The file must be named exactly `config.yml`. The lighthouse address `172.28.0.2` stays the same for every host in Docker.

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

You can also check the lighthouse logs for a handshake from the new host:

```powershell
docker compose logs lighthouse
```

You don't need to change or restart the lighthouse or other hosts.

---

# Notes

## Don't commit secrets

Add to `.gitignore`:

```
nebula-lab/certs/
nebula-lab/*/host.key
```

## Hosts outside this Docker network

The `172.28.0.2` address only exists inside Docker. A host on another computer needs:

1. The lighthouse port published: add `ports: ["4242:4242/udp"]` to the lighthouse service.
2. A Windows Firewall rule and a router port forward for UDP 4242.
3. The lighthouse's public IP in its `static_host_map` instead of `172.28.0.2`.

## Common problems

- **Container exits right away:** run `docker logs <name> --tail 20`.
- **Handshake never completes:** check `static_host_map` and that the lighthouse container is running.
- **Ping works but nothing else does:** the firewall only allows ICMP. Add inbound rules in each `config.yml`, for example `port: 22, proto: tcp, host: any`, then `docker compose restart`.
- **Certs expire after a year.** Re-sign them before then.