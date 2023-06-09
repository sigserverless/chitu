package server

import (
	"bytes"
	"differentiable/stub"
	"io"
	"log"
	"os"
	"os/exec"
)

type FunctionType func(stub stub.AgentStub, req []byte) ([]byte, error)

type Usercode interface {
	Handle(stub stub.AgentStub, req []byte) ([]byte, error)
}

type GoUsercode struct {
	Usercode FunctionType
}

func (u *GoUsercode) Handle(stub stub.AgentStub, req []byte) ([]byte, error) {
	return u.Usercode(stub, req)
}

type CmdUsercode struct {
	Cmd  string
	Args []string
}

func (u *CmdUsercode) Handle(s stub.AgentStub, req []byte) ([]byte, error) {
	fs := s.(*stub.FunchanStub)
	env := []string{}
	env = append(env, "INV_ID="+fs.InvId)
	env = append(env, "DAG_ID="+fs.DagId)

	cmd := exec.Command(u.Cmd, u.Args...)
	cmd.Env = env
	cmd.Stdin = bytes.NewReader(req)
	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)
	cmd.Stdout = mw
	cmd.Stderr = mw

	err := cmd.Run()

	if err != nil {
		log.Fatalf("CMD process failed: %v", err)
		return nil, err
	}

	return stdBuffer.Bytes(), nil
}
