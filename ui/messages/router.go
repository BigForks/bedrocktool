package messages

type router struct {
	handlers map[string]HandlerFunc
}

func (r *router) AddHandler(name string, handler HandlerFunc) {
	r.handlers[name] = handler
}

func (r *router) Handle(msg *Message) *Message {
	if msg.Target == "" {
		for _, handler := range r.handlers {
			handler(msg)
		}
		return nil
	}
	handler, ok := r.handlers[msg.Target]
	if !ok {
		return nil
	}

	return handler(msg)
}

var Router = router{
	handlers: make(map[string]func(msg *Message) *Message),
}
