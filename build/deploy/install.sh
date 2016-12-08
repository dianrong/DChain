#!/usr/bin/env bash

systemDir=/lib/systemd/system/

install_ca() {
  cp caserver.service $systemDir
  systemctl enable caserver
  systemctl start caserver
}

install_geth() {
  cp geth.service $systemDir
  systemctl enable geth
}

case $1 in
  ca)
    install_ca 
    ;;
  geth)
    install_geth 
    ;;
  *)
    echo "Usage: $0 {ca|geth}"
    exit
    ;;
esac
