# Port forwarding (REQUIRED for connecting from another network)

Your friend is on a different network, so the internet has to be able to reach the lighthouse laptop. Home routers block this by default. A "port forward" tells your router: "when something arrives on UDP port 4242, send it to my laptop."

Do every step in order. Do not skip any.

You will need:
- The lighthouse laptop, connected to your home network
- The router login (see "Can't log in" below if you don't have it)

---

## Part A: Prepare the laptop

### A1. Publish the port from Docker

Open `docker-compose.yml` in the `nebula-lab` folder. Find the `lighthouse:` section and add the `ports:` lines exactly like this (spaces matter, use spaces not tabs):

```yaml
  lighthouse:
    image: nebulaoss/nebula
    container_name: lighthouse
    ports:
      - "4242:4242/udp"
    cap_add: [NET_ADMIN]
    ...
```

Save the file, then in PowerShell, in the `nebula-lab` folder:

```powershell
docker compose up -d
docker ps
```

You should see `0.0.0.0:4242->4242/udp` in the PORTS column for the lighthouse. If you don't, the port isn't published and nothing below will work.

### A2. Allow the port in Windows Firewall

1. Click Start, type `PowerShell`.
2. Right-click **Windows PowerShell** and choose **Run as administrator**. Click Yes.
3. Paste this and press Enter:

```powershell
New-NetFirewallRule -DisplayName "Nebula" -Direction Inbound -Protocol UDP -LocalPort 4242 -Action Allow
```

It should print a block of text with `Enabled : True`.

### A3. Write down the laptop's local IP

In PowerShell:

```powershell
ipconfig
```

Scroll to the section for the adapter you're using. It's called **Wireless LAN adapter Wi-Fi** if you're on Wi-Fi, or **Ethernet adapter Ethernet** if you're plugged in. Ignore sections named vEthernet, WSL, or Docker.

Write down the **IPv4 Address**. It looks like `192.168.0.217`. This is your **laptop IP** for the rest of this guide. Also write down the **Default Gateway** in the same section, which is your router's address (probably `192.168.0.1`).

---

## Part B: Set up the TP-Link router

### B1. Open the router page

1. Make sure your laptop is on your home Wi-Fi (not a VPN or phone hotspot).
2. Open a browser.
3. Click the address bar, type `http://tplinkwifi.net` and press Enter. If that doesn't load, type `http://192.168.0.1` (or your Default Gateway from A3).
4. Log in with your router password.

**Can't log in?**
- Try the password printed on the sticker on the router.
- If the page asks you to *create* a password, make one and write it down.
- If you use the TP-Link **Deco** app, there is no web page. Do the same steps in the app (More > Advanced > Port Forwarding).
- Last resort: factory reset by holding the small Reset button on the router for 10 seconds. This erases your Wi-Fi name and password and every device must reconnect.

### B2. Reserve the laptop's IP (so it never changes)

If the laptop's IP changes later, the forward silently breaks. Prevent that:

1. Click **Advanced** at the top.
2. Click **Network** on the left, then **DHCP Server**.
3. Find **Address Reservation** and click **Add**.
4. Pick your laptop from the device list (or enter its MAC address) and set the IP to your laptop IP from A3.
5. Click **Save**.

(Some models put this under *Advanced > Network > LAN Settings* or *DHCP > Address Reservation*.)

### B3. Create the port forward

1. Click **Advanced** at the top.
2. Click **NAT Forwarding** on the left, then **Virtual Servers**. (Older models: **Forwarding > Virtual Servers**.)
3. Click **Add**.
4. Fill in exactly:

| Field | Enter |
|---|---|
| Service Name | nebula |
| External Port | 4242 |
| Internal IP | your laptop IP from A3 |
| Internal Port | 4242 |
| Protocol | **UDP** (not TCP, not All) |

5. Make sure the entry is enabled, then click **OK/Save**.

You should now see a row for `nebula` in the list with the status enabled.

### B4. Check whether another router is in front of this one

This is the step most people hit. In the TP-Link page:

1. Click **Status** (or **Basic > Network Map > Internet**).
2. Find **WAN IP** (also called Internet IP).
3. Open a new tab, search Google for `what is my ip`, and note the number it shows.

Compare them:

- **The numbers are the same:** you are done with the router. Go to Part D.
- **WAN IP starts with `192.168.`, `10.`, or `172.16` to `172.31`:** your Spectrum modem/router is also a router. Go to Part C.
- **WAN IP starts with `100.64` to `100.127`:** your provider is sharing one public address between customers (CGNAT). Port forwarding is impossible. You must run the lighthouse on a rented server (VPS) instead. Stop here.

---

## Part C: Forward on the Spectrum device too (only if B4 said so)

You need to forward UDP 4242 a second time, pointing at the **TP-Link's WAN IP** (the private number you saw in B4).

### C1. Try the My Spectrum app (easiest)

1. Install **My Spectrum** on your phone and log in with your Spectrum account.
2. Tap **Services**, then **Internet**.
3. Tap your network or router, then **Advanced Settings** (or **Network Settings**).
4. Find **Port Forwarding** and add a rule:
   - Protocol: **UDP**
   - External/Public port: **4242**
   - Internal/Private port: **4242**
   - Device or IP: the **TP-Link's WAN IP** from B4
5. Save.

### C2. Or the Spectrum router web page

Try `http://192.168.1.1` in a browser while on your home network. Spectrum's login is usually on the sticker or set in your account. If the page doesn't load, use the app. If there is no port forwarding option, your Spectrum device is modem-only (no extra router), and the extra router is something else in your setup.

### C3. If you can't do C1 or C2

Alternative: put the TP-Link in **Access Point mode** (Advanced > Operation Mode). Then do the port forward in the Spectrum device only, pointing at the laptop IP. The laptop's IP will change when you do this, so redo A3.

If neither works, use a VPS.

---

## Part D: Find the public IP to give your friend

1. On the lighthouse laptop's network, search Google for `what is my ip`.
2. Write down the number it shows (yours was `96.41.98.94` at the time of writing).
3. This is `LIGHTHOUSE_IP` in your friend's `config.yml`:

```yaml
static_host_map:
  "192.168.100.1": ["96.41.98.94:4242"]
```

Your public IP can change without warning. If your friend suddenly can't connect, re-check this number first.

---

## Part E: Test it

You cannot test a UDP port with online port checkers (they only test TCP). The real test is the friend's connection.

1. Make sure the lighthouse is running:
```powershell
   docker ps
```
2. Watch the lighthouse logs:
```powershell
   docker compose logs -f lighthouse
```
3. Have your friend start their container.
4. Within a few seconds you should see a line containing `Handshake message sent` or `Handshake` with your friend's public IP.

**Nothing shows up? Check these in order:**
1. `docker ps` shows `0.0.0.0:4242->4242/udp` on the lighthouse.
2. The Windows Firewall rule exists (A2).
3. The TP-Link forward says **UDP**, external and internal port both 4242, and the Internal IP equals the laptop's current IP (run `ipconfig` again to confirm).
4. If there's a second router, the forward in Part C exists.
5. The friend's `config.yml` has your **current** public IP and the right port.
6. The laptop is awake (not sleeping) and still on the same Wi-Fi.