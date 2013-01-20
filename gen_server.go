package node

import (
	"erlang/term"
)

type GenServer interface {
	Init(args ...interface{})
	HandleCast(message *term.Term)
	HandleCall(message *term.Term, from *term.Tuple)
	HandleInfo(message *term.Term)
	Terminate(reason interface{})
}

type GenServerImpl struct {

}

func (gs GenServerImpl) Options() (map[string]interface{}) {
	return map[string]interface{}{
		"chan-size": 100,
		"ctl-chan-size": 100,
	}
}

func (gs GenServerImpl) ProcessLoop(pid term.Pid, pcs procChannels, pd Process, args ...interface{}) {
	pd.(GenServer).Init(args...)
	pcs.ctl <- term.Tuple{term.Atom("$go_ctl"), term.Tuple{term.Atom("your-pid"), pid}}
	for {
		var message term.Term
		var fromPid term.Pid
		select {
		case msg := <-pcs.in:
			message = msg
		case msgFrom := <-pcs.inFrom:
			message = msgFrom[1]
			fromPid = msgFrom[0].(term.Pid)
		case ctlMsg := <-pcs.ctl:
			switch m := ctlMsg.(type) {
			case term.Tuple:
				switch mtag := m[0].(type) {
				case term.Atom:
					switch mtag {
					case term.Atom("$go_ctl"):
						nLog("Control message: %#v", m)
					default:
						nLog("Unknown message: %#v", m)
					}
				default:
					nLog("Unknown message: %#v", m)
				}
			default:
				nLog("Unknown message: %#v", m)
			}
			continue
		}
		nLog("Message from %#v", fromPid)
		switch m := message.(type) {
		case term.Tuple:
			switch mtag := m[0].(type) {
			case term.Atom:
				switch mtag {
				case term.Atom("$go_ctl"):
					nLog("Control message: %#v", message)
				case term.Atom("$gen_call"):
					fromTuple := m[1].(term.Tuple)
					pd.(GenServer).HandleCall(&m[2], &fromTuple)
				case term.Atom("$gen_cast"):
					pd.(GenServer).HandleCast(&m[1])
				default:
					pd.(GenServer).HandleInfo(&message)
				}
			default:
				nLog("mtag: %#v", mtag)
				pd.(GenServer).HandleInfo(&message)
			}
		default:
			nLog("m: %#v", m)
			pd.(GenServer).HandleInfo(&message)
		}
	}
}