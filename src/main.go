package main

import (

	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"

)

type clientInfo struct{
	IP string
	PORT string
	Ws *websocket.Conn
	status bool
}
var PORT string = ":8000"
var clients = make(map[string] *clientInfo)
var chanMessage = make( chan Message)
type Message struct{
	Receiver string `json:"receiver"`
	Sender string `json:"sender"`
	Message string `json:"message"`
}
//extra
var KHACHNum int = 1
func main() {
	//create a simple file server
	fs :=http.FileServer(http.Dir("../public"))
	http.Handle("/",fs)
	listenAndServe()
}

func listenAndServe(){

	log.Println("Server is listening on port ",resolveHostIp(),PORT)
	http.HandleFunc("/ws",handleConnections)
	err:=http.ListenAndServe(PORT,nil)
	if err != nil{
		log.Fatal("ListenAndServe: ",err) //fatal == print error and exit()
	}
}

func resolveHostIp() string {
	netInterfaceAddresses, err := net.InterfaceAddrs()
	if err != nil { return "" }
	for _, netInterfaceAddress := range netInterfaceAddresses {
		networkIp, ok := netInterfaceAddress.(*net.IPNet)
		if ok && networkIp.IP.To4() != nil&&!networkIp.IP.IsLoopback()&&networkIp.IP.To4()[0]==10 {
			//additional conditions ||networkIp.IP.To4()[0]==172||networkIp.IP.To4()[0]==192
			//!networkIp.IP.IsLoopback() && networkIp.IP.To4() != nil
			ip := networkIp.IP.String()
			return ip
		}
	}
		
	for _, netInterfaceAddress := range netInterfaceAddresses {
		networkIp, ok := netInterfaceAddress.(*net.IPNet)
		if ok && networkIp.IP.To4() != nil&&!networkIp.IP.IsLoopback()&&networkIp.IP.To4()[0]==172 {
			//additional conditions ||networkIp.IP.To4()[0]==172||networkIp.IP.To4()[0]==192
			//!networkIp.IP.IsLoopback() && networkIp.IP.To4() != nil
			ip := networkIp.IP.String()
			return ip
		}
	}

	for _, netInterfaceAddress := range netInterfaceAddresses {
		networkIp, ok := netInterfaceAddress.(*net.IPNet)
		if ok && networkIp.IP.To4() != nil&&!networkIp.IP.IsLoopback()&&networkIp.IP.To4()[0]==192 {
			//additional conditions ||networkIp.IP.To4()[0]==172||networkIp.IP.To4()[0]==192
			//!networkIp.IP.IsLoopback() && networkIp.IP.To4() != nil
			ip := networkIp.IP.String()
			return ip
		}
	}	
	return "__________"
}

func handleConnections(w http.ResponseWriter, r *http.Request){
	upgrader := websocket.Upgrader{}
	ws,err:=upgrader.Upgrade(w,r,nil)
	if err!= nil{
		log.Fatal(err)
	}
	log.Println("one new client connected")
	//receive hello from client
	var helloFromClient Message
	err = ws.ReadJSON(&helloFromClient)
	if err!=nil{log.Println("error read helloFromClient: ",err)}
	// for debugging
	//log.Println("sender:_"+helloFromClient.Sender+",receiver:_"+helloFromClient.Receiver+",Message:_"+helloFromClient.Message)

	//update client information
	a:=regexp.MustCompile(` `)
	contentInForm:=a.Split(helloFromClient.Message,3) //form: "INTRO" <ip> <port>
	sender:=helloFromClient.Sender
	if err != nil {
		log.Println("handleConnectionsInClient ",err)
	}
	response:=Message{sender,helloFromClient.Receiver,""}
	if  clients[sender]==nil{
		if sender =="KHACH"{
			response.Message="Do you need any help?"
		}else{
			response.Message="OK welcome to CHAT"
		}
		ws.WriteJSON(response)
		//regerist client
		if sender == "KHACH"{
			sender = sender+ strconv.Itoa(KHACHNum)
			KHACHNum+=1
		}
		clients[sender]=&clientInfo{contentInForm[1],contentInForm[2], ws,true}
	}else{
		if clients[sender].status==true{
			response.Message="ERR: login two different places"
			ws.WriteJSON(response)
			ws.Close()
			return
		}else {
			//update client
			clients[sender] = &clientInfo{contentInForm[1], contentInForm[2], ws, true}
			for clienti, _ := range clients {
				if clienti == sender {
					continue
				}
				updateStatus := Message{clienti, "server", "UPDATE " + sender + " " + clients[sender].IP + " " + clients[sender].PORT}
				sendMessage(clienti, updateStatus)
			}
			response.Message = "OK welcome back"
			ws.WriteJSON(response)
		}
	}

	go receiveMessage(ws,sender)
	go handleMessage()

}

func receiveMessage(ws *websocket.Conn,uname string){
	var msg Message
	for{
		err:= ws.ReadJSON(&msg)
		// if KHACH -> change to KHACH1, KHACH2,... else don't change
		if msg.Sender =="KHACH"{
			msg.Sender=uname
		}
		if err!= nil{
			log.Println("[receiveMessage] one client disconnected ")
			clients[uname].status=false
			ws.Close()
			return
		}
		chanMessage <-msg
	}
}

func handleMessage(){
	for {
		msg := <-chanMessage
		if msg.Receiver=="server"{ //received message
			log.Println("server receive from "+msg.Sender+":  "+msg.Message)
			//extra
			if len(msg.Sender)>=5{
				if msg.Sender[0:5] =="KHACH"{ //if receive Message from KHACHi
					msg.Message=msg.Sender+" " +msg.Message
					msg.Receiver="counselor"
					msg.Sender="server"
					sendMessage("counselor",msg) //forward message to counselor
					continue
				}
			}

			a := regexp.MustCompile(` `)
			contentInFormat:= a.Split(msg.Message, 2) //format: <MODE> <CONTENT>
			answer:=Message{msg.Sender,"server",""}
			switch contentInFormat[0] {
			case "REPKHACH": //extra <KHACH1> <hihihihihihi> ----> counselor reply KHACHi
				KhachContent:=a.Split(contentInFormat[1],2)
				khach:=KhachContent[0]
				answer.Message = KhachContent[1]
				sendMessage(khach,answer)

			case "CONNECT": //<STATUS> <NAME>
				NAME :=a.Split(contentInFormat[1],2)[0]
				if clients[NAME]==nil {
					answer.Message = "INFO_NO_ONE " + NAME
					sendMessage(msg.Sender,answer)
				}else if clients[NAME].status==false {
					answer.Message = "INFO_OFFLINE " + NAME
					sendMessage(msg.Sender,answer)
				}else{
					//ask partner for accepting
					forwardReq := Message{NAME,"server" , "REQUEST " + msg.Sender}
					sendMessage(NAME, forwardReq)
				}
			case "ACCEPT": //<1/0> <NAME>
				AGREE:=a.Split(contentInFormat[1],2)[0]
				NAME:=a.Split(contentInFormat[1],2)[1]
				forwardReq:=Message{NAME,"server",""}
				if AGREE=="1"{
					forwardReq.Message="INFO "+msg.Sender+" "+clients[msg.Sender].IP+" "+clients[msg.Sender].PORT
				}else{
					forwardReq.Message="INFO_CONFUSE "+msg.Sender
				}
				sendMessage(NAME, forwardReq)
			}

		}else{//sent message
			sendMessage(msg.Receiver,msg)
		}
	}
}

func sendMessage(uname string, m Message){
	if clients[uname]==nil{
		log.Println("send Message: client "+uname+" is offline")
		return
	}
	if clients[uname].status==false{
		log.Println("send Message: client "+uname+" is offline")
		return
	}
	err:=clients[uname].Ws.WriteJSON(m)
	if err != nil {
		log.Println("sendMessage: ",err)
		clients[uname].status=false
		return
	}
}
