#!/usr/bin/env bash

mdadm --stop /dev/md0 || true
mdadm --remove /dev/md0
mdadm --zero-superblock /dev/nvme0n1 /dev/nvme1n1

wipefs --all --force /dev/nvme0n1
wipefs --all --force /dev/nvme1n1
dd if=/dev/zero of=/dev/nvme0n1 bs=1M count=10
dd if=/dev/zero of=/dev/nvme1n1 bs=1M count=10
parted /dev/nvme0n1 mklabel gpt
parted /dev/nvme1n1 mklabel gpt

wget https://factory.talos.dev/image/3531bf15c8738b4bc46f2cdd7c5cd68fea388796b291117f0ee38b51a335fc47/v1.9.2/metal-amd64.raw.zst -O talos.raw.zst
zstd -d talos.raw.zst -o talos.raw

dd if=talos.raw of=/dev/nvme0n1 bs=4M status=progress
sync

reboot