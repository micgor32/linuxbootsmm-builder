# LinuxbootSMM PoC Builder
A simple script to build coreboot image with LinuxbootSMM.

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
linuxbootsmm-builder --deps
```
Then run:
```sh
linuxbootsmm-builder --fetch
```
Or alternatively if you already downloaded and unpacked the coreboot source:
```sh
linuxbootsmm-builder --build
```
