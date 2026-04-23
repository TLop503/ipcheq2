#!/usr/bin/env python3

import sys
import ipaddress


def parse_entries(filename):
    ips = set()
    networks = set()

    with open(filename, "r") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue

            try:
                if "/" in line:
                    net = ipaddress.ip_network(line, strict=False)
                    networks.add(net)
                else:
                    ip = ipaddress.ip_address(line)
                    ips.add(ip)
            except ValueError:
                print(f"Skipping invalid entry: {line}", file=sys.stderr)

    return ips, networks


def remove_covered_ips(ips, networks):
    remaining_ips = set()

    for ip in ips:
        if not any(ip in net for net in networks):
            remaining_ips.add(ip)

    return remaining_ips


def remove_redundant_subnets(networks):
    networks = list(networks)
    networks.sort(key=lambda n: n.prefixlen)  # smallest prefix (largest net) first

    pruned = []

    for net in networks:
        if not any(net.subnet_of(existing) for existing in pruned):
            pruned.append(net)

    return set(pruned)


def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <input_file>")
        sys.exit(1)

    ips, networks = parse_entries(sys.argv[1])

    # Remove subnets covered by larger subnets
    networks = remove_redundant_subnets(networks)

    # Remove IPs covered by remaining subnets
    ips = remove_covered_ips(ips, networks)

    # Output
    for net in sorted(networks, key=lambda n: (n.version, n.network_address, n.prefixlen)):
        print(net)

    for ip in sorted(ips, key=lambda i: (i.version, i)):
        print(ip)


if __name__ == "__main__":
    main()