package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	server, err := net.Listen("tcp", os.Getenv("LOCAL_IP"))
	if err != nil {
		fmt.Println("Listen failed: ", err)
		return
	}

	for {
		client, err := server.Accept()
		if err != nil {
			fmt.Println("Accept failed: ", err)
			continue
		}
		remote, err := net.Dial("tcp", os.Getenv("SOCKS5_IP"))
		if err != nil {
			fmt.Println("Connect remote failed: ", err)
			continue
		}
		go process(client, remote)
	}
}

func process(client net.Conn, remote net.Conn) {
	if err := Socks5Auth(client); err != nil {
		fmt.Println("Auth error:", err)
		client.Close()
		return
	}

	if err := RemoteSocks5Auth(remote); err != nil {
		fmt.Println("Remote auth error:", err)
		remote.Close()
		return
	}

	Socks5Forward(client, remote)
}

func Socks5Auth(client net.Conn) (err error) {
	buf := make([]byte, 256)

	// 读取 VER 和 NMETHODS
	n, err := io.ReadFull(client, buf[:2])
	if n != 2 {
		return errors.New("reading header: " + err.Error())
	}

	ver, nMethods := int(buf[0]), int(buf[1])
	if ver != 5 {
		return errors.New("invalid version")
	}

	// 读取 METHODS 列表
	n, err = io.ReadFull(client, buf[:nMethods])
	if n != nMethods {
		return errors.New("reading methods: " + err.Error())
	}

	//用户名密码认证
	n, err = client.Write([]byte{0x05, 0x02})
	if n != 2 || err != nil {
		return errors.New("write rsp: " + err.Error())
	}

	//读取VER USERNAME_LENGTH
	n, err = io.ReadFull(client, buf[:2])
	if n != 2 || err != nil {
		return errors.New("reading auth ver and username length: " + err.Error())
	}
	ver, usernameLength := int(buf[0]), int(buf[1])

	n, err = io.ReadFull(client, buf[:usernameLength])
	if n != usernameLength || err != nil {
		return errors.New("reading username: " + err.Error())
	}
	username := string(buf[:usernameLength])

	n, err = io.ReadFull(client, buf[:1])
	if n != 1 || err != nil {
		return errors.New("reading auth password length: " + err.Error())
	}
	passwordLength := int(buf[0])

	n, err = io.ReadFull(client, buf[:passwordLength])
	if n != passwordLength || err != nil {
		return errors.New("reading username: " + err.Error())
	}
	password := string(buf[:passwordLength])

	status := 0x00
	if username != os.Getenv("USERNAME") || password != os.Getenv("PASSWORD") {
		status = 0x01
	}

	n, err = client.Write([]byte{byte(ver), byte(status)})
	if n != 2 || err != nil {
		return errors.New("write rsp: " + err.Error())
	}

	return nil
}

func RemoteSocks5Auth(client net.Conn) (err error) {
	client.Write([]byte{0x05, 0x01, 0x00})

	buf := make([]byte, 256)

	// 读取 VER 和 NMETHODS
	n, err := io.ReadFull(client, buf[:2])
	if n != 2 {
		return errors.New("reading header: " + err.Error())
	}

	ver, nMethods := int(buf[0]), int(buf[1])
	if ver != 5 {
		return errors.New("invalid version")
	}

	// 读取 METHODS 列表
	n, err = io.ReadFull(client, buf[:nMethods])
	if n != nMethods {
		return errors.New("reading methods: " + err.Error())
	}

	return nil
}

func Socks5Forward(client, target net.Conn) {
	forward := func(src, dest net.Conn) {
		defer src.Close()
		defer dest.Close()
		io.Copy(src, dest)
	}
	go forward(client, target)
	go forward(target, client)
}
