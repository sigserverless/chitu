package server

import (
	"bufio"
	"context"
	ddt "differentiable/datatypes"
	"differentiable/pb/pany"
	"differentiable/pb/pdict"
	"differentiable/stub"
	"differentiable/stub/handle"
	"differentiable/utils"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"time"

	"google.golang.org/protobuf/proto"
)

// FunchanServer for usercode written in other languages
// Differences between go version and other languages version (e.g. py):
// 1. When serving a invocation request:
// - the go version forks a goroutine to execute the user code.
// - the py version forks a process to execute the python file.
// 2. The way user code communicates with state agent server:
// - the go version uses go channels
// - the py version uses Unix FIFOs
func NewCmdFunchanServer(period time.Duration, cmd string, args ...string) (*FunchanServer, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	consChan := make(chan ddt.DConsMsg, CHANNEL_BUFFER_SIZE)
	writeChan := make(chan ddt.DWriteMsg, CHANNEL_BUFFER_SIZE)
	endChan := make(chan ddt.DEndMsg, CHANNEL_BUFFER_SIZE)
	awaitChan := make(chan ddt.DAwaitMsg, CHANNEL_BUFFER_SIZE)
	readChan := make(chan ddt.DReadMsg, CHANNEL_BUFFER_SIZE)
	exportChan := make(chan ddt.ExportMsg, CHANNEL_BUFFER_SIZE)
	importChan := make(chan ddt.ImportMsg, CHANNEL_BUFFER_SIZE)
	triggerChan := make(chan ddt.TriggerMsg, CHANNEL_BUFFER_SIZE)
	s := &FunchanServer{
		objs:        map[string]handle.ObjectHandle{},
		invocations: map[string]stub.AgentStub{},
		// keyeds:      map[KeyWithDag]string{},
		keyedExports: map[KeyWithDag]string{},
		keyedImports: map[KeyWithDag]string{},
		consChan:     consChan,
		writeChan:    writeChan,
		endChan:      endChan,
		awaitChan:    awaitChan,
		readChan:     readChan,
		exportChan:   exportChan,
		importChan:   importChan,
		triggerChan:  triggerChan,
		period:       period,
	}
	forwardFifos(
		consChan,
		writeChan,
		endChan,
		awaitChan,
		readChan,
		exportChan,
		importChan,
		triggerChan,
	)

	go s.serveUserRequest(ctx)
	usercode := &CmdUsercode{
		Cmd:  cmd,
		Args: args,
	}
	s.serveOuterRequest(usercode)
	return s, cancel
}

func forwardFifos(
	consChan chan<- ddt.DConsMsg,
	writeChan chan<- ddt.DWriteMsg,
	endChan chan<- ddt.DEndMsg,
	awaitChan chan<- ddt.DAwaitMsg,
	readChan chan<- ddt.DReadMsg,
	exportChan chan<- ddt.ExportMsg,
	importChan chan<- ddt.ImportMsg,
	triggerChan chan<- ddt.TriggerMsg,
) {
	makeFifos([]string{
		"cons.req.fifo",
		"write.req.fifo",
		"end.req.fifo",
		"await.req.fifo",
		"read.req.fifo",
		"import.req.fifo",
		"export.req.fifo",
		"trigger.req.fifo",
		"cons.res.fifo",
		"write.res.fifo",
		"await.res.fifo",
		"read.res.fifo",
		"import.res.fifo",
		"pdict.write.req.fifo",
		"pdict.write.res.fifo",
		"pdict.await.req.fifo",
		"pdict.await.res.fifo",
		"pany.write.req.fifo",
		"pany.write.res.fifo",
		"pany.await.req.fifo",
		"pany.await.res.fifo",
	})
	go forwardFifo(endChan, "end")
	go forwardFifo(exportChan, "export")
	go forwardFifo(triggerChan, "trigger")
	// go forwardFifoWithRes(consChan, "cons")
	// go forwardFifoWithRes(writeChan, "write")
	// go forwardFifoWithRes(awaitChan, "await")
	// go forwardFifoWithRes(readChan, "read")
	// go forwardFifoWithRes(importChan, "import")
	go forwardFifoAwait(awaitChan)
	go forwardFifoRead(readChan)
	go forwardFifoWrite(writeChan)
	go forwardFifoImport(importChan)
	go forwardFifoCons(consChan)

	go forwardFifoPAwait(awaitChan)
	go forwardFifoPSet(writeChan)

	go forwardFifoPReplace(writeChan)
	go forwardFifoPAnyAwait(awaitChan)
	// TODO: these need refactoring
}

func makeFifos(names []string) {
	for _, name := range names {
		err := syscall.Mkfifo(name, 0666)
		if err != nil {
			panic(fmt.Sprintf("Error making fifo: %s, %v", name, err))
		}
	}
}

func readFifo(filename string) (*bufio.Reader, func() error) {
	pipe, err := os.OpenFile(filename, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		panic(fmt.Sprintf("Error opening named pipe for reading: %v", err))
	}

	return bufio.NewReader(pipe), pipe.Close
}

func writeFifo(filename string, res any) {
	pipe, err := os.OpenFile(filename, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		panic(fmt.Sprintf("Error opening named pipe for reading: %v", err))
	}
	resJson, err := json.Marshal(res)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling res: %v", err))
	}
	pipe.Write(resJson)
	pipe.WriteString("\n")
	pipe.Close()
}

// func forwardFifoWithRes[T any, R any](ch chan<- T, name string, handler func([]byte) R) {
// 	reader, close := readFifo(name + ".req.fifo")
// 	defer close()

// 	for {
// 		line, err := reader.ReadBytes('\n')
// 		if err == nil {
// 			res := handler(line)
// 			forwardRes(name+".res.fifo", res)
// 		} else {
// 			break
// 		}
// 	}
// }

func forwardFifoCons(ch chan<- ddt.DConsMsg) {
	reader, close := readFifo("cons.req.fifo")

	for {
		line, err := reader.ReadBytes('\n')
		if err == nil {
			var msg ddt.DConsMsg
			json.Unmarshal(line, &msg)
			resCh := make(chan ddt.DConsRes, 1)
			msg.Res = resCh
			ch <- msg

			res := <-resCh
			writeFifo("cons.res.fifo", res)
		} else {
			close()
			reader, close = readFifo("cons.req.fifo")
		}
	}
}

type FifoWriteMsg struct {
	Id  string          `json:"Id"`
	Wop json.RawMessage `json:"Wop"`
}

func forwardFifoWrite(ch chan<- ddt.DWriteMsg) {
	reader, close := readFifo("write.req.fifo")
	defer close()

	for {
		line, err := reader.ReadBytes('\n')
		if err == nil {
			var msgJson FifoWriteMsg
			json.Unmarshal(line, &msgJson)

			var msg ddt.DWriteMsg
			resCh := make(chan ddt.DWriteRes, 1)
			msg.Id = msgJson.Id
			wop, err := utils.Unmarshal(msgJson.Wop, ddt.WriteOpMapping)
			if err != nil {
				log.Fatalf("WriteFifo unmarshal error: %v", err)
			}
			msg.Wop = wop
			msg.Res = resCh
			ch <- msg

			res := <-resCh
			writeFifo("write.res.fifo", res)
		} else {
			close()
			reader, close = readFifo("write.req.fifo")
		}
	}
}

func forwardFifoRead(ch chan<- ddt.DReadMsg) {
	reader, close := readFifo("read.req.fifo")
	defer close()

	for {
		line, err := reader.ReadBytes('\n')
		if err == nil {
			var msg ddt.DReadMsg
			var rawMsg struct {
				Ids []string
				Rop struct {
					Kind string `json:"kind"`
				}
			}

			json.Unmarshal(line, &msg)
			json.Unmarshal(line, &rawMsg)

			ropKind := rawMsg.Rop.Kind

			resCh := make(chan ddt.DReadRes, 1)

			if ropKind == "MergeMean" {
				msg.Rop = &ddt.MergeMeansOp{
					Number: len(rawMsg.Ids),
				}
			}
			// msg.Rop = &ddt.MergeMeansOp{}
			msg.Res = resCh
			ch <- msg

			res := <-resCh
			writeFifo("read.res.fifo", res)
		} else {
			close()
			reader, close = readFifo("read.req.fifo")
		}
	}
}

func forwardFifoAwait(ch chan<- ddt.DAwaitMsg) {
	reader, close := readFifo("await.req.fifo")
	defer close()

	for {
		line, err := reader.ReadBytes('\n')
		if err == nil {
			var msg ddt.DAwaitMsg
			json.Unmarshal(line, &msg)
			resCh := make(chan ddt.DAwaitRes, 1)
			msg.Res = resCh
			ch <- msg

			res := <-resCh
			writeFifo("await.res.fifo", res)
		} else {
			close()
			reader, close = readFifo("await.req.fifo")
		}
	}
}

func forwardFifoImport(ch chan<- ddt.ImportMsg) {
	reader, close := readFifo("import.req.fifo")
	defer close()

	for {
		line, err := reader.ReadBytes('\n')
		if err == nil {
			var msg ddt.ImportMsg
			json.Unmarshal(line, &msg)
			resCh := make(chan ddt.ImportRes, 1)
			msg.Res = resCh
			ch <- msg

			res := <-resCh
			writeFifo("import.res.fifo", res)
		} else {
			close()
			reader, close = readFifo("import.req.fifo")
		}
	}
}

func forwardFifo[M any](ch chan<- M, fifo string) {
	reader, close := readFifo(fifo + ".req.fifo")

	for {
		line, err := reader.ReadBytes('\n')
		// if fifo == "trigger" {
		// 	fmt.Println("line:", string(line))
		// }

		if err == nil {
			var msg M
			json.Unmarshal(line, &msg)
			if fifo == "trigger" {
				fmt.Println("line:", string(line))
				fmt.Println("msg:", msg)
			}
			ch <- msg
		} else {
			close()
			reader, close = readFifo(fifo + ".req.fifo")
		}
	}
}

func forwardFifoPSet(ch chan<- ddt.DWriteMsg) {
	reader, close := readFifo("pdict.write.req.fifo")
	defer close()

	for {
		// line, err := reader.ReadBytes('\n')
		var length uint32
		err := binary.Read(reader, binary.BigEndian, &length)
		if err == nil && length > 0 {
			line := make([]byte, length)
			_, _ = io.ReadFull(reader, line)
			var msg ddt.DWriteMsg
			var rawMsg pdict.PDictSet
			resCh := make(chan ddt.DWriteRes, 1)
			proto.Unmarshal(line, &rawMsg)
			msg.Id = rawMsg.Id
			msg.Wop = &ddt.DictSetOp{
				K: rawMsg.Key,
				V: rawMsg.Values.Values,
			}
			msg.Res = resCh
			ch <- msg

			res := <-resCh
			writeFifo("pdict.write.res.fifo", res)
		} else {
			close()
			reader, close = readFifo("pdict.write.req.fifo")
		}
	}
}

func forwardFifoPAwait(ch chan<- ddt.DAwaitMsg) {
	reader, close := readFifo("pdict.await.req.fifo")
	defer close()

	for {
		line, err := reader.ReadBytes('\n')
		if err == nil {
			var msg ddt.DAwaitMsg
			json.Unmarshal(line, &msg)
			resCh := make(chan ddt.DAwaitRes, 1)
			msg.Res = resCh
			ch <- msg

			res := <-resCh
			pDictWriteFifo("pdict.await.res.fifo", res)
		} else {
			close()
			reader, close = readFifo("pdict.await.req.fifo")
		}
	}
}

func pDictWriteFifo(filename string, res ddt.DAwaitRes) {
	pipe, err := os.OpenFile(filename, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		panic(fmt.Sprintf("Error opening named pipe for reading: %v", err))
	}
	dictVal := res.Val.(*ddt.DDictVal)
	var pdictVal pdict.PDictVal
	pdictVal.Val = make(map[string]*pdict.Values)
	for k, v := range dictVal.Val {
		var values pdict.Values
		values.Values = utils.ConvertToSlice[float64](v)
		pdictVal.Val[k] = &values
	}
	resPb, err := proto.Marshal(&pdictVal)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling res: %v", err))
	}
	pipe.Write(resPb)
	pipe.Close()
}

func forwardFifoPAnyAwait(ch chan<- ddt.DAwaitMsg) {
	reader, close := readFifo("pany.await.req.fifo")
	defer close()

	for {
		line, err := reader.ReadBytes('\n')
		if err == nil {
			var msg ddt.DAwaitMsg
			json.Unmarshal(line, &msg)
			resCh := make(chan ddt.DAwaitRes, 1)
			msg.Res = resCh
			ch <- msg

			res := <-resCh
			pAnyWriteFifo("pany.await.res.fifo", res)
		} else {
			close()
			reader, close = readFifo("pany.await.req.fifo")
		}
	}
}

func pAnyWriteFifo(filename string, res ddt.DAwaitRes) {
	pipe, err := os.OpenFile(filename, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		panic(fmt.Sprintf("Error opening named pipe for reading: %v", err))
	}
	anyVal := res.Val.(*ddt.DAnyVal)
	var panyVal pany.PAnyVal
	if bytes, ok := anyVal.Val.([]byte); ok {
		panyVal.Val = bytes
	} else {
		panyVal.Val = []byte(anyVal.Val.(string))
	}
	resPb, err := proto.Marshal(&panyVal)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling res: %v", err))
	}
	length := uint32(len(resPb))
	err = binary.Write(pipe, binary.BigEndian, length)
	if err != nil {
		panic(fmt.Sprintf("Error writing length: %v", err))
	}
	pipe.Write(resPb)
	pipe.Close()
}

func forwardFifoPReplace(ch chan<- ddt.DWriteMsg) {
	reader, close := readFifo("pany.write.req.fifo")
	defer close()

	for {
		var length uint32
		err := binary.Read(reader, binary.BigEndian, &length)
		if err == nil && length > 0 {
			line := make([]byte, length)
			_, _ = io.ReadFull(reader, line)
			var msg ddt.DWriteMsg
			var rawMsg pany.PAnyReplace
			resCh := make(chan ddt.DWriteRes, 1)
			proto.Unmarshal(line, &rawMsg)
			msg.Id = rawMsg.Id
			msg.Wop = &ddt.ReplaceOp{
				Elem: rawMsg.Elem,
			}
			msg.Res = resCh
			ch <- msg

			res := <-resCh
			writeFifo("pany.write.res.fifo", res)
		} else {
			close()
			reader, close = readFifo("pany.write.req.fifo")
		}
	}
}
