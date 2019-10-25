package common

import (
	"encoding/json"
	"fmt"
	"github.com/b3log/gulu"
)

type Result struct {
	Cmd string `json:"cmd"`
	*gulu.Result
}

func NewResult() *Result {
	return &Result{"", &gulu.Result{Code: 1, Msg: "", Data: nil}}
}

func NewCmdResult(cmd string) *Result {
	ret := NewResult()
	ret.Cmd = cmd

	return ret
}

func (r *Result) Bytes() []byte {
	ret, err := json.Marshal(r)
	if nil != err {
		fmt.Println(fmt.Errorf("marshal result [%#v] failed [%s]", r, err))
	}
	return ret
}
