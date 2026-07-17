#!/bin/sh
set -e

echo "Setting up Layer 2 Bridge for Pure TUN/TAP Architecture..."

# 1. Extract current eth0 network details
IP=$(ip -4 addr show eth0 | grep inet | awk '{print $2}')
GW=$(ip route | grep default | awk '{print $3}')
MAC=$(cat /sys/class/net/eth0/address)

# 2. Create the TAP interface early
ip tuntap add dev tap0 mode tap
ip link set tap0 up

# 3. Create the Bridge interface (copying eth0's MAC to avoid ARP poisoning)
ip link add name br0 type bridge
ip link set dev br0 address $MAC
ip link set br0 up

# 4. Bind eth0 and tap0 to the Bridge
ip link set eth0 master br0
ip link set tap0 master br0

# 5. Migrate IP and Route to the Bridge
ip addr flush dev eth0
ip addr add $IP dev br0
if [ ! -z "$GW" ]; then
    ip route add default via $GW dev br0
fi

echo "Bridge setup complete! Executing VDR..."
exec "$@"
