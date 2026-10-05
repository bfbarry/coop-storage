# Connect a host from a different computer / Wi-Fi network

The new computer must be able to reach the lighthouse over the internet, so one of these must be true:

- **Same network (easy demo):** both computers are on the same Wi-Fi. Use the lighthouse laptop's local IP (run `ipconfig`, e.g. `192.168.0.217`).
- **Different network:** the lighthouse needs a public address. That means a router port forward (UDP 4242 to the lighthouse laptop) or running the lighthouse on a VPS. The public IP is whatever "what is my IP" shows from the lighthouse's network.

# Connect a host from a different computer / Wi-Fi network

The new computer reaches the lighthouse over the internet, so the lighthouse laptop must be reachable from outside its home network. That requires:

1. Publishing the lighthouse port from Docker
2. A Windows Firewall rule
3. A router port forward (UDP 4242) to the lighthouse laptop

Do all three before the friend tries to connect.

## On the lighthouse laptop

### 1. Publish the lighthouse port

Add to the lighthouse service in `docker-compose.yml`:

```yaml
    ports:
      - "4242:4242/udp"
```

```powershell
docker compose up -d
```

### 2. Allow it through Windows Firewall (run PowerShell as Administrator)

```powershell`
New-NetFirewallRule -DisplayName "Nebula" -Direction Inbound -Protocol UDP -LocalPort 4242 -Action Allow
```

### 3. Keep the lighthouse running

The laptop must be on and the lighthouse container up while the friend connects.

## On the new computer (needs Docker)

### 1. Pull the image

```powershell
docker pull nebulaoss/nebula
```

### 2. Generate a keypair (the private key never leaves this computer)

```powershell
mkdir nebula-host
cd nebula-host
docker run --rm -v "${PWD}:/out" -w /out --entrypoint /nebula-cert nebulaoss/nebula keygen -out-key host.key -out-pub host.pub
```

Send **only `host.pub`** to the lighthouse owner.

## Back on the lighthouse laptop

### 3. Sign the new host's public key

Put `host.pub` in `nebula-lab/certs`, then pick a unique Nebula IP (e.g. 192.168.100.20):

```powershell
docker run --rm -v "${PWD}/certs:/certs" -w /certs --entrypoint /nebula-cert nebulaoss/nebula sign -name friend1 -ip 192.168.100.20/24 -in-pub host.pub
```

### 4. Send back two files

- `certs/friend1.crt`
- `certs/ca.crt`

**Never send `ca.key`.**

## On the new computer

### 5. Put the files in `nebula-host`

Rename `friend1.crt` to `host.crt`. The folder should contain:

```
ca.crt
host.crt
host.key
config.yml
```

### 6. Write `config.yml`

Replace `LIGHTHOUSE_IP` with the lighthouse's local IP (same network) or public IP (different network):

```yaml
pki:
  ca: /config/ca.crt
  cert: /config/host.crt
  key: /config/host.key

static_host_map:
  "192.168.100.1": ["LIGHTHOUSE_IP:4242"]

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

### 7. Run the container

```powershell
docker run -d --name friend1 --cap-add NET_ADMIN --device /dev/net/tun -v "${PWD}:/config" nebulaoss/nebula
```

### 8. Check that it worked

```powershell
docker logs friend1 --tail 20
docker run --rm --network container:friend1 alpine ping -c 3 192.168.100.1
```

The lighthouse logs (`docker compose logs lighthouse`) should show a handshake from the new host.

## Troubleshooting

- **`no config files found`:** the file must be named exactly `config.yml` (not `.yaml` or `.txt`).
- **Handshake never completes:** check the `LIGHTHOUSE_IP`, that the lighthouse container is running, and that UDP 4242 is allowed through the firewall (and forwarded on the router for a different network).
- **Public IP changed:** update `static_host_map` on the new host and restart its container.
- **Ping works but nothing else:** the firewall only allows ICMP. Add inbound rules for other ports in each `config.yml`.