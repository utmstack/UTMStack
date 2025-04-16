package templates

const VlanUbuntu string = `
network:
  version: 2
  renderer: {{ .Renderer }}

  vlans:
    vlan10:
      id: 10
      link: {{ .Iface }}
      addresses: [10.21.199.3/24]
`
const VlanRedHat string = `
DEVICE=vlan10
TYPE=Vlan
VLAN_ID=10
PHYSDEV={{ .Iface }}
BOOTPROTO=none
IPADDR=10.21.199.3
PREFIX=24
VLAN=yes
ONBOOT=yes
`
