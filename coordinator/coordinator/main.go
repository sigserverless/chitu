package main

import (
	"bytes"
	"coordinator/channel"
	"coordinator/constant"
	"coordinator/object"
	"coordinator/reader"
	"coordinator/writer"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	redis "github.com/go-redis/redis/v8"
)

func main() {
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", 8080),
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	if channel.GlobalChannel == nil {
		channel.GlobalChannel = &channel.Channel{RedisClient: nil}

		channel.GlobalChannel.RedisClient = redis.NewClient(&redis.Options{
			Addr:     "redis-front.openfaas-lzy:6379",
			DB:       0,
			Password: "123456",
		})
	}

	http.HandleFunc("/get", HandleAgentGet)
	http.HandleFunc("/trigger", HandleAgentTrigger)
	http.HandleFunc("/put", HandleAgentPut)
	http.HandleFunc("/clear", HandleClear)
	http.HandleFunc("/ip", handle)
	http.HandleFunc("/", handle)

	s.ListenAndServe()
}

type GetPutReq struct {
	Key   string `json:"key"`
	DagID string `json:"dagId"`
}

type PutResponse struct {
	IPs []string `json:"ips"`
}

type Response struct {
	StatusCode int         `json:"statusCode"`
	StatusMsg  string      `json:"statusMsg"`
	Data       interface{} `json:"data"`
}

func HandleAgentPut(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		return
	}
	all, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	var request GetPutReq
	var response PutResponse
	err = json.Unmarshal(all, &request)
	if err != nil {
		panic(err)
	}
	log.Printf("HandleAgentPut: %v", request)

	ip := strings.Split(req.RemoteAddr, ":")[0]

	key := request.DagID + ":" + request.Key
	tobject, err := channel.GlobalChannel.GetObject(key)
	if err != nil {
		if err != redis.Nil {
			panic(err)
		}

		log.Printf("Key not exist: %v", key)

		object1 := object.NewObjectWithWriter(writer.NewWriter(ip), request.Key, request.DagID)

		err = channel.GlobalChannel.NewObject(object1, key)
		if err != nil {

			resp, err := json.Marshal(response)
			if err != nil {
				panic(err)
			}
			w.WriteHeader(http.StatusConflict)
			w.Write(resp)
			return
		}

		response.IPs = make([]string, 0)

	} else {

		response.IPs = make([]string, len(tobject.Readers))
		for i := 0; i < len(tobject.Readers); i++ {
			response.IPs[i] = tobject.Readers[i].IP
		}

		tobject.AddWriter(writer.NewWriter(ip))
	}
	resp, err := json.Marshal(response)
	if err != nil {

		panic(err)
	}
	w.Write(resp)

}

func HandleAgentGet(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		return
	}
	all, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	var request GetPutReq
	var response Response
	err = json.Unmarshal(all, &request)
	if err != nil {
		panic(err)
	}
	log.Printf("HandleAgentGet: %v", request)

	key := request.DagID + ":" + request.Key
	ip := strings.Split(req.RemoteAddr, ":")[0]
	tobject, err := channel.GlobalChannel.GetObject(key)
	if err == nil {
		if tobject.DagID != request.DagID {
			response.StatusCode = constant.ParameterInvalid
			response.StatusMsg = "dagId不一致"
		}

		tobject.AddReader(reader.NewReader(ip))

		req := GetNotifyReq{
			Key:   request.Key,
			DagId: request.DagID,
			IP:    ip,
		}

		if tobject.Writers != nil && len(tobject.Writers) > 0 {
			for i := 0; i < len(tobject.Writers); i++ {
				putState2Agent(req, tobject.Writers[i].IP)
			}

		}

		err = channel.GlobalChannel.SetObject(tobject, key)
		if err != nil {
			fmt.Println(err.Error())
			response.StatusCode = constant.ParameterInvalid
			response.StatusMsg = "key指代的对象不存在"
		}
	} else {

		if err == redis.Nil {
			object1 := object.NewObjectWithReader(reader.NewReader(ip), request.Key, request.DagID)
			err = channel.GlobalChannel.NewObject(object1, key)
			if err != nil {

				resp, err := json.Marshal(response)
				if err != nil {
					panic(err)
				}
				w.WriteHeader(http.StatusConflict)
				w.Write(resp)
				return
			}
		} else {

			w.WriteHeader(http.StatusConflict)
			return
		}

	}

	resp, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	w.Write(resp)
}

type TriggerReq struct {
	DagID string `json:"dagId"`
	Fname string `json:"fname"`
	Args  string `json:"args"`
}

type GetNotifyReq struct {
	Key   string `json:"key"`
	DagId string `json:"dagId"`
	IP    string `json:"ip"`
}

func putState2Agent(req GetNotifyReq, ip string) {
	reqJson, err := json.Marshal(req)
	log.Printf("Notifying agent: %s, request: %s", ip, string(reqJson))
	if err != nil {
		panic(err)
	}
	resp, err := http.Post("http://"+ip+":8080/get-notify", "application/json", bytes.NewReader(reqJson))
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	log.Println("Notifying agent response: ", string(all))
}

func HandleAgentTrigger(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		return
	}
	all, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	var request TriggerReq
	var response Response
	err = json.Unmarshal(all, &request)
	if err != nil {
		panic(err)
	}
	log.Printf("HandleAgentTrigger: %v", request)

	t := struct {
		DagId string `json:"dagId"`
		Req   string `json:"req"`
	}{
		DagId: request.DagID,
		Req:   request.Args,
	}
	marshal, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	reader := bytes.NewReader(marshal)
	_, err = http.Post("http://localhost:31112/function/"+request.Fname+"/async-invoke", "application/json", reader)
	if err != nil {
		response.StatusCode = constant.InternalServerError
		response.StatusMsg = err.Error()
		return
	}
	resp, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	w.Write(resp)
}

func HandleClear(w http.ResponseWriter, req *http.Request) {

	channel.GlobalChannel.Clear()
	w.Write([]byte("Cleared"))
}

func handle(w http.ResponseWriter, req *http.Request) {
	ip := strings.Split(req.RemoteAddr, ":")[0]
	w.Write([]byte(ip))
}
