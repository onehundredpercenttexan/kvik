package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fjukstad/kvik/kompute"
)

type Stage struct {
	Name      string
	Package   string
	Function  string
	Arguments map[string]string
	Depends   []Stage
	Session   *kompute.Session
}

type Pipeline struct {
	Name    string
	Kompute *kompute.Kompute
	Stages  []*Stage
}

func NewPipeline(name string, k *kompute.Kompute) Pipeline {
	p := Pipeline{name, k, nil}
	return p
}

func (p *Pipeline) AddStage(s Stage) {
	p.Stages = append(p.Stages, &s)
}

func NewStage(name, function, pkg string, argnames, args []string) Stage {

	argmap := make(map[string]string, 0)

	for i, argname := range argnames {
		arg := args[i]
		argmap[argname] = arg
	}

	s := Stage{name, pkg, function, argmap, []Stage{}, nil}

	return s
}

type Result struct {
	m     *sync.Mutex
	c     *sync.Cond
	Key   string
	Error error
}

func (p *Pipeline) Run() ([]*Result, error) {

	resultMap := make(map[string]*Result, 0)
	done := make(chan bool, len(p.Stages))

	for _, stage := range p.Stages {
		m := sync.Mutex{}
		c := sync.NewCond(&m)
		resultMap[stage.Name] = &Result{&m, c, "", nil}

		deps := stage.GetDependencies()
		r := resultMap[stage.Name]

		go func(r *Result, stage *Stage, deps []string) {
			if len(deps) != 0 {
				for _, dep := range deps {
					r := resultMap[dep]
					r.m.Lock()

					for r.Key == "" {
						r.c.Wait()
					}

					stage.ReplaceArg(dep, r.Key)
					r.m.Unlock()
				}
			}
			r.m.Lock()
			for r.Key == "" {
				r.Key, r.Error = p.ExecuteStage(stage)
			}
			r.c.Broadcast()
			r.m.Unlock()

			done <- true
		}(r, stage, deps)
	}

	for i := 0; i < len(p.Stages); i++ {
		<-done
	}

	results := []*Result{}
	for _, stage := range p.Stages {
		r := resultMap[stage.Name]
		results = append(results, r)
	}

	return results, nil
}

func (s *Stage) ReplaceArg(oldarg string, newarg string) {
	for i, arg := range s.Arguments {
		if strings.Contains(arg, oldarg) {
			s.Arguments[i] = newarg
			return
		}
	}
}

func (p *Pipeline) ExecuteStage(stage *Stage) (string, error) {

	args := stage.GetArguments()
	fun := stage.GetFullFunctionName()

	s, err := p.Kompute.Call(fun, args)

	if err != nil {
		s, err = p.Kompute.Call(fun, args)

		if err != nil {
			fmt.Println("ERROR", err)
			return "", err
		}
	}
	return s.Key, nil
}

func (p *Pipeline) GetResult(format string) string {
	lastStage := p.Stages[len(p.Stages)-1]
	res, _ := lastStage.Session.GetResult(p.Kompute, format)
	return res
}

func (p *Pipeline) Save() error {
	file, err := os.OpenFile(p.Name+".json", os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return err
	}

	b, err := json.Marshal(p)
	if err != nil {
		return err
	}

	_, err = file.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func (s *Stage) GetFullFunctionName() string {
	return s.Package + "/R/" + s.Function
}

func (s *Stage) GetArguments() string {
	str := ""
	i := 0
	for k, v := range s.Arguments {
		if strings.Contains(v, "/") || strings.Contains(v, ".") {
			v = "\"" + v + "\""
		}
		str = str + k + "=" + v
		if i < len(s.Arguments)-1 {
			str = str + "&"
		}
		i++
	}

	return str
}
func (s Stage) GetDependencies() []string {
	deps := []string{}
	for _, arg := range s.Arguments {
		if strings.Contains(arg, "from:") {
			argname := strings.Split(arg, "from:")[1]
			deps = append(deps, argname)
		}
	}
	return deps
}

func (s *Stage) Print() {
	fmt.Println("\tName:", s.Name)
	fmt.Println("\tPackage", s.Package)
	fmt.Println("\tFunction", s.Function)
	fmt.Println("\tArguments:")
	for k, v := range s.Arguments {
		fmt.Println("\t\t", k, "\t", v)
	}
	if s.Session != nil {
		fmt.Println("\tSession: ", s.Session.Key)
		fmt.Println("\tURL: /ocpu/tmp/" + s.Session.Key + "/R/")
	}

}
