package templates

const Vlan string = `
network:
  version: 2
  renderer: networkd

  vlans:
    vlan10:
      id: 10
      link: {{ .Iface }}
      addresses: [10.21.199.3/24]
`