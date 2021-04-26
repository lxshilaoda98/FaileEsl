package models

import (
	"bytes"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"os"
	"os/exec"
	"runtime"
)

//添加异常ip 到防火墙中
func addfw(ip string) {
	sysType := runtime.GOOS
	fmt.Println("当前系统：", sysType)
	if sysType == "linux" {
		// LINUX系统
		e, f, f1 := exec_shell(fmt.Sprintf("fail2ban-client set freeswitch banip %v", ip))
		if e != nil {
			fmt.Println("执行命令错误..>", e)
		} else {
			fmt.Println("成功执行命令.>", f, "...>", f1)
		}
	}
	if sysType == "windows" {
		// windows系统
		f, e := exec_cmd(fmt.Sprintf("netsh advfirewall firewall add rule name =\"des_%v\" remoteip=\"%v\" dir=in action=block", ip, ip))
		if e != nil {
			fmt.Println("执行命令错误..>", e)
		} else {
			fmt.Println("成功执行命令.>", f)
		}
	}

}

//运行bash脚本
func exec_shell(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}

//运行cmd脚本
func exec_cmd(command string) (str string, err error) {
	var whoami []byte

	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	//cmd := exec.Command("/bin/bash", "", s)
	fmt.Println("运行的cmd..>", command)
	cmd := exec.Command("cmd", "/C", command)
	if whoami, err = cmd.Output(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// 指定参数后过滤换行符
	ret, err := simplifiedchinese.GBK.NewDecoder().String(string(whoami))
	fmt.Println(ret)
	return ret, err
}
