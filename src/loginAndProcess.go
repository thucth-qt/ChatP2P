package main

import (
	"fmt"
	"io"
	"strings"
	"time"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"os"
	"tuilib"
)

const BUFFERSIZE = 1024

type ServerAddr= struct {
	IP string
	PORT string
}
var serverAddr = ServerAddr{
	IP:   "localhost",
	PORT: "8000",
}
var myName ="defaultname"
type friendInfo struct{
	status bool
	Ws *websocket.Conn
}
var friends = make(map[string] *friendInfo)
type Message struct{
	Receiver string `json:"receiver"`
	Sender string `json:"sender"`
	Message string `json:"message"` // format: "<receiver>_<content>"
}
var messageChan = make(chan Message)
var listenIP string
var listenPort string
var inputText = make(chan string)
var finishSession = make(chan bool)

func LoginThenProcess() {
	serverAddr.IP,serverAddr.PORT=tuilib.EnterServerAddr()
	myName=tuilib.Login()
	LoginFinish<-true

	//for{
	//	myName=tuilib.Login()
	//	if strings.TrimSpace(myName)!=""{break}
	//}

	Listener:=createListener()
	go startListen(Listener)
	//log.Println("____________________________created Listenter")
	//log.Println("____________________________inputed name")
	SocketToServer:= connectServer()
	//log.Println("____________________________connected server")
	defer SocketToServer.Close()
	go typeMessage()
	go handleMessageInClient()
	<-finishSession
}
func createListener() net.Listener{
	http.HandleFunc("/ws",handleConnectionsInClient)
	listener, err2 := net.Listen("tcp4", ":0")
	if err2 != nil {
		log.Fatal("Listen: ", err2)
	}
	myAddr := listener.Addr().(*net.TCPAddr) //cast type 
	listenIP = resolveHostIp()
	listenPort = strconv.Itoa(myAddr.Port)
	//OutPutChat<- "client listening on :"+ listenIP+":"+listenPort
	return listener

}
func resolveHostIp() string {
	netInterfaceAddresses, err := net.InterfaceAddrs()
	if err != nil { return "" }
	for _, netInterfaceAddress := range netInterfaceAddresses {
		networkIp, ok := netInterfaceAddress.(*net.IPNet)
		if ok && networkIp.IP.To4() != nil&&!networkIp.IP.IsLoopback()&&networkIp.IP.To4()[0]==10 {
			ip := networkIp.IP.String()
			return ip
		}
	}
		
	for _, netInterfaceAddress := range netInterfaceAddresses {
		networkIp, ok := netInterfaceAddress.(*net.IPNet)
		if ok && networkIp.IP.To4() != nil&&!networkIp.IP.IsLoopback()&&networkIp.IP.To4()[0]==172 {
			ip := networkIp.IP.String()
			return ip
		}
	}

	for _, netInterfaceAddress := range netInterfaceAddresses {
		networkIp, ok := netInterfaceAddress.(*net.IPNet)
		if ok && networkIp.IP.To4() != nil&&!networkIp.IP.IsLoopback()&&networkIp.IP.To4()[0]==192 {
			ip := networkIp.IP.String()
			return ip
		}
	}	
	return "__________"
}

func startListen(listener net.Listener) {
	//must register a handle func for defaultServeMux
	err3 := http.Serve(listener, nil)
	if err3 != nil {
		log.Fatal("ListenAndServe: ", err3)
	}
}

func connectServer() *websocket.Conn{
	return connectPartner("server",serverAddr.IP,serverAddr.PORT)
}
func connectPartner(name string, host string, port string) *websocket.Conn {
	//set URL for HTTP
	var u url.URL
	u = url.URL{Scheme: "ws", Host: host + ":" + port, Path: "/ws"}
	//fmt.Println("connecting to: ", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		OutPutChat<-"error dial:"+ err.Error()
		time.Sleep(5*time.Second)
	}
	if name!="server"{OutPutChat<-"... "+name +" is connected"}
	helloPartner(c)
	var response Message
	c.ReadJSON(&response)
	if response.Sender=="server"{
		if response.Message=="ERR: login two different places"{
			OutPutChat<-"OPP!! you are logging in another device.\n PLEASE LOG OUT THERE BEFORE LOG IN HERE"
			time.Sleep(2*time.Second)
			os.Exit(1)
		}else{
			OutPutChat<-"server: "+response.Message
		}
	}
	//fmt.Println("response: ",response.Message)
	friends[name] = &friendInfo{true, c} // add server to list friend
	go receiveMessageInClient(c)
	return c
}
func helloPartner(ws *websocket.Conn){

	introduce:=Message{"server",myName,"INTRO "+listenIP+" "+listenPort}
	err:= ws.WriteJSON(introduce)
	//log.Println("____________________________hello server")
	//log.Println("____________________________message hello: ",introduce)
	if err!=nil{log.Println("error register WriteJSON: ",err)}
}
func handleMessageInClient(){
	for{
		msg:=<-messageChan
		if msg.Receiver==myName { //received message => need to handle and then forward
			if msg.Sender=="server"{
				a:=regexp.MustCompile(" ")
				contentInForm:=a.Split(msg.Message,-1)
				mode:=contentInForm[0]
				switch mode{
				case "REQUEST":
					time.Sleep(100*time.Millisecond)
					OutPutChat<-"FROM server: "+contentInForm[1]+" request Connect to you? <Accept(1) or Confuse(0)> <name> "
				case "INFO":
					partName := contentInForm[1]
					partnerIP := contentInForm[2]
					partnerPort := contentInForm[3]
					connectPartner(partName,partnerIP,partnerPort)
				case "INFO_NO_ONE":
					partName := contentInForm[1]
					time.Sleep(100*time.Millisecond)
					OutPutChat<-"FROM server: No client named "+ partName +" in system"
				case "INFO_OFFLINE":
					partName := contentInForm[1]
					time.Sleep(100*time.Millisecond)
					OutPutChat<-"FROM server: Client named "+ partName +" is offline"
				case "INFO_CONFUSE":

					partName := contentInForm[1]
					time.Sleep(100*time.Millisecond)
					OutPutChat<-"FROM server: "+ partName +" confused request"
				case "UPDATE":
					if friends[contentInForm[1]] !=nil{
						partName := contentInForm[1]
						partnerIP := contentInForm[2]
						partnerPort := contentInForm[3]
						connectPartner(partName,partnerIP,partnerPort)
						time.Sleep(100*time.Millisecond)
						OutPutChat<-partName+" is online"
					}
				case "OK":
					OutPutChat<-"FROM server: "+msg.Message
				}
				//extra check KHACH
				if len(mode)>=5{
					if mode[0:5]=="KHACH"{
						OutPutChat<-msg.Message
					}
				}

			} else {
				a:=regexp.MustCompile(" ")
				contentInForm:=a.Split(msg.Message,2)
				mode:=contentInForm[0]
				if mode == "SENDFILE" {
					MsgAndIP := a.Split(contentInForm[1],2)
					connection, err := net.Dial("tcp", MsgAndIP[1] + ":9000")
					if err != nil {
						panic(err)
					}
					//defer connection.Close()
					//fmt.Println("Connected to server, start receiving the file name and file size")
					bufferFileName := make([]byte, 64)
					bufferFileSize := make([]byte, 10)

					connection.Read(bufferFileSize)
					fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

					connection.Read(bufferFileName)
					fileName := strings.Trim(string(bufferFileName), ":")

					newFile, err := os.Create(fileName)

					if err != nil {
						panic(err)
					}
					var receivedBytes int64

					for {
						if (fileSize - receivedBytes) < BUFFERSIZE {
							io.CopyN(newFile, connection, (fileSize - receivedBytes))
							connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
							break
						}
						io.CopyN(newFile, connection, BUFFERSIZE)
						receivedBytes += BUFFERSIZE
					}
					newFile.Close()
					OutPutChat<-"Received file completely!"
					connection.Close()
				} else {
					//fmt.Println(msg.Sender+": "+msg.Message)
					time.Sleep(100*time.Millisecond)
					OutPutChat<-"FROM "+msg.Sender+":\t\t"+msg.Message
				}
			}
		} else {// sent message => need to forward
			if msg.Receiver=="server"{
				a:=regexp.MustCompile(" ")
				contentInForm:=a.Split(msg.Message,-1)
				mode:=contentInForm[0]
				if mode!="CONNECT"{
					msg.Message="ACCEPT " + msg.Message
				}
				sendMessageInClient(msg.Receiver, msg)
			} else if len(msg.Receiver)>=5{ //extra receive from KHACHi
				if msg.Receiver[0:5]=="KHACH"{
					msg.Message="REPKHACH "+msg.Receiver+" "+msg.Message
					msg.Receiver="server"
					sendMessageInClient(msg.Receiver, msg)
				}

			}else{
				a:=regexp.MustCompile(" ")
				contentInForm:=a.Split(msg.Message,2)
				mode:=contentInForm[0]
				if mode == "SENDFILE"{
					msg.Message= msg.Message +" "+listenIP
					sendMessageInClient(msg.Receiver, msg)

					//send file to friend
					server, err := net.Listen("tcp", listenIP+":9000")
					if err != nil {
						fmt.Println("Error listetning: ", err)
						os.Exit(1)
					}
					//defer server.Close()
					//fmt.Println("Server started! Waiting for connections...")

					connection, err := server.Accept()
					if err != nil {
						fmt.Println("Error: ", err)
						os.Exit(1)
					}
					//fmt.Println("Client connected")
					sendFileToClient(connection, contentInForm[1])
					server.Close()
				} else {
					sendMessageInClient(msg.Receiver, msg)

				}
			}
			//sendMessageInClient(msg.Receiver, msg)
			//log.Println("____________________________sent message")
		}
	}
}
func typeMessage(){
	for{

		text:=<-InputChat
		a:= regexp.MustCompile(` `)
		textInForm:= a.Split(text,2) // <receiver> <content>
		m:=Message{ textInForm[0],myName,textInForm[1]}
		messageChan<-m
	}
}
func sendMessageInClient(uname string, m Message){
	if friends[uname]==nil{
		time.Sleep(50*time.Millisecond)
		OutPutChat<-"no friend named "+uname
		return
	}
	err:=friends[uname].Ws.WriteJSON(m)
	if err!= nil{
		log.Println(uname," is offline")
		friends[uname].status=false
		friends[uname].Ws.Close()
	}
}
func handleConnectionsInClient(w http.ResponseWriter, r *http.Request) {
	//Upgrade initial GET request to a websocket
	upgrader := websocket.Upgrader{}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err!=nil{log.Println("error upgrader: ",err)}
	//receive hello from partner
	helloFromPartner:=Message{"","",""}
	err = ws.ReadJSON(&helloFromPartner)
	if err!=nil{log.Println("error read helloFromPartner: ",err)}

	friends[helloFromPartner.Sender]=&friendInfo{true,ws}
	response:=Message{helloFromPartner.Sender,myName,myName+" connected"}
	ws.WriteJSON(response)
	go receiveMessageInClient(ws)
	if err != nil {
		log.Println("handleConnectionsInClient ",err)
	}
	go handleMessageInClient()

}
func receiveMessageInClient(ws *websocket.Conn){

	for{
		var msg Message
		err:= ws.ReadJSON(&msg)
		time.Sleep(100*time.Millisecond)
		//log.Println("receive MSG", msg)
		if err!= nil{
			for uname,info:=range friends{
				if info.Ws==ws{
					time.Sleep(100*time.Millisecond)
					OutPutChat<-uname+" exitted the CHATP2P "
				}
			}
			ws.Close()
			return
		}
		messageChan<- msg
		//log.Println("receive MSG", msg)
		//time.Sleep(1000*time.Millisecond)

	}
}

func sendFileToClient(conn net.Conn, filename string) {
	//fmt.Println("A client has connected!")

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := fillString(fileInfo.Name(), 64)
	//fmt.Println("Sending filename and filesize!")
	conn.Write([]byte(fileSize))
	conn.Write([]byte(fileName))
	sendBuffer := make([]byte, BUFFERSIZE)
	OutPutChat<-"Start sending file!"
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		conn.Write(sendBuffer)
	}
	conn.Close()
	OutPutChat<-"File has been sent, closing connection!"
	return
}

func fillString(returnString string, toLength int) string {
	for {
		lengtString := len(returnString)
		if lengtString < toLength {
			returnString = returnString + ":"
			continue
		}
		break
	}
	return returnString
}