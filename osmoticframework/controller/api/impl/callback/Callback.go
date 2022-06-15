package callback

import (
	"osmoticframework/controller/api/impl/request"
)

//These functions return API calls to the result channel.

func CallbackError(requestId string, err error) {
	_reply, _ := request.DeployTaskList.Load(requestId)
	reply := _reply.(request.RequestTask)
	reply.Result <- request.Result{
		ResultType: request.Error,
		Content:    err,
	}
}

func CallbackOk(requestId string, content interface{}) {
	_reply, _ := request.DeployTaskList.Load(requestId)
	reply := _reply.(request.RequestTask)
	reply.Result <- request.Result{
		ResultType: request.Ok,
		Content:    content,
	}
}
