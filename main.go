package main

import (

	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"regexp"

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
func main() {
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
	helloFromClient:=Message{"","",""}
	err = ws.ReadJSON(&helloFromClient)
	if err!=nil{log.Println("error read helloFromClient: ",err)}
	//update client information
	a:=regexp.MustCompile(` `)
	contentInForm:=a.Split(helloFromClient.Message,3) //form: "INTRO" <ip> <port>
	sender:=helloFromClient.Sender
	if err != nil {
		log.Println("handleConnectionsInClient ",err)
	}
	response:=Message{sender,helloFromClient.Receiver,""}
	if  clients[sender]==nil{
		response.Message="OK welcome to CHAT"
		ws.WriteJSON(response)
		//regerist client
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

	go receiveMessage(ws,helloFromClient.Sender)
	go handleMessage()

}
func receiveMessage(ws *websocket.Conn,uname string){
	var msg Message
	for{
		err:= ws.ReadJSON(&msg)
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
			log.Println(msg)
			a := regexp.MustCompile(` `)
			contentInFormat:= a.Split(msg.Message, -1) //format: <MODE> <CONTENT>
			answer:=Message{msg.Sender,"server",""}
			switch contentInFormat[0] {
			case "CONNECT":
				if clients[contentInFormat[1]]==nil {
					answer.Message = "INFO_NO_ONE " + contentInFormat[1]
					sendMessage(msg.Sender,answer)
				}else if clients[contentInFormat[1]].status==false {
					answer.Message = "INFO_OFFLINE " + contentInFormat[1]
					sendMessage(msg.Sender,answer)
				}else{
					//ask partner for accepting
					forwardReq := Message{contentInFormat[1],"server" , "REQUEST " + msg.Sender}
					sendMessage(contentInFormat[1], forwardReq)
				}
			case "ACCEPT":
				forwardReq:=Message{contentInFormat[2],"server",""}
				if contentInFormat[1]=="1"{
					forwardReq.Message="INFO "+msg.Sender+" "+clients[msg.Sender].IP+" "+clients[msg.Sender].PORT
				}else{
					forwardReq.Message="INFO_CONFUSE "+msg.Sender
				}
				sendMessage(contentInFormat[2], forwardReq)
			}

		}else{//sent message
			sendMessage(msg.Receiver,msg)
		}
	}
}

func sendMessage(uname string, m Message){
	if clients[uname].status==false{return}
	err:=clients[uname].Ws.WriteJSON(m)
	if err != nil {
		clients[uname].status=false
		return
	}
}
