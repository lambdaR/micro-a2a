package a2a

// A2A follows OpenAPI’s Authentication specification for authentication. Importantly, A2A agents
// do not exchange identity information within the A2A protocol. Instead, they obtain materials
// (such as tokens) out of band and transmit materials in HTTP headers and not in A2A payloads.
//
// Servers do send authentication requirements in A2A payloads. At minimum, servers are expected
// to publish their requirements in their Agent Card
//
// A2A servers should authenticate every request and reject or challenge requests with standard HTTP
// response codes (401, 403), and authentication-protocol-specific headers and bodies
// (such as a HTTP 401 response with a WWW-Authenticate header indicating the required authentication
// schema, or OIDC discovery document at a well-known path). More details discussed in Enterprise Ready.

// Discovering Agent Cards
//
// Open Discovery e.g https://DOMAIN/.well-known/agent.json via GET
// Curated Discovrey (Registry-Based)
// Private Discovrey (API-Based)

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	httpServer "github.com/micro/plugins/v5/server/http"

	"go-micro.dev/v5"
	"go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/server"
	"go-micro.dev/v5/store"

	"github.com/gin-gonic/gin"
)

type ResultChan chan JSONRPCResponse

type AgentHandler interface {
	TaskHandler(JSONRPCRequest) JSONRPCResponse
}

type AgentStreamHandler interface {
	StreamHandler(JSONRPCRequest, chan JSONRPCResponse)
}

// New event messages are broadcast to all registered client connection channels
type ClientChan chan string

type Agent struct {
	options AgentOptions
	Server  server.Server
}

// NewAgent creates new remote Agent (Server), if the WithStore option is not provided
// the store will default to the go-micro v5 memory store
func NewAgent(agentCard AgentCard, opts ...AgentOption) *Agent {
	re := regexp.MustCompile(`[ .]`) // Match spaces and periods
	agentName := re.ReplaceAllString(agentCard.Name, "")

	agent := &Agent{
		Server: httpServer.NewServer(
			server.Name(agentName),
			server.Address(agentCard.URL),
		),

		options: AgentOptions{
			AgentCard: agentCard,
		},
	}

	for _, o := range opts {
		o(&agent.options)
	}

	// set the default store
	if agent.options.Store == nil {
		agent.options.Store = store.NewMemoryStore()
	}

	// set the default logger
	if agent.options.Logger == nil {
		agent.options.Logger = logger.NewLogger()
	}

	return agent
}

func (a *Agent) SwitchOn() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Recovery())

	// TODO(daniel) check for Authentication, agent card provides information about the method
	// that should be used to Auth and then Use the proper Middleware
	// for example for Basic Authentication
	// 	authorized := router.Use(gin.BasicAuth(gin.Accounts{
	// 		"admin": "admin123", // username : admin, password : admin123
	// 	}))

	// check if the Agent supports streaming
	streamingSupported := a.options.AgentStreamHandler != nil && a.options.AgentCard.Capabilities.Streaming

	a.options.Logger.Log(logger.InfoLevel, fmt.Sprintf("streamingSupported: %v", streamingSupported))

	// build endpoints based on AgentCard.Name and AgentCard.Capabilities.Streaming
	var path, pathStream string
	re := regexp.MustCompile(`[ .]`) // Match spaces and periods
	modified := re.ReplaceAllString(a.options.AgentCard.Name, "")
	path, err := url.JoinPath("/", modified)
	if err != nil {
		log.Fatalln(err)
	}

	if streamingSupported {
		pathStream, err = url.JoinPath("/", modified, "/stream")
		if err != nil {
			log.Fatalln(err)
		}
	}

	router.POST(path, agentHandler(a))
	if streamingSupported {
		router.POST(pathStream, agentHandler(a))
		router.GET(pathStream, sseHeadersMiddleware(), streamHandlerMiddleware(a), agentStreamHandler(a))
	}

	hd := a.Server.NewHandler(router)
	if err := a.Server.Handle(hd); err != nil {
		log.Fatalln(err)
	}

	service := micro.NewService(
		micro.Server(a.Server),
		micro.Registry(registry.NewRegistry()),
		micro.Logger(a.options.Logger),
	)

	service.Init()
	service.Run()
}

func agentHandler(a *Agent) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r JSONRPCRequest

		if err := c.ShouldBindJSON(&r); err != nil {
			a.options.Logger.Log(logger.ErrorLevel, err)
			e := NewError(ErrorInternal, "error unmarshalling request body", nil)
			c.JSON(http.StatusInternalServerError, e)
			return
		}

		a.options.Logger.Log(logger.InfoLevel, r)

		switch r.Method {
		case TasksSend:
			_, ok := (r.Params).(TaskSendParams)
			if !ok {
				e := NewError(ErrorInvalidRequest, "request should include a TaskSendParams as params", nil)
				c.JSON(http.StatusBadRequest, e)
				return
			}

			if a.options.AgentHandler == nil {
				e := NewError(ErrorInternal, "the Agent doesn't implement AgentHandler", nil)
				c.JSON(http.StatusInternalServerError, e)
				return
			}

			res := (*a.options.AgentHandler).TaskHandler(r)
			if res.Error != nil {
				c.JSON(http.StatusBadRequest, res.Error)
				return
			}
			c.JSON(http.StatusOK, res)

		case TasksSendSubscribe:
			// check if id exitst in JSONRPCRequest
			if r.ID == nil {
				e := NewError(ErrorInvalidRequest, "ID shouldn't be nil", nil)
				c.JSON(http.StatusBadRequest, e)
				return
			}

			// save it in the store key=id | value=JSONRPCRequest
			rawReq, err := json.Marshal(r)
			if err != nil {
				e := NewError(ErrorInternal, "faild to Marshal request", nil)
				c.JSON(http.StatusInternalServerError, e)
				return
			}

			err = a.options.Store.Write(&store.Record{Key: fmt.Sprintf("%v", r.ID), Value: rawReq, Expiry: time.Second * 60})
			if err != nil {
				e := NewError(ErrorInternal, err.Error(), nil)
				c.JSON(http.StatusInternalServerError, e)
				return
			}

			// return OK
			c.JSON(http.StatusOK, nil)

		case TasksGet:
		case TasksCancel:
		default:
			e := NewError(ErrorInvalidRequest, "unsupported A2A method", nil)
			c.JSON(http.StatusInternalServerError, e)
			return
		}
	}
}

func agentStreamHandler(a *Agent) gin.HandlerFunc {
	return func(c *gin.Context) {
		rc, ok := c.Get("resultChan")
		if !ok {
			return
		}

		resultChan, ok := rc.(ResultChan)
		if !ok {
			return
		}

		c.Stream(func(w io.Writer) bool {
			if result, ok := <-resultChan; ok {
				a.options.Logger.Log(logger.InfoLevel, result)
				c.SSEvent("message", result)
				return true
			}
			return false
		})
	}
}

func streamHandlerMiddleware(a *Agent) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get id param from the request URL
		id, idExist := c.GetQuery("id")
		if !idExist {

			e := NewError(ErrorInvalidRequest, "URL is missing id query", nil)
			c.JSON(http.StatusBadRequest, e)
			return
		}

		// check if id exists in store
		record, err := a.options.Store.Read(id)
		if err != nil {
			e := NewError(ErrorInternal, "request id no found", nil)
			c.JSON(http.StatusInternalServerError, e)
			return
		}

		// if ID found, restructure JSONRPCRequest and launch StreamHandler
		var r JSONRPCRequest
		err = json.Unmarshal(record[0].Value, &r)
		if err != nil {
			e := NewError(ErrorInternal, "failed to Unmarshal req from store", nil)
			c.JSON(http.StatusInternalServerError, e)
			return
		}

		// a channel for sending back results
		results := make(ResultChan, 1)
		go (*a.options.AgentStreamHandler).StreamHandler(r, results)

		c.Set("resultChan", results)

		c.Next()
	}
}

func sseHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}

func updateRecord(key string, status string, rec []*store.Record, newValue Result) (*store.Record, error) {
	vRaw := *&rec[0].Value
	var v map[int64]Result
	err := json.Unmarshal(vRaw, &v)

	now := time.Now().Unix()

	v[now] = newValue
	valuesAsBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	record := store.Record{
		Key:      key,
		Expiry:   time.Second * 60,
		Value:    valuesAsBytes,
		Metadata: map[string]interface{}{"status": status},
	}

	return &record, nil
}

func newErrRes(r *JSONRPCRequest, e *JSONRPCError) *SSEResponse {
	res := new(JSONRPCResponse)
	res.JSONRPC = r.JSONRPC
	res.ID = r.ID
	res.Error = e

	finalRes := new(SSEResponse)
	finalRes.Data = *res

	return finalRes
}
