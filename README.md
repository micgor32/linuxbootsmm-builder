# LinuxbootSMM Builder
A simple script to build coreboot image with LinuxBootSMM as a payload.

Based on [corebootnerf](https://github.com/linuxboot/corebootnerf).

## Prerequisites
Please make sure you have Go >= 1.23, and your GOPATH is set up correctly.
When installing dependencies, the script assumes that it is being run on
either Arch Linux, Fedora or Debian/Ubuntu.

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
      --smp              Compile Linux with SMP support
      --depinstall       Install all dependencies
      --fetch            Fetch all the things we need
```
If no custom `site-local` is provided (i.e. no `--blobs` specified), the builder will create empty one in which the kernel image and initramfs are going to be placed after compilation. 
Please also note that when using default config, builder script assumes that it is being run under in `/tmp`!

## Considerations when creating custom configuration files (coreboot)
In general providing custom configs should not introduce and unexpected errors during the compilation (does **not** apply to the rom image being functional!), as long as the following configuration options are enabled:
- `CONFIG_SMM_PAYLOAD_INTERFACE`: this tells coreboot to enable SMM payload driver.
- `CONFIG_SMM_PAYLOAD_INTERFACE_PAYLOAD_MM`: enables MM payload interface, needed by Linux to preform post-boot unlock/lock.
- `CONFIG_PAYLOAD_LINUXBOOT`: enables LinuxBoot
- `CONFIG_LINUXBOOT_KERNEL_PATH="/tmp/coreboot-git/site-local/Image"`: the builder tool compiles patched Linux kernel separetely and places it under `site-local/Image`, without this path being specified, coreboot would compile unpatched Linux kernel during the its own compile process.
- `CONFIG_LINUXBOOT_INITRAMFS_PATH="/tmp/coreboot-git/site-local/initramfs_u-root.cpio"`: similarly as with the kernel, initramfs is build separately from coreboot by the builder.
- `CONFIG_DEBUG_SMI`: enables more verbose boot logs with regards to SMIs.

## Considerations when creating custom configuration files (Linux)
There are two configs provided: one with `CONFIG_SMP` enabled and one without. LinuxBoot traditionally does not come with SMP enabled (as there is no need real for SMP support), but
it can be enabled if there is a need for sending SMIs in parallel from different CPUs.
In general, regardless of SMP, the following options should stay enabled in a custom config:
 - `CONFIG_GOOGLE_FIRMWARE`: enables `drivers/firmware/google` module.
 - `CONFIG_GOOGLE_COREBOOT_TABLE`: enables interface for accessing coreboot table.
 - `CONFIG_GOOGLE_FRAMEBUFFER_COREBOOT`: enables the kernel to search for a framebuffer in the coreboot table.
 - `CONFIG_GOOGLE_MEMCONSOLE_COREBOOT`: enables the kernel to search for a firmware log in the coreboot table.
 - `CONFIG_SMM_DRIVER`: enables the support for SMM handling in Linux.
 - `CONFIG_DEBUG_KERNEL`: (optional) makes kernel more verbose, i.a. enables more logs when issuing SMIs.

## Example usage - QEMU Q35
In order to build an example of coreboot+LinuxBootSMM, one can use QEMU emulator:
```sh
# Without SMP
cd /tmp
linuxbootsmm-builder --fetch --config q35
cd coreboot-git/
qemu-system-x86_64 -bios build/coreboot.rom -M q35 -serial stdio

# With SMP
cd /tmp
linuxbootsmm-builder --fetch --config q35-smp
cd coreboot-git/
qemu-system-x86_64 -bios build/coreboot.rom -M q35 -serial stdio -smp NUM_CPUS
```

