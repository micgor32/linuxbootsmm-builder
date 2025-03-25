# LinuxbootSMM Builder
A simple script to build coreboot image with LinuxBootSMM as a payload.

Based on [corebootnerf](https://github.com/linuxboot/corebootnerf).

## Prerequisites
Please make sure you have Go >= 1.23, and your GOPATH is set up correctly.

## Usage
Install using go:
```sh
go install github.com/micgor32/linuxbootsmm-builder@latest
```
Check whether you have all dependencies mentioned [here](https://doc.coreboot.org/tutorial/part1.html#step-1-install-tools-and-libraries-needed-for-coreboot) installed, or run:
```sh
linuxbootsmm-builder --depinstall
```
For the remaining usage options please see:
```sh
Usage of linuxbootsmm-builder:
      --blobs string     Path to the custom site-local directory for coreboot (default "no")
      --build            Only build the image
      --config string    Path to config file for coreboot (default "default")
      --depinstall       Install all dependencies
      --fetch            Fetch all the things we need
```
It is highly recommended to provide your own config file - the default one is targeting RaptorLake-S board and would **not** work with other boards. Similarly, if no custom `site-local`
is provided, the builder will create empty one in which the kernel image and initramfs are going to be placed after compilation. Please note that when using default config, builder script assumes that its being run under in `/tmp`!

## Considerations when creating custom configuration files
In general providing custom configs should not introduce and unexpected errors during the compilation (does **not** apply to the rom image being functional!), as long as the following configuration options are enabled:
- `CONFIG_SMM_PAYLOAD_INTERFACE`: this tells coreboot to skip default SMM initialization, if disabled, Linux payload would not be able to perform SMM init.
- `CONFIG_PAYLOAD_LINUXBOOT`: enables LinuxBoot
- `CONFIG_LINUXBOOT_KERNEL_PATH="/tmp/coreboot-git/site-local/Image"`: the builder tool compiles patched Linux kernel separetely and places it under `site-local/Image`, without this path being specified, coreboot would compile unpatched Linux kernel during the its own compile process.
- `CONFIG_LINUXBOOT_INITRAMFS_PATH="/tmp/coreboot-git/site-local/initramfs_u-root.cpio"`: similarly as with the kernel, initramfs is build separately from coreboot by the builder.


