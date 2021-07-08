#!/usr/bin/env bash

echo "setting sysctl"
sudo sysctl -w net.core.rmem_max=21299200
sudo sysctl -w net.core.rmem_default=2500000
