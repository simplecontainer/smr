package commonv1

type State struct {
	Options []*Opts
}

type Opts struct {
	Name  string
	Value string
}

func (opt *Opts) IsEmpty() bool {
	return opt.Name == "" && opt.Value == ""
}

func (state *State) AddOpt(name string, value string) {
	state.Options = append(state.Options, &Opts{name, value})
}

func (state *State) GetOpt(name string) *Opts {
	if state.Options != nil {
		for _, v := range state.Options {
			if v.Name == name {
				return v
			}
		}
	}

	return &Opts{}
}
