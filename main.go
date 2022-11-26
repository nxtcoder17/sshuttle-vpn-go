package main

import (
  "embed"
  "errors"
  "flag"
  "fmt"
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

  runServerProxy := fmt.Sprintf("kubectl run %s --namespace %s --image=nxtcoder17/alpine.python3:nonroot --restart=Never -- sh -c 'tail -f /dev/null'", name, namespace)
  b, err := exec.Command("bash", "-c", runServerProxy).CombinedOutput()

  fmt.Printf("(kubectl): %s", b)
  if err != nil {
    fmt.Printf("trying to use existing pod, in case it already exists\n")
    // fmt.Println("err occurred while creating pod:", err)
  }

  dir := path.Join(os.TempDir(), "sshuttle-vpn-go")
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

  args := make([]string, 0, len(flag.Args())*2)
  args = append(args, "--dns", "-r", fmt.Sprintf("%s/%s", namespace, name), "-e", fmt.Sprintf("%v/k8s-helper.sh", dir))
  args = append(args, ipRanges...)
  c := exec.Command("sshuttle", args...)
  c.Stdin = os.Stdin
  c.Stdout = os.Stdout
  c.Stderr = os.Stderr
  c.Run()
}
