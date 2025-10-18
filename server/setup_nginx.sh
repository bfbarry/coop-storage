#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$(readlink -f "${BASH_SOURCE[0]}")" )" &> /dev/null && pwd )"
sudo ln -sf $SCRIPT_DIR/nginx.conf /opt/homebrew/etc/nginx/servers/upload-service.conf
sudo nginx -t
brew services restart nginx 