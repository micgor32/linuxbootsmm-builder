# LinuxbootSMM PoC Builder
A simple script to build coreboot image with LinuxBootSMM.

Based on [corebootnerf](https://github.com/linuxboot/corebootnerf), modified to work with newer coreboot release and LinuxBootSMM.

## Prerequisites
Please make sure you have Go >= 1.23, and your GOPATH is set up correctly.

## Usage
Install using go:
```sh
go install github.com/micgor32/linuxbootsmm-builder@latest
```
Check whether you have all dependencies mentioned [here](https://doc.coreboot.org/tutorial/part1.html#step-1-install-tools-and-libraries-needed-for-coreboot) installed, or run:
```sh
linuxbootsmm-builder --deps
```
For the remaining usage options please see:
```sh
Usage of linuxbootsmm-builder:
      --build            Only build the image
      --config string    Path to config file for coreboot (default "default")
      --depinstall       Install all dependencies
      --fetch            Fetch all the things we need
      --version string   Desired version of coreboot (default "24.12")
```
