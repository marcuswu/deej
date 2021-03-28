package deej

import (
	"os/exec"
	"strconv"
	"sync"

	"github.com/google/shlex"
	"go.uber.org/zap"
)

type Action struct {
	command string
	params  []string
}

func newAction(command string, params []string) *Action {
	a := &Action{
		command,
		params,
	}
	return a
}

func actionFromString(action string) *Action {
	params, _ := shlex.Split(action)
	return newAction(params[0], params[1:])
}

func actionsFromStrings(actions []string) []Action {
	result := []Action{}

	for _, a := range actions {
		result = append(result, *actionFromString(a))
	}

	return result
}

func (a *Action) Execute(logger *zap.SugaredLogger) {
	cmd := exec.Command(a.command, a.params...)
	err := cmd.Run()

	if err != nil {
		logger.Errorw("Failed to execute event handler action", "error", err)
	}
}

type eventMap struct {
	m    map[int][]Action
	lock sync.Locker
}

func newEventMap() *eventMap {

	m := &eventMap{
		m:    make(map[int][]Action),
		lock: &sync.Mutex{},
	}

	return m
}

func (em *eventMap) iterate(f func(int, []Action)) {
	em.lock.Lock()
	defer em.lock.Unlock()

	for key, value := range em.m {
		f(key, value)
	}
}

func (em *eventMap) get(key int) ([]Action, bool) {
	em.lock.Lock()
	defer em.lock.Unlock()

	value, ok := em.m[key]
	return value, ok
}

func (em *eventMap) set(key int, value []Action) {
	em.lock.Lock()
	defer em.lock.Unlock()

	em.m[key] = value
}

func eventMapFromConfigs(userMapping map[string][]string) *eventMap {
	resultMap := newEventMap()

	// copy targets from user config, ignoring empty values
	for eventIdxString, targets := range userMapping {
		eventIdx, _ := strconv.Atoi(eventIdxString)

		resultMap.set(eventIdx, actionsFromStrings(targets))
	}

	return resultMap
}

type eventHandler struct {
	deej   *Deej
	logger *zap.SugaredLogger
}

func newEventHandler(deej *Deej, logger *zap.SugaredLogger) (*eventHandler, error) {
	eh := &eventHandler{
		deej,
		logger,
	}

	return eh, nil
}

func (eh *eventHandler) initialize() error {
	eh.setupOnNumberedEvent()

	return nil
}

func (eh *eventHandler) setupOnNumberedEvent() {
	eventChannel := eh.deej.serial.SubscribeToNumberedEvent()

	go func() {
		for {
			select {
			case event := <-eventChannel:
				eh.handleNumberedEvent(event.EventID)
			}
		}
	}()
}

func (eh *eventHandler) handleNumberedEvent(id int) {
	actions, ok := eh.deej.config.EventMapping.get(id)
	if !ok {
		return
	}
	for _, action := range actions {
		eh.logger.Debugf("Executing event handler for event %d", id)
		action.Execute(eh.logger)
	}
}
