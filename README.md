# Overview

This is a simple control plan that manages a GCP firewall rule via Pulumi.

## Demo

Once server is running, and underlying cloud infra has been created,
curl /read to get state of firewall rule right now:

```
$ curl localhost:8080/read
2.2.2.0/24
 is allowed!
```

2.2.2.0/24 is not my public IP so pinging shouldn't work (firewall rule protocol = ICMP):

```
$ ping <public ip of cloud infra>
PING <public ip of cloud infra> (<public ip of cloud infra>): 56 data bytes
Request timeout for icmp_seq 0
Request timeout for icmp_seq 1
Request timeout for icmp_seq 2
Request timeout for icmp_seq 3
Request timeout for icmp_seq 4
Request timeout for icmp_seq 5
^C
--- <public ip of cloud infra> ping statistics ---
7 packets transmitted, 0 packets received, 100.0% packet loss
```

Now we get the control plane ready to write my public IP to the underlying GCP firewall:

```
$ curl localhost:8080/prepare -X POST --data "<my public ip>"
<my public ip> added!
$ curl localhost:8080/read
<my public ip>
 is allowed!
$ curl localhost:8080/diff
Here's the plan!
Previewing update (dev):
  pulumi:pulumi:Stack: (same)
...
Resources:
    ~ 1 to update
    3 unchanged
...
```

Plan looks good. Now we tell pulumi to execute it:

```
$ curl localhost:8080/write
Cloud provider updated!
```

Now we can ping!!

```
$ ping <public ip of cloud infra>
PING <public ip of cloud infra> (<public ip of cloud infra>): 56 data bytes
64 bytes from <public ip of cloud infra>: icmp_seq=0 ttl=56 time=59.500 ms
64 bytes from <public ip of cloud infra>: icmp_seq=1 ttl=56 time=151.206 ms
64 bytes from <public ip of cloud infra>: icmp_seq=2 ttl=56 time=55.875 ms
```
