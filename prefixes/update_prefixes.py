#!/usr/bin/env python3

from hashlib import sha256
import ipaddress
from urllib.request import urlopen

PREFIXES_URL = "https://mask-api.icloud.com/egress-ip-ranges.csv"
HASH_FILE = "upstream-list.hash"

if __name__ == "__main__":
    response = urlopen(PREFIXES_URL)
    prefixes_csv = response.read()
    prefixes_hash = sha256(prefixes_csv).hexdigest()

    with open(HASH_FILE) as f:
        previous_hash = f.read()

    # only update if upstream list has changed
    if prefixes_hash != previous_hash:
        print("hashes differ, updating prefix files")

        ipv4_raw = []
        ipv6_raw = []

        lines = prefixes_csv.split(b"\n")
        for line in lines:
            prefix_bytes = line.split(b",", 1)[0]
            prefix = ipaddress.ip_network(prefix_bytes.decode("utf-8"))

            if prefix.version == 4:
                ipv4_raw.append(prefix)
            else:
                ipv6_raw.append(prefix)

        # collapse each set of prefixes
        ipv4 = list(ipaddress.collapse_addresses(ipv4_raw))
        ipv6 = list(ipaddress.collapse_addresses(ipv6_raw))

        # write out to file, with first line being number of collapsed prefixes
        for file, prefixes in (("ipv4.txt", ipv4), ("ipv6.txt", ipv6)):
            with open(file, "w") as out:
                out.write(f"{len(prefixes)}\n")

                for prefix in prefixes:
                    out.write(f"{prefix}\n")

        # update prefixes hash in file
        with open(HASH_FILE, "w") as f:
            f.write(prefixes_hash)

    else:
        print("skipping update since list hasn't changed")

