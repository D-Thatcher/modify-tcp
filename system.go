package main

import (
	"fmt"
	"os/exec"
)

func changeUFW(enable bool) {
	app := "/usr/sbin/ufw"

	arg0 := "disable"
	if enable {
		arg0 = "enable"
	}
	cmd := exec.Command(app, arg0)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	printLn(string(stdout))
}

func changeIpTableRule(add bool, iface string, queueNum string) {
	app := "/sbin/iptables"

	arg0 := "-A"
	if !add {
		arg0 = "-D"
	}
	arg1 := "INPUT"
	arg2 := "-i"
	arg3 := iface
	arg4 := "-j"
	arg5 := "NFQUEUE"
	arg6 := "--queue-num"
	arg7 := queueNum

	cmd := exec.Command(app, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	printLn(string(stdout))
}
