package main

import (
	dpfm_api_caller "data-platform-api-article-deletes-rmq-kube/DPFM_API_Caller"
	dpfm_api_input_reader "data-platform-api-article-deletes-rmq-kube/DPFM_API_Input_Reader"
	dpfm_api_output_formatter "data-platform-api-article-deletes-rmq-kube/DPFM_API_Output_Formatter"
	"data-platform-api-article-deletes-rmq-kube/config"
	"encoding/json"
	"fmt"
	"time"

	"github.com/latonaio/golang-logging-library-for-data-platform/logger"
	database "github.com/latonaio/golang-mysql-network-connector"
	rabbitmq "github.com/latonaio/rabbitmq-golang-client-for-data-platform"
)

func main() {
	l := logger.NewLogger()
	conf := config.NewConf()
	db, err := database.NewMySQL(conf.DB)
	if err != nil {
		l.Fatal(err.Error())
	}
	rmq, err := rabbitmq.NewRabbitmqClient(conf.RMQ.URL(), conf.RMQ.QueueFrom(), conf.RMQ.SessionControlQueue(), conf.RMQ.QueueToSQL(), 0)
	if err != nil {
		l.Fatal(err.Error())
	}
	defer rmq.Close()
	iter, err := rmq.Iterator()
	if err != nil {
		l.Fatal(err.Error())
	}
	defer rmq.Stop()

	caller := dpfm_api_caller.NewDPFMAPICaller(conf, rmq, db)

	for msg := range iter {
		start := time.Now()
		err = callProcess(rmq, caller, conf, msg)
		if err != nil {
			msg.Fail()
			continue
		}
		msg.Success()
		l.Info("process time %v\n", time.Since(start).Milliseconds())
	}
}

func recovery(l *logger.Logger, err *error) {
	if e := recover(); e != nil {
		*err = fmt.Errorf("error occurred: %w", e)
		l.Error(err)
		return
	}
}
func getSessionID(data map[string]interface{}) string {
	id := fmt.Sprintf("%v", data["runtime_session_id"])
	return id
}

func callProcess(rmq *rabbitmq.RabbitmqClient, caller *dpfm_api_caller.DPFMAPICaller, conf *config.Conf, msg rabbitmq.RabbitmqMessage) (err error) {
	l := logger.NewLogger()
	defer recovery(l, &err)

	l.AddHeaderInfo(map[string]interface{}{"runtime_session_id": getSessionID(msg.Data())})
	var input dpfm_api_input_reader.SDC
	var output dpfm_api_output_formatter.SDC

	err = json.Unmarshal(msg.Raw(), &input)
	if err != nil {
		l.Error(err)
		return
	}
	err = json.Unmarshal(msg.Raw(), &output)
	if err != nil {
		l.Error(err)
		return
	}

	accepter := getAccepter(&input)
	res, errs := caller.AsyncDeletes(accepter, &input, &output, l)
	if len(errs) != 0 {
		for _, err := range errs {
			l.Error(err)
		}
		output.APIProcessingResult = getBoolPtr(false)
		output.APIProcessingError = errs[0].Error()
		output.Message = res
		rmq.Send(conf.RMQ.QueueToResponse(), output)
		return errs[0]
	}
	output.APIProcessingResult = getBoolPtr(true)
	output.Message = res

	l.JsonParseOut(output)
	rmq.Send(conf.RMQ.QueueToResponse(), output)

	return nil
}

func getAccepter(input *dpfm_api_input_reader.SDC) []string {
	accepter := input.Accepter
	if len(input.Accepter) == 0 {
		accepter = []string{"All"}
	}

	if accepter[0] == "All" {
		accepter = []string{
			"Header",
		}
	}
	return accepter
}

func getBoolPtr(b bool) *bool {
	return &b
}
