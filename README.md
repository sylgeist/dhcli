# dhcli
CLI utility for interacting with Kea DHCP

Stork requires authentication through ENV variables:

```
STORK_USER
STORK_PASS
```

You can find the right values in Vault under:

`stork-dhcp/tools`

#### Known Bugs

The reservation search doesn't work with Stage2 region names at this time.

`dhcli res S2R4`

## Production releases

The latest production version of the `dhcli` CLI command is available as follows:

|         | Intel 64-bit  | ARM 64-bit |
| ------- | ------------- | ---------- |
| MacOS   | [dhcli][1]     | [dhcli][2]  |
| Linux   | [dhcli][3]     | [dhcli][4]  |
| Windows | [dhcli.exe][5] |

[1]: https://artifactory/dhcli/production/darwin/amd64/dhcli
[2]: https://artifactory/dhcli/production/darwin/arm64/dhcli
[3]: https://artifactory/dhcli/production/linux/amd64/dhcli
[4]: https://artifactory/dhcli/production/linux/arm64/dhcli
[5]: https://artifactory/dhcli/production/windows/amd64/dhcli.exe

## Staging releases

The latest staging version of the `dhcli` CLI command is available as follows:

|         | Intel 64-bit   | ARM 64-bit |
| ------- | -------------- | ---------- |
| MacOS   | [dhcli][11]     | [dhcli][12] |
| Linux   | [dhcli][13]     | [dhcli][14] |
| Windows | [dhcli.exe][15] |

[11]: https://artifactoryg/dhcli/staging/darwin/amd64/dhcli
[12]: https://artifactoryg/dhcli/staging/darwin/arm64/dhcli
[13]: https://artifactoryg/dhcli/staging/linux/amd64/dhcli
[14]: https://artifactoryg/dhcli/staging/linux/arm64/dhcli
[15]: https://artifactoryg/dhcli/staging/windows/amd64/dhcli.exe
