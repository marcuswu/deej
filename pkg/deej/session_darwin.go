package deej

import (
	"fmt"

	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
	"go.uber.org/zap"
)

const maxVolume = 100

type darwinSession struct {
	baseSession

	localizedName string
}

type masterSession struct {
	baseSession
}

func newDarwinSession(logger *zap.SugaredLogger, localizedName string) *darwinSession {
	s := &darwinSession{}

	s.localizedName = localizedName
	s.name = localizedName
	s.humanReadableDesc = localizedName

	// use a self-identifying session name e.g. deej.sessions.chrome
	s.logger = logger.Named(s.Key())
	s.logger.Debugw("Creating session for", "localizedName", localizedName)
	s.logger.Debugw(sessionCreationLogMessage, "session", s)

	return s
}

func newMasterSession(logger *zap.SugaredLogger) *masterSession {
	s := &masterSession{}
	key := masterSessionName

	s.logger = logger.Named(key)
	s.master = true
	s.name = key
	s.humanReadableDesc = key

	s.logger.Debugw(sessionCreationLogMessage, "session", s)

	return s
}

func (ds *darwinSession) GetVolume() float32 {
	var error *core.NSDictionary = nil
	code := core.NSString_FromString("tell application \"Background Music\" to get vol of (a reference to (the first audio application whose name is equal to \"" + ds.name + "\"))")
	script := objc.Get("NSAppleScript").Alloc().Send("initWithSource:", code)
	result := script.Send("executeAndReturnError:", &error)
	volume := result.Float()

	return float32(volume) / float32(maxVolume)
}

func (ds *darwinSession) SetVolume(volume float32) error {
	vol := uint32(volume * maxVolume)
	var error *core.NSDictionary = nil
	code := core.NSString_FromString(fmt.Sprintf("tell application \"Background Music\" to set vol of (a reference to (the first audio application whose name is equal to \"%s\")) to %d", ds.name, vol))
	script := objc.Get("NSAppleScript").Alloc().Send("initWithSource:", code)
	_ = script.Send("executeAndReturnError:", &error)

	ds.logger.Debugf("Adjusting session volume to %f", volume)
	return nil
}

func (ds *darwinSession) Release() {

}

func (ms *masterSession) GetVolume() float32 {
	var error *core.NSDictionary = nil
	code := core.NSString_FromString("tell application \"Background Music\" to get output volume")
	script := objc.Get("NSAppleScript").Alloc().Send("initWithSource:", code)
	result := script.Send("executeAndReturnError:", &error)
	volume := result.Float()

	return float32(volume)
}

func (ms *masterSession) SetVolume(volume float32) error {
	error := objc.ObjectPtr(0)
	code := core.NSString_FromString(fmt.Sprintf("tell application \"Background Music\" to set output volume to %f", volume))
	script := objc.Get("NSAppleScript").Alloc().Send("initWithSource:", code)
	_ = script.Send("executeAndReturnError:", &error)
	ms.logger.Debugf("Adjusting master volume to %.2f", volume)

	return nil
}

func (ms *masterSession) Release() {

}
