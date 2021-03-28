package deej

import (
	"fmt"

	"github.com/progrium/macdriver/objc"
	"go.uber.org/zap"
)

type darwinSessionFinder struct {
	logger        *zap.SugaredLogger
	sessionLogger *zap.SugaredLogger
}

func newSessionFinder(logger *zap.SugaredLogger) (SessionFinder, error) {
	sf := &darwinSessionFinder{
		logger:        logger.Named("session_finder"),
		sessionLogger: logger.Named("sessions"),
	}

	sf.logger.Debug("Created OSX session finder instance")

	return sf, nil
}

func (dsf *darwinSessionFinder) GetAllSessions() ([]Session, error) {
	sessions := []Session{}

	masterSink, err := dsf.getMasterSession()
	if err == nil {
		sessions = append(sessions, masterSink)
	} else {
		dsf.logger.Warnw("Failed to get master audio session", "error", err)
	}

	// create sessions from list of apps
	if err := dsf.enumerateAndAddSessions(&sessions); err != nil {
		dsf.logger.Warnw("Failed to enumerate audio sessions", "error", err)
		return nil, fmt.Errorf("enumerate audio sessions: %w", err)
	}

	return sessions, nil
}

func (dsf *darwinSessionFinder) Release() error {
	return nil
}

func runningApplications() []string {
	applications := []string{}
	apps := objc.Get("NSWorkspace").Get("sharedWorkspace").Get("runningApplications")

	for i := int64(0); i < apps.Get("count").Int(); i++ {
		applications = append(applications, apps.Send("objectAtIndex:", i).Send("localizedName").String())
	}

	return applications
}

func (dsf *darwinSessionFinder) getMasterSession() (Session, error) {
	// ...
	// create the master source session
	source := newMasterSession(dsf.sessionLogger)

	return source, nil
}

func (dsf *darwinSessionFinder) enumerateAndAddSessions(sessions *[]Session) error {
	apps := runningApplications()

	for _, name := range apps {

		newSession := newDarwinSession(dsf.logger, name)
		*sessions = append(*sessions, newSession)
	}

	return nil
}
