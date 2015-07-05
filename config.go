package main

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/opencontainers/runc/libcontainer/configs"
)

const defaultMountFlags = syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

// expected format for bindmounts: /source/to/mount:/dest/to/mount[:ro|rw]
func loadConfig(uid, rootfs, hostname string, bindMounts []string) (*configs.Config, error) {
	var config = &configs.Config{
		Rootfs:            rootfs,
		ParentDeathSignal: int(syscall.SIGKILL),
		Capabilities: []string{
			"CHOWN",
			"DAC_OVERRIDE",
			"FSETID",
			"FOWNER",
			"MKNOD",
			"NET_RAW",
			"SETGID",
			"SETUID",
			"SETFCAP",
			"SETPCAP",
			"NET_BIND_SERVICE",
			"SYS_CHROOT",
			"KILL",
			"AUDIT_WRITE",
		},
		Namespaces: configs.Namespaces([]configs.Namespace{
			{Type: configs.NEWNS},
			{Type: configs.NEWUTS},
			{Type: configs.NEWIPC},
			{Type: configs.NEWPID},
		}),
		Cgroups: &configs.Cgroup{
			Name:            uid,
			Parent:          "psdock",
			AllowAllDevices: false,
			AllowedDevices:  configs.DefaultAllowedDevices,
		},
		Hostname: hostname,
		Devices:  configs.DefaultAutoCreatedDevices,
		MaskPaths: []string{
			"/proc/kcore",
		},
		ReadonlyPaths: []string{
			"/proc/sys", "/proc/sysrq-trigger", "/proc/irq", "/proc/bus",
		},
		Mounts: []*configs.Mount{
			{
				Source:      "proc",
				Destination: "/proc",
				Device:      "proc",
				Flags:       defaultMountFlags,
			},
			{
				Source:      "tmpfs",
				Destination: "/dev",
				Device:      "tmpfs",
				Flags:       syscall.MS_NOSUID | syscall.MS_STRICTATIME,
				Data:        "mode=755",
			},
			{
				Source:      "devpts",
				Destination: "/dev/pts",
				Device:      "devpts",
				Flags:       syscall.MS_NOSUID | syscall.MS_NOEXEC,
				Data:        "newinstance,ptmxmode=0666,mode=0620,gid=5",
			},
			{
				Device:      "tmpfs",
				Source:      "shm",
				Destination: "/dev/shm",
				Data:        "mode=1777,size=65536k",
				Flags:       defaultMountFlags,
			},
			{
				Source:      "mqueue",
				Destination: "/dev/mqueue",
				Device:      "mqueue",
				Flags:       defaultMountFlags,
			},
			{
				Source:      "sysfs",
				Destination: "/sys",
				Device:      "sysfs",
				Flags:       defaultMountFlags | syscall.MS_RDONLY,
			},
			{
				Source:      "/etc/resolv.conf",
				Destination: "/etc/resolv.conf",
				Device:      "bind",
				Flags:       syscall.MS_BIND | syscall.MS_REC | syscall.MS_RDONLY,
			},
		},
		Rlimits: []configs.Rlimit{
			{
				Type: syscall.RLIMIT_NOFILE,
				Hard: uint64(4096),
				Soft: uint64(4096),
			},
		},
	}

	//abb bind mounts if any
	for _, rawBind := range bindMounts {
		mount := &configs.Mount{
			Device: "bind",
			Flags:  syscall.MS_BIND | syscall.MS_REC,
		}
		parts := strings.SplitN(rawBind, ":", 3)
		switch len(parts) {
		default:
			return nil, fmt.Errorf("invalid bind mount %s", rawBind)
		case 2:
			mount.Source, mount.Destination = parts[0], parts[1]
		case 3:
			mount.Source, mount.Destination = parts[0], parts[1]
			switch parts[2] {
			case "ro":
				mount.Flags |= syscall.MS_RDONLY
			case "rw":
			default:
				return nil, fmt.Errorf("invalid bind mount mode %s", parts[2])
			}
		}
		config.Mounts = append(config.Mounts, mount)
	}

	return config, nil
}
