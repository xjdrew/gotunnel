## Run with systemd service

1. copy gotunnel binary to `/usr/local/bin/gotunnel`
```bash
sudo cp $GOPATH/bin/gotunnel /usr/local/bin/gotunnel

```

2. copy this repo's `systemd/gotunnel-server.service` or `systemd/gotunnel-client.service` to `/etc/systemd/system`

3. start service
```bash
systemctl daemon-reload

systemctl start gotunnel-server.service

# or 

systemctl start gotunnel-client.service

```


