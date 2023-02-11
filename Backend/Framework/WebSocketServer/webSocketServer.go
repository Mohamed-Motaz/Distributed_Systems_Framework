package main

import (
	cache "Framework/Cache"
	logger "Framework/Logger"
	mq "Framework/MessageQueue"
	utils "Framework/Utils"
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

func newClient(id string, webSocketConn *websocket.Conn) *Client {
	return &Client{
		id:            id,
		finishedJobs:  make(chan string),
		webSocketConn: webSocketConn,
	}
}

func NewWebSocketServer() (*WebSocketServer, error) {

	webSocketServer := &WebSocketServer{
		id:      uuid.NewString(),
		cache:   cache.NewCache(cache.CreateCacheAddress(CacheHost, CachePort)),
		queue:   mq.NewMQ(mq.CreateMQAddress(MqUsername, MqPassword, MqHost, MqPort)),
		clients: make(map[string]*Client),
		mu:      sync.Mutex{},
	}

	serveMux := mux.NewRouter()
	serveMux.HandleFunc("/openWS/{clientId}", webSocketServer.handleJobRequests)
	serveMux.HandleFunc("/uploadBinary", webSocketServer.handleUploadBinaryRequests).Methods("POST")
	serveMux.HandleFunc("/getAllBinaries", webSocketServer.handleGetAllBinariesRequests).Methods("POST")
	serveMux.HandleFunc("/deleteBinary", webSocketServer.handleDeleteBinaryRequests).Methods("POST")
	serveMux.HandleFunc("/getSystemProgress", webSocketServer.handleGetSystemProgressRequests).Methods("POST")
	serveMux.HandleFunc("/getAllFinishedJobs", webSocketServer.handleGetAllFinishedJobsRequests).Methods("POST")

	webSocketServer.requestHandler = cors.AllowAll().Handler(middlewareLogger(serveMux))

	go webSocketServer.listenAndServe()
	go webSocketServer.deliverJobs()

	return webSocketServer, nil
}

func (webSocketServer *WebSocketServer) listenAndServe() {

	logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Listening on %v:%v", MyHost, MyPort)

	if err := http.ListenAndServe(MyHost+":"+MyPort, webSocketServer.requestHandler); err != nil {
		logger.FailOnError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Failed in listening on port} -> error : %v", err)
	}
}

// Middleware to log all incoming requests
func middlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.LogRequest(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Request received from %v to %v", r.RemoteAddr, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

// this method is responsible for listening on the specific websocket connection
// this method returns only when the connection is closed

func (webSocketServer *WebSocketServer) listenForJobs(client *Client) {

	defer func() {
		webSocketServer.mu.Lock()
		delete(webSocketServer.clients, client.id)
		webSocketServer.mu.Unlock()
	}()
	defer client.webSocketConn.Close()

	for {

		_, message, err := client.webSocketConn.ReadMessage()
		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Unable to read from the websocket %+v with err:\n%v", client.webSocketConn.RemoteAddr(), err)
			webSocketServer.writeError(client, utils.Error{Err: true, ErrMsg: "Unable to read from the websocket"})
			return
		}

		newJobRequest := &JobRequest{}

		err = json.Unmarshal(message, newJobRequest)

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Error with client while unmarshaling json %v\n%v", client.webSocketConn.RemoteAddr(), err)
			webSocketServer.writeError(client, utils.Error{Err: true, ErrMsg: "Invalid format"})
			continue
		}
		logger.LogInfo(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "This is the message received on the websocket connection %+v", newJobRequest)
		//immediately attempt to send the optional files to the lockserver
		if !webSocketServer.handleSendOptionalFiles(client, newJobRequest) { //no need to send the errors since this method is responsible for this
			continue
		}

		modifiedJobRequest := &mq.AssignedJob{
			ClientId: client.id,
		}

		webSocketServer.modifyJobRequest(newJobRequest, modifiedJobRequest)

		jobToAssign := new(bytes.Buffer)

		err = json.NewEncoder(jobToAssign).Encode(modifiedJobRequest)
		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Error with encoding data %+v for client %+v\n%+v", modifiedJobRequest, client.webSocketConn.RemoteAddr(), err)
			webSocketServer.writeError(client, utils.Error{Err: true, ErrMsg: "Can't encode the job request and send it to the message queue at the moment"})
			//DONE, send an rpc to the lockserver telling it to delete the files
			webSocketServer.handleDeleteOptionalFiles(modifiedJobRequest.JobId)
			continue
		}

		err = webSocketServer.queue.Enqueue(mq.ASSIGNED_JOBS_QUEUE, jobToAssign.Bytes())

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{New job not enqeued to jobs assigned queue} -> error : %+v", err)
			webSocketServer.writeError(client, utils.Error{Err: true, ErrMsg: "Message queue unavailable"})
			webSocketServer.handleDeleteOptionalFiles(modifiedJobRequest.JobId)
		} else {
			logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "New job successfully enqeued to jobs assigned queue")
		}

	}
}

func (webSocketServer *WebSocketServer) deliverJobs() {

	finishedJobsChan, err := webSocketServer.queue.Dequeue(mq.FINISHED_JOBS_QUEUE)

	if err != nil {
		logger.FailOnError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Server can't consume finished Jobs} -> error : %v", err)
	}

	for {
		time.Sleep(time.Second * 5)

		for finishedJobObj := range finishedJobsChan {

			finishedJob := &mq.FinishedJob{}
			err := json.Unmarshal(finishedJobObj.Body, finishedJob)
			if err != nil {
				logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to unMarshal job %v } -> error : %v\n Will be discarded", string(finishedJobObj.Body), err)
				finishedJobObj.Ack(false)
				continue
			}

			//DONE: send the appropriate response to the user

			//CASES
			//cache is dead				p1 client is alive     			   --send the response, and ack
			//							p2 client isn't alive  			   --nack
			//
			//cache is alive			p1 client is mine and alive	   	   --send the response, cache, and ack
			//							p2 client is mine and dead		   --cache only, and ack
			//							p3 client isn't mine			   --nack
			//DONE check if job has error or not

			webSocketServer.mu.Lock()
			client, clientIsAlive := webSocketServer.clients[finishedJob.ClientId]
			webSocketServer.mu.Unlock()

			if finishedJob.Err && clientIsAlive {
				webSocketServer.writeError(client, utils.Error{Err: true, ErrMsg: "There was an Error while processing the job"})
				continue
			}

			var clientData *cache.CacheValue
			clientData, err = webSocketServer.cache.Get(finishedJob.ClientId)

			if err != nil && err != redis.Nil { //case 1 -- cache is dead

				logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to connect to cache at the moment} -> error : %v", err)
				if clientIsAlive { //p1
					go webSocketServer.writeFinishedJob(client, *finishedJob)
					finishedJobObj.Ack(false)

				} else { //p2
					finishedJobObj.Nack(false, true)
				}
			} else { //case 2 -- cache is alive

				if clientData.ServerID == webSocketServer.id {
					if clientIsAlive { //p1
						go webSocketServer.writeFinishedJob(client, *finishedJob)
					}
					//p2
					finishedJobToCache := &cache.FinishedJob{
						JobId:     finishedJob.JobId,
						JobResult: finishedJob.Result,
					}

					clientData.FinishedJobs = append(clientData.FinishedJobs, *finishedJobToCache)
					err := webSocketServer.cache.Set(finishedJob.ClientId, clientData, MAX_IDLE_CACHE_TIME)

					if err != nil {
						logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to connect to cache at the moment} -> error : %v", err)
					}
					finishedJobObj.Ack(false)

				} else {
					finishedJobObj.Nack(false, true)
				}
			}

		}

	}
}
