package timer2

//Note: currently if the kill method is invoked then calls to the public
//methods will block indefinitely. Some work arounds are needed, for example
//waiting for outstanding ends to be called before ending the go routine,
//or buffering up

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

//Functions of the timer are accessed using commands
type command struct {
	opcode string
	args   []interface{}
}

//Opcodes for commands
const (
	killOp             = "kill"
	stopOp             = "stop"
	durationOp         = "duration"
	startContributorOp = "start-contrib"
	endContributorOp   = "end-contrib"
	errorFreeOpCode    = "error-free"
	getErrorOpCode     = "get-error"
	contribTimeOp      = "contrib-time"
	contribErrOp       = "contrib-err-op"
	startServiceCallOp = "start-svc-call"
	endServiceCallOp   = "end-svc-call"
	toJsonOp           = "to-json"
)

//EndToEnd timer is an opaque data type handed out to timer consumers. It exposes
//several methods, but allows direct access to the data members only from a goroutine
//spawned to manage the timer state.
type EndToEndTimer struct {
	Name         string
	TxnId        string
	start        time.Time
	c            chan command
	r            chan interface{}
	Duration     time.Duration
	Error        string
	ErrorFree    bool
	Contributors []*Contributor
}

type Contributor struct {
	timer        *EndToEndTimer
	Name         string
	start        time.Time
	Duration     time.Duration
	Error        string
	ServiceCalls []*ServiceCall
}

type ServiceCall struct {
	Name        string
	endpoint    string
	Duration    time.Duration
	Error       string
	start       time.Time
	contributor *Contributor
}

func (t *EndToEndTimer) handleStop(cmd command) {
	t.Duration = time.Now().Sub(t.start)
	log.Println("args 0", cmd.args[0])
	if len(cmd.args) > 0 {
		theErr, ok := cmd.args[0].(error)
		if ok {
			log.Println("set error to", theErr.Error())
			t.Error = theErr.Error()
			t.ErrorFree = false
		} else {
			log.Print("set error to nil")
			t.Error = ""
		}
	}
}

func (t *EndToEndTimer) handleDuration() {
	t.r <- t.Duration
}

func (t *EndToEndTimer) handleContribError(cmd command) {
	contributor := cmd.args[0].(*Contributor)
	t.r <- contributor.Error
}

func (t *EndToEndTimer) handleStartServiceCall(cmd command) {
	contributor := cmd.args[0].(*Contributor)
	name := cmd.args[1].(string)
	endpoint := cmd.args[2].(string)
	start := cmd.args[3].(time.Time)

	svcCall := &ServiceCall{
		Name:        name,
		endpoint:    endpoint,
		start:       start,
		contributor: contributor,
	}

	contributor.ServiceCalls = append(contributor.ServiceCalls, svcCall)

	t.r <- svcCall
}

func (t *EndToEndTimer) handleEndServiceCall(cmd command) {
	sc := cmd.args[0].(*ServiceCall)

	var err error
	theErr, ok := cmd.args[1].(error)
	if ok {
		err = theErr
	}

	end := cmd.args[2].(time.Time)

	if err != nil {
		sc.Error = err.Error()
	}
	sc.Duration = end.Sub(sc.start)

}

func (t *EndToEndTimer) handleToJson() {
	log.Println("handleToJson")
	s, err := json.Marshal(t)
	if err != nil {
		s = []byte("{}")
	}
	t.r <- string(s)
}

//handleTimerOps is the internal go routine responsible for accessing the timer internals
func (t *EndToEndTimer) handleTimerOps() {
	for {
		cmd := <-t.c
		log.Println("handle command", cmd.opcode)
		switch cmd.opcode {
		case contribErrOp:
			t.handleContribError(cmd)
		case startContributorOp:
			t.handleStartContributor(cmd)
		case endContributorOp:
			t.handleStopContributor(cmd)
		case killOp:
			return
		case stopOp:
			t.handleStop(cmd)
		case durationOp:
			t.handleDuration()
		case getErrorOpCode:
			t.handleGetError(cmd)
		case contribTimeOp:
			t.handleContribTime(cmd)
		case errorFreeOpCode:
			t.handleErrorFree(cmd)
		case startServiceCallOp:
			t.handleStartServiceCall(cmd)
		case endServiceCallOp:
			t.handleEndServiceCall(cmd)
		case toJsonOp:
			t.handleToJson()
		default:
			fmt.Println("command", cmd.opcode)
		}
	}
}

func NewEndToEndTimer(name string) *EndToEndTimer {
	e2e := &EndToEndTimer{
		Name:      name,
		start:     time.Now(),
		TxnId:     makeTxnId(),
		c:         make(chan command),
		r:         make(chan interface{}),
		ErrorFree: true,
	}

	go e2e.handleTimerOps()

	return e2e
}

func (t *EndToEndTimer) handleErrorFree(cmd command) {
	t.r <- t.ErrorFree
}

func (t *EndToEndTimer) handleContribTime(cmd command) {
	contributor := cmd.args[0].(*Contributor)
	t.r <- contributor.Duration
}

func (t *EndToEndTimer) handleGetError(cmd command) {
	t.r <- t.Error
}

func (t *EndToEndTimer) handleStartContributor(cmd command) {
	name, start := getContributorArgs(cmd.args)
	contributor := &Contributor{
		timer: t,
		Name:  name,
		start: start,
	}

	t.Contributors = append(t.Contributors, contributor)

	t.r <- contributor
}

func (t *EndToEndTimer) handleStopContributor(cmd command) {
	ct, err, stopTime := extractEndContributorArgs(cmd.args)
	ct.Duration = stopTime.Sub(ct.start)
	if err != nil {
		ct.Error = err.Error()
		t.ErrorFree = false
	}
}

func (t *EndToEndTimer) Stop(err error) {
	log.Println("Stop called with", err)
	t.c <- command{
		opcode: stopOp,
		args:   []interface{}{err},
	}
}

func (t *EndToEndTimer) EndToEndDuration() time.Duration {
	t.c <- command{opcode: durationOp}
	r := <-t.r
	d, ok := r.(time.Duration)

	if ok {
		return d
	} else {
		fmt.Println("Was not able to coerce restult to a Duration")
		return 0 * time.Millisecond
	}
}

func (t *EndToEndTimer) Kill() {
	t.c <- command{opcode: killOp}
}

func setContributorArgs(name string) []interface{} {
	return []interface{}{name, time.Now()}
}

func getContributorArgs(args []interface{}) (string, time.Time) {
	return args[0].(string), args[1].(time.Time)
}

func (t *EndToEndTimer) StartContributor(name string) *Contributor {
	t.c <- command{
		opcode: startContributorOp,
		args:   setContributorArgs(name),
	}

	r := <-t.r
	contributor, ok := r.(*Contributor)
	if ok {
		return contributor
	} else {
		return nil
	}
}

func (t *EndToEndTimer) IsErrorFree() bool {
	t.c <- command{
		opcode: errorFreeOpCode,
	}
	r := <-t.r
	errorFree, ok := r.(bool)
	if !ok {
		return false
	}

	return errorFree
}

func (t *EndToEndTimer) GetError() string {
	t.c <- command{
		opcode: getErrorOpCode,
	}
	r := <-t.r
	err, ok := r.(string)
	if !ok {
		return ""
	}

	return err
}

func setEndContributorArgs(ct *Contributor, err error) []interface{} {
	return []interface{}{ct, err, time.Now()}
}

func extractEndContributorArgs(args []interface{}) (*Contributor, error, time.Time) {
	ct := args[0].(*Contributor)
	var err error
	if args[1] != nil {
		err = args[1].(error)
	}
	stopTime := args[2].(time.Time)
	return ct, err, stopTime
}

func (ct *Contributor) End(err error) {
	ct.timer.c <- command{
		opcode: endContributorOp,
		args:   setEndContributorArgs(ct, err),
	}
}

func (ct *Contributor) Time() time.Duration {
	ct.timer.c <- command{
		opcode: contribTimeOp,
		args:   []interface{}{ct},
	}

	r := <-ct.timer.r
	contributor, ok := r.(time.Duration)
	if ok {
		return contributor
	} else {
		return 0 * time.Millisecond
	}
}

func (ct *Contributor) GetError() string {
	ct.timer.c <- command{
		opcode: contribErrOp,
		args:   []interface{}{ct},
	}

	r := <-ct.timer.r
	errmsg, ok := r.(string)
	if ok {
		return errmsg
	} else {
		return ""
	}
}

func makeTxnId() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (ct *Contributor) StartServiceCall(name string, endpoint string) *ServiceCall {
	ct.timer.c <- command{
		opcode: startServiceCallOp,
		args:   []interface{}{ct, name, endpoint, time.Now()},
	}

	svcCall := <-ct.timer.r
	return svcCall.(*ServiceCall)
}

func (sc *ServiceCall) End(err error) {
	sc.contributor.timer.c <- command{
		opcode: endServiceCallOp,
		args:   []interface{}{sc, err, time.Now()},
	}
}

func (t *EndToEndTimer) ToJSONString() string {
	t.c <- command{
		opcode: toJsonOp,
	}

	json := <-t.r
	return json.(string)
}
