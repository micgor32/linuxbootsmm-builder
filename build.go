// Copyright 2012-2018 the u-root Authors,
// Copyright 2025 9elements. 
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"path/filepath"
	"github.com/go-ini/ini"
	
	cp "github.com/otiai10/copy"
	flag "github.com/spf13/pflag"
)

var (
	configTxt = `loglevel=1
	init=/init
rootwait
`
	deps    	  = flag.Bool("depinstall", false, "Install all dependencies")
	fetch         = flag.Bool("fetch", false, "Fetch all the things we need")
	build 		  = flag.Bool("build", false, "Only build the image (you have to provide Linux image and initramfs)")
	configPath 	  = flag.String("config", "default", "Path to config file for coreboot") 
	smpEnabled    = flag.Bool("smp", false, "Compile Linux with SMP support")
	corebootVer   = "git" // hardcoded for now, we only need it to avoid situation when someone have "coreboot" dir already, we do not want to overwrite it
	blobsPath     = flag.String("blobs", "no", "Path to the custom site-local directory for coreboot")
	threads       = runtime.NumCPU() + 4 // Number of threads to use when calling make.
	// based on coreboot docs requirements
	packageListDebian   = []string{ 
		"bison",
		"git",
		"golang",
		"build-essential",
		"curl",
		"gnat",
		"flex",
		"gnat",
		"libncurses-dev",
		"libssl-dev",
		"zlib1g-dev",
		"pkgconf",
		"qemu-system-x86",
	}
	packageListArch = []string{
		"base-devel",
		"curl",
		"git",
		"gcc-ada",
		"ncurses",
		"zlib",
		"qemu-full",
	}
	packageListRedhat = []string{
		"git",
		"make",
		"gcc-gnat",
		"flex",
		"bison",
		"xz",
		"bzip2",
		"gcc",
		"g++",
		"ncurses-devel",
		"wget",
		"zlib-devel",
		"patch",
		"qemu",
	}
)

func getGitVersion() error {
	var args = []string{"clone", "https://github.com/micgor32/coreboot.git", "coreboot-" + corebootVer}
	fmt.Printf("-------- Getting the coreboot via git %v\n", args)
	cmd := exec.Command("git", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("didn't clone coreboot %v", err)
		return err
	}
	return nil
}

func corebootGet() error {
	getGitVersion()
	
	cmd := exec.Command("make", "-j"+strconv.Itoa(threads), "crossgcc-i386", "CPUS=$(nproc)")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "coreboot-" + corebootVer
	if err := cmd.Run(); err != nil {
		fmt.Printf("toolchain build failed %v", err)
		return err
	}

	if *configPath == "default" {
		var config = []string{"-O", "defconfig", "https://raw.githubusercontent.com/micgor32/linuxbootsmm-builder/refs/heads/master/defconfig-raptorlake"}
		cmd = exec.Command("wget", config...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		cmd.Dir = "coreboot-" + corebootVer
		if err := cmd.Run(); err != nil {
			fmt.Printf("obtaining config failed %v", err)
			return err
		}
	} else if *configPath == "q35" {
		var config = []string{"-O", "defconfig", "https://raw.githubusercontent.com/micgor32/linuxbootsmm-builder/refs/heads/master/defconfig-qemu"}
		cmd = exec.Command("wget", config...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		cmd.Dir = "coreboot-" + corebootVer
		if err := cmd.Run(); err != nil {
			fmt.Printf("obtaining config failed %v", err)
			return err
		}
	} else {
		os.Link(*configPath, "coreboot-" + corebootVer + "/defconfig")
	}

	if *blobsPath != "no" {
		if err := cp.Copy(*blobsPath, "coreboot-" + corebootVer + "/site-local"); err != nil {
			fmt.Printf("copying custom site-local failed %v", err)
			return err
		}
	} else {
		newpath := filepath.Join("coreboot-" + corebootVer, "site-local")
		if err := os.MkdirAll(newpath, os.ModePerm); err != nil {
			fmt.Printf("error creating site-local %v", err)
			return err
		}
	}

	cmd = exec.Command("make", "defconfig", "KBUILD_DEFCONFIG=defconfig")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "coreboot-" + corebootVer
	if err := cmd.Run(); err != nil {
		fmt.Printf("generating config failed %v", err)
		return err
	}
	
	return nil
}

func patchKernel() error {
	// TODO: consider also checking the patch correctness before applying (i.e. run git apply --check *path_to_patch*).
	var patchParsers = []string{"https://raw.githubusercontent.com/9elements/LinuxBootSMM/refs/heads/main/poc/patches/patch-0001-drivers-firmware-smm-parsing-SMM-related-informations-from-coreboot-table.diff"}
	cmd := exec.Command("wget", patchParsers...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "linux-smm"
	if err := cmd.Run(); err != nil {
		fmt.Printf("obtaining patch failed %v", err)
		return err
	}

	var patchLoader = []string{"https://raw.githubusercontent.com/9elements/LinuxBootSMM/refs/heads/main/poc/patches/patch-0002-loader-for-linux-owned-smi-handler.diff"}
	cmd = exec.Command("wget", patchLoader...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "linux-smm"
	if err := cmd.Run(); err != nil {
		fmt.Printf("obtaining patch failed %v", err)
		return err
	}

	var applyParsers = []string{"am", "patch-0001-drivers-firmware-smm-parsing-SMM-related-informations-from-coreboot-table.diff"}
	cmd = exec.Command("git", applyParsers...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "linux-smm"
	if err := cmd.Run(); err != nil {
		fmt.Printf("applying patch failed %v", err)
		return err
	}

	var applyLoader = []string{"am", "patch-0002-loader-for-linux-owned-smi-handler.diff"}
	cmd = exec.Command("git", applyLoader...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "linux-smm"
	if err := cmd.Run(); err != nil {
		fmt.Printf("applying patch failed %v", err)
		return err
	}
	
	return nil
}

func getKernel() error {
	var args = []string{"clone", "https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git", "linux-smm"}
	fmt.Printf("-------- Getting the kernel via git %v\n", args)
	cmd := exec.Command("git", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("didn't cloned the kernel %v", err)
		//return err
	}

	patchKernel();

	if *smpEnabled {
		var config = []string{"-O", ".config", "https://raw.githubusercontent.com/micgor32/linuxbootsmm-builder/refs/heads/master/defconfig-linux-smp"}
		cmd = exec.Command("wget", config...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		cmd.Dir = "linux-smm"
		if err := cmd.Run(); err != nil {
			fmt.Printf("obtaining config failed %v", err)
			return err
		}
	} else {
		var config = []string{"-O", ".config", "https://raw.githubusercontent.com/micgor32/linuxbootsmm-builder/refs/heads/master/defconfig-linux"}
		cmd = exec.Command("wget", config...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		cmd.Dir = "linux-smm"
		if err := cmd.Run(); err != nil {
			fmt.Printf("obtaining config failed %v", err)
			return err
		}
	}

	cmd = exec.Command("make", "olddefconfig")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "linux-smm"
	if err := cmd.Run(); err != nil {
		fmt.Printf("generating config failed %v", err)
		return err
	}

	return nil
}

func kernelBuild() error {
	// if err := getKernel(); err != nil {
	// 	fmt.Printf("didn't cloned the kernel %v", err)
	// 	return err
	// }
	getKernel() // do not bound it with error handling, in case of sources being already there the build fails, FIXME: add proper handling of this case
	
	cmd := exec.Command("make", "-j"+strconv.Itoa(threads))
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "linux-smm"
	if err := cmd.Run(); err != nil {
		fmt.Printf("error when builing the kernel %v", err)
		return err
	}

	if err := cp.Copy("linux-smm/arch/x86/boot/bzImage", "coreboot-" + corebootVer + "/site-local/Image"); err != nil {
		fmt.Printf("error copying the kernel image %v", err)
		return err
	}
	return nil
}

func initramfsGen() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cmd := exec.Command(pwd + "/uroot/u-root","-build", "bb", "-initcmd", "init", "-uinitcmd", "boot", "-defaultsh", "gosh", "-o", "initramfs_u-root.cpio", "core", "boot", "coreboot-app")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "uroot"
	if err := cmd.Run(); err != nil {
		fmt.Printf("error when builing the initramfs %v", err)
		return err
	}

	cmd = exec.Command("xz" ,"--check", "crc32", "--lzma2=dict=512KiB", "initramfs_u-root.cpio")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "uroot"
	if err := cmd.Run(); err != nil {
		fmt.Printf("error when builing the initramfs %v", err)
		return err
	}
	
	if err := cp.Copy("uroot/initramfs_u-root.cpio.xz", "coreboot-" + corebootVer + "/site-local/initramfs_u-root.cpio.xz"); err != nil {
		fmt.Printf("error copying the initramfs %v", err)
		return err
	}

	return nil
}

func buildCoreboot() error {
	// Let's check whether the config is there
	if _, err := os.Stat("coreboot-" + corebootVer + "/.config"); err != nil {
		return err
	}

	kernelBuild()
	initramfsGen()

	cmd := exec.Command("make", "-j"+strconv.Itoa(threads))
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "coreboot-" + corebootVer
	err := cmd.Run()
	if err != nil {
		return err
	}
	
	if _, err := os.Stat("coreboot-" + corebootVer + "/build/coreboot.rom"); err != nil {
		return err
	}
	fmt.Printf("build/coreboot.rom created")
	return nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v %v: %v", name, args, err)
	}
	return nil
}

func check() error {
	if os.Getenv("GOPATH") == "" {
		return fmt.Errorf("You have to set GOPATH.")
	}
	return nil
}

// Ugly, but fast way to deal with getting u-root up to run
func urootInstall() error {
	var args = []string{"clone", "https://github.com/u-root/u-root", "uroot"}
	cmd := exec.Command("git", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("didn't cloned the kernel %v", err)
		return err
	}

	cmd = exec.Command("go", "build")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "uroot"
	err := cmd.Run()
	if err != nil {
		return err
	}
	 
	return nil
}

func pacmaninstall() error {
	missing := []string{}
	for _, packageName := range packageListArch {
		cmd := exec.Command("pacman", "-Ql", packageName)
		if err := cmd.Run(); err != nil {
			missing = append(missing, packageName)
		}
	}

	if len(missing) == 0 {
		fmt.Println("No missing dependencies to install")
		return nil
	}

	fmt.Printf("Using pacman to get %v\n", missing)
	get := []string{"pacman", "-S", "--noconfirm"}
	get = append(get, missing...)
	cmd := exec.Command("sudo", get...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}


func dnfinstall() error {
	missing := []string{}
	for _, packageName := range packageListRedhat {
		cmd := exec.Command("dnf", "info", packageName)
		if err := cmd.Run(); err != nil {
			missing = append(missing, packageName)
		}
	}

	if len(missing) == 0 {
		fmt.Println("No missing dependencies to install")
		return nil
	}

	fmt.Printf("Using dnf to get %v\n", missing)
	get := []string{"dnf", "-y", "install"}
	get = append(get, missing...)
	cmd := exec.Command("sudo", get...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

func aptget() error {
	missing := []string{}
	for _, packageName := range packageListDebian {
		cmd := exec.Command("dpkg", "-s", packageName)
		if err := cmd.Run(); err != nil {
			missing = append(missing, packageName)
		}
	}

	if len(missing) == 0 {
		fmt.Println("No missing dependencies to install")
		return nil
	}

	fmt.Printf("Using apt-get to get %v\n", missing)
	get := []string{"apt-get", "-y", "install"}
	get = append(get, missing...)
	cmd := exec.Command("sudo", get...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()

}

func depinstall() error {
	// Regardless of the distro, we need u-root
	urootInstall()

	cfg, err := ini.Load("/etc/os-release")
    if err != nil {
        log.Fatal("Fail to read file: ", err)
    }

    ConfigParams := make(map[string]string)
    ConfigParams["ID"] = cfg.Section("").Key("ID").String()
	osID := ConfigParams["ID"]

	switch osID {
		case "fedora":
			dnfinstall()
		case "rhel":
			dnfinstall()
		case "debian":
			aptget()
		case "ubuntu":
			aptget()
		case "arch":
			pacmaninstall()
		default:
			log.Fatal("No matching OS found")
	}

	return nil
}

func allFunc() error {
	var cmds = []struct {
		f      func() error
		skip   bool
		ignore bool
		n      string
	}{
		{f: depinstall, skip: !*deps, ignore: false, n: "Install dependencies"},
		{f: corebootGet, skip: *build || !*fetch, ignore: false, n: "Download coreboot"},
		{f: buildCoreboot, skip: *deps, ignore: false, n: "build coreboot"},
	}

	for _, c := range cmds {
		log.Printf("-----> Step %v: ", c.n)
		if c.skip {
			log.Printf("-------> Skip")
			continue
		}
		log.Printf("----------> Start")
		err := c.f()
		if c.ignore {
			log.Printf("----------> Ignore result")
			continue
		}
		if err != nil {
			return fmt.Errorf("%v: %v", c.n, err)
		}
		log.Printf("----------> Finished %v\n", c.n)
	}
	return nil
}

func main() {
	flag.Parse()
	log.Printf("Using patched kernel for LinuxBootSMM\n")
	if err := allFunc(); err != nil {
		log.Fatalf("fail error is : %v", err)
	}
	log.Printf("execution completed successfully\n")
}
