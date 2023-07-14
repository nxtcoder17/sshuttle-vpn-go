package main

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

var (
	//go:embed "lib"
	libDir embed.FS
)

func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func main() {
	var isVerbose bool

	flag.BoolVar(&isVerbose, "verbose", false, "--verbose")
	flag.Parse()

	splits := strings.Split(os.Args[1], "/")
	if len(splits) != 2 {
		panic("first argument must be a namespaced-name in format <namespace>/<name>")
	}

	namespace := splits[0]
	name := splits[1]

	ipRanges := make([]string, 0, len(os.Args))
	if len(os.Args) > 2 {
		ipRanges = os.Args[2:]
	} else {
		ipRanges = append(ipRanges, "0.0.0.0/0")
	}

	fmt.Printf("forwarded IP ranges: %+v\n", ipRanges)

	dir := path.Join(os.TempDir(), "sshuttle-vpn-go")
	kubectlPodScript := path.Join(dir, "ensure-reverse-proxy-pod.sh")

	if !fileExists(kubectlPodScript) {
		if err := os.Mkdir(dir, 0777); err != nil {
			if !errors.Is(err, os.ErrExist) {
				panic(err)
			}
		}

		hBytes, err := libDir.ReadFile("lib/ensure-reverse-proxy-pod.sh")
		if err != nil {
			panic(err)
		}

		if err := os.WriteFile(path.Join(dir, "ensure-reverse-proxy-pod.sh"), hBytes, 0744); err != nil {
			panic(err)
		}
	}

	cmd := exec.Command("bash", kubectlPodScript)
	outputBuff := new(bytes.Buffer)
	cmd.Stdout = outputBuff
	cmd.Stderr = nil
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("POD_NAME=%s", name),
		fmt.Sprintf("NAMESPACE=%s", namespace),
		fmt.Sprintf("KUBECONFIG=%s", os.Getenv("KUBECONFIG")),
	)

	if err := cmd.Run(); err != nil {
		fmt.Println("err occurred while creating pod:", err)
	}

	fmt.Printf("(kubectl): %s", outputBuff.String())

	k8sHelperPath := path.Join(dir, "k8s-helper.sh")
	if !fileExists(k8sHelperPath) {
		if err := os.Mkdir(dir, 0777); err != nil {
			if !errors.Is(err, os.ErrExist) {
				panic(err)
			}
		}

		hBytes, err := libDir.ReadFile("lib/sshuttle-k8s-helper.sh")
		if err != nil {
			panic(err)
		}

		if err := os.WriteFile(path.Join(dir, "k8s-helper.sh"), hBytes, 0744); err != nil {
			panic(err)
		}
	}

	args := make([]string, 0, len(os.Args)*2)
	if isVerbose {
		args = append(args, "-vv")
	}
	args = append(args, "--dns", "-r", fmt.Sprintf("%s/%s", namespace, name), "-l", "127.0.0.1:12301", "-e", fmt.Sprintf("%v/k8s-helper.sh", dir))
	args = append(args, ipRanges...)
	c := exec.Command("sshuttle", args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
}
