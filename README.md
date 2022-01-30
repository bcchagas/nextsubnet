![nextsubnet logo](./logo/nextsubnet-banner.jpeg)

# NEXTSUBNET
`nextsubnet` is a tool for helping you find out what the next subnet available is in a given virtual network, specially when it already has some subnets created and you want to fill in the gaps when possible.

## INSTALATION

Go to [releases](https://github.com/bcchagas/nextsubnet/releases) and download the version of you choosing.

## USAGE

```bash
Usage:
  nextsubnet -n network -m mask [--ignore-list list | --ignore-file file] [flags]

Examples:
  # Find the next /24 subnet in the network 10.0.0.0/22
  # that doesn't overlap any of the two existent subnets
  nextsubnet --network 10.0.0.0/22 --subnet-mask 25 --ignore-list 10.0.0.0/24,10.0.1.128/25

  # You can also pass in a file containing the subnets in use
  nextsubnet --network 10.0.0.0/22  --subnet-mask 24 --ignore-file subnets.txt
```

## MOTIVATION
When using the Azure Portal to provision a subnet it automatically calculates the subnet address space based on the mask you provide. On the other hand, when provisioning via `az-cli`, `ARM Template` or `REST API` you have to calculate the subnet range yourself.

## USE CASE SCENARIO

Let's say you have the following virtual network: `10.0.0.1/22` and two subnets whithin it: `10.0.0.0/24`, `10.0.1.128/25`

```yaml
vnet: 10.0.0.1/22
  subnet: 10.0.0.0/24
  subnet: 10.0.1.128/25
```

If you were to provision a `/25` subnet, the next available address space would fall between the two existing ones:

```yaml
vnet: 10.0.0.1/22
  subnet: 10.0.0.0/24
  subnet: 10.0.1.0/25
  subnet: 10.0.1.128/25
```

In the ever changing environment that is the cloud, it's often the case that a subnet is no longer needed and is eventually removed, leaving a gap. Using `nextsubnet` you can optimize the virtual network address space usage by filling those gaps whenever possible.

