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
	"github.com/go-ini/ini"

	flag "github.com/spf13/pflag"
)

var (
	configTxt = `loglevel=1
	init=/init
rootwait
`
	deps    	  = flag.Bool("depinstall", false, "Install all dependencies")
	fetch         = flag.Bool("fetch", false, "Fetch all the things we need")
	build 		  = flag.Bool("build", false, "Only build the image")
	kernelVersion = "6.12.12" // will be omitted for now, add later possibility to use modified version with enabled custom driver
	parseVersion  = flag.String("version", "24.12", "Desired version of coreboot") // TODO: add the possibility to use git version and/or the modified version for LinuxbootSMM. for now valid are either release tags or "latest"
	configPath 	  = flag.String("config", "default", "Path to config file for coreboot") 
	corebootVer   = *parseVersion
	workingDir    = ""
	linuxVersion  = "linux-stable"
	homeDir       = ""
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
	}
	packageListArch = []string{
		"base-devel",
		"curl",
		"git",
		"gcc-ada",
		"ncurses",
		"zlib",
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
	}
)

func getGitVersion() error {
	var args = []string{"clone", "https://review.coreboot.org/coreboot", "coreboot-latest"} // hardcode the "latest" for now
	fmt.Printf("-------- Getting the coreboot via git %v\n", args)
	cmd := exec.Command("git", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("didn't clone coreboot %v", err)
		return err
	}
	return nil
}

func corebootPatchConfig_x86_64() error {
	var patch = []string{"https://raw.githubusercontent.com/micgor32/linuxbootsmm-builder/refs/heads/master/defconfig-linuxboot-x86_64.patch"}
	cmd := exec.Command("wget", patch...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "coreboot-" + corebootVer + "/payloads/external/LinuxBoot/x86_64"
	if err := cmd.Run(); err != nil {
		fmt.Printf("obtaining patch failed %v", err)
		return err
	}

	cmd = exec.Command("patch", "-u", "defconfig", "-i", "defconfig-linuxboot-x86_64.patch")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "coreboot-" + corebootVer + "/payloads/external/LinuxBoot/x86_64"
	if err := cmd.Run(); err != nil {
		fmt.Printf("patching failed %v", err)
		return err
	}
	
	return nil
}

func corebootPatchConfig_i386() error {
	var patch = []string{"https://raw.githubusercontent.com/micgor32/linuxbootsmm-builder/refs/heads/master/defconfig-linuxboot-i386.patch"}
	cmd := exec.Command("wget", patch...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "coreboot-" + corebootVer + "/payloads/external/LinuxBoot/i386"
	if err := cmd.Run(); err != nil {
		fmt.Printf("obtaining patch failed %v", err)
		return err
	}

	cmd = exec.Command("patch", "-u", "defconfig", "-i", "defconfig-linuxboot-i386.patch")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "coreboot-" + corebootVer + "/payloads/external/LinuxBoot/i386"
	if err := cmd.Run(); err != nil {
		fmt.Printf("patching failed %v", err)
		return err
	}
	
	return nil
}

// Essentially what we do here is to modify coreboot's makefile
// to use patched kernel. Probably there exists more elegant solution,
// hence, TODO: improve in the future.
func kernelPatch() error {
	var patch = []string{"https://raw.githubusercontent.com/micgor32/linuxbootsmm-builder/refs/heads/master/linux-makefile.patch"}
	cmd := exec.Command("wget", patch...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "coreboot-" + corebootVer + "/payloads/external/LinuxBoot/targets"
	if err := cmd.Run(); err != nil {
		fmt.Printf("obtaining patch failed %v", err)
		return err
	}

	cmd = exec.Command("patch", "-u", "linux.mk", "-i", "linux-makefile.patch")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "coreboot-" + corebootVer + "/payloads/external/LinuxBoot/targets"
	if err := cmd.Run(); err != nil {
		fmt.Printf("patching failed %v", err)
		return err
	}
	
	return nil
	
}

func corebootGet() error {
	if corebootVer == "latest" {
		getGitVersion()
	} else {
		var args = []string{"https://coreboot.org/releases/coreboot-" + corebootVer + ".tar.xz"}
		fmt.Printf("-------- Getting coreboot via wget %v\n", "https://coreboot.org/releases/coreboot-" + corebootVer + ".tar.xz")
		cmd := exec.Command("wget", args...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("didn't wget coreboot %v", err)
			return err
		}
	

		cmd = exec.Command("tar", "xf", "coreboot-" + corebootVer + ".tar.xz")
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
		fmt.Printf("untar failed %v", err)
			return err
		}
	}

	corebootPatchConfig_x86_64()
	corebootPatchConfig_i386()
	kernelPatch()

	cmd := exec.Command("make", "-j"+strconv.Itoa(threads), "crossgcc-i386", "CPUS=$(nproc)")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "coreboot-" + corebootVer
	if err := cmd.Run(); err != nil {
		fmt.Printf("toolchain build failed %v", err)
		return err
	}

	cmd = exec.Command("make", "-C", "payloads/coreinfo", "olddefconfig")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "coreboot-" + corebootVer
	if err := cmd.Run(); err != nil {
		fmt.Printf("build failed %v", err)
		return err
	}

	cmd = exec.Command("make", "-C", "payloads/coreinfo")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "coreboot-" + corebootVer
	if err := cmd.Run(); err != nil {
		fmt.Printf("build failed %v", err)
		return err
	}

	if *configPath == "default" {
		var config = []string{"https://raw.githubusercontent.com/micgor32/linuxbootsmm-builder/refs/heads/master/defconfig"}
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

	cmd = exec.Command("make", "defconfig", "KBUILD_DEFCONFIG=defconfig")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = "coreboot-" + corebootVer
	if err := cmd.Run(); err != nil {
		fmt.Printf("generating config failed %v", err)
		return err
	}
	
	return nil
}

func buildCoreboot() error {
	os.Link("config", "coreboot-" + corebootVer + "/.config")

	cmd := exec.Command("make", "-j"+strconv.Itoa(threads))
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = append(os.Environ(), "ARCH=x86_64")
	cmd.Dir = "coreboot-" + corebootVer
	err := cmd.Run()
	if err != nil {
		// This is absolutely the ugliest solution, but for now
		// when initramfs generation fails in the first run (which is almost certain)
		// we just restart the compilation. TODO: fix this issue properly
		cmd := exec.Command("make", "-j"+strconv.Itoa(threads))
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		cmd.Env = append(os.Environ(), "ARCH=x86_64")
		cmd.Dir = "coreboot-" + corebootVer
		err := cmd.Run()
		if err != nil {
			return err
		}
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
	get := []string{"pacman", "-S", "--noconfim"}
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

	fmt.Printf("Using dng to get %v\n", missing)
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
		{f: check, skip: false, ignore: false, n: "check environment"},
		{f: depinstall, skip: !*deps, ignore: false, n: "Install dependencies"},
		{f: corebootGet, skip: *build || !*fetch, ignore: false, n: "Download coreboot"},
		{f: buildCoreboot, skip: false, ignore: false, n: "build coreboot"},
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
	log.Printf("Using kernel %v\n", kernelVersion)
	if err := allFunc(); err != nil {
		log.Fatalf("fail error is : %v", err)
	}
	log.Printf("execution completed successfully\n")
}
