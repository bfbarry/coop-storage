#!/bin/sh
ngrok_file=~/.config/ngrok/ngrok.yml
if ! grep -q "web_addr" $ngrok_file; then
    ngrok config add-authtoken $NGROK_AUTH_TOKEN &&
    echo '    web_addr: 0.0.0.0:4040' >> $ngrok_file
fi
ngrok http --region=us --log=stdout http://api:8181