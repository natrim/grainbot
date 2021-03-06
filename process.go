package main

/**
DUPLICATED from goagain + modded to use net.Conn
It enables process restart without droping net.Conn
*/

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

// SignalChan accepts SIGINT, SIGTERM, SIGQUIT resp. SIGUSR2 signals for quit resp. restart
var SignalChan chan os.Signal

// Just alias some syscall signals
const (
	SIGINT  = syscall.SIGINT
	SIGTERM = syscall.SIGTERM
	SIGQUIT = syscall.SIGQUIT
	SIGUSR2 = syscall.SIGUSR2
)

// Getpid return's current process pid
func Getpid() int {
	return syscall.Getpid()
}

// Quit terminates bot
func Quit() {
	SignalChan <- syscall.SIGQUIT
}

// Restart restart's bot
func Restart() {
	SignalChan <- syscall.SIGUSR2
}

// Kill process specified in the environment with the signal specified in the
// environment; default to SIGQUIT.
func killParentAfterRestart() error {
	var pid int
	_, err := fmt.Sscan(os.Getenv("RESTART_PID"), &pid)
	if io.EOF == err {
		_, err = fmt.Sscan(os.Getenv("RESTART_PPID"), &pid)
	}
	if nil != err {
		return err
	}

	log.Infoln("Sending parent GRAIN (pid: ", pid, ") QUIT signal")
	return syscall.Kill(pid, syscall.SIGQUIT)
}

// Reconstruct a net.Conn from a file descriptior and name specified in the
// environment.  Deal with Go's insistence on dup(2)ing file descriptors.
func findSocket() (l net.Conn, err error) {
	var fd uintptr
	if _, err = fmt.Sscan(os.Getenv("RESTART_FD"), &fd); nil != err {
		return
	}
	l, err = net.FileConn(os.NewFile(fd, os.Getenv("RESTART_NAME")))
	if nil != err {
		return
	}
	switch l.(type) {
	case *net.TCPConn, *net.UnixConn:
	default:
		err = fmt.Errorf(
			"file descriptor is %T not *net.TCPConn or *net.UnixConn",
			l,
		)
		return
	}
	if err = syscall.Close(int(fd)); nil != err {
		return
	}
	return
}

// WaitOnSignals will block process until receives quit or restart signal
func (bot *Bot) WaitOnSignals(l net.Conn) error {
	SignalChan = make(chan os.Signal, 2)
	signal.Notify(
		SignalChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGUSR2,
	)
	forked := false
	for {
		sig := <-SignalChan
		switch sig {
		case syscall.SIGTERM:
			return nil //just unblock
		case syscall.SIGINT:
			return nil //just unblock
		case syscall.SIGQUIT:
			return nil //just unblock
		case syscall.SIGUSR2:
			if forked { //druhy a dalsi jen ukonci
				return nil
			}
			forked = true

			if err := bot.beforeFork(); nil != err {
				log.Errorln("BeforeForkError:", err)
			}

			if err := forkAndExec(l); nil != err { //pri prvnim signalu udelej fork, vrat hodnotu jen kdyz bude chyba
				return err
			}
		}
	}
}

// Fork and exec this same image without dropping the net.Conn.
func forkAndExec(l net.Conn) error {
	argv0, err := lookPath()
	if nil != err {
		return err
	}
	wd, err := os.Getwd()
	if nil != err {
		return err
	}
	fd, err := setEnvs(l)
	if nil != err {
		return err
	}
	if err := os.Setenv("RESTART_PID", ""); nil != err {
		return err
	}
	if err := os.Setenv(
		"RESTART_PPID",
		fmt.Sprint(syscall.Getpid()),
	); nil != err {
		return err
	}
	files := make([]*os.File, fd+1)
	files[syscall.Stdin] = os.Stdin
	files[syscall.Stdout] = os.Stdout
	files[syscall.Stderr] = os.Stderr
	addr := l.RemoteAddr()
	files[fd] = os.NewFile(
		fd,
		fmt.Sprintf("%s:%s->", addr.Network(), addr.String()),
	)
	p, err := os.StartProcess(argv0, os.Args, &os.ProcAttr{
		Dir:   wd,
		Env:   os.Environ(),
		Files: files,
		Sys:   &syscall.SysProcAttr{},
	})
	if nil != err {
		return err
	}
	log.Infoln("Spawned new GRAIN child (pid: ", p.Pid, ")")
	if err = os.Setenv("RESTART_PID", fmt.Sprint(p.Pid)); nil != err {
		return err
	}
	return nil
}

func lookPath() (argv0 string, err error) {
	argv0, err = exec.LookPath(os.Args[0])
	if nil != err {
		return
	}
	if _, err = os.Stat(argv0); nil != err {
		return
	}
	return
}

func setEnvs(l net.Conn) (fd uintptr, err error) {
	v := reflect.ValueOf(l).Elem().FieldByName("fd").Elem()
	fd = uintptr(v.FieldByName("sysfd").Int())
	_, _, e1 := syscall.Syscall(syscall.SYS_FCNTL, fd, syscall.F_SETFD, 0)
	if 0 != e1 {
		err = e1
		return
	}
	if err = os.Setenv("RESTART_FD", fmt.Sprint(fd)); nil != err {
		return
	}
	addr := l.RemoteAddr()
	if err = os.Setenv(
		"RESTART_NAME",
		fmt.Sprintf("%s:%s->", addr.Network(), addr.String()),
	); nil != err {
		return
	}
	return
}
