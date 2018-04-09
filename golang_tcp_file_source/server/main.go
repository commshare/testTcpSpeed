// tut_tcpclient_filereceiver project main.go
// Made by Gilles Van Vlasselaer
// More info about it on www.mrwaggel.be/post/golang-sending-a-file-over-tcp/
/*
http://www.mrwaggel.be/post/golang-transfer-a-file-over-a-tcp-socket/
*/
package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"log"
	"github.com/dustin/go-humanize"
	"math"
)

//Define that the binairy data of the file will be sent 1024 bytes at a time
const BUFFERSIZE = 1024
const (
	ADDR =""
)
func main() {
	/*服务器是接收用的，所以虽然这个拨号，但是这个是服务器*/
	connection, err := net.Dial("tcp", ADDR)
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	fmt.Println("Connected to server, start receiving the file name and file size")
	//Create buffer to read in the name and size of the file
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)
	//Get the filesize
	connection.Read(bufferFileSize)
	//Strip the ':' from the received size, convert it to a int64
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)
	//Get the filename
	connection.Read(bufferFileName)
	//Strip the ':' once again but from the received file name now
	fileName := strings.Trim(string(bufferFileName), ":")
	fmt.Printf("filename %v fileSize %v \n", fileName, fileSize)
	//Create a new file to write in
	newFile, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("create file fail  %v \n", fileName)
		panic(err)
	}
	defer newFile.Close()
	//Create a variable to store in the total amount of data that we received already
	var receivedBytes int64
	//Start writing in the file
	startTime := time.Now()
	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(newFile, connection, (fileSize - receivedBytes))
			//Empty the remaining bytes that we don't need from the network buffer
			connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			//We are done writing the file, break out of the loop
			break
		}
		io.CopyN(newFile, connection, BUFFERSIZE)
		//Increment the counter
		receivedBytes += BUFFERSIZE
	}
	fmt.Printf("Received file completely! cost %v \n", time.Now().Sub(startTime).Seconds())
}

type BytesPerTime struct {
	Bytes    uint64
	Duration time.Duration
}

func SpeedMeter(input chan BytesPerTime, bytesPerSec chan BytesPerTime) {
	go func() {
		bpt := BytesPerTime{}
		for {
			select {
			case newBpt := <-input: /*收集到的信息，存到bpt，然后bpt存到bytesPerSec里，之后，bpt恢复初始值*/
				bpt.Bytes += newBpt.Bytes
				bpt.Duration += newBpt.Duration
			case bytesPerSec <- bpt:
				bpt = BytesPerTime{}
			}
		}
	}()
}

func SpeedReporter(input chan BytesPerTime, interval time.Duration) {
	go func() {
		for {
			select {
			case <-time.After(interval): /*这个是为了周期性的上报*/
				bpt, ok := <-input
				if !ok {
					log.Println("Reporting stopped")
					return
				}
				if bpt.Duration.Seconds() != 0 {
					log.Printf("Throughput: %s/s", humanize.IBytes(uint64(math.Ceil(float64(bpt.Bytes)/bpt.Duration.Seconds()))))
				} else {
					log.Println("No throughput")
				}
			}
		}
	}()
}