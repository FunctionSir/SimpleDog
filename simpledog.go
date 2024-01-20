/*
 * @Author: FunctionSir
 * @License: AGPLv3
 * @Date: 2024-01-17 14:58:08
 * @LastEditTime: 2024-01-20 21:36:23
 * @LastEditors: FunctionSir
 * @Description: -
 * @FilePath: /SimpleDog/simpledog.go
 */

package main

import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-ini/ini"
)

const VER string = "0.1.0-alpha"
const CODENAME = "NunotabaShinobu"

var GuidePath string = "/dev/null"
var Listen string = "127.0.0.1:2180"
var Quiet bool = false
var NoTime bool = false
var DogMsgs []string = nil

func outln(str string) {
	if Quiet {
		return
	}
	if !NoTime {
		str = time.Now().String() + " " + str
	}
	println(str)
}

func err_handler(err error) bool {
	if Quiet {
		return err != nil
	}
	if err != nil {
		bark("[E] " + err.Error())
	}
	return err != nil
}

func read_lines(name string) []string {
	var r = []string{}
	f, e := os.Open(name)
	err_handler(e)
	defer func() {
		e := f.Close()
		c := 0
		for err_handler(e) && c <= 8 {
			e = f.Close()
			c++
		}
	}()
	fileScanner := bufio.NewScanner(f)
	for fileScanner.Scan() {
		r = append(r, fileScanner.Text())
	}
	return r
}

func args_parser() {
	args := os.Args
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-g", "--guide-file":
			GuidePath = args[i+1]
			i++
		case "-l", "--listen":
			Listen = args[i+1]
			i++
		case "-q", "--quiet":
			Quiet = true
		case "-n", "--no-time":
			NoTime = true
		}
	}
}

func bark(str string) {
	DogMsgs = append(DogMsgs, str)
}

func hello() {
	if Quiet {
		return
	}
	println("SimpleDog Watchdog Software")
	println("Version: " + VER + " (" + CODENAME + ")")
	println("This is a liber software under AGPLv3")
}

func watchdog(confPath string) {
	var triggerExecGap int = 10 * 1000
	var triggerWanted int = 0
	var triggerMode string = "NEQ"
	confFile, err := ini.Load(confPath)
	if err_handler(err) {
		bark("[E] A watch dog failed to start because can't load conf file properly.")
		bark("[I] Error caused by conf file \"" + confPath + "\".")
		return
	}
	general, err := confFile.GetSection("General")
	if err_handler(err) {
		bark("[E] A watch dog failed to start because it does not have section \"General\".")
		bark("[I] Error caused by conf file \"" + confPath + "\".")
		return
	}
	trigger, err := confFile.GetSection("Trigger")
	if err_handler(err) {

		bark("[E] A watch dog failed to start because it does not have section \"Trigger\".")
		bark("[I] Error caused by conf file \"" + confPath + "\".")
		return
	}
	action, err := confFile.GetSection("Action")
	if err_handler(err) {
		bark("[E] A watch dog failed to start because it does not have section \"Action\".")
		bark("[I] Error caused by conf file \"" + confPath + "\".")
		return
	}
	confVer, err := general.GetKey("Version")
	if err_handler(err) {
		bark("[E] A watch dog failed to start because section \"General\" does not have key \"Version\".")
		bark("[I] Error caused by conf file \"" + confPath + "\".")
		return
	}
	if confVer.String() != "0.1" {
		bark("[E] A watch dog failed to start because this version of conf file is not supported.")
		bark("[I] Error caused by conf file \"" + confPath + "\".")
		return
	}
	triggerExec, err := trigger.GetKey("Exec")
	if err_handler(err) {
		bark("[E] A watch dog failed to start because section \"Trigger\" does not have key \"Exec\".")
		bark("[I] Error caused by conf file \"" + confPath + "\".")
		return
	}
	actionExec, err := action.GetKey("Exec")
	if err_handler(err) {
		bark("[E] A watch dog failed to start because section \"Action\" does not have key \"Exec\".")
		bark("[I] Error caused by conf file \"" + confPath + "\".")
		return
	}
	if trigger.HasKey("Gap") {
		tmpKey, _ := trigger.GetKey("Gap")
		tmpStr := tmpKey.String()
		triggerExecGap, err = strconv.Atoi(tmpStr)
	}
	if trigger.HasKey("Wanted") {
		tmpKey, _ := trigger.GetKey("Wanted")
		tmpStr := tmpKey.String()
		triggerWanted, err = strconv.Atoi(tmpStr)
	}
	if trigger.HasKey("Mode") {
		tmpKey, _ := trigger.GetKey("Mode")
		triggerMode = tmpKey.String()
	}
	splitedTriggerExec := strings.Split(triggerExec.String(), " ")
	splitedActionExec := strings.Split(actionExec.String(), " ")
	for {
		cmd := exec.Command(splitedTriggerExec[0], splitedTriggerExec[1:]...)
		cmd.Run()
		if (cmd.ProcessState.ExitCode() == triggerWanted && triggerMode == "EQ") ||
			(cmd.ProcessState.ExitCode() != triggerWanted && triggerMode == "NEQ") {
			bark("[I] Ran \"" + triggerExec.String() + "\" and got exit code \"" + strconv.Itoa(cmd.ProcessState.ExitCode()) + "\"!")
			bark("[B] Dog defined by \"" + confPath + "\" barking!")
			cmd := exec.Command(splitedActionExec[0], splitedActionExec[1:]...)
			cmd.Run()
		}
		time.Sleep(time.Duration(triggerExecGap) * time.Millisecond)
	}
}

func main() {
	args_parser()
	hello()
	guideFile, err := os.ReadFile(GuidePath)
	if err_handler(err) {
		panic(err)
	}
	guideLines := read_lines(string(guideFile))
	for i := range guideLines {
		if guideLines[i][0] != '#' {
			go watchdog(guideLines[i])
		}
	}
	for {
		if len(DogMsgs) > 0 {
			outln(DogMsgs[0])
			DogMsgs = DogMsgs[1:]
		}
	}
}
