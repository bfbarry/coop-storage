# Nebula lab

A small Nebula mesh that runs entirely in Docker: one lighthouse and two hosts.
Everything it needs (certs, keys, configs) is already in this folder, so you only need Docker.

> The certs and keys here are for this test lab only. Don't reuse them for anything real.

## Requirements

- Docker Desktop (or Docker Engine + Compose), running
- Git

## Start it

```powershell
cd nebula-lab
docker compose up -d
```

This works the same in PowerShell, Mac/Linux terminals, and Git Bash.

## Check that it works

1. Confirm all three containers are running:

```powershell
   docker compose ps
```

   You should see `lighthouse`, `host1` with status `Up`.

2. Look for handshakes in the lighthouse logs:

```powershell
   docker compose logs lighthouse
```

   You should see `Handshake message sent` lines mentioning `host1` and `host2`.

3. Ping across the mesh. The Nebula image has no `ping`, so this uses a throwaway container on the host's network:

```powershell
   docker run --rm --network container:host1 alpine ping -c 3 192.168.100.1
   docker run --rm --network container:host1 alpine ping -c 3 192.168.100.9
```

   The first pings the lighthouse, the second pings `host2` through the overlay. Both should get replies.

## Stop it

```powershell
docker compose down
```

## Addresses

| Node        | Nebula IP       | Docker IP  |
|-------------|-----------------|------------|
| lighthouse1 | 192.168.100.1   | 172.28.0.2 |
| host1       | 192.168.100.5   | 172.28.0.3 |
| host2       | 192.168.100.9   | 172.28.0.4 |

## Notes

- This runs your own copy of the mesh on your own computer. It doesn't connect to anyone else's machine.
- The firewall only allows ping (ICMP) between hosts. To allow other traffic, add rules to the `inbound` section of each node's `config.yml`, then run `docker compose restart`.
- Certs expire one year after they were created. If the containers stop connecting after that, the certs need to be re-signed.

## Troubleshooting

| Problem | Fix |
|---|---|
| A container exits right away | `docker logs <name> --tail 20` and read the last lines |
| `no config files found at /config/config.yml` | The file must be named exactly `config.yml` (not `.yaml` or `.txt`) |
| `/dev/net/tun` or "operation not permitted" errors | Make sure Docker Desktop is running Linux containers (WSL2 on Windows). On Linux, run `sudo modprobe tun` |
| Port or subnet conflict | Another Docker network is using `172.28.0.0/24`. Run `docker network ls` and remove the conflicting one |
| Ping fails but containers are up | Wait 10 seconds for the handshake, then check `docker compose logs` |