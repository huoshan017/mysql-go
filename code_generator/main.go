package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	mysql_base "github.com/huoshan017/mysql-go/base"
	mysql_generate "github.com/huoshan017/mysql-go/generate"
)

//var protoc_root = os.Getenv("GOPATH") + "/mysql-go/_external/"

var protoc_dest_map = map[string]string{
	"windows": "windows/protoc.exe",
	"linux":   "linux/protoc",
	"darwin":  "darwin/protoc",
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "args num not enough\n")
		return
	}

	var arg_config_file, arg_dest_path, arg_protoc_path *string
	// 代碼生成配置路徑
	arg_config_file = flag.String("c", "", "config file path")
	// 目標代碼生成目錄
	arg_dest_path = flag.String("d", "", "dest source path")
	// protoc根目錄
	arg_protoc_path = flag.String("p", "", "protoc file root path")
	flag.Parse()

	var config_path string
	if nil != arg_config_file && *arg_config_file != "" {
		config_path = *arg_config_file
		fmt.Fprintf(os.Stdout, "config file path %v\n", config_path)
	} else {
		fmt.Fprintf(os.Stderr, "not found config file arg\n")
		return
	}

	var dest_path string
	if nil != arg_dest_path && *arg_dest_path != "" {
		dest_path = *arg_dest_path
		fmt.Fprintf(os.Stdout, "dest path %v\n", dest_path)
	} else {
		fmt.Fprintf(os.Stderr, "not found dest path arg\n")
		return
	}

	var protoc_path string
	if nil != arg_protoc_path && *arg_protoc_path != "" {
		go_os := runtime.GOOS //os.Getenv("GOOS")
		protoc_path = *arg_protoc_path + "/" + protoc_dest_map[go_os]
	} else {
		fmt.Fprintf(os.Stderr, "not found dest protoc file root path\n")
		return
	}

	fmt.Fprintf(os.Stdout, "protoc path %v\n", protoc_path)

	var config_loader mysql_generate.ConfigLoader
	if !config_loader.Load(config_path) {
		return
	}

	if !config_loader.Generate(dest_path) {
		return
	}

	fmt.Fprintf(os.Stdout, "generated source\n")

	proto_dest_path, config_file := path.Split(config_path)
	proto_dest_path += ".proto/"
	mysql_base.CreateDirs(proto_dest_path)
	proto_file := strings.Replace(config_file, "json", "proto", -1)

	fmt.Fprintf(os.Stdout, "proto_dest_path: %v    proto_file: %v\n", proto_dest_path, proto_file)

	if !config_loader.GenerateFieldStructsProto(proto_dest_path + proto_file) {
		fmt.Fprintf(os.Stderr, "generate proto file failed\n")
		return
	}

	fmt.Fprintf(os.Stdout, "generated proto\n")

	cmd := exec.Command(protoc_path, "--go_out", dest_path /*+"/"+config_loader.DBPkg*/, "--proto_path", proto_dest_path, proto_file)
	var out bytes.Buffer

	fmt.Fprintf(os.Stdout, "--go_out=%v  --proto_path=%v  proto_file=%v\n", dest_path /*+"/"+config_loader.DBPkg*/, proto_dest_path, proto_file)

	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cmd run err: %v\n", err.Error())
		return
	}
	fmt.Printf("%s", out.String())

	if !config_loader.GenerateInitFunc(dest_path) {
		fmt.Fprintf(os.Stderr, "generate init func failed\n")
		return
	}

	fmt.Fprintf(os.Stdout, "generated init funcs\ngenerated all\n")
}
