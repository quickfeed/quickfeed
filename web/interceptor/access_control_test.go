package interceptor_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"google.golang.org/grpc"
)

const qfServiceName = "qf.QuickFeedService"

func TestAccessControlMethods(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qlog.Logger(t)
	ags := web.NewQuickFeedService(logger.Desugar(), db, &scm.Manager{}, web.BaseHookOptions{}, &ci.Local{})

	s := grpc.NewServer()
	qf.RegisterQuickFeedServiceServer(s, ags)

	access := interceptor.GetAccessTable()
	qfServiceInfo, ok := s.GetServiceInfo()[qfServiceName]
	if !ok {
		t.Fatalf("failed to read service info (%s)", qfServiceName)
	}

	for _, method := range qfServiceInfo.Methods {
		_, ok := access[method.Name]
		if !ok {
			t.Errorf("access control table missing method %s", method.Name)
		}
	}
}

func TestAccessControl(t *testing.T) {

}
